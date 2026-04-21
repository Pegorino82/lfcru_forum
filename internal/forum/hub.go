package forum

import (
	"context"
	"errors"
	"sync"
)

// MaxSubscribersPerTopic is the per-topic SSE subscriber cap (CON-05).
const MaxSubscribersPerTopic = 200
const channelBufferSize = 16

// ErrTooManySubscribers is returned by Subscribe when the per-topic limit is reached.
var ErrTooManySubscribers = errors.New("too many SSE subscribers for this topic")

type subscriber struct {
	userID int64
	ch     chan string
}

// Hub is an in-process SSE broadcast hub.
// Safe for concurrent use from multiple goroutines.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[int64][]subscriber
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[int64][]subscriber),
	}
}

// Subscribe registers a new SSE subscriber for topicID.
// Returns a receive-only channel that receives rendered HTML fragments.
// The channel is closed automatically when ctx is cancelled.
// Returns ErrTooManySubscribers when the per-topic limit (200) is exceeded.
// Anonymous users should pass userID = 0.
func (h *Hub) Subscribe(ctx context.Context, topicID, userID int64) (<-chan string, error) {
	h.mu.Lock()
	if len(h.subscribers[topicID]) >= MaxSubscribersPerTopic {
		h.mu.Unlock()
		return nil, ErrTooManySubscribers
	}
	ch := make(chan string, channelBufferSize)
	h.subscribers[topicID] = append(h.subscribers[topicID], subscriber{userID: userID, ch: ch})
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.unsubscribe(topicID, ch)
	}()

	return ch, nil
}

// Broadcast sends fragment to all subscribers of topicID except the post author.
// Anonymous subscribers (userID == 0) always receive the event.
// Slow subscribers (full buffer) silently drop the message.
func (h *Hub) Broadcast(topicID, authorUserID int64, fragment string) {
	h.mu.RLock()
	subs := make([]subscriber, len(h.subscribers[topicID]))
	copy(subs, h.subscribers[topicID])
	h.mu.RUnlock()

	for _, sub := range subs {
		if sub.userID != 0 && sub.userID == authorUserID {
			continue
		}
		sub := sub // capture
		go func() {
			defer func() { recover() }() //nolint:errcheck
			select {
			case sub.ch <- fragment:
			default:
				// slow subscriber: drop
			}
		}()
	}
}

// unsubscribe removes ch from the subscriber list and closes the channel.
// If no subscribers remain for topicID, the map key is deleted.
func (h *Hub) unsubscribe(topicID int64, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.subscribers[topicID]
	for i, sub := range subs {
		if sub.ch == ch {
			subs[i] = subs[len(subs)-1]
			subs = subs[:len(subs)-1]
			break
		}
	}
	if len(subs) == 0 {
		delete(h.subscribers, topicID)
	} else {
		h.subscribers[topicID] = subs
	}
	close(ch)
}
