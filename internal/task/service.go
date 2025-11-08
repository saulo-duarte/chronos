package task

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	googlecalendar "github.com/saulo-duarte/chronos-lambda/internal/google_calendar"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	studytopic "github.com/saulo-duarte/chronos-lambda/internal/study_topic"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
	util "github.com/saulo-duarte/chronos-lambda/internal/utils"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrProjectNotFound    = project.ErrProjectNotFound
	ErrStudyTopicNotFound = studytopic.ErrStudyTopicNotFound
	ErrInvalidID          = errors.New("invalid id format")
	ErrProjectRequired    = errors.New("projectId is required for PROJECT tasks")
)

const dashboardTaskLimit = 5

type TaskService interface {
	CreateTask(ctx context.Context, t *Task) (*Task, error)
	FindAllByUser(ctx context.Context) ([]*Task, error)
	FindByID(ctx context.Context, id string) (*Task, error)
	DeleteByID(ctx context.Context, id string) error
	FindAllByProjectID(ctx context.Context, projectID string) ([]*Task, error)
	FindAllByTopicID(ctx context.Context, topicID string) ([]*Task, error)
	UpdateTask(ctx context.Context, dto *TaskUpdateDTO) (*Task, error)
	GetDashboardStats(ctx context.Context) (*DashboardStatsResponse, error)
}

type taskService struct {
	repo            TaskRepository
	projectService  project.ProjectService
	userRepo        user.UserRepository
	studyTopicRepo  studytopic.StudyTopicRepository
	calendarManager googlecalendar.CalendarManager
}

func NewService(
	repo TaskRepository,
	projectService project.ProjectService,
	userRepo user.UserRepository,
	studyTopicRepo studytopic.StudyTopicRepository,
	calendarManager googlecalendar.CalendarManager,
) TaskService {
	return &taskService{
		repo:            repo,
		projectService:  projectService,
		userRepo:        userRepo,
		studyTopicRepo:  studyTopicRepo,
		calendarManager: calendarManager,
	}
}

func (s *taskService) CreateTask(ctx context.Context, t *Task) (*Task, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	t.UserID = userID

	if err := s.validateTaskDependencies(ctx, t); err != nil {
		return nil, err
	}

	if err := s.repo.Create(t); err != nil {
		config.WithContext(ctx).WithError(err).Error("Failed to create task")
		return nil, err
	}

	s.syncWithCalendar(ctx, userID, t)
	config.WithContext(ctx).WithField("task_id", t.ID).Info("Task created successfully")
	return t, nil
}

func (s *taskService) FindAllByUser(ctx context.Context) ([]*Task, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	tasks, err := s.repo.ListByUser(userID)
	if err != nil {
		config.WithContext(ctx).WithError(err).Error("Failed to list tasks")
		return nil, err
	}

	return tasks, nil
}

func (s *taskService) FindByID(ctx context.Context, id string) (*Task, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	taskID, err := s.parseUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.getTaskByID(ctx, taskID, userID)
}

func (s *taskService) DeleteByID(ctx context.Context, id string) error {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return err
	}

	taskID, err := s.parseUUID(ctx, id)
	if err != nil {
		return err
	}

	task, err := s.getTaskByID(ctx, taskID, userID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(taskID, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrTaskNotFound
		}
		config.WithContext(ctx).WithError(err).Error("Failed to delete task")
		return err
	}

	if task.GoogleCalendarEventID != "" {
		s.calendarManager.RemoveTask(ctx, userID, task.GoogleCalendarEventID)
	}

	config.WithContext(ctx).WithField("task_id", id).Info("Task deleted successfully")
	return nil
}

func (s *taskService) FindAllByProjectID(ctx context.Context, projectID string) ([]*Task, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	pid, err := s.parseUUID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if err := s.validateProjectExists(ctx, pid); err != nil {
		return nil, err
	}

	tasks, err := s.repo.ListByProjectAndUser(pid, userID)
	if err != nil {
		config.WithContext(ctx).WithError(err).Error("Failed to list tasks by project")
		return nil, err
	}

	return tasks, nil
}

