package runtime

import (
	"testing"
)

func TestBrokerEnqueueIdle(t *testing.T) {
	b := NewBroker()

	result := b.Enqueue("session1", "task1", "fix bug")

	if result.Queued {
		t.Error("should not be queued when idle")
	}
	if !b.IsBusy("session1") {
		t.Error("should be busy after enqueue")
	}
}

func TestBrokerEnqueueBusy(t *testing.T) {
	b := NewBroker()

	b.Enqueue("session1", "task1", "first task")
	result := b.Enqueue("session1", "task2", "second task")

	if !result.Queued {
		t.Error("should be queued when busy")
	}
	if result.Position != 1 {
		t.Errorf("position = %d, want 1", result.Position)
	}
}

func TestBrokerComplete(t *testing.T) {
	b := NewBroker()

	b.Enqueue("session1", "task1", "first")
	b.Enqueue("session1", "task2", "second")

	next := b.Complete("session1")

	if next == nil {
		t.Fatal("expected next task to start")
	}
	if next.ID != "task2" {
		t.Errorf("next task ID = %q, want task2", next.ID)
	}
	if !b.IsBusy("session1") {
		t.Error("should still be busy with next task")
	}
}

func TestBrokerCompleteNoQueue(t *testing.T) {
	b := NewBroker()

	b.Enqueue("session1", "task1", "only task")
	next := b.Complete("session1")

	if next != nil {
		t.Error("should return nil when no more tasks")
	}
	if b.IsBusy("session1") {
		t.Error("should not be busy when all tasks done")
	}
}

func TestBrokerFail(t *testing.T) {
	b := NewBroker()

	b.Enqueue("session1", "task1", "first")
	b.Enqueue("session1", "task2", "second")

	next := b.Fail("session1")

	if next == nil {
		t.Fatal("expected next task to start after fail")
	}
	if next.ID != "task2" {
		t.Errorf("next task ID = %q, want task2", next.ID)
	}
}

func TestBrokerCancel(t *testing.T) {
	b := NewBroker()

	b.Enqueue("session1", "task1", "first")
	b.Enqueue("session1", "task2", "second")

	b.Cancel("session1")

	if b.IsBusy("session1") {
		t.Error("should not be busy after cancel")
	}
	if b.GetQueueLength("session1") != 0 {
		t.Error("queue should be empty after cancel")
	}
}

func TestBrokerGetRunning(t *testing.T) {
	b := NewBroker()

	if b.GetRunning("session1") != nil {
		t.Error("should return nil when no running task")
	}

	b.Enqueue("session1", "task1", "fix bug")

	running := b.GetRunning("session1")
	if running == nil {
		t.Fatal("expected running task")
	}
	if running.ID != "task1" {
		t.Errorf("running task ID = %q, want task1", running.ID)
	}
}

func TestBrokerMultipleSessions(t *testing.T) {
	b := NewBroker()

	b.Enqueue("session1", "task1", "first")
	b.Enqueue("session2", "task2", "second")

	if !b.IsBusy("session1") {
		t.Error("session1 should be busy")
	}
	if !b.IsBusy("session2") {
		t.Error("session2 should be busy")
	}

	b.Complete("session1")
	if b.IsBusy("session1") {
		t.Error("session1 should not be busy after complete")
	}
	if !b.IsBusy("session2") {
		t.Error("session2 should still be busy")
	}
}
