// internal/repository/postgres/task_repository.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/makokhawanjala/realtime-taskboard/internal/domain"
)

// TaskRepository handles task database operations
type TaskRepository struct {
	db *sqlx.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create creates a new task
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (
			id, board_id, title, description, status, priority,
			assigned_to, created_by, due_date, position, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		task.ID, task.BoardID, task.Title, task.Description,
		task.Status, task.Priority, task.AssignedTo, task.CreatedBy,
		task.DueDate, task.Position, task.CreatedAt, task.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// GetByID retrieves a task by ID
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	var task domain.Task
	query := `
		SELECT id, board_id, title, description, status, priority,
		       assigned_to, created_by, due_date, position, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &task, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// GetByBoardID retrieves all tasks for a specific board
func (r *TaskRepository) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]domain.Task, error) {
	var tasks []domain.Task
	query := `
		SELECT id, board_id, title, description, status, priority,
		       assigned_to, created_by, due_date, position, created_at, updated_at
		FROM tasks
		WHERE board_id = $1
		ORDER BY position ASC, created_at DESC
	`

	err := r.db.SelectContext(ctx, &tasks, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by board: %w", err)
	}

	return tasks, nil
}

// GetByStatus retrieves all tasks with a specific status
func (r *TaskRepository) GetByStatus(ctx context.Context, boardID uuid.UUID, status domain.TaskStatus) ([]domain.Task, error) {
	var tasks []domain.Task
	query := `
		SELECT id, board_id, title, description, status, priority,
		       assigned_to, created_by, due_date, position, created_at, updated_at
		FROM tasks
		WHERE board_id = $1 AND status = $2
		ORDER BY position ASC, created_at DESC
	`

	err := r.db.SelectContext(ctx, &tasks, query, boardID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by status: %w", err)
	}

	return tasks, nil
}

// Update updates an existing task
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4,
		    assigned_to = $5, due_date = $6, position = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(
		ctx, query,
		task.Title, task.Description, task.Status, task.Priority,
		task.AssignedTo, task.DueDate, task.Position, task.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// Delete deletes a task by ID
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// UpdatePosition updates the position of a task
func (r *TaskRepository) UpdatePosition(ctx context.Context, id uuid.UUID, position int) error {
	query := `UPDATE tasks SET position = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, position, id)
	if err != nil {
		return fmt.Errorf("failed to update task position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// GetAll retrieves all tasks (useful for admin/debugging)
func (r *TaskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {
	var tasks []domain.Task
	query := `
		SELECT id, board_id, title, description, status, priority,
		       assigned_to, created_by, due_date, position, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &tasks, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all tasks: %w", err)
	}

	return tasks, nil
}
