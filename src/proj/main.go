package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"proj/tasks"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {

	taskManager := &tasks.TaskManager{
		Tasks: make(map[string]*tasks.Task),
	}

	jar := tls_client.NewCookieJar()
	clientOptions := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_117),
		tls_client.WithCookieJar(jar),
	}
	client, _ := tls_client.NewHttpClient(tls_client.NewLogger(), clientOptions...)
	fmt.Println(client)

	http.HandleFunc("/status", func(writer http.ResponseWriter, request *http.Request) {
		response := map[string]string{"status": "Connected"}
		json.NewEncoder(writer).Encode(response)
	})

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

	http.HandleFunc("/tasks/create", func(writer http.ResponseWriter, request *http.Request) {
		var task tasks.Task
		if err := json.NewDecoder(request.Body).Decode(&task); err != nil {
			http.Error(writer, "Invalid task data", http.StatusBadRequest)
			return
		}

		taskManager.AddTask(&task)
		response := map[string]string{"message": "Task created"}
		fmt.Println(task)
		json.NewEncoder(writer).Encode(response)
	})

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

	http.HandleFunc("/tasks/all", func(writer http.ResponseWriter, request *http.Request) {
		tasks := taskManager.GetAllTasks()
		json.NewEncoder(writer).Encode(tasks)
	})
	http.ListenAndServe(":1942", nil)
}
