package task

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
