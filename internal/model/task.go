package model

import (
	"fmt"
	"time"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in-progress"
	StatusDone       TaskStatus = "done"
	StatusBlocked    TaskStatus = "blocked"
)

func (s TaskStatus) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone, StatusBlocked:
		return true
	}
	return false
}

func ParseStatus(s string) (TaskStatus, error) {
	v := TaskStatus(s)
	if !v.Valid() {
		return "", fmt.Errorf("invalid status %q (want todo|in-progress|done|blocked)", s)
	}
	return v, nil
}

type TaskPriority string

const (
	PriorityLow      TaskPriority = "low"
	PriorityMedium   TaskPriority = "medium"
	PriorityHigh     TaskPriority = "high"
	PriorityCritical TaskPriority = "critical"
)

func (p TaskPriority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	}
	return false
}

func ParsePriority(s string) (TaskPriority, error) {
	v := TaskPriority(s)
	if !v.Valid() {
		return "", fmt.Errorf("invalid priority %q (want low|medium|high|critical)", s)
	}
	return v, nil
}

type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Assignee    string       `json:"assignee,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	DueDate     string       `json:"dueDate,omitempty"`
}

type Board struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Tasks       []Task    `json:"tasks"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
