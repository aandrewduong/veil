package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goquery"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

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
	Client        tls_client.HttpClient
	Session       Session
	WebhookURL    string
	HomepageURL   string
	SSOManagerURL string
	CRNs          []string
}

type TaskManager struct {
	Tasks map[string]*Task
	mutex sync.Mutex
}

func (tm *TaskManager) GetTaskStatus(id string) string {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if task, exists := tm.Tasks[id]; exists {
		return task.Status
	}
	return "Task not found"
}

func (tm *TaskManager) AddTask(task *Task) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.Tasks[task.ID] = task
}

func (tm *TaskManager) DeleteTask(id string) bool {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if _, exists := tm.Tasks[id]; exists {
		delete(tm.Tasks, id)
		return true
	}
	return false
}

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

			task.Username = "20482280"
			task.Password = "Poke20031!"
			if task.Mode == "Watch" {
				task.Watch()
			} else if task.Mode == "Signup" {
				task.Signup()
			}

			// Lock the mutex only when updating the status
			tm.mutex.Lock()
			//task.Status = "Completed"
			tm.mutex.Unlock()
		}()
		return true
	}
	return false
}

func (tm *TaskManager) GetAllTasks() []*Task {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tasks := make([]*Task, 0, len(tm.Tasks))
	for _, task := range tm.Tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (t *Task) InitClient() {
	jar := tls_client.NewCookieJar()
	client_options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_117),
		tls_client.WithCookieJar(jar),
	}
	t.Client, _ = tls_client.NewHttpClient(tls_client.NewLogger(), client_options...)
}

func (t *Task) MakeReq(method string, url string, headers [][2]string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
	}
	for _, header := range headers {
		req.Header.Add(header[0], header[1])
	}
	return req
}

func (t *Task) DoReq(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	resp, err := t.Client.Do(req)
	return resp, err
}

func discardResp(resp *http.Response) {
	if resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		defer resp.Body.Close()
	}
}

func readBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	return body, err
}

func getSelectorAttr(document *goquery.Document, selector string, attr string) string {
	value := ""
	document.Find(selector).Each(func(index int, element *goquery.Selection) {
		_value, exists := element.Attr(attr)
		if exists {
			value = _value
		}
	})
	return value
}

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

func formatDuration(time time.Duration) string {
	totalSeconds := int64(time.Seconds())

	days := totalSeconds / (60 * 60 * 24)
	hours := (totalSeconds % (60 * 60 * 24)) / (60 * 60)
	minutes := (totalSeconds % (60 * 60)) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
}

func extractModel(jsonData []byte) (map[string]interface{}, error) {
	var data AddCourse
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}
	return data.Model, nil
}
