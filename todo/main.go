package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Todo struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	Done bool   `json:"done"`
}

const dataFile = "todos.json"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	todos := loadTodos()
	cmd := os.Args[1]

	switch cmd {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo add <task>")
			return
		}
		task := strings.Join(os.Args[2:], " ")
		todos = addTodo(todos, task)
		saveTodos(todos)

	case "list", "ls":
		listTodos(todos)

	case "done":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo done <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID")
			return
		}
		todos = markDone(todos, id)
		saveTodos(todos)

	case "rm", "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo rm <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID")
			return
		}
		todos = removeTodo(todos, id)
		saveTodos(todos)

	case "clear":
		todos = []Todo{}
		saveTodos(todos)
		fmt.Println("All todos cleared")

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`todo - simple task manager

Usage:
  todo add <task>   Add a new todo
  todo list         List all todos
  todo done <id>    Mark todo as done
  todo rm <id>      Remove a todo
  todo clear        Remove all todos`)
}

func loadTodos() []Todo {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return []Todo{}
	}
	var todos []Todo
	json.Unmarshal(data, &todos)
	return todos
}

func saveTodos(todos []Todo) {
	data, _ := json.MarshalIndent(todos, "", "  ")
	os.WriteFile(dataFile, data, 0644)
}

func addTodo(todos []Todo, task string) []Todo {
	id := 1
	if len(todos) > 0 {
		id = todos[len(todos)-1].ID + 1
	}
	todo := Todo{ID: id, Task: task, Done: false}
	todos = append(todos, todo)
	fmt.Printf("Added: [%d] %s\n", id, task)
	return todos
}

func listTodos(todos []Todo) {
	if len(todos) == 0 {
		fmt.Println("No todos yet. Add one with: todo add <task>")
		return
	}
	fmt.Println()
	for _, t := range todos {
		status := " "
		if t.Done {
			status = "x"
		}
		fmt.Printf("  [%s] %d. %s\n", status, t.ID, t.Task)
	}
	fmt.Println()
}

func markDone(todos []Todo, id int) []Todo {
	for i, t := range todos {
		if t.ID == id {
			todos[i].Done = true
			fmt.Printf("Done: [%d] %s\n", id, t.Task)
			return todos
		}
	}
	fmt.Println("Todo not found")
	return todos
}

func removeTodo(todos []Todo, id int) []Todo {
	for i, t := range todos {
		if t.ID == id {
			fmt.Printf("Removed: [%d] %s\n", id, t.Task)
			return append(todos[:i], todos[i+1:]...)
		}
	}
	fmt.Println("Todo not found")
	return todos
}
