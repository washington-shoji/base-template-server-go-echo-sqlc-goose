package events

import (
	"context"
	"sync"

	"go-echo-server-template/internal/platform/logging"
)

// Event is a domain event payload.
type Event struct {
	Name    string
	Payload any
}

type Handler func(ctx context.Context, event Event) error

// Bus is a synchronous in-process event bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

func (b *Bus) Subscribe(name string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[name] = append(b.handlers[name], h)
}

func (b *Bus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	hs := append([]Handler(nil), b.handlers[event.Name]...)
	b.mu.RUnlock()

	log := logging.WithContext(ctx, "events")
	for _, h := range hs {
		if err := h(ctx, event); err != nil {
			log.Error("event handler failed", err, map[string]interface{}{"event": event.Name})
		}
	}
}
