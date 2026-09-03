package models

import "time"

type DailyProgress struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index" json:"user_id"`
	Date           time.Time `gorm:"type:date;not null;index" json:"date"` 
	NewAyahsCount  int       `gorm:"default:0" json:"new_ayahs_count"`    
	ReviewedCount  int       `gorm:"default:0" json:"reviewed_count"`    
	GoalTarget     int       `gorm:"default:20" json:"goal_target"`        
	IsGoalAchieved bool      `gorm:"default:false" json:"is_goal_achieved"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}