// internal/service/task_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/makokhawanjala/realtime-taskboard/internal/domain"
	"github.com/makokhawanjala/realtime-taskboard/internal/repository/postgres"
	"github.com/makokhawanjala/realtime-taskboard/internal/repository/redis"
)

// TaskService handles task business logic
type TaskService struct {
	taskRepo     *postgres.TaskRepository
	pubsubClient *redis.PubSubClient
}

// NewTaskService creates a new task service
func NewTaskService(taskRepo *postgres.TaskRepository, pubsubClient *redis.PubSubClient) *TaskService {
	return &TaskService{
		taskRepo:     taskRepo,
		pubsubClient: pubsubClient,
	}
}

// CreateTask creates a new task and publishes an event
func (s *TaskService) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Create task entity
	now := time.Now()
	task := &domain.Task{
		ID:          uuid.New(),
		BoardID:     req.BoardID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		AssignedTo:  req.AssignedTo,
		DueDate:     req.DueDate,
		Position:    req.Position,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Save to database
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Publish task created event
	event := &domain.WebSocketEvent{
		Type:      "task_created",
		BoardID:   task.BoardID,
		Task:      task,
		Timestamp: time.Now(),
	}

	if err := s.pubsubClient.PublishTaskEvent(ctx, event); err != nil {
		// Log error but don't fail the request
		fmt.Printf("⚠️ Failed to publish task_created event: %v\n", err)
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (s *TaskService) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetTasksByBoard retrieves all tasks for a board
func (s *TaskService) GetTasksByBoard(ctx context.Context, boardID uuid.UUID) ([]domain.Task, error) {
	tasks, err := s.taskRepo.GetByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTasksByStatus retrieves tasks by status for a board
func (s *TaskService) GetTasksByStatus(ctx context.Context, boardID uuid.UUID, status domain.TaskStatus) ([]domain.Task, error) {
	tasks, err := s.taskRepo.GetByStatus(ctx, boardID, status)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateTask updates a task and publishes an event
func (s *TaskService) UpdateTask(ctx context.Context, id uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get existing task
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.AssignedTo != nil {
		task.AssignedTo = req.AssignedTo
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.Position != nil {
		task.Position = *req.Position
	}

	// Save to database
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Refresh task from database to get updated timestamp
	task, err = s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Publish task updated event
	event := &domain.WebSocketEvent{
		Type:      "task_updated",
		BoardID:   task.BoardID,
		Task:      task,
		Timestamp: time.Now(),
	}

	if err := s.pubsubClient.PublishTaskEvent(ctx, event); err != nil {
		fmt.Printf("⚠️ Failed to publish task_updated event: %v\n", err)
	}

	return task, nil
}

// DeleteTask deletes a task and publishes an event
func (s *TaskService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	// Get task first to retrieve board ID for the event
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := s.taskRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Publish task deleted event
	event := &domain.WebSocketEvent{
		Type:      "task_deleted",
		BoardID:   task.BoardID,
		TaskID:    &id,
		Timestamp: time.Now(),
	}

	if err := s.pubsubClient.PublishTaskEvent(ctx, event); err != nil {
		fmt.Printf("⚠️ Failed to publish task_deleted event: %v\n", err)
	}

	return nil
}

// UpdateTaskPosition updates the position of a task
func (s *TaskService) UpdateTaskPosition(ctx context.Context, id uuid.UUID, position int) error {
	// Get task first to retrieve board ID for the event
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update position
	if err := s.taskRepo.UpdatePosition(ctx, id, position); err != nil {
		return fmt.Errorf("failed to update task position: %w", err)
	}

	// Get updated task
	task, err = s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Publish task updated event
	event := &domain.WebSocketEvent{
		Type:      "task_updated",
		BoardID:   task.BoardID,
		Task:      task,
		Timestamp: time.Now(),
	}

	if err := s.pubsubClient.PublishTaskEvent(ctx, event); err != nil {
		fmt.Printf("⚠️ Failed to publish task position update event: %v\n", err)
	}

	return nil
}

// GetAllTasks retrieves all tasks (admin/debug)
func (s *TaskService) GetAllTasks(ctx context.Context) ([]domain.Task, error) {
	tasks, err := s.taskRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
