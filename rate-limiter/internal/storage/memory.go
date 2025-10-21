package storage

import (
	"context"
	"sync"
	"time"
)

type MemoryItem struct {
	Value     int64
	ExpiresAt time.Time
}

type MemoryStorage struct {
	data   map[string]*MemoryItem
	mutex  sync.RWMutex
	ticker *time.Ticker
	done   chan bool
}

func NewMemoryStorage() *MemoryStorage {
	ms := &MemoryStorage{
		data:   make(map[string]*MemoryItem),
		ticker: time.NewTicker(time.Minute),
		done:   make(chan bool),
	}

	go ms.cleanup()

	return ms
}

func (m *MemoryStorage) cleanup() {
	for {
		select {
		case <-m.ticker.C:
			m.mutex.Lock()
			now := time.Now()
			for key, item := range m.data {
				if now.After(item.ExpiresAt) {
					delete(m.data, key)
				}
			}
			m.mutex.Unlock()
		case <-m.done:
			return
		}
	}
}

func (m *MemoryStorage) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	expiresAt := now.Add(ttl)

	if item, exists := m.data[key]; exists && now.Before(item.ExpiresAt) {
		item.Value++
		return item.Value, nil
	}

	m.data[key] = &MemoryItem{
		Value:     1,
		ExpiresAt: expiresAt,
	}

	return 1, nil
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (int64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		return 0, nil
	}

	return item.Value, nil
}

func (m *MemoryStorage) Set(ctx context.Context, key string, value int64, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data[key] = &MemoryItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	return nil
}

func (m *MemoryStorage) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if item, exists := m.data[key]; exists {
		item.ExpiresAt = time.Now().Add(ttl)
	}

	return nil
}

func (m *MemoryStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return 0, nil
	}

	remaining := time.Until(item.ExpiresAt)
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

func (m *MemoryStorage) Close() error {
	m.ticker.Stop()
	close(m.done)
	return nil
}
