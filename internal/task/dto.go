package task

import (
	"github.com/google/uuid"
	util "github.com/saulo-duarte/chronos-lambda/internal/utils"
)

type TaskStats struct {
	Total      int `json:"total"`
	Todo       int `json:"todo"`
	InProgress int `json:"in_progress"`
	Done       int `json:"done"`
	Overdue    int `json:"overdue"`
}

type TaskTypeStats struct {
	Event   int `json:"event"`
	Study   int `json:"study"`
	Project int `json:"project"`
}

type DashboardStatsResponse struct {
	Stats     TaskStats     `json:"stats"`
	Type      TaskTypeStats `json:"type"`
	Month     []*Task       `json:"month"`
	LastTasks []*Task       `json:"last_tasks"`
}

type TaskUpdateDTO struct {
	ID            uuid.UUID          `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Status        TaskStatus         `json:"status"`
	Priority      TaskPriority       `json:"priority"`
	StartDate     util.LocalDateTime `json:"startDate"`
	DueDate       util.LocalDateTime `json:"dueDate"`
	RemoveDueDate bool               `json:"removeDueDate"`
	DoneAt        util.LocalDateTime `json:"doneAt"`
}
