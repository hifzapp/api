package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GoogleID  string    `gorm:"uniqueIndex;not null" json:"google_id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Hearts    int       `gorm:"default:5" json:"hearts"`   
	Streak    int       `gorm:"default:0" json:"streak"`  
	DailyGoal   int       `gorm:"default:5" json:"daily_goal"`
	Onboarded   bool      `gorm:"default:false" json:"onboarded"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	HifzProgresses []HifzProgress `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	DailyProgresses []DailyProgress `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"daily_progresses,omitempty"`
	HifzSessions    []HifzSession   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"hifz_sessions,omitempty"`
}