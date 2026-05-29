package runtime

import (
	"fmt"
	"sync"
)

// TaskStatus represents the status of a task in the queue.
type TaskStatus string

const (
	TaskQueued     TaskStatus = "queued"
	TaskRunning    TaskStatus = "running"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
	TaskCancelled  TaskStatus = "cancelled"
)

// Task represents a task in the broker.
type Task struct {
	ID         string
	SessionKey string
	Prompt     string
	Status     TaskStatus
}

// Broker manages task queuing and session state.
type Broker struct {
	mu       sync.Mutex
	running  map[string]*Task // sessionKey -> running task
	queue    map[string][]*Task // sessionKey -> queued tasks
}

// NewBroker creates a new session broker.
func NewBroker() *Broker {
	return &Broker{
		running: make(map[string]*Task),
		queue:   make(map[string][]*Task),
	}
}

// EnqueueResult represents the result of enqueueing a task.
type EnqueueResult struct {
	Queued   bool
	Position int
	Message  string
}

// Enqueue adds a task to the queue. If the session is idle, the task starts immediately.
// If the session is busy, the task is queued and the user is informed.
func (b *Broker) Enqueue(sessionKey, taskID, prompt string) EnqueueResult {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if session is idle
	if _, running := b.running[sessionKey]; !running {
		// Start immediately
		b.running[sessionKey] = &Task{
			ID:         taskID,
			SessionKey: sessionKey,
			Prompt:     prompt,
			Status:     TaskRunning,
		}
		return EnqueueResult{
			Queued:  false,
			Message: "立即执行",
		}
	}

	// Session is busy → queue the task
	position := len(b.queue[sessionKey]) + 1
	b.queue[sessionKey] = append(b.queue[sessionKey], &Task{
		ID:         taskID,
		SessionKey: sessionKey,
		Prompt:     prompt,
		Status:     TaskQueued,
	})

	return EnqueueResult{
		Queued:   true,
		Position: position,
		Message:  fmt.Sprintf("当前有任务在执行，已排队（位置 %d）。回复\"停止\"可取消当前任务。", position),
	}
}

// Complete marks the running task as completed and starts the next queued task.
func (b *Broker) Complete(sessionKey string) *Task {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Mark current as completed
	if t, ok := b.running[sessionKey]; ok {
		t.Status = TaskCompleted
		delete(b.running, sessionKey)
	}

	// Start next queued task
	if len(b.queue[sessionKey]) > 0 {
		next := b.queue[sessionKey][0]
		b.queue[sessionKey] = b.queue[sessionKey][1:]
		next.Status = TaskRunning
		b.running[sessionKey] = next
		return next
	}

	return nil
}

// Fail marks the running task as failed and starts the next queued task.
func (b *Broker) Fail(sessionKey string) *Task {
	b.mu.Lock()
	defer b.mu.Unlock()

	if t, ok := b.running[sessionKey]; ok {
		t.Status = TaskFailed
		delete(b.running, sessionKey)
	}

	// Start next queued task
	if len(b.queue[sessionKey]) > 0 {
		next := b.queue[sessionKey][0]
		b.queue[sessionKey] = b.queue[sessionKey][1:]
		next.Status = TaskRunning
		b.running[sessionKey] = next
		return next
	}

	return nil
}

// Cancel cancels the running task and clears the queue.
func (b *Broker) Cancel(sessionKey string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if t, ok := b.running[sessionKey]; ok {
		t.Status = TaskCancelled
		delete(b.running, sessionKey)
	}

	// Clear queue
	for _, t := range b.queue[sessionKey] {
		t.Status = TaskCancelled
	}
	delete(b.queue, sessionKey)
}

// IsBusy checks if a session has a running task.
func (b *Broker) IsBusy(sessionKey string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	_, running := b.running[sessionKey]
	return running
}

// GetRunning returns the running task for a session, or nil.
func (b *Broker) GetRunning(sessionKey string) *Task {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.running[sessionKey]
}

// GetQueueLength returns the number of queued tasks for a session.
func (b *Broker) GetQueueLength(sessionKey string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return len(b.queue[sessionKey])
}
