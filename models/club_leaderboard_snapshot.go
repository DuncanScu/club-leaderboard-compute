package models

import (
	"time"

	"gorm.io/gorm"
)

type ClubLeaderboardSnapshot struct {
	gorm.Model
	ClubID            uint      `gorm:"not null"`
	TotalPoints       int       `gorm:"not null;default:0"`
	MemberCount       int       `gorm:"not null;default:0"`
	ActiveMemberCount int       `gorm:"not null;default:0"`
	PeriodType        string    `gorm:"size:20;not null"` // weekly, monthly, annual, all_time
	PeriodStart       time.Time `gorm:"not null"`
	PeriodEnd         time.Time `gorm:"not null"`
	GlobalRank        *int
	LocalRank         *int
	City              string    `gorm:"size:100"`
}

func (ClubLeaderboardSnapshot) TableName() string {
	return "club_leaderboard_snapshots"
}
