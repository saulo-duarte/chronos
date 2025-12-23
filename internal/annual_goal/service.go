package annual_goal

import (
	"errors"

	"github.com/google/uuid"
)

type Service interface {
	Create(userID uuid.UUID, dto CreateAnnualGoalDTO) (*AnnualGoalResponse, error)
	List(userID uuid.UUID) ([]AnnualGoalResponse, error)
	Update(id uuid.UUID, userID uuid.UUID, dto UpdateAnnualGoalDTO) (*AnnualGoalResponse, error)
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(userID uuid.UUID, dto CreateAnnualGoalDTO) (*AnnualGoalResponse, error) {
	goal := AnnualGoal{
		UserID:      userID,
		Title:       dto.Title,
		Description: dto.Description,
		Year:        dto.Year,
		Status:      AnnualGoalStatusActive,
	}

	if err := s.repo.Create(&goal); err != nil {
		return nil, err
	}

	return s.toResponse(&goal), nil
}

func (s *service) List(userID uuid.UUID) ([]AnnualGoalResponse, error) {
	goals, err := s.repo.FindAllByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []AnnualGoalResponse
	for _, goal := range goals {
		responses = append(responses, *s.toResponse(&goal))
	}
	return responses, nil
}

func (s *service) Update(id uuid.UUID, userID uuid.UUID, dto UpdateAnnualGoalDTO) (*AnnualGoalResponse, error) {
	goal, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if goal.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if dto.Title != nil {
		goal.Title = *dto.Title
	}
	if dto.Description != nil {
		goal.Description = *dto.Description
	}
	if dto.Year != nil {
		goal.Year = *dto.Year
	}
	if dto.Status != nil {
		goal.Status = *dto.Status
	}

	if err := s.repo.Update(goal); err != nil {
		return nil, err
	}

	return s.toResponse(goal), nil
}

func (s *service) Delete(id uuid.UUID, userID uuid.UUID) error {
	goal, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if goal.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.repo.Delete(id)
}

func (s *service) toResponse(goal *AnnualGoal) *AnnualGoalResponse {
	return &AnnualGoalResponse{
		ID:          goal.ID,
		Title:       goal.Title,
		Description: goal.Description,
		Year:        goal.Year,
		Status:      goal.Status,
		UserID:      goal.UserID,
		CreatedAt:   goal.CreatedAt,
		UpdatedAt:   goal.UpdatedAt,
	}
}
