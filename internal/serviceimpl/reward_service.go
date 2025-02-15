package serviceimpl

import (
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type rewardService struct {
	DB *gorm.DB
}

//var _ service.rewardService = &rewardService{}

func NewRewardService(db *gorm.DB) *rewardService {
	return &rewardService{DB: db}
}

func (s *rewardService) GetTotalRewards(req request.GetRewardRequest) (decimal.Decimal, error) {
	var totalAmount decimal.Decimal

	// Build the query
	query := s.DB.Model(&models.Reward{}).Select("COALESCE(SUM(amount), 0) AS total")

	// Apply filters
	query = request.ApplyGetRewardRequest(req, query)

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Calculate the sum
	if err := query.Scan(&totalAmount).Error; err != nil {
		return decimal.Zero, fmt.Errorf("failed to calculate total rewards: %w", err)
	}

	// Ensure totalAmount is set to zero if no records are found
	if totalAmount.IsZero() {
		return decimal.Zero, nil
	}

	return totalAmount.Round(6), nil
}

// GetRewards fetches rewards based on the provided request
func (s *rewardService) GetRewards(req request.GetRewardRequest) ([]models.Reward, int64, error) {
	var rewards []models.Reward
	var count int64

	// Start query
	query := s.DB.Model(&models.Reward{})

	// Apply filters
	query = request.ApplyGetRewardRequest(req, query)

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count rewards: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Preload("RewardedMember").Preload("RelatedMember").Find(&rewards).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch rewards: %w", err)
	}

	return rewards, count, nil
}

func (s *rewardService) GetNewReferrerCount(req request.GetRewardRequest) (int64, error) {
	var count int64

	// Fetch start and end dates if not provided
	if req.PaginationConditions.StartDate == nil || req.PaginationConditions.EndDate == nil {
		var dateRangeStartStr, dateRangeEndStr string

		// Fetch the earliest and latest created_at values from the database
		if err := s.DB.Table("referral_rewards").Select("MIN(created_at)").Row().Scan(&dateRangeStartStr); err != nil {
			return 0, fmt.Errorf("failed to fetch earliest created_at date: %w", err)
		}
		if err := s.DB.Table("referral_rewards").Select("MAX(created_at)").Row().Scan(&dateRangeEndStr); err != nil {
			return 0, fmt.Errorf("failed to fetch latest created_at date: %w", err)
		}

		parseTimestamp := func(ts string) (*time.Time, error) {
			parsed, err := time.Parse("2006-01-02 15:04:05-07:00", ts)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
			return &parsed, nil
		}

		if req.PaginationConditions.StartDate == nil {
			parsed, err := parseTimestamp(dateRangeStartStr)
			if err != nil {
				return 0, fmt.Errorf("failed to parse earliest created_at date: %w", err)
			}
			req.PaginationConditions.StartDate = parsed
		}
		if req.PaginationConditions.EndDate == nil {
			parsed, err := parseTimestamp(dateRangeEndStr)
			if err != nil {
				return 0, fmt.Errorf("failed to parse latest created_at date: %w", err)
			}
			req.PaginationConditions.EndDate = parsed
		}
	}

	// Query to find unique ReferredByMemberReferenceID within the provided date range
	subQuery := s.DB.Table("referral_rewards").
		Select("referred_by_member_reference_id").
		Where("created_at < ?", req.PaginationConditions.StartDate)

	query := s.DB.Table("referral_rewards r").
		Distinct("r.referred_by_member_reference_id").
		Where("r.created_at BETWEEN ? AND ?", req.PaginationConditions.StartDate, req.PaginationConditions.EndDate).
		Where("r.referred_by_member_reference_id NOT IN (?)", subQuery)

	query = request.ApplyGetRewardRequest(req, query)

	// Execute the query to count distinct new referred_by_member_reference_id
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count new referred_by_member_reference_id: %w", err)
	}

	return count, nil
}

func (s *rewardService) GetNewRefereeCount(req request.GetRewardRequest) (int64, error) {
	var count int64

	// Fetch start and end dates if not provided
	if req.PaginationConditions.StartDate == nil || req.PaginationConditions.EndDate == nil {
		var dateRangeStartStr, dateRangeEndStr string

		// Fetch the earliest and latest created_at values from the database
		if err := s.DB.Table("referral_rewards").Select("MIN(created_at)").Row().Scan(&dateRangeStartStr); err != nil {
			return 0, fmt.Errorf("failed to fetch earliest created_at date: %w", err)
		}
		if err := s.DB.Table("referral_rewards").Select("MAX(created_at)").Row().Scan(&dateRangeEndStr); err != nil {
			return 0, fmt.Errorf("failed to fetch latest created_at date: %w", err)
		}

		parseTimestamp := func(ts string) (*time.Time, error) {
			parsed, err := time.Parse("2006-01-02 15:04:05-07:00", ts)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
			return &parsed, nil
		}

		if req.PaginationConditions.StartDate == nil {
			parsed, err := parseTimestamp(dateRangeStartStr)
			if err != nil {
				return 0, fmt.Errorf("failed to parse earliest created_at date: %w", err)
			}
			req.PaginationConditions.StartDate = parsed
		}
		if req.PaginationConditions.EndDate == nil {
			parsed, err := parseTimestamp(dateRangeEndStr)
			if err != nil {
				return 0, fmt.Errorf("failed to parse latest created_at date: %w", err)
			}
			req.PaginationConditions.EndDate = parsed
		}
	}

	// Query to find unique RefereeMemberReferenceID within the provided date range
	subQuery := s.DB.Table("referral_rewards").
		Select("referee_member_reference_id").
		Where("created_at < ?", req.PaginationConditions.StartDate)

	query := s.DB.Table("referral_rewards r").
		Distinct("r.referee_member_reference_id").
		Where("r.created_at BETWEEN ? AND ?", req.PaginationConditions.StartDate, req.PaginationConditions.EndDate).
		Where("r.referee_member_reference_id NOT IN (?)", subQuery)

	// Apply filters
	query = request.ApplyGetRewardRequest(req, query)

	// Execute the query to count distinct new referee_reference_ids
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count new referee_member_reference_id: %w", err)
	}

	return count, nil
}
