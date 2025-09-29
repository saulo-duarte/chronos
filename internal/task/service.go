package task

import (
	"context"
	"errors"
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

type TaskService interface {
	CreateTask(ctx context.Context, t *Task) (*Task, error)
	FindAllByUser(ctx context.Context) ([]*Task, error)
	FindByID(ctx context.Context, id string) (*Task, error)
	DeleteByID(ctx context.Context, id string) error
	FindAllByProjectID(ctx context.Context, projectID string) ([]*Task, error)
	FindAllByTopicID(ctx context.Context, topicID string) ([]*Task, error)
	UpdateTask(ctx context.Context, t *Task) (*Task, error)
}

type taskService struct {
	repo            TaskRepository
	projectService  project.ProjectService
	userRepo        user.UserRepository
	studyTopicRepo  studytopic.StudyTopicRepository
	calendarService googlecalendar.CalendarService
}

func NewService(repo TaskRepository, projectService project.ProjectService, userRepo user.UserRepository, studyTopicRepo studytopic.StudyTopicRepository, calendarService googlecalendar.CalendarService) TaskService {
	return &taskService{
		repo:            repo,
		projectService:  projectService,
		userRepo:        userRepo,
		studyTopicRepo:  studyTopicRepo,
		calendarService: calendarService,
	}
}

func getUserIDFromContext(ctx context.Context, log logrus.FieldLogger, action string) (uuid.UUID, error) {
	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warnf("Attempt to %s without authentication", action)
		return uuid.Nil, ErrUnauthorized
	}
	return uuid.MustParse(claims.UserID), nil
}

func parseUUID(log logrus.FieldLogger, id string, entityName string) (uuid.UUID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		log.WithError(err).Warnf("Invalid %s ID", entityName)
		return uuid.Nil, ErrInvalidID
	}
	return parsedID, nil
}

func (s *taskService) validateTaskDependencies(ctx context.Context, log logrus.FieldLogger, t *Task) error {
	if t.Type == "PROJECT" && t.ProjectId == nil {
		return errors.New("projectId is required for PROJECT tasks")
	}

	if t.ProjectId != nil {
		if _, err := s.projectService.GetProjectByID(ctx, t.ProjectId.String()); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"project_id": t.ProjectId,
				"user_id":    t.UserID,
			}).Error("Project not found or does not belong to the user")
			return ErrProjectNotFound
		}
	}

	if t.StudyTopicId != nil {
		if _, err := s.studyTopicRepo.GetByID(t.StudyTopicId.String()); err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"study_topic_id": *t.StudyTopicId,
				"user_id":        t.UserID,
			}).Error("Study topic not found or does not belong to the user")
			return ErrStudyTopicNotFound
		}
	}

	return nil
}

func (s *taskService) CreateTask(ctx context.Context, t *Task) (*Task, error) {
	log := config.WithContext(ctx)
	userID, err := getUserIDFromContext(ctx, log, "create task")
	if err != nil {
		return nil, err
	}

	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	t.UserID = userID

	if err := s.validateTaskDependencies(ctx, log, t); err != nil {
		return nil, err
	}

	if err := s.repo.Create(t); err != nil {
		log.WithError(err).Error("Failed to create task")
		return nil, err
	}

	if t.StartDate != nil || t.DueDate != nil {
		calTask := googlecalendar.CalendarTask{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			StartDate:   util.ToTimePtr(t.StartDate),
			DueDate:     util.ToTimePtr(t.DueDate),
		}

		eventID, err := s.calendarService.AddEventToCalendar(ctx, t.UserID, &calTask)
		if err != nil {
			log.WithError(err).Warnf("Failed to add task %s to Google Calendar", t.ID)
		} else if eventID != "" {
			t.GoogleCalendarEventID = eventID
			if err := s.repo.Update(t); err != nil {
				log.WithError(err).Error("Failed to update task with Google Calendar Event ID")
			}
		}
	}

	log.WithField("task_id", t.ID).Info("Task created successfully")
	return t, nil
}

func (s *taskService) FindAllByUser(ctx context.Context) ([]*Task, error) {
	log := config.WithContext(ctx)
	userID, err := getUserIDFromContext(ctx, log, "list tasks")
	if err != nil {
		return nil, err
	}

	tasks, err := s.repo.ListByUser(userID)
	if err != nil {
		log.WithError(err).Error("Failed to list tasks by user")
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) FindByID(ctx context.Context, id string) (*Task, error) {
	log := config.WithContext(ctx)
	userID, err := getUserIDFromContext(ctx, log, "find task")
	if err != nil {
		return nil, err
	}

	taskID, err := parseUUID(log, id, "task")
	if err != nil {
		return nil, errors.New("invalid task id")
	}

	task, err := s.repo.FindByIdAndUserId(taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithFields(logrus.Fields{
				"task_id": id,
				"user_id": userID,
			}).Warn("Task not found or does not belong to user")
			return nil, ErrTaskNotFound
		}
		log.WithError(err).Error("Error finding task by ID")
		return nil, err
	}
	return task, nil
}

