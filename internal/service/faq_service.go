package service

import (
	"context"
	"errors"
	"strings"

	"github.com/AYaSmyslov/faqapi/internal/models"
	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrBadRequest     = errors.New("bad request")
	ErrNoSuchQuestion = errors.New("question does not exitst")
)

type FAQService struct {
	db *gorm.DB
}

func NewFAQService(db *gorm.DB) *FAQService {
	return &FAQService{db: db}
}

func (s *FAQService) CreateQuestion(ctx context.Context, text string) (*models.Question, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ErrBadRequest
	}

	question := &models.Question{Text: text}
	if err := s.db.WithContext(ctx).Create(question).Error; err != nil {
		return nil, err
	}

	return question, nil
}

func (s *FAQService) ListQuestions(ctx context.Context) ([]models.Question, error) {
	var questions []models.Question

	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&questions).Error; err != nil {
		return nil, err
	}

	return questions, nil
}

func (s *FAQService) GetQuestionWithAnswers(ctx context.Context, id uint) (*models.Question, error) {
	var question models.Question

	if err := s.db.WithContext(ctx).
		Preload("Answers", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &question, nil
}

func (s *FAQService) DeleteQuestion(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&models.Question{}, id)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *FAQService) CreateAnswer(ctx context.Context, questionID uint, userID, text string) (*models.Answer, error) {
	userID = strings.TrimSpace(userID)
	text = strings.TrimSpace(text)
	if userID == "" || text == "" {
		return nil, ErrBadRequest
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Question{}).
		Where("id = ?", questionID).
		Count(&count).Error; err != nil {
		return nil, err
	}

	answer := &models.Answer{
		QuestionID: questionID,
		UserID:     userID,
		Text:       text,
	}

	if err := s.db.WithContext(ctx).Create(answer).Error; err != nil {
		return nil, err
	}

	return answer, nil

}

func (s *FAQService) GetAnswer(ctx context.Context, id uint) (*models.Answer, error) {
	var answer models.Answer

	if err := s.db.WithContext(ctx).First(&answer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &answer, nil
}

func (s *FAQService) DeleteAnswer(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&models.Answer{}, id)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
