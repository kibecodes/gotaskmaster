package tasks

import (
	"context"
	"sync"
	"time"
)

type Task struct {
	ID         string        `json:"id"`
	Status     string        `json:"status"`
	RetryCount int           `json:"retry_count"`
	CancelChan chan struct{} `json:"-"`
}

type TaskManager struct {
	tasks      map[string]*Task
	mu         sync.Mutex
	taskQueue  chan Task
	wg         sync.WaitGroup
	shutdown   chan struct{}
	maxRetries int
	retryDelay time.Duration
}

func NewTaskManager(workerCount, maxRetries int, retryDelay time.Duration) *TaskManager {
	tm := &TaskManager{
		tasks:      make(map[string]*Task),
		taskQueue:  make(chan Task, 100),
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}

	for i := 0; i < workerCount; i++ {
		go tm.worker()
	}

	return tm
}

func (tm *TaskManager) AddTask(id string) {
	tm.mu.Lock()
	task := &Task{
		ID:         id,
		Status:     "queued",
		RetryCount: 0,
		CancelChan: make(chan struct{}),
	}
	tm.tasks[id] = task
	tm.mu.Unlock()
	tm.taskQueue <- *task
}

func (tm *TaskManager) UpdateTaskStatus(id string, status string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if task, exists := tm.tasks[id]; exists {
		task.Status = status
	}
}

func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	task, exists := tm.tasks[id]
	return task, exists
}

func (tm *TaskManager) CancelTask(id string) bool {
	tm.mu.Lock()
	task, exists := tm.tasks[id]
	tm.mu.Unlock()

	if !exists || task.Status == "Completed" {
		return false
	}

	close(task.CancelChan)
	task.Status = "Cancelled"
	return true
}

func (tm *TaskManager) ListTasks() map[string]*Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.tasks
}

func (tm *TaskManager) worker() {
	for {
		select {
		case task := <-tm.taskQueue:
			tm.wg.Add(1)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			tm.ProcessTask(ctx, &task)
			cancel()
			tm.wg.Done()

		case <-tm.shutdown:
			return
		}
	}
}

func (tm *TaskManager) ProcessTask(ctx context.Context, task *Task) {
	for task.RetryCount <= tm.maxRetries {
		select {
		case <-task.CancelChan:
			task.Status = "Cancelled"
			return
		default:
			time.Sleep(5 * time.Second)
			tm.mu.Lock()
			if ctx.Err() == nil {
				task.Status = "Completed"
				tm.mu.Unlock()
				return
			} else {
				task.RetryCount++
				task.Status = "Failed, Retrying..."
				tm.mu.Unlock()
				time.Sleep(tm.retryDelay)
			}
		}
	}
}

func (tm *TaskManager) Stop() {
	close(tm.shutdown)
	tm.wg.Wait()
}

func (tm *TaskManager) RetryTask(id string) bool {
	tm.mu.Lock()
	task, exists := tm.tasks[id]
	if !exists || task.Status != "Failed" || task.RetryCount <= tm.maxRetries {
		tm.mu.Unlock()
		return false
	}

	task.RetryCount++
	task.Status = "Retrying.."
	tm.mu.Unlock()

	tm.taskQueue <- *task
	return true
}
