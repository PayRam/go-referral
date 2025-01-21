package serviceimpl

import (
	"fmt"
	"github.com/PayRam/go-referral/request"
	"github.com/PayRam/go-referral/response"
	"gorm.io/gorm"
	"time"
)

type aggregatorService struct {
	DB *gorm.DB
}

// var _ service.aggregatorService = &aggregatorService{}

func NewAggregatorService(db *gorm.DB) *aggregatorService {
	return &aggregatorService{DB: db}
}

func (s *aggregatorService) GetReferrersWithStats(req request.GetReferrerRequest) ([]response.ReferrerStats, int64, error) {
	var result []response.ReferrerStats
	var totalCount int64

	// Build base query for referrers
	query := s.DB.Table("referral_referrer r").
		Select(`
			r.id AS id,
			r.project AS project,
			r.email AS email,
			r.reference_id AS reference_id,
			r.code AS code,
			COUNT(DISTINCT rr.id) AS referee_count,
			COALESCE(SUM(re.amount), 0) AS total_rewards,
			r.created_at AS created_at,
			r.updated_at AS updated_at,
			r.deleted_at AS deleted_at
		`).
		Joins(`
			LEFT JOIN referral_referee rr ON r.id = rr.referrer_id AND r.project = rr.project
		`).
		Joins(`
			LEFT JOIN referral_rewards re ON r.id = re.referrer_id AND r.project = re.project
		`)

	// Apply campaign IDs filter if provided
	if req.CampaignIDs != nil && len(req.CampaignIDs) > 0 {
		query = query.Joins(`
			JOIN referral_referrer_campaigns rc ON rc.referrer_id = r.id AND rc.project = r.project
		`).Where("rc.campaign_id IN (?)", req.CampaignIDs)
	}

	// Group the results to avoid duplicates
	query = query.Group("r.id, r.project, r.email, r.reference_id, r.code, r.created_at, r.updated_at, r.deleted_at")

	// Apply filters from request
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("r.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("r.id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("r.reference_id = ?", *req.ReferenceID)
	}
	if req.Email != nil {
		query = query.Where("email = ?", *req.Email)
	}
	if req.Code != nil {
		query = query.Where("r.code = ?", *req.Code)
	}

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Execute the query
	if err := query.Scan(&result).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch referrers with stats: %w", err)
	}

	return result, totalCount, nil
}

func (s *aggregatorService) GetRewardsStats(req request.GetRewardRequest) ([]response.RewardStats, error) {
	var results []response.RewardStats

	// Parse timezone-aware timestamps if needed
	parseTimestamp := func(ts string) (*time.Time, error) {
		parsed, err := time.Parse("2006-01-02 15:04:05-07:00", ts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}
		return &parsed, nil
	}

	// Handle date range logic
	if req.PaginationConditions.CreatedAfter == nil || req.PaginationConditions.CreatedBefore == nil {
		var dateRangeStartStr, dateRangeEndStr string

		// Fetch the earliest and latest created_at values from the database
		if err := s.DB.Table("referral_rewards").Select("MIN(created_at)").Row().Scan(&dateRangeStartStr); err != nil {
			return nil, fmt.Errorf("failed to fetch earliest created_at date: %w", err)
		}
		if err := s.DB.Table("referral_rewards").Select("MAX(created_at)").Row().Scan(&dateRangeEndStr); err != nil {
			return nil, fmt.Errorf("failed to fetch latest created_at date: %w", err)
		}

		if req.PaginationConditions.CreatedAfter == nil {
			parsed, err := parseTimestamp(dateRangeStartStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse earliest created_at date: %w", err)
			}
			req.PaginationConditions.CreatedAfter = parsed
		}
		if req.PaginationConditions.CreatedBefore == nil {
			parsed, err := parseTimestamp(dateRangeEndStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse latest created_at date: %w", err)
			}
			req.PaginationConditions.CreatedBefore = parsed
		}
	}

	// Construct the query
	query := s.DB.Table("referral_rewards r").
		Select(`
			CASE 
				WHEN 33 THEN printf('%s %d', 
					substr('JanFebMarAprMayJunJulAugSepOctNovDec', (strftime('%m', r.created_at) - 1) * 3 + 1, 3),
					CAST(strftime('%d', r.created_at) AS INTEGER))
				WHEN 190 THEN printf('%s %d',
					substr('JanFebMarAprMayJunJulAugSepOctNovDec',
						(strftime('%m', r.created_at, 'weekday 1', '-7 days') - 1) * 3 + 1, 3),
					CAST(strftime('%d', r.created_at, 'weekday 1', '-7 days') AS INTEGER))
				ELSE printf('%s %d',
					substr('JanFebMarAprMayJunJulAugSepOctNovDec', (strftime('%m', r.created_at) - 1) * 3 + 1, 3),
					cast(strftime('%Y', r.created_at) as integer))
			END AS date,
			SUM(r.amount) AS total_rewards,
			COUNT(DISTINCT r.referrer_reference_id) AS unique_referrers
		`).
		Where(`
			r.created_at BETWEEN
				COALESCE(?, (SELECT MIN(created_at) FROM referral_rewards)) AND
				COALESCE(?, (SELECT MAX(created_at) FROM referral_rewards))
		`, req.PaginationConditions.CreatedAfter, req.PaginationConditions.CreatedBefore).
		Group("date").
		Order("r.created_at ASC")

	// Apply filters
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("r.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("r.id = ?", *req.ID)
	}
	if req.RefereeID != nil {
		query = query.Where("r.referee_id = ?", *req.RefereeID)
	}
	if req.ReferrerID != nil {
		query = query.Where("r.referrer_id = ?", *req.ReferrerID)
	}
	if req.ReferrerReferenceID != nil {
		query = query.Where("r.referrer_reference_id = ?", *req.ReferrerReferenceID)
	}
	if req.ReferrerCode != nil {
		query = query.Where("r.referrer_code = ?", *req.ReferrerCode)
	}
	if req.CampaignIDs != nil && len(req.CampaignIDs) > 0 {
		query = query.Where("r.campaign_id IN (?)", req.CampaignIDs)
	}
	if req.Status != nil {
		query = query.Where("r.status = ?", *req.Status)
	}

	// Execute the query
	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch rewards stats: %w", err)
	}

	return results, nil
}
