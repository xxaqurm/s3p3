package buffer

import (
	"encoding/json"
	"fmt"
	"os"
	"agent/collector"
	"sync"
	//"time"
)

type Buffer struct {
	events     []collector.Event
	maxSize    int
	mu         sync.RWMutex
	backupFile string
}

func NewBuffer(maxSize int, backupFile string) *Buffer {
	b := &Buffer{
		events:     make([]collector.Event, 0, maxSize),
		maxSize:    maxSize,
		backupFile: backupFile,
	}
	b.restoreFromBackup()
	return b
}

func (b *Buffer) Add(event collector.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.events = append(b.events, event)
	
	// Если буфер переполнен, сохраняем на диск
	if len(b.events) >= b.maxSize {
		b.backupToDisk()
	}
}

func (b *Buffer) GetBatch(size int) []collector.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if len(b.events) == 0 {
		return nil
	}
	
	batchSize := size
	if batchSize > len(b.events) {
		batchSize = len(b.events)
	}
	
	batch := b.events[:batchSize]
	b.events = b.events[batchSize:]
	
	return batch
}

func (b *Buffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.events)
}

func (b *Buffer) backupToDisk() {
	if b.backupFile == "" {
		return
	}
	
	data, err := json.Marshal(b.events)
	if err != nil {
		fmt.Printf("Ошибка маршалинга буфера: %v\n", err)
		return
	}
	
	err = os.WriteFile(b.backupFile, data, 0644)
	if err != nil {
		fmt.Printf("Ошибка записи буфера на диск: %v\n", err)
	}
}

func (b *Buffer) restoreFromBackup() {
	if b.backupFile == "" {
		return
	}
	
	if _, err := os.Stat(b.backupFile); os.IsNotExist(err) {
		return
	}
	
	data, err := os.ReadFile(b.backupFile)
	if err != nil {
		fmt.Printf("Ошибка чтения бэкапа: %v\n", err)
		return
	}
	
	var events []collector.Event
	err = json.Unmarshal(data, &events)
	if err != nil {
		fmt.Printf("Ошибка демаршалинга бэкапа: %v\n", err)
		return
	}
	
	b.events = events
	
	// Удаляем бэкап после восстановления
	os.Remove(b.backupFile)
}