package models

import (
	"time"

	"gorm.io/gorm"
)

type UserClubLeaderboardSnapshot struct {
	gorm.Model
	UserID            uint      `gorm:"not null"`
	ClubID            uint      `gorm:"not null"`
	PointsContributed int       `gorm:"not null;default:0"`
	PeriodType        string    `gorm:"size:20;not null"` // weekly, monthly, annual, all_time
	PeriodStart       time.Time `gorm:"not null"`
	PeriodEnd         time.Time `gorm:"not null"`
	ClubRank          *int
}

func (UserClubLeaderboardSnapshot) TableName() string {
	return "user_club_leaderboard_snapshots"
}
