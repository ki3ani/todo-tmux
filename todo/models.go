package main

import "time"

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Todo struct {
	ID        int64     `json:"id"`
	Task      string    `json:"task"`
	Done      bool      `json:"done"`
	Priority  Priority  `json:"priority"`
	Category  string    `json:"category"`
	DueDate   string    `json:"due_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoFilter struct {
	Status   string // all, done, pending
	Priority string // all, low, medium, high
	Category string
	Search   string
}
