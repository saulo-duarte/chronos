package project

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
	ErrProjectNotFound = errors.New("project not found")
	ErrUnauthorized    = errors.New("unauthorized")
)

type ProjectService interface {
	CreateProject(ctx context.Context, p *Project) (*Project, error)
	GetProjectByID(ctx context.Context, id string) (*Project, error)
	ListProjectsByUser(ctx context.Context) ([]*Project, error)
	UpdateProject(ctx context.Context, id string, dto *UpdateProjectDTO) (*Project, error)
	DeleteProject(ctx context.Context, id string) error
}

type projectService struct {
	repo ProjectRepository
}

func NewService(repo ProjectRepository) ProjectService {
	return &projectService{repo: repo}
}

func (s *projectService) CreateProject(ctx context.Context, p *Project) (*Project, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Tentativa de criar projeto sem autenticação")
		return nil, ErrUnauthorized
	}

	if p.Title == "" {
		log.Warn("Título do projeto não pode ser vazio")
		return nil, errors.New("project title cannot be empty")
	}

	if p.Status == "" {
		p.Status = ProjectStatus(NOT_INITIALIZED)
	}

	p.ID = uuid.New()
	p.UserID = uuid.MustParse(claims.UserID)
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	if err := s.repo.Create(p); err != nil {
		log.WithError(err).Error("Falha ao criar projeto")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"project_id": p.ID,
		"user_id":    p.UserID,
	}).Info("Projeto criado com sucesso")

	return p, nil
}

func (s *projectService) GetProjectByID(ctx context.Context, id string) (*Project, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Tentativa de acessar projeto sem autenticação")
		return nil, ErrUnauthorized
	}

	project, err := s.repo.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar projeto por ID")
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	if project.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"project_id": project.ID,
			"user_id":    claims.UserID,
		}).Warn("Usuário tentou acessar projeto de outro usuário")
		return nil, ErrUnauthorized
	}

	return project, nil
}

func (s *projectService) ListProjectsByUser(ctx context.Context) ([]*Project, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Tentativa de listar projetos sem autenticação")
		return nil, ErrUnauthorized
	}

	userID, _ := uuid.Parse(claims.UserID)
	projects, err := s.repo.ListByUser(userID)
	if err != nil {
		log.WithError(err).Error("Erro ao listar projetos do usuário")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"user_id": claims.UserID,
		"count":   len(projects),
	}).Info("Projetos listados com sucesso")

	return projects, nil
}

func (s *projectService) UpdateProject(ctx context.Context, id string, dto *UpdateProjectDTO) (*Project, error) {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Tentativa de atualizar projeto sem autenticação")
		return nil, ErrUnauthorized
	}

	if err := dto.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar projeto para atualização")
		return nil, err
	}
	if existing == nil {
		return nil, ErrProjectNotFound
	}

	if existing.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"project_id": existing.ID,
			"user_id":    claims.UserID,
		}).Warn("Usuário tentou atualizar projeto de outro usuário")
		return nil, ErrUnauthorized
	}

	existing.Title = dto.Title
	existing.Description = dto.Description

	existing.UpdatedAt = time.Now()

	if err := s.repo.Update(existing); err != nil {
		log.WithError(err).Error("Falha ao atualizar projeto")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"project_id": existing.ID,
		"user_id":    claims.UserID,
	}).Info("Projeto atualizado com sucesso")

	return existing, nil
}

func (s *projectService) DeleteProject(ctx context.Context, id string) error {
	log := config.WithContext(ctx)

	claims, err := auth.GetUserClaimsFromContext(ctx)
	if err != nil {
		log.WithError(err).Warn("Tentativa de deletar projeto sem autenticação")
		return ErrUnauthorized
	}

	project, err := s.repo.GetByID(id)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar projeto para exclusão")
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}

	if project.UserID.String() != claims.UserID {
		log.WithFields(logrus.Fields{
			"project_id": project.ID,
			"user_id":    claims.UserID,
		}).Warn("Usuário tentou deletar projeto de outro usuário")
		return ErrUnauthorized
	}

	if err := s.repo.Delete(id); err != nil {
		log.WithError(err).Error("Falha ao deletar projeto")
		return err
	}

	log.WithFields(logrus.Fields{
		"project_id": id,
		"user_id":    claims.UserID,
	}).Info("Projeto deletado com sucesso")

	return nil
}
