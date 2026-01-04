package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"agent/buffer"
	"agent/collector"
	"agent/processor"
	"agent/sender"
)

// Agent главный координатор SIEM агента
type Agent struct {
	config      Config
	collectors  []collector.Collector
	buffer      buffer.Buffer
	processor   processor.Processor
	sender      sender.Sender

	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	running    bool
}

// Config конфигурация агента
type Config struct {
	AgentID          string
	ServerHost       string
	ServerPort       int
	CollectionInterval  int // миллисекунды
	SenderInterval      int // миллисекунды
	BatchSize        int
	BufferMaxSize    int
}

// NewAgent создаёт новый агент
func NewAgent(config Config, bufferInstance buffer.Buffer, processorInstance processor.Processor, senderInstance sender.Sender) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		config:      config,
		collectors:   make([]collector.Collector, 0),
		buffer:      bufferInstance,
		processor:   processorInstance,
		sender:      senderInstance,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// RegisterCollector регистрирует источник логов
func (a *Agent) RegisterCollector(col collector.Collector) {
	a.collectors = append(a.collectors, col)
	log.Printf("[Agent] Collector registered: %s (%s)", col.GetSourceName(), col.GetSourceType())
}

// Start запускает агент
func (a *Agent) Start() error {
	a.mu.Lock()
	if a.running {
		a. mu.Unlock()
		return fmt.Errorf("agent is already running")
	}
	a.running = true
	a.mu.Unlock()

	log. Println("[Agent] Starting SIEM Agent...")
	log.Printf("[Agent] Agent ID: %s", a.config.AgentID)
	log.Printf("[Agent] Server: %s:%d", a.config.ServerHost, a.config. ServerPort)

	// Запускаем горутины для каждого компонента
	a.wg.Add(3)
	go a.collectorLoop()
	go a.processorLoop()
	go a.senderLoop()

	log.Println("[Agent] Agent started successfully")
	return nil
}

// Stop останавливает агент
func (a *Agent) Stop() {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return
	}
	a.running = false
	a.mu. Unlock()

	log.Println("[Agent] Stopping SIEM Agent...")

	a.cancel()
	a.wg.Wait()

	// Закрываем соединение с сервером
	a.sender.Close()

	log.Println("[Agent] Agent stopped")
}

// IsRunning возвращает статус агента
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// GetBufferSize возвращает текущий размер буфера
func (a *Agent) GetBufferSize() int {
	return a.buffer.Size()
}

// collectorLoop основной цикл сборщика
func (a *Agent) collectorLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Duration(a.config.CollectionInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			log.Println("[Collector] Stopping collector loop")
			return
		case <-ticker.C:
			a.collectLogs()
		}
	}
}

// collectLogs собирает логи из всех источников
func (a *Agent) collectLogs() {
	for _, col := range a.collectors {
		events, err := col. Collect()
		if err != nil {
			log.Printf("[Collector] Error collecting from %s: %v", col. GetSourceName(), err)
			continue
		}

		if len(events) > 0 {
			// Конвертируем в buffer. Event
			bufferEvents := convertEvents(events)
			a.buffer.Push(bufferEvents)
			log.Printf("[Collector] Collected %d events from %s", len(events), col.GetSourceName())
		}
	}
}

// processorLoop основной цикл обработчика
func (a *Agent) processorLoop() {
	defer a.wg.Done()

	for {
		select {
		case <-a.ctx.Done():
			log.Println("[Processor] Stopping processor loop")
			return
		default:
			// Извлекаем из буфера
			events := a.buffer.Pop(a.config.BatchSize)
			if len(events) == 0 {
				time.Sleep(100 * time. Millisecond)
				continue
			}

			// Конвертируем в processor.Event
			procEvents := convertBufferToProcessor(events)

			// Обрабатываем
			processedEvents := a.processor.ProcessBatch(procEvents)

			if len(processedEvents) > 0 {
				// Конвертируем обратно в buffer.Event
				bufferEvents := convertProcessorToBuffer(processedEvents)
				a.buffer.Push(bufferEvents)
				log.Printf("[Processor] Processed %d events", len(processedEvents))
			}
		}
	}
}

// senderLoop основной цикл отправителя
func (a *Agent) senderLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Duration(a.config.SenderInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			log.Println("[Sender] Stopping sender loop")
			return
		case <-ticker.C: 
			a.sendEvents()
		}
	}
}

// sendEvents отправляет события на сервер
func (a *Agent) sendEvents() {
	if !a.sender.IsConnected() {
		log.Println("[Sender] Server is not connected")
		return
	}

	events := a.buffer.Pop(a.config.BatchSize)
	if len(events) == 0 {
		return
	}

	// Конвертируем в sender.Event
	senderEvents := convertBufferToSender(events)

	if err := a.sender.Send(senderEvents); err != nil {
		log.Printf("[Sender] Error sending events: %v", err)
		// Возвращаем в буфер при ошибке
		a. buffer.Push(events)
	} else {
		log.Printf("[Sender] Sent %d events to server", len(senderEvents))
	}
}

// ==================== Вспомогательные функции конвертации ====================

// convertEvents конвертирует collector.Event в buffer.Event
func convertEvents(events []collector.Event) []buffer.Event {
	result := make([]buffer.Event, len(events))
	for i, e := range events {
		result[i] = buffer.Event{
			Timestamp: e.Timestamp,
			Hostname:  e.Hostname,
			Source:    e.Source,
			EventType: e.EventType,
			Severity:  e.Severity,
			User:      e.User,
			Process:   e.Process,
			Command:   e.Command,
			RawLog:    e.RawLog,
		}
	}
	return result
}

// convertBufferToProcessor конвертирует buffer.Event в processor.Event
func convertBufferToProcessor(events []buffer.Event) []processor.Event {
	result := make([]processor.Event, len(events))
	for i, e := range events {
		result[i] = processor.Event{
			Timestamp: e.Timestamp,
			Hostname:  e. Hostname,
			Source:    e.Source,
			EventType: e.EventType,
			Severity:  e.Severity,
			User:      e. User,
			Process:   e.Process,
			Command:   e.Command,
			RawLog:    e.RawLog,
		}
	}
	return result
}

// convertProcessorToBuffer конвертирует processor.Event в buffer.Event
func convertProcessorToBuffer(events []processor.Event) []buffer.Event {
	result := make([]buffer.Event, len(events))
	for i, e := range events {
		result[i] = buffer.Event{
			Timestamp: e.Timestamp,
			Hostname:  e. Hostname,
			Source:    e.Source,
			EventType: e.EventType,
			Severity:  e.Severity,
			User:      e. User,
			Process:   e.Process,
			Command:   e.Command,
			RawLog:    e.RawLog,
		}
	}
	return result
}

// convertBufferToSender конвертирует buffer.Event в sender. Event
func convertBufferToSender(events []buffer.Event) []sender.Event {
	result := make([]sender.Event, len(events))
	for i, e := range events {
		result[i] = sender.Event{
			Timestamp: e. Timestamp,
			Hostname:   e.Hostname,
			Source:    e.Source,
			EventType: e.EventType,
			Severity:  e.Severity,
			User:      e.User,
			Process:   e.Process,
			Command:   e.Command,
			RawLog:    e. RawLog,
		}
	}
	return result
}