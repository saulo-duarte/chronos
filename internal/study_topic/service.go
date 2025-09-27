package studytopic

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	studysubject "github.com/saulo-duarte/chronos-lambda/internal/study_subject"
	"github.com/sirupsen/logrus"
)

var (
	ErrStudyTopicNotFound   = errors.New("study topic not found")
	ErrStudySubjectNotFound = studysubject.ErrStudySubjectNotFound
	ErrUnauthorized         = errors.New("unauthorized")
)

type StudyTopicService interface {
	CreateStudyTopic(ctx context.Context, topic *StudyTopic) (*StudyTopic, error)
	GetStudyTopicByID(ctx context.Context, id string) (*StudyTopic, error)
	ListStudyTopicsBySubject(ctx context.Context, studySubjectID string) ([]*StudyTopic, error)
	UpdateStudyTopic(ctx context.Context, topic *StudyTopic) (*StudyTopic, error)
	DeleteStudyTopic(ctx context.Context, id string) error
}

type studyTopicService struct {
	repo        StudyTopicRepository
	subjectRepo studysubject.StudySubjectRepository
}

func NewService(repo StudyTopicRepository, subjectRepo studysubject.StudySubjectRepository) StudyTopicService {
	return &studyTopicService{repo: repo, subjectRepo: subjectRepo}
}

func (s *studyTopicService) CreateStudyTopic(ctx context.Context, topic *StudyTopic) (*StudyTopic, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to create study topic without authentication")
		return nil, ErrUnauthorized
	}

	if topic.Name == "" {
		log.Warn("Study topic name cannot be empty")
		return nil, errors.New("study topic name cannot be empty")
	}

	subject, err := s.subjectRepo.GetByID(topic.StudySubjectID.String())
	if err != nil {
		log.WithError(err).Error("Error fetching study subject for topic creation")
		return nil, err
	}
	if subject == nil {
		return nil, ErrStudySubjectNotFound
	}
	if subject.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"subject_id": subject.ID,
			"user_id":    claims.UserID,
		}).Warn("User attempted to create a topic for another user's subject")
		return nil, ErrUnauthorized
	}

	if topic.Position != 0 {
		if err := s.validateUniquePosition(topic.Position, topic.StudySubjectID.String(), claims.UserID, ""); err != nil {
			return nil, err
		}
	}

	topic.ID = uuid.New()
	topic.UserID = uuid.MustParse(claims.UserID)
	topic.CreatedAt = time.Now()
	topic.UpdatedAt = time.Now()

	if err := s.repo.Create(topic); err != nil {
		log.WithError(err).Error("Failed to create study topic")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"topic_id":   topic.ID,
		"user_id":    topic.UserID,
		"subject_id": topic.StudySubjectID,
	}).Info("Study topic created successfully")

	return topic, nil
}

func (s *studyTopicService) GetStudyTopicByID(ctx context.Context, id string) (*StudyTopic, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to get study topic without authentication")
		return nil, ErrUnauthorized
	}

	topic, err := s.repo.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Error fetching study topic by ID")
		return nil, err
	}
	if topic == nil {
		return nil, ErrStudyTopicNotFound
	}

	if topic.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"topic_id": topic.ID,
			"user_id":  claims.UserID,
		}).Warn("User attempted to access another user's study topic")
		return nil, ErrUnauthorized
	}

	return topic, nil
}

func (s *studyTopicService) ListStudyTopicsBySubject(ctx context.Context, studySubjectID string) ([]*StudyTopic, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to list study topics without authentication")
		return nil, ErrUnauthorized
	}

	subject, err := s.subjectRepo.GetByID(studySubjectID)
	if err != nil {
		log.WithError(err).Error("Error fetching study subject for topic listing")
		return nil, err
	}
	if subject == nil {
		return nil, ErrStudySubjectNotFound
	}
	if subject.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"subject_id": studySubjectID,
			"user_id":    claims.UserID,
		}).Warn("User attempted to list topics for another user's subject")
		return nil, ErrUnauthorized
	}

	topics, err := s.repo.ListBySubject(studySubjectID)
	if err != nil {
		log.WithError(err).Error("Error listing study topics by subject")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"subject_id": studySubjectID,
		"user_id":    claims.UserID,
		"count":      len(topics),
	}).Info("Study topics listed successfully")

	return topics, nil
}

func (s *studyTopicService) UpdateStudyTopic(ctx context.Context, topic *StudyTopic) (*StudyTopic, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to update study topic without authentication")
		return nil, ErrUnauthorized
	}

	if topic.Name == "" {
		log.Warn("Study topic name cannot be empty")
		return nil, errors.New("study topic name cannot be empty")
	}

	existing, err := s.repo.GetByID(topic.ID.String())
	if err != nil {
		log.WithError(err).Error("Error fetching study topic for update")
		return nil, err
	}
	if existing == nil {
		return nil, ErrStudyTopicNotFound
	}

	if existing.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"topic_id": existing.ID,
			"user_id":  claims.UserID,
		}).Warn("User attempted to update another user's study topic")
		return nil, ErrUnauthorized
	}

	if topic.Position != existing.Position {
		if err := s.validateUniquePosition(topic.Position, existing.StudySubjectID.String(), claims.UserID, existing.ID.String()); err != nil {
			return nil, err
		}
	}

	existing.Name = topic.Name
	existing.Description = topic.Description
	existing.Position = topic.Position
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		log.WithError(err).Error("Failed to update study topic")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"topic_id": existing.ID,
		"user_id":  claims.UserID,
	}).Info("Study topic updated successfully")

	return existing, nil
}

func (s *studyTopicService) DeleteStudyTopic(ctx context.Context, id string) error {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to delete study topic without authentication")
		return ErrUnauthorized
	}

	topic, err := s.repo.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Error fetching study topic for deletion")
		return err
	}
	if topic == nil {
		return ErrStudyTopicNotFound
	}

	if topic.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"topic_id": topic.ID,
			"user_id":  claims.UserID,
		}).Warn("User attempted to delete another user's study topic")
		return ErrUnauthorized
	}

	if err := s.repo.Delete(id); err != nil {
		log.WithError(err).Error("Failed to delete study topic")
		return err
	}

	log.WithFields(logrus.Fields{
		"topic_id": id,
		"user_id":  claims.UserID,
	}).Info("Study topic deleted successfully")

	return nil
}

func (s *studyTopicService) validateUniquePosition(position int, studySubjectID string, userID string, excludeID string) error {
	topics, err := s.repo.ListBySubject(studySubjectID)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		if topic.Position == position && topic.UserID.String() == userID {
			if excludeID != "" && topic.ID.String() == excludeID {
				continue
			}
			return errors.New("já existe um tópico com esta posição neste assunto")
		}
	}

	return nil
}
