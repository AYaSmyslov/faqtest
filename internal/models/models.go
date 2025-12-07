package models

import "time"

type Question struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Text      string    `json:"text" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;"`
	Answers   []Answer  `json:"answers,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
}

type Answer struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	QuestionID uint      `json:"qusetion_id" gorm:"not null;index"`
	UserID     string    `json:"user_id" gorm:"type:varchar(64);not null"`
	Text       string    `json:"text" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null;"`
}
