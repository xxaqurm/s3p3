package processor

import (
	"regexp"
	"strings"
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

// Processor интерфейс для обработчика событий
type Processor interface {
	// Process обрабатывает одно событие
	Process(event Event) (Event, bool)

	// ProcessBatch обрабатывает пакет событий
	ProcessBatch(events []Event) []Event
}

// LogProcessor реализация обработчика логов
type LogProcessor struct {
	anomalyRules []AnomalyRule
	filters      []Filter
}

// AnomalyRule правило обнаружения аномалии
type AnomalyRule struct {
	Name     string
	Pattern  *regexp.Regexp
	Severity string
	EventType string
}

// Filter правило фильтрации
type Filter struct {
	Name    string
	Pattern *regexp.Regexp
	Action  string // "drop", "alert", "pass"
}

// NewLogProcessor создаёт новый обработчик логов
func NewLogProcessor() *LogProcessor {
	return &LogProcessor{
		anomalyRules: initializeAnomalyRules(),
		filters:      initializeFilters(),
	}
}

// Process обрабатывает одно событие
func (lp *LogProcessor) Process(event Event) (Event, bool) {
	// Проверяем фильтры
	if ! lp.shouldProcess(event) {
		return event, false
	}

	// Нормализуем событие
	event = lp.normalize(event)

	// Обогащаем событие
	event = lp.enrich(event)

	// Обнаруживаем аномалии
	lp.detectAnomaly(&event)

	return event, true
}

// ProcessBatch обрабатывает пакет событий
func (lp *LogProcessor) ProcessBatch(events []Event) []Event {
	var result []Event

	for _, event := range events {
		processedEvent, ok := lp. Process(event)
		if ok {
			result = append(result, processedEvent)
		}
	}

	return result
}

// shouldProcess проверяет, нужно ли обрабатывать событие
func (lp *LogProcessor) shouldProcess(event Event) bool {
	// Проверяем фильтры
	for _, filter := range lp.filters {
		if filter.Pattern.MatchString(event.RawLog) {
			if filter.Action == "drop" {
				return false
			}
		}
	}

	// Не обрабатываем пустые сообщения
	if strings.TrimSpace(event.RawLog) == "" {
		return false
	}

	return true
}

// normalize нормализует формат события
func (lp *LogProcessor) normalize(event Event) Event {
	// Нормализуем уровень серьёзности
	event.Severity = strings.ToUpper(event.Severity)
	if ! isValidSeverity(event. Severity) {
		event.Severity = "INFO"
	}

	// Нормализуем тип события
	event.EventType = strings.ToLower(event.EventType)

	// Очищаем сообщение от лишних символов
	event. RawLog = strings.TrimSpace(event.RawLog)

	return event
}

// enrich обогащает событие дополнительной информацией
func (lp *LogProcessor) enrich(event Event) Event {
	// Извлекаем IP адреса если есть
	ipPattern := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	ips := ipPattern.FindAllString(event.RawLog, -1)
	if len(ips) > 0 {
		// Можно добавить IP в metadata если понадобится
	}

	// Извлекаем port
	portPattern := regexp.MustCompile(`port\s+(\d+)`)
	ports := portPattern.FindAllStringSubmatch(event.RawLog, -1)
	if len(ports) > 0 {
		// Можно добавить port в metadata если понадобится
	}

	return event
}

// detectAnomaly обнаруживает подозрительную активность
func (lp *LogProcessor) detectAnomaly(event *Event) bool {
	for _, rule := range lp.anomalyRules {
		if rule.Pattern.MatchString(event.RawLog) {
			event.Severity = rule.Severity
			event.EventType = rule.EventType
			return true
		}
	}
	return false
}

// initializeAnomalyRules инициализирует правила обнаружения аномалий
func initializeAnomalyRules() []AnomalyRule {
	return []AnomalyRule{
		{
			Name:      "SQL_Injection_Attempt",
			Pattern:   regexp. MustCompile(`(?i)(union|select|insert|drop|delete|update).*(from|into|table)`),
			Severity: "CRITICAL",
			EventType: "sql_injection",
		},
		{
			Name:      "Brute_Force_SSH",
			Pattern:   regexp. MustCompile(`(?i)failed password.*ssh`),
			Severity: "CRITICAL",
			EventType: "brute_force_attack",
		},
		{
			Name:      "Privilege_Escalation",
			Pattern:   regexp.MustCompile(`(?i)(sudo|su\s).*(root|wheel)`),
			Severity: "WARNING",
			EventType: "privilege_escalation",
		},
		{
			Name:      "Unauthorized_Access",
			Pattern:   regexp.MustCompile(`(?i)(unauthorized|denied|permission|forbidden)`),
			Severity: "WARNING",
			EventType: "unauthorized_access",
		},
		{
			Name:      "File_Integrity_Violation",
			Pattern:    regexp.MustCompile(`(?i)(chmod|chown).*(777|755)`),
			Severity: "WARNING",
			EventType: "file_integrity",
		},
		{
			Name:      "Dangerous_Command",
			Pattern:   regexp.MustCompile(`(?i)(rm\s+-rf|dd\s+if=|mkfs|fdisk|parted)`),
			Severity: "CRITICAL",
			EventType: "dangerous_command",
		},
	}
}

// initializeFilters инициализирует фильтры событий
func initializeFilters() []Filter {
	return []Filter{
		{
			Name:     "Ignore_Systemd_Messages",
			Pattern: regexp.MustCompile(`(?i)systemd\[.*\]:`),
			Action:  "pass", // пропускаем systemd сообщения
		},
		{
			Name:    "Ignore_CRON",
			Pattern: regexp.MustCompile(`(?i)CRON\[`),
			Action:  "pass", // пропускаем CRON события
		},
	}
}

// isValidSeverity проверяет, корректен ли уровень серьёзности
func isValidSeverity(severity string) bool {
	validSeverities := map[string]bool{
		"INFO":      true,
		"WARNING":  true,
		"CRITICAL":  true,
	}
	return validSeverities[severity]
}