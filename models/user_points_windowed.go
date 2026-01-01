package models

import (
	"time"

	"gorm.io/gorm"
)

type UserPointsWindowed struct {
	gorm.Model
	UserID      uint      `gorm:"not null"`
	ClubID      uint      `gorm:"not null"`
	Points      int       `gorm:"not null"`
	Source      string    `gorm:"size:50"`
	ReferenceID *uint
	WeekStart   time.Time `gorm:"type:date;not null"`
	MonthStart  time.Time `gorm:"type:date;not null"`
	YearStart   time.Time `gorm:"type:date;not null"`
}

func (UserPointsWindowed) TableName() string {
	return "user_points_windowed"
}
