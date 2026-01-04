package buffer

import (
	"sync"
)

// Event импортируем из collector
type Event struct {
	Timestamp string `json:"timestamp"`
	Hostname  string `json:"hostname"`
	Source    string `json:"source"`
	EventType string `json:"event_type"`
	Severity  string `json:"severity"`
	User      string `json:"user,omitempty"`
	Process   string `json:"process,omitempty"`
	Command   string `json:"command,omitempty"`
	RawLog    string `json:"raw_log"`
}

// Buffer интерфейс для буфера событий
type Buffer interface {
	// Push добавляет события в буфер
	Push(events []Event) error

	// Pop извлекает события из буфера
	Pop(maxCount int) []Event

	// Size возвращает текущий размер буфера
	Size() int

	// IsEmpty проверяет, пуст ли буфер
	IsEmpty() bool

	// Clear очищает буфер
	Clear()
}

// RingBuffer реализация кольцевого буфера
type RingBuffer struct {
	entries   []Event
	maxSize   int
	head      int
	tail      int
	count     int
	mu        sync.RWMutex
	notEmpty  *sync.Cond
	notFull   *sync.Cond
}

// NewRingBuffer создаёт новый кольцевой буфер
func NewRingBuffer(maxSize int) *RingBuffer {
	rb := &RingBuffer{
		entries:  make([]Event, maxSize),
		maxSize: maxSize,
	}
	rb.notEmpty = sync.NewCond(&rb.mu)
	rb.notFull = sync. NewCond(&rb.mu)
	return rb
}

// Push добавляет события в буфер
func (rb *RingBuffer) Push(events []Event) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// Ждём, пока будет свободное место
	for len(events) > rb.maxSize-rb.count {
		rb. notFull.Wait()
	}

	// Добавляем события в буфер
	for _, event := range events {
		rb.entries[rb.tail] = event
		rb.tail = (rb.tail + 1) % rb.maxSize
		rb.count++
	}

	// Сигнализируем, что буфер не пуст
	rb.notEmpty. Broadcast()

	return nil
}

// Pop извлекает события из буфера
func (rb *RingBuffer) Pop(maxCount int) []Event {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// Ждём, пока в буфере появятся события
	for rb.count == 0 {
		rb. notEmpty.Wait()
	}

	// Определяем количество событий для извлечения
	if maxCount > rb.count {
		maxCount = rb.count
	}

	// Извлекаем события из буфера
	result := make([]Event, maxCount)
	for i := 0; i < maxCount; i++ {
		result[i] = rb.entries[rb.head]
		rb. head = (rb.head + 1) % rb.maxSize
		rb.count--
	}

	// Сигнализируем, что в буфере есть свободное место
	rb.notFull.Broadcast()

	return result
}

// Size возвращает текущий размер буфера
func (rb *RingBuffer) Size() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// IsEmpty проверка на пустоту
func (rb *RingBuffer) IsEmpty() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count == 0
}

// Clear очищает буфер
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu. Unlock()
	rb.head = 0
	rb.tail = 0
	rb.count = 0
	rb.notFull.Broadcast()
}