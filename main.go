package main

import (
	"gotaskmaster/internal/handlers"
	"gotaskmaster/internal/tasks"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	tm := tasks.NewTaskManager(5, 3, 2*time.Second)
	h := handlers.NewHandler(tm)

	http.HandleFunc("/submit", h.SubmitHandler)
	http.HandleFunc("/status", h.StatusHandler)
	http.HandleFunc("/cancel", h.CancelHandler)
	http.HandleFunc("/tasks", h.ListTasksHandler)
	http.HandleFunc("/retry", h.RetryHandler)

	log.Println("Server is running at port 8080...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down server gracefully...")
	tm.Stop()
	log.Println("Server stopped.")
}
