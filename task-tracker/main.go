package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TaskList struct {
	Tasks  []Task `json:"tasks"`
	NextID int    `json:"nextId"`
}

const filePath = "tasks.json"

func loadTasks() (*TaskList, error) {
	data, err := os.ReadFile(filePath)

	if os.IsNotExist(err) {
		return &TaskList{Tasks: []Task{}, NextID: 1}, nil
	}

	if err != nil {
		return nil, err
	}

	var taskList TaskList
	if err := json.Unmarshal(data, &taskList); err != nil {
		return nil, err
	}

	return &taskList, nil
}

func saveTasks(taskList *TaskList) error {
	data, err := json.MarshalIndent(taskList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func addTask(taskList *TaskList, description string) Task {
	now := time.Now().UTC()

	task := Task{
		ID:          taskList.NextID,
		Description: description,
		Status:      StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	taskList.Tasks = append(taskList.Tasks, task)
	taskList.NextID++

	return task
}

func listTasks(taskList *TaskList, filterStatus string) {
	for _, task := range taskList.Tasks {
		if filterStatus != "" && string(task.Status) != filterStatus {
			continue
		}

		fmt.Printf("[%d] %s (%s)\n", task.ID, task.Description, task.Status)
	}
}

func updateTask(taskList *TaskList, id int, description string) *Task {
	for i := range taskList.Tasks {
		if taskList.Tasks[i].ID == id {
			taskList.Tasks[i].Description = description
			taskList.Tasks[i].UpdatedAt = time.Now().UTC()
			return &taskList.Tasks[i]
		}
	}
	return nil
}

func deleteTask(taskList *TaskList, id int) bool {
	for i, task := range taskList.Tasks {
		if task.ID == id {
			taskList.Tasks = append(taskList.Tasks[:i], taskList.Tasks[i+1:]...)
			return true
		}
	}
	return false
}

func markTask(taskList *TaskList, id int, status Status) *Task {
	for i := range taskList.Tasks {
		if taskList.Tasks[i].ID == id {
			taskList.Tasks[i].Status = status
			taskList.Tasks[i].UpdatedAt = time.Now().UTC()
			return &taskList.Tasks[i]
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: clcli <command> [arguments]")
		return
	}

	command := os.Args[1]

	taskList, err := loadTasks()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	switch command {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: task-cli add <description>")
			return
		}
		description := os.Args[2]
		task := addTask(taskList, description)
		if err := saveTasks(taskList); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Task added successfully (ID: %d)\n", task.ID)
	case "list":
		filterStatus := ""
		if len(os.Args) >= 3 {
			filterStatus = os.Args[2]
		}
		listTasks(taskList, filterStatus)
	case "update":
		if len(os.Args) < 4 {
			fmt.Println("Usage: task-cli update <id> <new description>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID")
			return
		}
		description := os.Args[3]
		task := updateTask(taskList, id, description)
		if task == nil {
			fmt.Printf("Task not found (ID: %d)\n", id)
			return
		}
		if err := saveTasks(taskList); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Task updated successfully (ID: %d)\n", task.ID)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Usage: task-cli delete <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID")
			return
		}
		if !deleteTask(taskList, id) {
			fmt.Printf("Task not found (ID: %d)\n", id)
			return
		}
		if err := saveTasks(taskList); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Task deleted successfully (ID: %d)\n", id)
	case "mark-in-progress":
		if len(os.Args) < 3 {
			fmt.Println("Usage: task-cli mark-in-progress <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID")
			return
		}
		task := markTask(taskList, id, StatusInProgress)
		if task == nil {
			fmt.Printf("Task not found (ID: %d)\n", id)
			return
		}
		if err := saveTasks(taskList); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Task marked as in-progress successfully (ID:", task.ID, ")")
	case "mark-done":
		if len(os.Args) < 3 {
			fmt.Println("Usage: task-cli mark-done <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID")
			return
		}
		task := markTask(taskList, id, StatusDone)
		if task == nil {
			fmt.Printf("Task not found (ID: %d)\n", id)
			return
		}
		if err := saveTasks(taskList); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Task marked as done successfully (ID:", task.ID, ")")
	default:
		fmt.Println("Unknown command:", command)
	}
}
