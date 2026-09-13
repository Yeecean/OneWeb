// Package events 实现进程内事件总线。
package events

import (
	"sync"

	"github.com/yeecean/oneweb/internal/domain"
)

// EventBus 提供订阅/发布模式的内存事件总线。
// Channel 满时非阻塞丢弃（避免慢消费者阻塞生产者），并记录告警。
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[domain.EventType][]chan domain.DomainEvent
	BufferSize  int
}

// NewEventBus 创建事件总线，默认缓冲 64。
func NewEventBus(bufferSize int) *EventBus {
	if bufferSize <= 0 {
		bufferSize = 64
	}
	return &EventBus{
		subscribers: map[domain.EventType][]chan domain.DomainEvent{},
		BufferSize:  bufferSize,
	}
}

// Subscribe 订阅指定事件类型，返回接收 channel。
func (b *EventBus) Subscribe(eventType domain.EventType) <-chan domain.DomainEvent {
	ch := make(chan domain.DomainEvent, b.BufferSize)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
	return ch
}

// Unsubscribe 移除订阅。
func (b *EventBus) Unsubscribe(ch <-chan domain.DomainEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for et, chans := range b.subscribers {
		for i, c := range chans {
			if c == ch {
				b.subscribers[et] = append(chans[:i], chans[i+1:]...)
				break
			}
		}
	}
}

// Publish 向指定事件类型的全部订阅者广播。
func (b *EventBus) Publish(event domain.DomainEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers[event.Type] {
		select {
		case ch <- event:
		default:
			// 缓冲区满：丢弃该事件避免阻塞
		}
	}
}
