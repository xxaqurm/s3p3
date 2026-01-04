package collector

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Event - структура нормализованного события
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

// Collector интерфейс для сборщиков логов
type Collector interface {
	// Collect собирает события из источника
	Collect() ([]Event, error)

	// GetSourceName возвращает имя источника логов
	GetSourceName() string

	// GetSourceType возвращает тип источника (syslog, audit, bash_history и т.д.)
	GetSourceType() string
}


func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func extractUser(log string) string {
	patterns := []string{
		`user[=\s]+([a-zA-Z0-9\-_]+)`,
		`for ([a-zA-Z0-9\-_]+)`,
		`sudo:\s+([a-zA-Z0-9\-_]+)\s+:`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(log)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func determineSeverity(log string) string {
	logUpper := strings.ToUpper(log)

	if strings.Contains(logUpper, "ERROR") || strings.Contains(logUpper, "CRITICAL") || strings.Contains(logUpper, "FATAL") {
		return "CRITICAL"
	}
	if strings.Contains(logUpper, "WARN") || strings.Contains(logUpper, "ALERT") {
		return "WARNING"
	}
	if strings.Contains(logUpper, "NOTICE") || strings.Contains(logUpper, "INFO") {
		return "INFO"
	}

	return "INFO"
}

func extractAuthLogUser(log string) string {
	patterns := []string{
		`(?:for invalid user|for)\s+([a-zA-Z0-9\-_]+)`,
		`sudo:\s+([a-zA-Z0-9\-_]+)\s+:`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(log)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func extractAuthLogProcess(log string) string {
	re := regexp. MustCompile(`^.*\s([a-zA-Z0-9\-_]+)\[?\d*\]? :\s`)
	matches := re.FindStringSubmatch(log)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func determineBashCommandSeverity(command string) string {
	commandUpper := strings.ToUpper(command)

	dangerousPatterns := []string{
		"RM -RF", "DD IF=", "MKFS", "FDISK", "PARTED",
		"CHMOD 777", "CHOWN", "SUDO", "SU ",
		"PASSWD", "USERADD", "USERDEL", "GROUPADD",
		"IPTABLES", "UFW", "FIREWALL", "SYSTEMCTL STOP",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(commandUpper, pattern) {
			return "WARNING"
		}
	}

	return "INFO"
}

func determineAuditSeverity(eventType string) string {
	criticalEvents := []string{"EXECVE", "OPEN", "CONNECT", "SOCKET", "PTRACE", "KILL"}

	for _, event := range criticalEvents {
		if eventType == event {
			return "WARNING"
		}
	}

	return "INFO"
}

func extractAuditEventType(log string) string {
	re := regexp.MustCompile(`type=([A-Z_]+)`)
	matches := re.FindStringSubmatch(log)
	if len(matches) > 1 {
		return matches[1]
	}
	return "UNKNOWN"
}

func extractAuditTimestamp(log string) string {
	// Извлекаем timestamp из audit:  msg=audit(1234567890.123: 4567)
	re := regexp.MustCompile(`audit\((\d+)\.(\d+):`)
	matches := re.FindStringSubmatch(log)
	if len(matches) > 2 {
		unixTime, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			t := time.Unix(unixTime, 0)
			return t.Format(time.RFC3339)
		}
	}
	return time.Now().Format(time.RFC3339)
}

func extractAuditCommand(log string) string {
	// Ищет команду в EXECVE событиях:  a0="command" a1="arg1" ... 
	re := regexp.MustCompile(`a0="([^"]+)"`)
	matches := re.FindStringSubmatch(log)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractAuditUser(log string) string {
	// Ищет uid
	re := regexp.MustCompile(`uid=(\d+)`)
	matches := re.FindStringSubmatch(log)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractSyslogService(log string) (service, eventType string) {
	// Ищем паттерны типа:  "sudo:", "kernel:", "systemd:", "sshd:" и т.д.
	re := regexp.MustCompile(`^.*\s([a-zA-Z0-9\-_]+)\[?\d*\]?:\s`)
	matches := re.FindStringSubmatch(log)
	if len(matches) > 1 {
		service = matches[1]
	}

	// Определяем тип события по сервису
	switch service {
	case "kernel":
		eventType = "kernel_message"
	case "sudo":
		eventType = "sudo_execution"
	case "sshd":
		eventType = "ssh_event"
	case "systemd": 
		eventType = "systemd_event"
	default: 
		eventType = "system_event"
	}

	return service, eventType
}