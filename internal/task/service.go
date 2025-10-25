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
	"github.com/sirupsen/logrus"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrProjectNotFound    = project.ErrProjectNotFound
	ErrStudyTopicNotFound = studytopic.ErrStudyTopicNotFound
	ErrInvalidID          = errors.New("invalid id format")
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

	if err := s.validateDependencies(ctx, t); err != nil {
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

	taskID, err := s.parseID(ctx, id, "task")
	if err != nil {
		return nil, err
	}

	return s.findTask(ctx, taskID, userID)
}

func (s *taskService) DeleteByID(ctx context.Context, id string) error {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return err
	}

	taskID, err := s.parseID(ctx, id, "task")
	if err != nil {
		return err
	}

	task, err := s.findTask(ctx, taskID, userID)
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

	pid, err := s.parseID(ctx, projectID, "project")
	if err != nil {
		return nil, err
	}

	if err := s.validateProject(ctx, &pid); err != nil {
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

	tid, err := s.parseID(ctx, topicID, "study topic")
	if err != nil {
		return nil, err
	}

	if err := s.validateStudyTopic(ctx, &tid); err != nil {
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

	task, err := s.findTask(ctx, dto.ID, userID)
	if err != nil {
		return nil, err
	}

	needsCalendarSync := s.applyUpdates(task, dto)
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

func (s *taskService) getUserID(ctx context.Context) (uuid.UUID, error) {
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		config.WithContext(ctx).WithError(err).Warn("Unauthorized access attempt")
		return uuid.Nil, ErrUnauthorized
	}
	return uuid.MustParse(claims.UserID), nil
}

func (s *taskService) parseID(ctx context.Context, id, entityName string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		config.WithContext(ctx).WithError(err).Warnf("Invalid %s ID: %s", entityName, id)
		return uuid.Nil, ErrInvalidID
	}
	return parsedID, nil
}

func (s *taskService) findTask(ctx context.Context, taskID, userID uuid.UUID) (*Task, error) {
	task, err := s.repo.FindByIdAndUserId(taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			config.WithContext(ctx).WithFields(logrus.Fields{
				"task_id": taskID,
				"user_id": userID,
			}).Warn("Task not found or unauthorized")
			return nil, ErrTaskNotFound
		}
		config.WithContext(ctx).WithError(err).Error("Error finding task")
		return nil, err
	}

	return task, nil
}

func (s *taskService) validateProject(ctx context.Context, projectID *uuid.UUID) error {
	if projectID == nil {
		return nil
	}

	if _, err := s.projectService.GetProjectByID(ctx, projectID.String()); err != nil {
		config.WithContext(ctx).WithError(err).WithField("project_id", projectID).Error("Project validation failed")
		return ErrProjectNotFound
	}

	return nil
}

func (s *taskService) validateStudyTopic(ctx context.Context, topicID *uuid.UUID) error {
	if topicID == nil {
		return nil
	}

	if _, err := s.studyTopicRepo.GetByID(topicID.String()); err != nil {
		config.WithContext(ctx).WithError(err).WithField("study_topic_id", topicID).Error("Study topic validation failed")
		return ErrStudyTopicNotFound
	}

	return nil
}

func (s *taskService) validateDependencies(ctx context.Context, t *Task) error {
	if t.Type == "PROJECT" && t.ProjectId == nil {
		return errors.New("projectId is required for PROJECT tasks")
	}

	if err := s.validateProject(ctx, t.ProjectId); err != nil {
		return err
	}

	return s.validateStudyTopic(ctx, t.StudyTopicId)
}

func (s *taskService) toCalendarTask(t *Task) *googlecalendar.CalendarTask {
	var eventID *string
	if t.GoogleCalendarEventID != "" {
		eventID = &t.GoogleCalendarEventID
	}

	return &googlecalendar.CalendarTask{
		ID:                    t.ID,
		Name:                  t.Name,
		Description:           t.Description,
		StartDate:             util.ToTimePtr(t.StartDate),
		DueDate:               util.ToTimePtr(t.DueDate),
		GoogleCalendarEventID: eventID,
	}
}

