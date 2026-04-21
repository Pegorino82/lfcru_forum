package forum

import (
	"context"
	"testing"
	"time"
)

func TestHub_Subscribe_ReceivesBroadcast(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := h.Subscribe(ctx, 1, 42)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	h.Broadcast(1, 99, "hello")

	select {
	case msg := <-ch:
		if msg != "hello" {
			t.Errorf("expected %q, got %q", "hello", msg)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for broadcast")
	}
}

func TestHub_Broadcast_SkipsAuthor(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := h.Subscribe(ctx, 1, 42)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Broadcast with authorUserID == subscriber.userID — should be skipped
	h.Broadcast(1, 42, "should not arrive")

	select {
	case msg := <-ch:
		t.Errorf("author should not receive own broadcast, got %q", msg)
	case <-time.After(100 * time.Millisecond):
		// expected: no message
	}
}

func TestHub_Broadcast_AnonymousAlwaysReceives(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Anonymous subscriber (userID == 0)
	ch, err := h.Subscribe(ctx, 1, 0)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	// Even if authorUserID == 0, anon subscribers must receive
	h.Broadcast(1, 99, "for anon")

	select {
	case msg := <-ch:
		if msg != "for anon" {
			t.Errorf("expected %q, got %q", "for anon", msg)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for broadcast")
	}
}

// TestHub_CtxCancel verifies that cancelling ctx removes the subscriber and deletes the topic key (CHK-03).
func TestHub_CtxCancel_RemovesSubscriber(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())

	_, err := h.Subscribe(ctx, 1, 10)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	cancel()
	// Allow the cleanup goroutine to run
	time.Sleep(50 * time.Millisecond)

	h.mu.RLock()
	_, exists := h.subscribers[1]
	h.mu.RUnlock()

	if exists {
		t.Error("topic key should be deleted after last subscriber cancels")
	}
}

// TestHub_TopicIsolation verifies that broadcast to topic A does not reach topic B subscribers.
func TestHub_TopicIsolation(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chA, _ := h.Subscribe(ctx, 1, 10)
	chB, _ := h.Subscribe(ctx, 2, 20)

	h.Broadcast(2, 99, "for topic 2")

	select {
	case msg := <-chB:
		if msg != "for topic 2" {
			t.Errorf("topic B: expected %q, got %q", "for topic 2", msg)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("topic B: timed out waiting for broadcast")
	}

	select {
	case msg := <-chA:
		t.Errorf("topic A should not receive topic B broadcast, got %q", msg)
	case <-time.After(100 * time.Millisecond):
		// expected: no message
	}
}

// TestHub_SubscriberLimit verifies that the 201st subscriber gets ErrTooManySubscribers.
func TestHub_SubscriberLimit(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < MaxSubscribersPerTopic; i++ {
		_, err := h.Subscribe(ctx, 1, int64(i))
		if err != nil {
			t.Fatalf("subscriber %d: unexpected error: %v", i, err)
		}
	}

	_, err := h.Subscribe(ctx, 1, 999)
	if err != ErrTooManySubscribers {
		t.Errorf("expected ErrTooManySubscribers, got %v", err)
	}
}

// TestHub_MultipleSubscribers_AllReceive verifies all non-author subscribers receive the broadcast.
func TestHub_MultipleSubscribers_AllReceive(t *testing.T) {
	h := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1, _ := h.Subscribe(ctx, 1, 10)
	ch2, _ := h.Subscribe(ctx, 1, 20)
	ch3, _ := h.Subscribe(ctx, 1, 30)

	h.Broadcast(1, 99, "group message")

	for i, ch := range []<-chan string{ch1, ch2, ch3} {
		select {
		case msg := <-ch:
			if msg != "group message" {
				t.Errorf("subscriber %d: expected %q, got %q", i+1, "group message", msg)
			}
		case <-time.After(200 * time.Millisecond):
			t.Errorf("subscriber %d: timed out", i+1)
		}
	}
}
