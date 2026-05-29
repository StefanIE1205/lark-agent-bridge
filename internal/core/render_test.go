package core

import (
	"testing"
	"time"
)

func TestProgressReporterWrites(t *testing.T) {
	var sent []string
	r := NewProgressReporter(50*time.Millisecond, func(text string) {
		sent = append(sent, text)
	})

	r.Write("hello")
	r.Write(" world")

	// Should not have sent yet (interval hasn't passed)
	time.Sleep(60 * time.Millisecond)

	r.Write(" more")

	// Now it should trigger a flush
	if len(sent) == 0 {
		t.Error("expected at least one progress update")
	}
}

func TestProgressReporterFinal(t *testing.T) {
	var sent []string
	r := NewProgressReporter(time.Hour, func(text string) {
		sent = append(sent, text)
	})

	r.Write("hello")
	r.Write(" world")

	// Note: first Write may trigger flush due to zero lastSent — that's OK.
	// The important thing is Final() itself does NOT call sendFn.
	beforeFinal := len(sent)
	final := r.Final()

	if final != "hello world" {
		t.Errorf("Final() = %q, want %q", final, "hello world")
	}
	// Final() should NOT add more sends
	if len(sent) != beforeFinal {
		t.Errorf("Final() should not call sendFn: had %d sends before, now %d", beforeFinal, len(sent))
	}
}

func TestProgressReporterEmptyFinal(t *testing.T) {
	r := NewProgressReporter(time.Hour, nil)
	final := r.Final()
	if final != "" {
		t.Errorf("empty reporter Final() should be empty, got %q", final)
	}
}

func TestProgressReporterDefaultInterval(t *testing.T) {
	r := NewProgressReporter(0, nil)
	if r.interval != 3*time.Second {
		t.Errorf("default interval = %v, want 3s", r.interval)
	}
}

func TestProgressReporterNoSendFn(t *testing.T) {
	r := NewProgressReporter(1*time.Millisecond, nil)
	r.Write("test")
	time.Sleep(2 * time.Millisecond)
	r.Write("more")
	// Should not panic even without sendFn
	r.Final()
}
