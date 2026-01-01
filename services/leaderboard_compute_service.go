package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Duncanscu/leaderboard-computer/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaderboardComputeService interface {
	ComputeAllLeaderboards(ctx context.Context) error
}

type leaderboardComputeService struct {
	db *gorm.DB
}

func NewLeaderboardComputeService(db *gorm.DB) LeaderboardComputeService {
	return &leaderboardComputeService{
		db: db,
	}
}

func (s *leaderboardComputeService) ComputeAllLeaderboards(ctx context.Context) error {
	now := time.Now()

	// Compute for all period types
	periods := []struct {
		periodType  string
		periodStart time.Time
		periodEnd   time.Time
		windowField string
	}{
		{
			periodType:  "weekly",
			periodStart: getWeekStart(now),
			periodEnd:   getWeekEnd(now),
			windowField: "week_start",
		},
		{
			periodType:  "monthly",
			periodStart: getMonthStart(now),
			periodEnd:   getMonthEnd(now),
			windowField: "month_start",
		},
		{
			periodType:  "annual",
			periodStart: getYearStart(now),
			periodEnd:   getYearEnd(now),
			windowField: "year_start",
		},
	}

	for _, period := range periods {
		log.Printf("[COMPUTE] Computing %s leaderboards for %s to %s",
			period.periodType, period.periodStart.Format("2006-01-02"), period.periodEnd.Format("2006-01-02"))

		// Compute club leaderboards
		if err := s.computeClubLeaderboard(period.periodType, period.periodStart, period.periodEnd, period.windowField); err != nil {
			log.Printf("[ERROR] Failed to compute club leaderboard for %s: %v", period.periodType, err)
			return err
		}

		// Compute user club contributor leaderboards
		if err := s.computeUserClubLeaderboard(period.periodType, period.periodStart, period.periodEnd, period.windowField); err != nil {
			log.Printf("[ERROR] Failed to compute user club leaderboard for %s: %v", period.periodType, err)
			return err
		}

		log.Printf("[SUCCESS] Completed %s leaderboards", period.periodType)
	}

	return nil
}

func (s *leaderboardComputeService) computeClubLeaderboard(periodType string, periodStart, periodEnd time.Time, windowField string) error {
	// Aggregate points per club for the period
	type ClubPoints struct {
		ClubID            uint
		TotalPoints       int
		ActiveMemberCount int
		City              string
	}

	var clubPoints []ClubPoints
	query := fmt.Sprintf(`
		SELECT
			c.id as club_id,
			COALESCE(SUM(upw.points), 0) as total_points,
			COUNT(DISTINCT upw.user_id) as active_member_count,
			c.address_city as city
		FROM clubs c
		LEFT JOIN user_points_windowed upw ON c.id = upw.club_id
			AND upw.%s = ?
		GROUP BY c.id, c.address_city
		ORDER BY total_points DESC
	`, windowField)

	if err := s.db.Raw(query, periodStart).Scan(&clubPoints).Error; err != nil {
		return fmt.Errorf("failed to aggregate club points: %w", err)
	}

	log.Printf("[COMPUTE] Found %d clubs for %s period", len(clubPoints), periodType)

	// Compute global ranks
	for globalRank, cp := range clubPoints {
		rank := globalRank + 1
		snapshot := models.ClubLeaderboardSnapshot{
			ClubID:            cp.ClubID,
			TotalPoints:       cp.TotalPoints,
			ActiveMemberCount: cp.ActiveMemberCount,
			PeriodType:        periodType,
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			GlobalRank:        &rank,
			City:              cp.City,
		}

		// Upsert snapshot
		if err := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "club_id"}, {Name: "period_type"}, {Name: "period_start"}},
			DoUpdates: clause.AssignmentColumns([]string{"total_points", "active_member_count", "global_rank", "city", "period_end", "updated_at"}),
		}).Create(&snapshot).Error; err != nil {
			return fmt.Errorf("failed to save club leaderboard snapshot: %w", err)
		}
	}

	// Compute local/city ranks
	cities := make(map[string]bool)
	for _, cp := range clubPoints {
		if cp.City != "" {
			cities[cp.City] = true
		}
	}

	for city := range cities {
		var cityClubs []models.ClubLeaderboardSnapshot
		if err := s.db.Where("city = ? AND period_type = ? AND period_start = ?", city, periodType, periodStart).
			Order("total_points DESC").
			Find(&cityClubs).Error; err != nil {
			return fmt.Errorf("failed to get city clubs: %w", err)
		}

		for localRank, club := range cityClubs {
			rank := localRank + 1
			if err := s.db.Model(&models.ClubLeaderboardSnapshot{}).
				Where("id = ?", club.ID).
				Update("local_rank", rank).Error; err != nil {
				return fmt.Errorf("failed to update local rank: %w", err)
			}
		}

		log.Printf("[COMPUTE] Computed local ranks for %s: %d clubs", city, len(cityClubs))
	}

	return nil
}

func (s *leaderboardComputeService) computeUserClubLeaderboard(periodType string, periodStart, periodEnd time.Time, windowField string) error {
	// Get all clubs
	var clubs []models.Club
	if err := s.db.Find(&clubs).Error; err != nil {
		return fmt.Errorf("failed to get clubs: %w", err)
	}

	for _, club := range clubs {
		type UserPoints struct {
			UserID uint
			Points int
		}

		var userContributions []UserPoints
		query := fmt.Sprintf(`
			SELECT
				user_id,
				COALESCE(SUM(points), 0) as points
			FROM user_points_windowed
			WHERE club_id = ? AND %s = ?
			GROUP BY user_id
			ORDER BY points DESC
		`, windowField)

		if err := s.db.Raw(query, club.ID, periodStart).Scan(&userContributions).Error; err != nil {
			return fmt.Errorf("failed to aggregate user contributions for club %d: %w", club.ID, err)
		}

		// Save snapshots with ranks
		for rank, uc := range userContributions {
			clubRank := rank + 1
			snapshot := models.UserClubLeaderboardSnapshot{
				UserID:            uc.UserID,
				ClubID:            club.ID,
				PointsContributed: uc.Points,
				PeriodType:        periodType,
				PeriodStart:       periodStart,
				PeriodEnd:         periodEnd,
				ClubRank:          &clubRank,
			}

			// Upsert snapshot
			if err := s.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "club_id"}, {Name: "period_type"}, {Name: "period_start"}},
				DoUpdates: clause.AssignmentColumns([]string{"points_contributed", "club_rank", "period_end", "updated_at"}),
			}).Create(&snapshot).Error; err != nil {
				return fmt.Errorf("failed to save user club leaderboard snapshot: %w", err)
			}
		}

		if len(userContributions) > 0 {
			log.Printf("[COMPUTE] Computed %d user rankings for club %d (%s)", len(userContributions), club.ID, periodType)
		}
	}

	return nil
}

// Time helper functions
func getWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	daysToSubtract := int(weekday) - int(time.Monday)
	weekStart := t.AddDate(0, 0, -daysToSubtract)
	return time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
}

func getWeekEnd(t time.Time) time.Time {
	weekStart := getWeekStart(t)
	return weekStart.AddDate(0, 0, 7).Add(-time.Second)
}

func getMonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func getMonthEnd(t time.Time) time.Time {
	return getMonthStart(t).AddDate(0, 1, 0).Add(-time.Second)
}

func getYearStart(t time.Time) time.Time {
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
}

func getYearEnd(t time.Time) time.Time {
	return getYearStart(t).AddDate(1, 0, 0).Add(-time.Second)
}
