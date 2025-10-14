// internal/domain/task.go
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Common domain errors
var (
	ErrTaskNotFound   = errors.New("task not found")
	ErrBoardNotFound  = errors.New("board not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInternalError  = errors.New("internal server error")
	ErrDuplicateEntry = errors.New("duplicate entry")
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

// TaskPriority represents the priority of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// Task represents a task in the system
type Task struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	BoardID     uuid.UUID    `json:"board_id" db:"board_id"`
	Title       string       `json:"title" db:"title"`
	Description string       `json:"description" db:"description"`
	Status      TaskStatus   `json:"status" db:"status"`
	Priority    TaskPriority `json:"priority" db:"priority"`
	AssignedTo  *uuid.UUID   `json:"assigned_to,omitempty" db:"assigned_to"`
	CreatedBy   *uuid.UUID   `json:"created_by,omitempty" db:"created_by"`
	DueDate     *time.Time   `json:"due_date,omitempty" db:"due_date"`
	Position    int          `json:"position" db:"position"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// CreateTaskRequest represents the payload for creating a new task
type CreateTaskRequest struct {
	BoardID     uuid.UUID    `json:"board_id" binding:"required"`
	Title       string       `json:"title" binding:"required,min=1,max=255"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	AssignedTo  *uuid.UUID   `json:"assigned_to,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	Position    int          `json:"position"`
}

// UpdateTaskRequest represents the payload for updating a task
type UpdateTaskRequest struct {
	Title       *string       `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string       `json:"description,omitempty"`
	Status      *TaskStatus   `json:"status,omitempty"`
	Priority    *TaskPriority `json:"priority,omitempty"`
	AssignedTo  *uuid.UUID    `json:"assigned_to,omitempty"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
	Position    *int          `json:"position,omitempty"`
}

// Board represents a task board
type Board struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// WebSocketEvent represents an event to be broadcast via WebSocket
type WebSocketEvent struct {
	Type      string     `json:"type"` // "task_created", "task_updated", "task_deleted"
	BoardID   uuid.UUID  `json:"board_id"`
	Task      *Task      `json:"task,omitempty"`
	TaskID    *uuid.UUID `json:"task_id,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// Validate validates the CreateTaskRequest
func (r *CreateTaskRequest) Validate() error {
	if r.Title == "" {
		return ErrInvalidInput
	}
	if r.BoardID == uuid.Nil {
		return ErrInvalidInput
	}

	// Set defaults if not provided
	if r.Status == "" {
		r.Status = TaskStatusTodo
	}
	if r.Priority == "" {
		r.Priority = TaskPriorityMedium
	}

	// Validate status
	if !isValidStatus(r.Status) {
		return ErrInvalidInput
	}

	// Validate priority
	if !isValidPriority(r.Priority) {
		return ErrInvalidInput
	}

	return nil
}

// Validate validates the UpdateTaskRequest
func (r *UpdateTaskRequest) Validate() error {
	// At least one field must be provided
	if r.Title == nil && r.Description == nil && r.Status == nil &&
		r.Priority == nil && r.AssignedTo == nil && r.DueDate == nil && r.Position == nil {
		return ErrInvalidInput
	}

	// Validate status if provided
	if r.Status != nil && !isValidStatus(*r.Status) {
		return ErrInvalidInput
	}

	// Validate priority if provided
	if r.Priority != nil && !isValidPriority(*r.Priority) {
		return ErrInvalidInput
	}

	return nil
}

// isValidStatus checks if the given status is valid
func isValidStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone:
		return true
	default:
		return false
	}
}

// isValidPriority checks if the given priority is valid
func isValidPriority(priority TaskPriority) bool {
	switch priority {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	default:
		return false
	}
}