func (s *taskService) DeleteByID(ctx context.Context, id string) error {
	log := config.WithContext(ctx)
	userID, err := getUserIDFromContext(ctx, log, "delete task")
	if err != nil {
		return err
	}

	taskID, err := parseUUID(log, id, "task")
	if err != nil {
		return errors.New("invalid task id")
	}

	task, err := s.repo.FindByIdAndUserId(taskID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithFields(logrus.Fields{
				"task_id": id,
				"user_id": userID,
			}).Warn("Task not found or does not belong to user for deletion")
			return ErrTaskNotFound
		}
		log.WithError(err).Error("Error finding task before deletion")
		return err
	}

	googleEventID := task.GoogleCalendarEventID

	if err := s.repo.Delete(taskID, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrTaskNotFound
		}
		log.WithError(err).Error("Failed to delete task")
		return err
	}

	if googleEventID != "" {
		err := s.calendarService.DeleteEventFromCalendar(ctx, userID, googleEventID)
		if err != nil {
			log.WithError(err).Warnf("Failed to delete Google Calendar event %s for task %s", googleEventID, id)
		}
	}

	log.WithField("task_id", id).Info("Task deleted successfully")
	return nil
}

func (s *taskService) FindAllByProjectID(ctx context.Context, projectID string) ([]*Task, error) {
	log := config.WithContext(ctx)
	userID, err := getUserIDFromContext(ctx, log, "list tasks by project")
	if err != nil {
		return nil, err
	}

	pid, err := parseUUID(log, projectID, "project")
	if err != nil {
		return nil, errors.New("invalid project id")
	}

	if _, err := s.projectService.GetProjectByID(ctx, projectID); err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			log.WithFields(logrus.Fields{
				"project_id": projectID,
				"user_id":    userID,
			}).Warn("Project not found or does not belong to user")
			return nil, ErrProjectNotFound
		}
		log.WithError(err).Error("Error finding project by ID")
		return nil, err
	}

	tasks, err := s.repo.ListByProjectAndUser(pid, userID)
	if err != nil {
		log.WithError(err).Error("Failed to list tasks by project")
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) FindAllByTopicID(ctx context.Context, topicID string) ([]*Task, error) {
	log := config.WithContext(ctx)
	userID, err := getUserIDFromContext(ctx, log, "list tasks by topic")
	if err != nil {
		return nil, err
	}

	tid, err := parseUUID(log, topicID, "study topic")
	if err != nil {
		return nil, errors.New("invalid study topic id")
	}

	if _, err := s.studyTopicRepo.GetByID(tid.String()); err != nil {
		if errors.Is(err, studytopic.ErrStudyTopicNotFound) {
			log.WithFields(logrus.Fields{
				"topic_id": topicID,
				"user_id":  userID,
			}).Warn("Study topic not found or does not belong to user")
			return nil, ErrStudyTopicNotFound
		}
		log.WithError(err).Error("Error finding study topic by ID")
		return nil, err
	}

	tasks, err := s.repo.ListByStudyTopicAndUser(tid, userID)
	if err != nil {
		log.WithError(err).Error("Failed to list tasks by study topic")
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) UpdateTask(ctx context.Context, t *Task) (*Task, error) {
	log := config.WithContext(ctx)
	userID, err := getUserIDFromContext(ctx, log, "update task")
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByIdAndUserId(t.ID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			log.WithFields(logrus.Fields{
				"task_id": t.ID,
				"user_id": userID,
			}).Warn("Task not found for update")
			return nil, ErrTaskNotFound
		}
		log.WithError(err).Error("Error finding task for update")
		return nil, err
	}

	updateCalendar := false

	if t.StartDate != nil && (existing.StartDate == nil || !t.StartDate.Equal(*existing.StartDate)) {
		existing.StartDate = t.StartDate
		updateCalendar = true
	}
	if t.DueDate != nil && (existing.DueDate == nil || !t.DueDate.Equal(*existing.DueDate)) {
		existing.DueDate = t.DueDate
		updateCalendar = true
	}
	if t.Name != "" && existing.Name != t.Name {
		existing.Name = t.Name
		updateCalendar = true
	}
	if t.Description != "" && existing.Description != t.Description {
		existing.Description = t.Description
		updateCalendar = true
	}

	if t.Status != "" {
		existing.Status = t.Status
	}
	if t.Priority != "" {
		existing.Priority = t.Priority
	}
	if !t.DoneAt.IsZero() {
		existing.DoneAt = t.DoneAt
	}

	existing.UpdatedAt = time.Now()

	if updateCalendar {
		calTask := googlecalendar.CalendarTask{
			ID:          existing.ID,
			Name:        existing.Name,
			Description: existing.Description,
			StartDate:   util.ToTimePtr(existing.StartDate),
			DueDate:     util.ToTimePtr(existing.DueDate),
		}

		if existing.GoogleCalendarEventID != "" {
			calTask.GoogleCalendarEventID = &existing.GoogleCalendarEventID
			err = s.calendarService.UpdateEventInCalendar(ctx, userID, &calTask)
			if err != nil {
				log.WithError(err).Warnf("Failed to update Google Calendar event for task %s", existing.ID)
			}
		} else if existing.StartDate != nil || existing.DueDate != nil {
			eventID, err := s.calendarService.AddEventToCalendar(ctx, userID, &calTask)
			if err != nil {
				log.WithError(err).Warnf("Failed to add task %s to Google Calendar on update", existing.ID)
			} else if eventID != "" {
				existing.GoogleCalendarEventID = eventID
			}
		}
	}

	if err := s.repo.Update(existing); err != nil {
		log.WithError(err).Error("Failed to update task")
		return nil, err
	}

	log.WithField("task_id", existing.ID).Info("Task updated successfully")
	return existing, nil
}
