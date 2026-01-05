package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if err := InitDB(); err != nil {
		fmt.Println("Error initializing database:", err)
		os.Exit(1)
	}
	defer CloseDB()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "add":
		handleAdd()
	case "list", "ls":
		handleList()
	case "done":
		handleDone()
	case "undone":
		handleUndone()
	case "rm", "remove":
		handleRemove()
	case "clear":
		handleClear()
	case "server":
		runServer()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`todo - simple task manager

Usage:
  todo add <task> [-p high|medium|low] [-c category] [-d due-date]
  todo list [-s done|pending] [-p priority] [-c category]
  todo done <id>      Mark todo as done
  todo undone <id>    Mark todo as not done
  todo rm <id>        Remove a todo
  todo clear          Remove all todos
  todo server         Start web UI on port 8080

Examples:
  todo add "Buy milk" -p high -c shopping
  todo add "Finish report" -d 2024-12-31
  todo list -s pending -p high`)
}

func handleAdd() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: todo add <task> [-p priority] [-c category] [-d due-date]")
		return
	}

	args := os.Args[2:]
	task := ""
	priority := PriorityMedium
	category := ""
	dueDate := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p":
			if i+1 < len(args) {
				priority = Priority(args[i+1])
				i++
			}
		case "-c":
			if i+1 < len(args) {
				category = args[i+1]
				i++
			}
		case "-d":
			if i+1 < len(args) {
				dueDate = args[i+1]
				i++
			}
		default:
			if task == "" {
				task = args[i]
			} else {
				task += " " + args[i]
			}
		}
	}

	if task == "" {
		fmt.Println("Please provide a task")
		return
	}

	todo, err := CreateTodo(task, priority, category, dueDate)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Added: [%d] %s\n", todo.ID, todo.Task)
	if category != "" {
		fmt.Printf("  Category: %s\n", category)
	}
	if dueDate != "" {
		fmt.Printf("  Due: %s\n", dueDate)
	}
}

func handleList() {
	filter := TodoFilter{}

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-s":
			if i+1 < len(os.Args) {
				filter.Status = os.Args[i+1]
				i++
			}
		case "-p":
			if i+1 < len(os.Args) {
				filter.Priority = os.Args[i+1]
				i++
			}
		case "-c":
			if i+1 < len(os.Args) {
				filter.Category = os.Args[i+1]
				i++
			}
		}
	}

	todos, err := GetTodos(filter)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

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
		priorityIcon := ""
		switch t.Priority {
		case PriorityHigh:
			priorityIcon = "!"
		case PriorityLow:
			priorityIcon = "-"
		}
		extra := ""
		if t.Category != "" {
			extra += fmt.Sprintf(" [%s]", t.Category)
		}
		if t.DueDate != "" {
			extra += fmt.Sprintf(" (due: %s)", t.DueDate)
		}
		fmt.Printf("  [%s]%s %d. %s%s\n", status, priorityIcon, t.ID, t.Task, extra)
	}
	fmt.Println()
}

func handleDone() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: todo done <id>")
		return
	}
	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}
	todo, err := GetTodo(id)
	if err != nil {
		fmt.Println("Todo not found")
		return
	}
	MarkTodoDone(id, true)
	fmt.Printf("Done: [%d] %s\n", id, todo.Task)
}

func handleUndone() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: todo undone <id>")
		return
	}
	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}
	todo, err := GetTodo(id)
	if err != nil {
		fmt.Println("Todo not found")
		return
	}
	MarkTodoDone(id, false)
	fmt.Printf("Undone: [%d] %s\n", id, todo.Task)
}

func handleRemove() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: todo rm <id>")
		return
	}
	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}
	todo, err := GetTodo(id)
	if err != nil {
		fmt.Println("Todo not found")
		return
	}
	DeleteTodo(id)
	fmt.Printf("Removed: [%d] %s\n", id, todo.Task)
}

func handleClear() {
	ClearTodos()
	fmt.Println("All todos cleared")
}

// Placeholder - actual server code is in server.go
func runServer() {
	startServer()
}

// Ignore unused import warning for strings
var _ = strings.TrimSpace
