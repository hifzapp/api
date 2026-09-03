package models

import "time"

type HifzSession struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	SessionType string    `gorm:"size:50;not null" json:"session_type"`
	Duration    int       `json:"duration"`
	SurahNumber int       `json:"surah_number"`
	StartAyah   int       `json:"start_ayah"`
	EndAyah     int       `json:"end_ayah"`
	Score       int       `json:"score"`
	Mistakes    int       `gorm:"default:0" json:"mistakes"`
	HeartsLost  int       `gorm:"default:0" json:"hearts_lost"`
	IsComplete  bool      `gorm:"default:false" json:"is_complete"`
	CreatedAt   time.Time `json:"created_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}