package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"goquery"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type Task struct {
	ID            string `json:"id"`
	Mode          string `json:"mode"`
	Term          string `json:"term"`
	Crns          string `json:"crns"`
	Status        string `json:"status"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	WebhookURL    string `json:"webhook_url"`
	Client        tls_client.HttpClient
	Session       Session
	HomepageURL   string
	SSOManagerURL string
	CRNs          []string
}

type SanitizedTask struct {
	ID            string `json:"id"`
	Mode          string `json:"mode"`
	Term          string `json:"term"`
	Crns          string `json:"crns"`
	Status        string `json:"status"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	WebhookURL    string `json:"webhook_url"`
	HomepageURL   string `json:"homepage_url"`
	SSOManagerURL string `json:"sso_manager_url"`
}

type TaskManager struct {
	Tasks map[string]*Task
	mutex sync.Mutex
}

// GetTaskStatus returns the status of a task by its ID.
func (tm *TaskManager) GetTaskStatus(id string) string {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if task, exists := tm.Tasks[id]; exists {
		return task.Status
	}
	return "Task not found"
}

// AddTask adds a new task to the TaskManager.
func (tm *TaskManager) AddTask(task *Task) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.Tasks[task.ID] = task
}

// DeleteTask removes a task from the TaskManager by its ID.
func (tm *TaskManager) DeleteTask(id string) bool {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if _, exists := tm.Tasks[id]; exists {
		delete(tm.Tasks, id)
		return true
	}
	return false
}

// RunTask runs a task by its ID.
func (tm *TaskManager) RunTask(id string) bool {
	tm.mutex.Lock()
	task, exists := tm.Tasks[id]
	if exists {
		task.Status = "Running"
	}
	tm.mutex.Unlock()

	if exists {
		go func() {
			// Perform the task's work without holding the mutex
			task.InitClient()
			task.CRNs = strings.Split(task.Crns, ",")
			if task.Mode == "Watch" {
				task.Watch()
			} else if task.Mode == "Signup" {
				task.Signup()
			}

			// Lock the mutex only when updating the status
			tm.mutex.Lock()
			defer tm.mutex.Unlock()

			/*
				if task.Mode != "Watch" {
					task.Status = "Completed"
				}
			*/
		}()
		return true
	}
	return false
}

// SanitizeTask creates a sanitized version of the task.
func SanitizeTask(task *Task) *SanitizedTask {
	return &SanitizedTask{
		ID:            task.ID,
		Mode:          task.Mode,
		Term:          task.Term,
		Crns:          task.Crns,
		Status:        task.Status,
		Username:      task.Username,
		Password:      task.Password,
		WebhookURL:    task.WebhookURL,
		HomepageURL:   task.HomepageURL,
		SSOManagerURL: task.SSOManagerURL,
	}
}

// GetAllSanitizedTasks returns all tasks in a sanitized format.
func (tm *TaskManager) GetAllSanitizedTasks() []*SanitizedTask {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tasks := make([]*SanitizedTask, 0, len(tm.Tasks))
	for _, task := range tm.Tasks {
		tasks = append(tasks, SanitizeTask(task))
	}
	return tasks
}

// InitClient initializes the HTTP client for the task.
func (t *Task) InitClient() {
	jar := tls_client.NewCookieJar()
	clientOptions := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_117),
		tls_client.WithCookieJar(jar),
	}
	t.Client, _ = tls_client.NewHttpClient(tls_client.NewLogger(), clientOptions...)
}

// MakeReq creates a new HTTP request with the given method, URL, headers, and body.
func (t *Task) MakeReq(method, url string, headers [][2]string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
	}
	for _, header := range headers {
		req.Header.Add(header[0], header[1])
	}
	return req
}

// DoReq executes the given HTTP request.
func (t *Task) DoReq(req *http.Request) (*http.Response, error) {
	return t.Client.Do(req)
}

// discardResp discards the response body to free up resources.
func discardResp(resp *http.Response) {
	if resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		defer resp.Body.Close()
	}
}

// readBody reads the response body and returns it as a byte slice.
func readBody(resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// getSelectorAttr gets the attribute value of the first element that matches the selector.
func getSelectorAttr(document *goquery.Document, selector, attr string) string {
	value := ""
	document.Find(selector).Each(func(index int, element *goquery.Selection) {
		if attrValue, exists := element.Attr(attr); exists {
			value = attrValue
		}
	})
	return value
}

// generateRandomString generates a random string of the given length.
func generateRandomString(length int) string {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	result := make([]byte, length)
	for i := range result {
		result[i] = characters[random.Intn(len(characters))]
	}
	return string(result)
}

// formatDuration formats a time.Duration into a human-readable string.
func formatDuration(duration time.Duration) string {
	totalSeconds := int64(duration.Seconds())

	days := totalSeconds / (60 * 60 * 24)
	hours := (totalSeconds % (60 * 60 * 24)) / (60 * 60)
	minutes := (totalSeconds % (60 * 60)) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
}

// extractModel extracts the model from the JSON data.
func extractModel(jsonData []byte) (map[string]interface{}, error) {
	var data AddCourse
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}
	return data.Model, nil
}

// SendNotification sends a notification with the given action and message.
func (t *Task) SendNotification(action string, message string) error {
	payload := WebhookPayload{
		Username: "veil",
		Embeds: []Embed{
			{
				Title:       action,
				Description: message,
				Footer: &Footer{
					Text: "Veil",
				},
				Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			},
		},
	}
	jsonData, _ := json.Marshal(payload)
	headers := [][2]string{
		{"accept", "application/json"},
		{"accept-language", "en-US,en;q=0.9"},
		{"content-type", "application/json"},
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"},
	}
	t.DoReq(t.MakeReq("POST", t.WebhookURL, headers, []byte(string(jsonData))))
	return nil
}