func (s *taskService) syncWithCalendar(ctx context.Context, userID uuid.UUID, t *Task) {
	calTask := s.toCalendarTask(t)
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

func (s *taskService) applyUpdates(task *Task, dto *TaskUpdateDTO) bool {
	needsCalendarSync := false

	needsCalendarSync = s.updateStringField(&task.Name, dto.Name) || needsCalendarSync
	needsCalendarSync = s.updateStringField(&task.Description, dto.Description) || needsCalendarSync
	s.updateTaskStatus(&task.Status, dto.Status)
	s.updateTaskPriority(&task.Priority, dto.Priority)

	needsCalendarSync = s.updateDateField(&task.StartDate, dto.StartDate) || needsCalendarSync
	needsCalendarSync = s.updateDueDate(task, dto) || needsCalendarSync

	if !dto.DoneAt.IsZero() {
		task.DoneAt = dto.DoneAt
	}

	return needsCalendarSync
}

func (s *taskService) updateStringField(field *string, newValue string) bool {
	if newValue != "" && newValue != *field {
		*field = newValue
		return true
	}
	return false
}

func (s *taskService) updateTaskStatus(field *TaskStatus, newValue TaskStatus) {
	if newValue != "" && newValue != *field {
		*field = newValue
	}
}

func (s *taskService) updateTaskPriority(field *TaskPriority, newValue TaskPriority) {
	if newValue != "" && newValue != *field {
		*field = newValue
	}
}

func (s *taskService) updateDateField(field **util.LocalDateTime, newValue time.Time) bool {
	if newValue.IsZero() {
		return false
	}

	newDate := util.LocalDateTime{Time: newValue}
	if *field == nil || !(*field).Equal(newDate) {
		*field = &newDate
		return true
	}
	return false
}

func (s *taskService) updateDueDate(task *Task, dto *TaskUpdateDTO) bool {
	if dto.RemoveDueDate {
		if task.DueDate != nil {
			task.DueDate = nil
			return true
		}
		return false
	}

	return s.updateDateField(&task.DueDate, dto.DueDate)
}

func (s *taskService) buildDashboardStats(tasks []*Task) *DashboardStatsResponse {
	stats := TaskStats{Total: len(tasks)}
	typeStats := TaskTypeStats{}
	var tasksThisMonth []*Task

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()

	for _, task := range tasks {
		s.updateTaskStats(task, &stats, &typeStats, now)
		s.addTaskIfCurrentMonth(task, &tasksThisMonth, currentYear, currentMonth)
	}

	return &DashboardStatsResponse{
		Stats:     stats,
		Type:      typeStats,
		Month:     tasksThisMonth,
		LastTasks: s.getLastTasks(tasks),
	}
}

func (s *taskService) updateTaskStats(task *Task, stats *TaskStats, typeStats *TaskTypeStats, now time.Time) {
	s.updateStatusStats(task, stats, now)
	s.updateTypeStats(task, typeStats)
}

func (s *taskService) updateStatusStats(task *Task, stats *TaskStats, now time.Time) {
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

func (s *taskService) updateTypeStats(task *Task, typeStats *TaskTypeStats) {
	switch task.Type {
	case "EVENT":
		typeStats.Event++
	case "STUDY":
		typeStats.Study++
	case "PROJECT":
		typeStats.Project++
	}
}

func (s *taskService) addTaskIfCurrentMonth(task *Task, tasksThisMonth *[]*Task, year int, month time.Month) {
	if task.DueDate == nil {
		return
	}

	dueYear, dueMonth, _ := task.DueDate.Time.Date()
	if dueYear == year && dueMonth == month {
		*tasksThisMonth = append(*tasksThisMonth, task)
	}
}

func (s *taskService) getLastTasks(tasks []*Task) []*Task {
	if len(tasks) == 0 {
		return []*Task{}
	}

	sortedTasks := s.sortTasksByCreatedAt(tasks)

	if len(sortedTasks) > dashboardTaskLimit {
		return sortedTasks[:dashboardTaskLimit]
	}

	return sortedTasks
}

func (s *taskService) sortTasksByCreatedAt(tasks []*Task) []*Task {
	sorted := make([]*Task, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	return sorted
}
