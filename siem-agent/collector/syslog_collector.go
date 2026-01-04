package collector

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type SyslogCollector struct {
	filePath string
	lastPos  int64
}

func NewSyslogCollector(filePath string) *SyslogCollector {
	return &SyslogCollector{
		filePath: filePath,
		lastPos:  0,
	}
}

func (sc *SyslogCollector) Collect() ([]Event, error) {
	var events []Event

	file, err := os.Open(sc.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open syslog file: %w", err)
	}
	defer file.Close()

	// Ищем последнюю прочитанную позицию
	file.Seek(sc.lastPos, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			event := sc.parseSyslogLine(line)
			events = append(events, event)
		}
	}

	// Сохраняем новую позицию
	if pos, err := file.Seek(0, 1); err == nil {
		sc.lastPos = pos
	}

	return events, scanner.Err()
}

func (sc *SyslogCollector) GetSourceName() string {
	return "syslog"
}

func (sc *SyslogCollector) GetSourceType() string {
	return "syslog"
}

func (sc *SyslogCollector) parseSyslogLine(rawLog string) Event {
	event := Event{
		RawLog:     rawLog,
		Source:    "syslog",
		Timestamp: time.Now().Format(time.RFC3339),
		Hostname:  getHostname(),
		Severity:  determineSeverity(rawLog),
	}

	// Извлекаем сервис/процесс
	process, eventType := extractSyslogService(rawLog)
	event.Process = process
	event.EventType = eventType

	// Извлекаем пользователя если есть
	if user := extractUser(rawLog); user != "" {
		event.User = user
	}

	return event
}