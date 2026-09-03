package models

import "time"

type HifzProgress struct {
	ID uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null;index:idx_user_surah" json:"user_id"`
	SurahNumber int `gorm:"not null;index:idx_user_surah" json:"surah_number"`
	StartAyah int `gorm:"not null" json:"start_ayah"`
	EndAyah int `gorm:"not null" json:"end_ayah"`
	MasteryLevel int `gorm:"not null;default:1" json:"mastery_level"`
	Repetitions int `gorm:"not null;default:0" json:"repetitions"`
	LastReviewedAt *time.Time `gorm:"index" json:"last_reviewed_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`

}