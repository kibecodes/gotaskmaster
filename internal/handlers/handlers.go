package handlers

import (
	"encoding/json"
	"gotaskmaster/internal/tasks"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	TaskManager *tasks.TaskManager
}

func NewHandler(tm *tasks.TaskManager) *Handler {
	return &Handler{TaskManager: tm}
}

func (h *Handler) SubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post method is allowed", http.StatusMethodNotAllowed)
		return
	}

	id := uuid.New().String()
	h.TaskManager.AddTask(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"task_id": id})
}

func (h *Handler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	task, exists := h.TaskManager.GetTask(id)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) CancelHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	success := h.TaskManager.CancelTask(id)
	if !success {
		http.Error(w, "Task not found or already completed", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task cancelled"})
}

func (h *Handler) ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks := h.TaskManager.ListTasks()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) RetryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only Post method is allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	success := h.TaskManager.RetryTask(id)
	if !success {
		http.Error(w, "Task not found or cannot be retried", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task retried"})
}