func (s *taskService) FindAllByTopicID(ctx context.Context, topicID string) ([]*Task, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	tid, err := s.parseUUID(ctx, topicID)
	if err != nil {
		return nil, err
	}

	if err := s.validateTopicExists(ctx, tid); err != nil {
		return nil, err
	}

	tasks, err := s.repo.ListByStudyTopicAndUser(tid, userID)
	if err != nil {
		config.WithContext(ctx).WithError(err).Error("Failed to list tasks by study topic")
		return nil, err
	}

	return tasks, nil
}

func (s *taskService) UpdateTask(ctx context.Context, dto *TaskUpdateDTO) (*Task, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	task, err := s.getTaskByID(ctx, dto.ID, userID)
	if err != nil {
		return nil, err
	}

	needsCalendarSync := s.applyTaskUpdates(task, dto)
	task.UpdatedAt = time.Now()

	if err := s.repo.Update(task); err != nil {
		config.WithContext(ctx).WithError(err).Error("Failed to update task")
		return nil, err
	}

	if needsCalendarSync {
		s.syncWithCalendar(ctx, userID, task)
	}

	config.WithContext(ctx).WithField("task_id", task.ID).Info("Task updated successfully")
	return task, nil
}

func (s *taskService) GetDashboardStats(ctx context.Context) (*DashboardStatsResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	tasks, err := s.repo.ListByUser(userID)
	if err != nil {
		config.WithContext(ctx).WithError(err).Error("Failed to list tasks for dashboard")
		return nil, err
	}

	return s.buildDashboardStats(tasks), nil
}

// ============= Helper Methods =============

func (s *taskService) getUserID(ctx context.Context) (uuid.UUID, error) {
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		config.WithContext(ctx).WithError(err).Warn("Unauthorized access attempt")
		return uuid.Nil, ErrUnauthorized
	}
	return uuid.MustParse(claims.UserID), nil
}

func (s *taskService) parseUUID(ctx context.Context, id string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		config.WithContext(ctx).WithError(err).Warnf("Invalid ID: %s", id)
		return uuid.Nil, ErrInvalidID
	}
	return parsedID, nil
}

func (s *taskService) getTaskByID(ctx context.Context, taskID, userID uuid.UUID) (*Task, error) {
	task, err := s.repo.FindByIdAndUserId(taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrTaskNotFound
		}
		config.WithContext(ctx).WithError(err).Error("Error finding task")
		return nil, err
	}
	return task, nil
}

func (s *taskService) validateProjectExists(ctx context.Context, projectID uuid.UUID) error {
	if _, err := s.projectService.GetProjectByID(ctx, projectID.String()); err != nil {
		config.WithContext(ctx).WithError(err).WithField("project_id", projectID).Error("Project not found")
		return ErrProjectNotFound
	}
	return nil
}

func (s *taskService) validateTopicExists(ctx context.Context, topicID uuid.UUID) error {
	if _, err := s.studyTopicRepo.GetByID(topicID.String()); err != nil {
		config.WithContext(ctx).WithError(err).WithField("study_topic_id", topicID).Error("Study topic not found")
		return ErrStudyTopicNotFound
	}
	return nil
}

func (s *taskService) validateTaskDependencies(ctx context.Context, t *Task) error {
	if t.Type == "PROJECT" && t.ProjectId == nil {
		return ErrProjectRequired
	}

	if t.ProjectId != nil {
		if err := s.validateProjectExists(ctx, *t.ProjectId); err != nil {
			return err
		}
	}

	if t.StudyTopicId != nil {
		if err := s.validateTopicExists(ctx, *t.StudyTopicId); err != nil {
			return err
		}
	}

	return nil
}

