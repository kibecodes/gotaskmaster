package handlers

import (
	"encoding/json"
	"gotaskmaster/internal/tasks"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSubmitHandler(t *testing.T) {
	tm := tasks.NewTaskManager(5, 3, 2*time.Second)
	h := NewHandler(tm)

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	rec := httptest.NewRecorder()

	h.SubmitHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var resBody map[string]string
	json.NewDecoder(resp.Body).Decode(&resBody)

	if _, exists := resBody["task_id"]; !exists {
		t.Fatalf("Expected task_id in response")
	}
}

func TestStatusHandler(t *testing.T) {
	tm := tasks.NewTaskManager(5, 3, 2*time.Second)
	h := NewHandler(tm)

	taskID := "test-task-1"
	h.TaskManager.AddTask(taskID)

	req := httptest.NewRequest(http.MethodGet, "/status?id="+taskID, nil)
	rec := httptest.NewRecorder()

	h.StatusHandler(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var resBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resBody)

	if resBody["id"] != taskID {
		t.Fatalf("Expected task ID %s, got %s", taskID, resBody["id"])
	}
}

func TestRetryHandler(t *testing.T) {
	tm := tasks.NewTaskManager(5, 3, 2*time.Second)
	h := NewHandler(tm)

	taskID := "failed-task-1"
	h.TaskManager.AddTask(taskID)
	h.TaskManager.UpdateTaskStatus(taskID, "Failed")

	mux := http.NewServeMux()
	mux.HandleFunc("/retry", h.RetryHandler)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/retry?id="+taskID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var resBody map[string]string
	json.NewDecoder(resp.Body).Decode(&resBody)

	if resBody["message"] != "Task retried" {
		t.Fatalf("Expected 'Task retried' message, got %s", resBody["message"])
	}
}
