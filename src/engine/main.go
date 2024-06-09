package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"proj/tasks"
)

func main() {
	// Initialize TaskManager
	taskManager := &tasks.TaskManager{
		Tasks: make(map[string]*tasks.Task),
	}

	// Health check endpoint
	http.HandleFunc("/status", func(writer http.ResponseWriter, request *http.Request) {
		response := map[string]string{"status": "Connected"}
		json.NewEncoder(writer).Encode(response)
	})

	// Get task status by ID
	http.HandleFunc("/tasks/status", func(writer http.ResponseWriter, request *http.Request) {
		id := request.URL.Query().Get("id")
		if id == "" {
			http.Error(writer, "Missing task ID", http.StatusBadRequest)
			return
		}

		status := taskManager.GetTaskStatus(id)
		response := map[string]string{"status": status}
		json.NewEncoder(writer).Encode(response)
	})

	// Create a new task
	http.HandleFunc("/tasks/create", func(writer http.ResponseWriter, request *http.Request) {
		var task tasks.Task
		if err := json.NewDecoder(request.Body).Decode(&task); err != nil {
			http.Error(writer, "Invalid task data", http.StatusBadRequest)
			return
		}

		taskManager.AddTask(&task)
		response := map[string]string{"message": "Task created"}
		json.NewEncoder(writer).Encode(response)
	})

	// Delete a task by ID
	http.HandleFunc("/tasks/delete", func(writer http.ResponseWriter, request *http.Request) {
		id := request.URL.Query().Get("id")
		if id == "" {
			http.Error(writer, "Missing task ID", http.StatusBadRequest)
			return
		}

		if taskManager.DeleteTask(id) {
			response := map[string]string{"message": "Task deleted"}
			json.NewEncoder(writer).Encode(response)
		} else {
			http.Error(writer, "Task not found", http.StatusNotFound)
		}
	})

	// Run a task by ID
	http.HandleFunc("/tasks/run", func(writer http.ResponseWriter, request *http.Request) {
		id := request.URL.Query().Get("id")
		if id == "" {
			http.Error(writer, "Missing task ID", http.StatusBadRequest)
			return
		}

		if taskManager.RunTask(id) {
			response := map[string]string{"message": "Task is running"}
			json.NewEncoder(writer).Encode(response)
		} else {
			http.Error(writer, "Task not found", http.StatusNotFound)
		}
	})

	// Get all sanitized tasks
	http.HandleFunc("/tasks/all", func(writer http.ResponseWriter, request *http.Request) {
		tasks := taskManager.GetAllSanitizedTasks()
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(tasks); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			fmt.Printf("Error encoding tasks: %v\n", err)
		}
	})

	// Start HTTP server
	if err := http.ListenAndServe(":1942", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
