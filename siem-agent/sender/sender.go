package sender

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// Event структура события (копируем для независимости пакета)
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

// Response ответ от сервера
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// Sender интерфейс для отправителя событий
type Sender interface {
	// Send отправляет события на сервер
	Send(events []Event) error

	// IsConnected проверяет, доступен ли сервер
	IsConnected() bool

	// Close закрывает соединение
	Close() error
}

// TCPSender отправляет события по TCP
type TCPSender struct {
	host    string
	port    int
	conn    net.Conn
	timeout time.Duration
}

// NewTCPSender создаёт новый TCP отправитель
func NewTCPSender(host string, port int) *TCPSender {
	return &TCPSender{
		host:    host,
		port:    port,
		timeout: 10 * time.Second,
	}
}

// Send отправляет события на сервер
func (ts *TCPSender) Send(events []Event) error {
	if len(events) == 0 {
		return nil
	}

	// Подключаемся к серверу если не подключены
	if ts.conn == nil {
		if err := ts.connect(); err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
	}

	// Отправляем каждое событие
	for _, event := range events {
		if err := ts.sendEvent(event); err != nil {
			// Закрываем соединение при ошибке
			ts. conn.Close()
			ts.conn = nil
			return fmt.Errorf("failed to send event: %w", err)
		}
	}

	return nil
}

// IsConnected проверяет, доступен ли сервер
func (ts *TCPSender) IsConnected() bool {
	if ts.conn != nil {
		// Проверяем, живо ли соединение, отправляя пинг
		ts.conn.SetReadDeadline(time.Now().Add(ts.timeout))
		buf := make([]byte, 1)
		_, err := ts.conn.Read(buf)
		ts.conn.SetReadDeadline(time.Time{})

		if err == nil {
			return true
		}

		// Проверяем, был ли это timeout или реальная ошибка
		if os.IsTimeout(err) {
			return true
		}

		ts.conn = nil
		return false
	}

	// Пытаемся подключиться
	return ts.checkConnection()
}

// Close закрывает соединение
func (ts *TCPSender) Close() error {
	if ts.conn != nil {
		return ts.conn.Close()
	}
	return nil
}

// connect подключается к серверу
func (ts *TCPSender) connect() error {
	addr := fmt.Sprintf("%s:%d", ts.host, ts. port)
	log.Printf("[Sender] Connecting to server at %s", addr)
	conn, err := net. DialTimeout("tcp", addr, ts.timeout)
	if err != nil {
		log.Printf("[Sender] Failed to connect:  %v", err)
		return err
	}
	log.Printf("[Sender] Connected to server")
	ts.conn = conn
	return nil
}

// checkConnection проверяет, возможно ли подключиться к серверу
func (ts *TCPSender) checkConnection() bool {
	addr := fmt.Sprintf("%s:%d", ts.host, ts. port)
	conn, err := net.DialTimeout("tcp", addr, ts.timeout)
	if err != nil {
		log.Printf("[Sender] Server check failed: %v", err)
		return false
	}
	conn.Close()
	return true
}

// sendEvent отправляет одно событие на сервер
func (ts *TCPSender) sendEvent(event Event) error {
	// Формируем команду в формате: database collection action jsonArg
	database := "security_db"
	collection := "security_events.json"
	action := "insert"

	// Сериализуем событие в JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Формируем команду
	command := fmt.Sprintf("%s %s %s %s\n", database, collection, action, string(jsonData))

	log.Printf("[Sender] Sending command: %s", command[: 100]) // Логируем первые 100 символов

	// Отправляем команду
	_, err = ts.conn.Write([]byte(command))
	if err != nil {
		return err
	}

	// Читаем ответ от сервера
	response, err := ts.readResponse()
	if err != nil {
		return err
	}

	log.Printf("[Sender] Server response: %s - %s", response.Status, response.Message)

	// Проверяем статус ответа
	if response.Status != "success" {
		return fmt.Errorf("server error: %s", response.Message)
	}

	return nil
}

// readResponse читает ответ от сервера
func (ts *TCPSender) readResponse() (Response, error) {
	var response Response
	var buf strings.Builder

	// Читаем до символа новой строки
	for {
		b := make([]byte, 1)
		n, err := ts. conn.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			return response, err
		}

		if n == 0 {
			break
		}

		if b[0] == '\n' {
			break
		}

		buf.WriteByte(b[0])
	}

	// Парсим JSON ответ
	err := json.Unmarshal([]byte(buf.String()), &response)
	if err != nil {
		return response, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, nil
}