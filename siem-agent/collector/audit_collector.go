package collector

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type AuditCollector struct {
	filePath string
	lastPos  int64
}

func NewAuditCollector(filePath string) *AuditCollector {
	return &AuditCollector{
		filePath: filePath,
		lastPos:  0,
	}
}

func (ac *AuditCollector) Collect() ([]Event, error) {
	var events []Event

	file, err := os.Open(ac.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer file.Close()

	// Ищем последнюю прочитанную позицию
	file. Seek(ac.lastPos, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			event := ac.parseAuditLine(line)
			events = append(events, event)
		}
	}

	// Сохраняем новую позицию
	if pos, err := file.Seek(0, 1); err == nil {
		ac.lastPos = pos
	}

	return events, scanner.Err()
}

func (ac *AuditCollector) GetSourceName() string {
	return "audit"
}

func (ac *AuditCollector) GetSourceType() string {
	return "audit"
}

func (ac *AuditCollector) parseAuditLine(rawLog string) Event {
	event := Event{
		RawLog:    rawLog,
		Source:    "audit",
		Timestamp: extractAuditTimestamp(rawLog),
		Hostname:  getHostname(),
		Process:   "auditd",
	}

	// Извлекаем тип события
	event.EventType = extractAuditEventType(rawLog)

	// Определяем серьёзность
	event.Severity = determineAuditSeverity(event.EventType)

	// Извлекаем команду
	if command := extractAuditCommand(rawLog); command != "" {
		event.Command = command
	}

	// Извлекаем пользователя
	if user := extractAuditUser(rawLog); user != "" {
		event.User = user
	}

	return event
}