func (s *taskService) syncWithCalendar(ctx context.Context, userID uuid.UUID, t *Task) {
	calTask := &googlecalendar.CalendarTask{
		ID:                    t.ID,
		Name:                  t.Name,
		Description:           t.Description,
		StartDate:             util.ToTimePtr(t.StartDate),
		DueDate:               util.ToTimePtr(t.DueDate),
		GoogleCalendarEventID: s.getEventIDPtr(t.GoogleCalendarEventID),
	}

	eventID, err := s.calendarManager.SyncTask(ctx, userID, calTask)
	if err != nil {
		config.WithContext(ctx).WithError(err).Warnf("Calendar sync failed for task %s", t.ID)
		return
	}

	if eventID != t.GoogleCalendarEventID {
		t.GoogleCalendarEventID = eventID
		if err := s.repo.Update(t); err != nil {
			config.WithContext(ctx).WithError(err).Error("Failed to update task with calendar event ID")
		}
	}
}

func (s *taskService) getEventIDPtr(eventID string) *string {
	if eventID == "" {
		return nil
	}
	return &eventID
}

func (s *taskService) applyTaskUpdates(task *Task, dto *TaskUpdateDTO) bool {
	needsSync := false

	if dto.Name != "" && dto.Name != task.Name {
		task.Name = dto.Name
		needsSync = true
	}

	if dto.Description != "" && dto.Description != task.Description {
		task.Description = dto.Description
		needsSync = true
	}

	if dto.Status != "" && dto.Status != task.Status {
		task.Status = dto.Status
	}

	if dto.Priority != "" && dto.Priority != task.Priority {
		task.Priority = dto.Priority
	}

	if s.updateDate(&task.StartDate, &dto.StartDate) {
		needsSync = true
	}

	if dto.RemoveDueDate {
		if task.DueDate != nil {
			task.DueDate = nil
			needsSync = true
		}
	} else if s.updateDate(&task.DueDate, &dto.DueDate) {
		needsSync = true
	}

	if !dto.DoneAt.IsZero() {
		if t := util.ToTimePtr(&dto.DoneAt); t != nil {
			task.DoneAt = *t
		}
	}

	return needsSync
}

func (s *taskService) updateDate(field **util.LocalDateTime, newValue *util.LocalDateTime) bool {
	if newValue == nil || newValue.IsZero() {
		return false
	}

	if *field == nil || !(*field).Equal(*newValue) {
		copied := *newValue
		*field = &copied
		return true
	}

	return false
}

func (s *taskService) buildDashboardStats(tasks []*Task) *DashboardStatsResponse {
	stats := TaskStats{Total: len(tasks)}
	typeStats := TaskTypeStats{}
	var tasksThisMonth []*Task

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()

	for _, task := range tasks {
		s.countTaskByStatus(task, &stats, now)
		s.countTaskByType(task, &typeStats)

		if s.isTaskInCurrentMonth(task, currentYear, currentMonth) {
			tasksThisMonth = append(tasksThisMonth, task)
		}
	}

	return &DashboardStatsResponse{
		Stats:     stats,
		Type:      typeStats,
		Month:     tasksThisMonth,
		LastTasks: s.getRecentTasks(tasks),
	}
}

func (s *taskService) countTaskByStatus(task *Task, stats *TaskStats, now time.Time) {
	switch task.Status {
	case "TODO":
		stats.Todo++
	case "IN_PROGRESS":
		stats.InProgress++
	case "DONE":
		stats.Done++
	default:
		stats.Todo++
	}

	if task.Status != "DONE" && task.DueDate != nil && task.DueDate.Time.Before(now) {
		stats.Overdue++
	}
}

func (s *taskService) countTaskByType(task *Task, typeStats *TaskTypeStats) {
	switch task.Type {
	case "EVENT":
		typeStats.Event++
	case "STUDY":
		typeStats.Study++
	case "PROJECT":
		typeStats.Project++
	}
}

func (s *taskService) isTaskInCurrentMonth(task *Task, year int, month time.Month) bool {
	if task.DueDate == nil {
		return false
	}
	dueYear, dueMonth, _ := task.DueDate.Time.Date()
	return dueYear == year && dueMonth == month
}

func (s *taskService) getRecentTasks(tasks []*Task) []*Task {
	if len(tasks) == 0 {
		return []*Task{}
	}

	sorted := make([]*Task, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	if len(sorted) > dashboardTaskLimit {
		return sorted[:dashboardTaskLimit]
	}

	return sorted
}
