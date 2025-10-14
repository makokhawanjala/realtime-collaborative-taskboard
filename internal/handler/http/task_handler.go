// internal/handler/http/task_handler.go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/makokhawanjala/realtime-taskboard/internal/domain"
	"github.com/makokhawanjala/realtime-taskboard/internal/service"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask handles POST /api/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req domain.CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskService.CreateTask(c.Request.Context(), &req)
	if err != nil {
		if err == domain.ErrInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTask handles GET /api/tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetTasksByBoard handles GET /api/boards/:board_id/tasks
func (h *TaskHandler) GetTasksByBoard(c *gin.Context) {
	boardIDParam := c.Param("board_id")
	boardID, err := uuid.Parse(boardIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	// Check if status filter is provided
	statusParam := c.Query("status")
	if statusParam != "" {
		status := domain.TaskStatus(statusParam)
		tasks, err := h.taskService.GetTasksByStatus(c.Request.Context(), boardID, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
			return
		}
		c.JSON(http.StatusOK, tasks)
		return
	}

	// Get all tasks for the board
	tasks, err := h.taskService.GetTasksByBoard(c.Request.Context(), boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetAllTasks handles GET /api/tasks
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.taskService.GetAllTasks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// UpdateTask handles PUT /api/tasks/:id
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req domain.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskService.UpdateTask(c.Request.Context(), id, &req)
	if err != nil {
		if err == domain.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if err == domain.ErrInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTaskPosition handles PATCH /api/tasks/:id/position
func (h *TaskHandler) UpdateTaskPosition(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		Position int `json:"position" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.UpdateTaskPosition(c.Request.Context(), id, req.Position); err != nil {
		if err == domain.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task position updated successfully"})
}

// DeleteTask handles DELETE /api/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.taskService.DeleteTask(c.Request.Context(), id); err != nil {
		if err == domain.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
