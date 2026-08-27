// Package events provides small in-process pub/sub hubs that let
// background jobs notify HTTP handlers (typically SSE streams) without
// either side depending on the other's internals.
package events

import "sync"

// subscriberBuffer is generous enough that a normal doujinshi archive's
// entire page count can be published before a subscriber starts reading
// without dropping events; see Publish.
const subscriberBuffer = 32

// ThumbnailEvent describes a change in an archive's thumbnail generation
// progress, delivered to SSE subscribers.
type ThumbnailEvent struct {
	// Index is the page index that just became ready. Meaningless when Done.
	Index int
	// Done is true once thumbnail generation for the archive has reached a
	// terminal state (succeeded, or permanently failed after exhausting
	// retries) — no further events will be published for it after this.
	Done bool
	// Error is set when Done is true and generation ended in permanent
	// failure (all retries exhausted). Empty on success.
	Error string
}

// ThumbnailBroadcaster is an in-process pub/sub hub for thumbnail
// generation progress, keyed by archive ID.
type ThumbnailBroadcaster struct {
	mu   sync.Mutex
	subs map[string]map[chan ThumbnailEvent]struct{}
}

func NewThumbnailBroadcaster() *ThumbnailBroadcaster {
	return &ThumbnailBroadcaster{
		subs: make(map[string]map[chan ThumbnailEvent]struct{}),
	}
}

// Subscribe registers a new listener for archiveID's events. The caller
// must call the returned cancel func (typically via defer) once it stops
// reading, or the channel will leak.
func (b *ThumbnailBroadcaster) Subscribe(archiveID string) (<-chan ThumbnailEvent, func()) {
	ch := make(chan ThumbnailEvent, subscriberBuffer)

	b.mu.Lock()
	if b.subs[archiveID] == nil {
		b.subs[archiveID] = make(map[chan ThumbnailEvent]struct{})
	}

	b.subs[archiveID][ch] = struct{}{}
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if subs, ok := b.subs[archiveID]; ok {
			if _, ok := subs[ch]; ok {
				delete(subs, ch)
				close(ch)
			}

			if len(subs) == 0 {
				delete(b.subs, archiveID)
			}
		}
	}

	return ch, cancel
}

// Publish sends event to every current subscriber of archiveID.
// Non-blocking: a subscriber that isn't keeping up has the event dropped
// rather than stalling the publisher (a background job worker).
func (b *ThumbnailBroadcaster) Publish(archiveID string, event ThumbnailEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.subs[archiveID] {
		select {
		case ch <- event:
		default:
		}
	}
}
