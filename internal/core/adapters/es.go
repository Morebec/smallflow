package adapters

import (
	"context"
	"github.com/morebec/go-misas/misas"
	"sync"
)

type InMemoryEventStore struct {
	mu     sync.Mutex
	events []any
}

func (i *InMemoryEventStore) Record(ctx context.Context, event any) misas.Error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.events = append(i.events, event)
	return nil
}

func (i *InMemoryEventStore) Events() []any {
	i.mu.Lock()
	defer i.mu.Unlock()
	eventsCopy := make([]any, len(i.events))
	copy(eventsCopy, i.events)
	return eventsCopy
}
