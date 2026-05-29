package core

import (
	"strings"
	"sync"
	"time"
)

// ProgressReporter buffers agent output and sends throttled updates.
type ProgressReporter struct {
	mu       sync.Mutex
	buf      strings.Builder
	interval time.Duration
	lastSent time.Time
	sendFn   func(text string)
}

func NewProgressReporter(interval time.Duration, sendFn func(string)) *ProgressReporter {
	if interval <= 0 {
		interval = 3 * time.Second
	}
	return &ProgressReporter{
		interval: interval,
		sendFn:   sendFn,
	}
}

// Write adds text to the buffer. If enough time has passed since the last
// send, it triggers a progress update.
func (r *ProgressReporter) Write(text string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buf.WriteString(text)

	if time.Since(r.lastSent) >= r.interval {
		r.flushLocked()
	}
}

// Final returns the full accumulated output without sending it.
// The caller is responsible for sending the final message.
func (r *ProgressReporter) Final() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	text := r.buf.String()
	r.buf.Reset()
	return text
}

func (r *ProgressReporter) flushLocked() {
	if r.buf.Len() == 0 {
		return
	}
	if r.sendFn != nil {
		r.sendFn(strings.TrimSpace(r.buf.String()))
	}
	r.lastSent = time.Now()
}
