package collector

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BashCollector struct {
	homeDir string
	user    string
	lastPos int64
}

func NewBashCollector(user, homeDir string) *BashCollector {
	return &BashCollector{
		homeDir:  homeDir,
		user:    user,
		lastPos: 0,
	}
}

func (bc *BashCollector) Collect() ([]Event, error) {
	var events []Event

	bashHistoryPath := filepath.Join(bc.homeDir, ".bash_history")

	file, err := os.Open(bashHistoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bash history file: %w", err)
	}
	defer file.Close()

	// Ищем последнюю прочитанную позицию
	file.Seek(bc. lastPos, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			event := bc. parseBashHistoryLine(line)
			events = append(events, event)
		}
	}

	// Сохраняем новую позицию
	if pos, err := file.Seek(0, 1); err == nil {
		bc.lastPos = pos
	}

	return events, scanner.Err()
}

func (bc *BashCollector) GetSourceName() string {
	return "bash_history_" + bc.user
}

func (bc *BashCollector) GetSourceType() string {
	return "bash_history"
}

func (bc *BashCollector) parseBashHistoryLine(rawLog string) Event {
	event := Event{
		RawLog:    rawLog,
		Source:    "bash_history",
		EventType: "command_execution",
		Severity:  determineBashCommandSeverity(rawLog),
		Timestamp: time.Now().Format(time.RFC3339),
		Hostname:  getHostname(),
		User:      bc.user,
		Process:   "bash",
		Command:   rawLog,
	}

	return event
}