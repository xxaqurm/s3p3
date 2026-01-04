package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent/agent"
	"agent/buffer"
	"agent/collector"
	"agent/processor"
	"agent/sender"
)

type Config struct {
	Agent struct {
		ID string `yaml:"id"`
	} `yaml:"agent"`
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Logging struct {
		Sources            []string `yaml:"sources"`
		CollectionInterval int      `yaml:"collection_interval"`
		SendInterval       int      `yaml:"send_interval"`
		BatchSize          int      `yaml:"batch_size"`
		BufferMaxSize      int      `yaml:"buffer_max_size"`
	} `yaml:"logging"`
}

func main() {
	// Парсим флаги командной строки
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	// Загружаем конфигурацию
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("========================================")
	log.Println("SIEM Agent v1.0")
	log.Println("========================================")
	log.Printf("Agent ID: %s\n", config.Agent. ID)
	log.Printf("Server: %s:%d\n", config.Server.Host, config.Server.Port)
	log.Printf("Collection Interval: %d ms\n", config. Logging.CollectionInterval)
	log.Printf("Send Interval: %d ms\n", config. Logging.SendInterval)
	log.Println("========================================")

	// Создаем компоненты
	rbuffer := buffer.NewRingBuffer(config.Logging.BufferMaxSize)
	processorInstance := processor.NewLogProcessor()
	senderInstance := sender.NewTCPSender(config.Server. Host, config.Server.Port)

	// Создаём конфигурацию агента
	agentConfig := agent.Config{
		AgentID:            config.Agent.ID,
		ServerHost:         config.Server. Host,
		ServerPort:         config.Server.Port,
		CollectionInterval: config.Logging.CollectionInterval,
		SenderInterval:     config.Logging.SendInterval,
		BatchSize:          config.Logging.BatchSize,
		BufferMaxSize:      config.Logging.BufferMaxSize,
	}

	// Создаём агент
	siem := agent. NewAgent(agentConfig, rbuffer, processorInstance, senderInstance)

	// Регистрируем сборщики логов
	log.Println("\n[Main] Registering log collectors...")

	// Syslog collector
	if len(config.Logging.Sources) > 0 {
		for _, source := range config.Logging.Sources {
			switch source {
			case "syslog":
				siem.RegisterCollector(collector.NewSyslogCollector("/var/log/syslog"))
				log. Println("[Main] Syslog collector registered")
			case "auditd":
				siem.RegisterCollector(collector.NewAuditCollector("/var/log/audit/audit.log"))
				log.Println("[Main] Audit collector registered")
			case "bash_history":
				// Регистрируем bash_history для текущего пользователя
				currentUser := os.Getenv("USER")
				homeDir := os.Getenv("HOME")
				if currentUser != "" && homeDir != "" {
					siem.RegisterCollector(collector.NewBashCollector(currentUser, homeDir))
					log.Println("[Main] Bash history collector registered")
				}
			}
		}
	}

	// Обработчик сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем агент
	if err := siem.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// Горутина для мониторинга статуса
	go monitorAgent(siem)

	// Ждём сигнала завершения
	sig := <-sigChan
	log.Printf("\n[Main] Received signal: %v", sig)

	// Останавливаем агент
	siem. Stop()

	log.Println("[Main] Exiting...")
}

// monitorAgent мониторит статус агента и выводит статистику
func monitorAgent(siem *agent.Agent) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !siem.IsRunning() {
			break
		}
		bufferSize := siem.GetBufferSize()
		log.Printf("[Monitor] Buffer size: %d events", bufferSize)
	}
}

// loadConfig загружает конфигурацию из YAML файла
func loadConfig(path string) (Config, error) {
	// Для простоты используем встроенную конфиг
	// В реальности здесь была бы загрузка из YAML файла
	return Config{
		Agent: struct {
			ID string `yaml:"id"`
		}{ID: "agent-ubuntu-01"},
		Server: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{Host: "127.0.0.1", Port: 8080},
		Logging: struct {
			Sources            []string `yaml:"sources"`
			CollectionInterval int      `yaml:"collection_interval"`
			SendInterval       int      `yaml:"send_interval"`
			BatchSize          int      `yaml:"batch_size"`
			BufferMaxSize      int      `yaml:"buffer_max_size"`
		}{
			Sources:            []string{"syslog", "auditd", "bash_history"},
			CollectionInterval: 5000,
			SendInterval:       10000,
			BatchSize:          100,
			BufferMaxSize:       10000,
		},
	}, nil
}