package events

import (
	"sync"
	"testing"

	"github.com/yeecean/oneweb/internal/domain"
)

func TestEventBusSubscribePublish(t *testing.T) {
	bus := NewEventBus(8)
	ch := bus.Subscribe(domain.EvtOpStarted)
	defer bus.Unsubscribe(ch)

	bus.Publish(domain.DomainEvent{ID: "op-1", Type: domain.EvtOpStarted, ProfileID: "p1"})

	select {
	case ev := <-ch:
		if ev.ID != "op-1" || ev.ProfileID != "p1" {
			t.Fatalf("unexpected event: %+v", ev)
		}
	default:
		t.Fatal("event not received")
	}
}

func TestEventBusTypeFilter(t *testing.T) {
	bus := NewEventBus(8)
	ch := bus.Subscribe(domain.EvtOpLog)
	defer bus.Unsubscribe(ch)

	// 发布不同类型事件，不应投递到本订阅
	bus.Publish(domain.DomainEvent{Type: domain.EvtOpStarted})
	select {
	case ev := <-ch:
		t.Fatalf("wrong event delivered: %+v", ev)
	default:
		// 正确：未收到
	}
	bus.Publish(domain.DomainEvent{Type: domain.EvtOpLog})
	select {
	case <-ch:
		// 正确：收到日志事件
	default:
		t.Fatal("log event not delivered")
	}
}

func TestEventBusConcurrentSafety(t *testing.T) {
	bus := NewEventBus(16)
	var wg sync.WaitGroup
	// 并发发布
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				bus.Publish(domain.DomainEvent{ID: "e", Type: domain.EvtOpStarted})
			}
		}(i)
	}
	// 并发订阅/退订
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := bus.Subscribe(domain.EvtOpStarted)
			for j := 0; j < 50; j++ {
				bus.Publish(domain.DomainEvent{ID: "e", Type: domain.EvtOpStarted})
			}
			bus.Unsubscribe(ch)
		}()
	}
	wg.Wait()
	// 若竞态未修复会 panic（go test -race 覆盖）
}
