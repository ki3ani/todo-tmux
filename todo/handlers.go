package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func handleAPITodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		filter := TodoFilter{
			Status:   r.URL.Query().Get("status"),
			Priority: r.URL.Query().Get("priority"),
			Category: r.URL.Query().Get("category"),
			Search:   r.URL.Query().Get("search"),
		}
		todos, err := GetTodos(filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if todos == nil {
			todos = []Todo{}
		}
		json.NewEncoder(w).Encode(todos)

	case "POST":
		var input struct {
			Task     string `json:"task"`
			Priority string `json:"priority"`
			Category string `json:"category"`
			DueDate  string `json:"due_date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(input.Task) == "" {
			http.Error(w, "Task is required", http.StatusBadRequest)
			return
		}
		priority := Priority(input.Priority)
		if priority == "" {
			priority = PriorityMedium
		}
		todo, err := CreateTodo(input.Task, priority, input.Category, input.DueDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(todo)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAPITodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from path: /api/todos/123
	path := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		todo, err := GetTodo(id)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(todo)

	case "PUT":
		var input struct {
			Task     string `json:"task"`
			Done     bool   `json:"done"`
			Priority string `json:"priority"`
			Category string `json:"category"`
			DueDate  string `json:"due_date"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		err := UpdateTodo(id, input.Task, input.Done, Priority(input.Priority), input.Category, input.DueDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todo, _ := GetTodo(id)
		json.NewEncoder(w).Encode(todo)

	case "PATCH":
		// Quick toggle done status
		var input struct {
			Done *bool `json:"done"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if input.Done != nil {
			MarkTodoDone(id, *input.Done)
		}
		todo, _ := GetTodo(id)
		json.NewEncoder(w).Encode(todo)

	case "DELETE":
		err := DeleteTodo(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAPICategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	categories, err := GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if categories == nil {
		categories = []string{}
	}
	json.NewEncoder(w).Encode(categories)
}
