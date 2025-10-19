package quiz

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"gorm.io/gorm"
)

type QuizService interface {
	CreateQuizWithQuestions(ctx context.Context, quiz *Quiz, questions []*QuizQuestion) error
	DeleteQuiz(ctx context.Context, quizID string) error
	AddQuestionToQuiz(ctx context.Context, quizID string, question *QuizQuestion) error
	RemoveQuestion(ctx context.Context, questionID string) error
	GetQuizWithQuestions(ctx context.Context, quizID string) (*QuizWithQuestionsDTO, error)
	ListQuizzesByUser(ctx context.Context, userID string) ([]*Quiz, error)
}

type quizService struct {
	repo QuizRepository
	db   *gorm.DB
}

func NewService(db *gorm.DB, repo QuizRepository) QuizService {
	return &quizService{
		repo: repo,
		db:   db,
	}
}

func (s *quizService) CreateQuizWithQuestions(ctx context.Context, quiz *Quiz, questions []*QuizQuestion) error {
	log := config.WithContext(ctx)
	log.Info("Criando novo quiz...")

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(quiz).Error; err != nil {
			log.Errorf("Erro ao criar quiz: %v", err)
			return err
		}

		for i := range questions {
			questions[i].QuizID = quiz.ID
		}

		if len(questions) > 0 {
			if err := tx.Create(&questions).Error; err != nil {
				log.Errorf("Erro ao criar perguntas do quiz: %v", err)
				return err
			}
		}

		log.Info("Quiz criado com sucesso", "quiz_id", quiz.ID.String())
		return nil
	})
}

func (s *quizService) DeleteQuiz(ctx context.Context, quizID string) error {
	log := config.WithContext(ctx)
	log.Info("Deletando quiz...")

	if err := s.repo.Delete(quizID); err != nil {
		log.Errorf("Erro ao deletar quiz: %v", err)
		return err
	}

	log.Info("Quiz deletado com sucesso")
	return nil
}

func (s *quizService) AddQuestionToQuiz(ctx context.Context, quizID string, question *QuizQuestion) error {
	log := config.WithContext(ctx)
	log.Info("Adicionando nova pergunta ao quiz...")

	qz, err := s.repo.GetByID(quizID)
	if err != nil {
		log.Errorf("Erro ao buscar quiz: %v", err)
		return err
	}
	if qz == nil {
		err := errors.New("quiz não encontrado")
		log.Warnf("Quiz não encontrado: %v", err)
		return err
	}

	question.QuizID = qz.ID
	if question.ID == uuid.Nil {
		question.ID = uuid.New()
	}

	if err := s.repo.AddQuestions([]*QuizQuestion{question}); err != nil {
		log.Errorf("Erro ao adicionar pergunta: %v", err)
		return err
	}

	log.Info("Pergunta adicionada com sucesso", "question_id", question.ID.String())
	return nil
}

func (s *quizService) RemoveQuestion(ctx context.Context, questionID string) error {
	log := config.WithContext(ctx)
	log.Info("Removendo pergunta...", "question_id", questionID)

	if err := s.repo.DeleteQuestion(questionID); err != nil {
		log.Errorf("Erro ao remover pergunta: %v", err)
		return err
	}

	log.Info("Pergunta removida com sucesso", "question_id", questionID)
	return nil
}

func (s *quizService) GetQuizWithQuestions(ctx context.Context, quizID string) (*QuizWithQuestionsDTO, error) {
	log := config.WithContext(ctx)
	log.Info("Buscando quiz com perguntas...", "quiz_id", quizID)

	quiz, err := s.repo.GetByID(quizID)
	if err != nil {
		log.Errorf("Erro ao buscar quiz: %v", err)
		return nil, err
	}
	if quiz == nil {
		return nil, nil
	}

	questions, err := s.repo.ListQuestionsByQuiz(quizID)
	if err != nil {
		log.Errorf("Erro ao listar perguntas do quiz: %v", err)
		return nil, err
	}

	return &QuizWithQuestionsDTO{
		Quiz:      quiz,
		Questions: questions,
	}, nil
}

func (s *quizService) ListQuizzesByUser(ctx context.Context, userID string) ([]*Quiz, error) {
	log := config.WithContext(ctx)
	log.Info("Listando quizzes do usuário...", "user_id", userID)

	quizzes, err := s.repo.ListQuizzesByUser(userID)
	if err != nil {
		log.Errorf("Erro ao listar quizzes do usuário: %v", err)
		return nil, err
	}

	return quizzes, nil
}
