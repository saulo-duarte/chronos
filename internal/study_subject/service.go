package studysubject

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/sirupsen/logrus"
)

var (
	ErrStudySubjectNotFound = errors.New("study subject not found")
	ErrUnauthorized         = errors.New("unauthorized")
)

type StudySubjectService interface {
	CreateStudySubject(ctx context.Context, subj *StudySubject) (*StudySubject, error)
	ListStudySubjectsByUser(ctx context.Context, userID string) ([]*StudySubject, error)
	UpdateStudySubject(ctx context.Context, subj *StudySubject) (*StudySubject, error)
	DeleteStudySubject(ctx context.Context, id string) error
}

type studySubjectService struct {
	repo StudySubjectRepository
}

func NewService(repo StudySubjectRepository) StudySubjectService {
	return &studySubjectService{repo: repo}
}

func (s *studySubjectService) CreateStudySubject(ctx context.Context, subj *StudySubject) (*StudySubject, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to create study subject without authentication")
		return nil, ErrUnauthorized
	}

	if subj.Name == "" {
		log.Warn("study subject name cannot be empty")
		return nil, errors.New("study subject name cannot be empty")
	}

	subj.UserID = uuid.MustParse(claims.UserID)
	subj.ID = uuid.New()
	subj.CreatedAt = time.Now()
	subj.UpdatedAt = time.Now()

	if err := s.repo.Create(subj); err != nil {
		log.WithError(err).Error("failed to create study subject")
		return nil, err
	}
	return subj, nil
}

func (s *studySubjectService) ListStudySubjectsByUser(ctx context.Context, userID string) ([]*StudySubject, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to list study subjects without authentication")
		return nil, ErrUnauthorized
	}

	if claims.UserID != userID {
		log.WithFields(logrus.Fields{
			"user_id_from_path":   userID,
			"user_id_from_claims": claims.UserID,
		}).Warn("User attempted to list another user's study subjects")
		return nil, ErrUnauthorized
	}

	subjects, err := s.repo.ListByUser(userID)
	if err != nil {
		log.WithError(err).Error("failed to list study subjects by user")
		return nil, err
	}
	return subjects, nil
}

func (s *studySubjectService) UpdateStudySubject(ctx context.Context, subj *StudySubject) (*StudySubject, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to update study subject without authentication")
		return nil, ErrUnauthorized
	}

	existing, err := s.repo.GetByID(subj.ID.String())
	if err != nil {
		log.WithError(err).Error("Error fetching study subject for update")
		return nil, err
	}
	if existing == nil {
		return nil, ErrStudySubjectNotFound
	}

	if existing.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"subject_id": existing.ID,
			"user_id":    claims.UserID,
		}).Warn("User attempted to update another user's study subject")
		return nil, ErrUnauthorized
	}

	if subj.Name == "" {
		log.Warn("study subject name cannot be empty")
		return nil, errors.New("study subject name cannot be empty")
	}

	existing.Name = subj.Name
	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		log.WithError(err).Error("failed to update study subject")
		return nil, err
	}
	return existing, nil
}

func (s *studySubjectService) DeleteStudySubject(ctx context.Context, id string) error {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Attempt to delete study subject without authentication")
		return ErrUnauthorized
	}

	subject, err := s.repo.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Error fetching study subject for deletion")
		return err
	}
	if subject == nil {
		return ErrStudySubjectNotFound
	}

	if subject.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"subject_id": subject.ID,
			"user_id":    claims.UserID,
		}).Warn("User attempted to delete another user's study subject")
		return ErrUnauthorized
	}

	if err := s.repo.Delete(id); err != nil {
		log.WithError(err).Error("failed to delete study subject")
		return err
	}
	return nil
}
