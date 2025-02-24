package serviceimpl

import (
	"database/sql"
	"fmt"
	"github.com/PayRam/go-referral/request"
	"github.com/PayRam/go-referral/response"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
	"time"
)

type aggregatorService struct {
	DB *gorm.DB
}

// var _ service.aggregatorService = &aggregatorService{}

func NewAggregatorService(db *gorm.DB) *aggregatorService {
	return &aggregatorService{DB: db}
}

func (s *aggregatorService) GetReferrerMembersStats(req request.GetMemberRequest) ([]response.ReferrerStats, int64, error) {
	var result []response.ReferrerStats
	var totalCount int64

	// Build base query for referrers
	query := s.DB.Table("referral_members").
		Select(`
			referral_members.id AS id,
			referral_members.project AS project,
			referral_members.email AS email,
			referral_members.reference_id AS reference_id,
			referral_members.code AS code,
			COUNT(DISTINCT rr.id) AS referee_count,
			COALESCE(CAST(SUM(re.amount) AS TEXT), '0') AS total_rewards,
			CASE 
				WHEN referral_members.referred_by_member_id IS NOT NULL AND referral_members.referred_by_member_id > 0 
				THEN TRUE 
				ELSE FALSE 
			END AS is_referred,
			referral_members.created_at AS created_at,
			referral_members.updated_at AS updated_at,
			COALESCE(CAST(referral_members.deleted_at AS TEXT), '') AS deleted_at 
		`).
		Joins(`
			LEFT JOIN referral_members rr ON referral_members.id = rr.referred_by_member_id AND referral_members.project = rr.project
		`).
		Joins(`
			LEFT JOIN referral_rewards re ON referral_members.id = re.rewarded_member_id AND referral_members.project = re.project
		`)

	// Apply campaign IDs filter if provided
	if req.CampaignIDs != nil && len(req.CampaignIDs) > 0 {
		query = query.Joins(`
			JOIN referral_members_campaigns rc ON rc.member_id = referral_members.id AND rc.project = referral_members.project
		`).Where("rc.campaign_id IN (?)", req.CampaignIDs)
	}

	// **Fix Grouping Issues**
	query = query.Group(`
		referral_members.id, referral_members.project, referral_members.email, referral_members.reference_id,
		referral_members.code, referral_members.created_at, referral_members.updated_at, referral_members.deleted_at
	`)

	// Apply filters
	query = request.ApplyGetMemberRequest(req, query)

	// **Fix Count Query to Avoid Pagination**
	countQuery := s.DB.Table("(?) AS subquery", query).Select("COUNT(*)")
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count referrer stats: %w", err)
	}

	// Apply pagination after counting
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Execute the query and scan results
	rows, err := query.Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch referrers with stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var referrer response.ReferrerStats
		var totalRewardsStr string
		var email sql.NullString
		var deletedAt sql.NullString // ✅ Handling possible NULLs

		err := rows.Scan(
			&referrer.ID, &referrer.Project, &email, &referrer.ReferenceID, &referrer.Code,
			&referrer.RefereeCount, &totalRewardsStr, &referrer.IsReferred,
			&referrer.CreatedAt, &referrer.UpdatedAt, &deletedAt, // ✅ Added deletedAt
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan referrer stats: %w", err)
		}

		// Convert email NULL handling
		if email.Valid {
			referrer.Email = &email.String
		} else {
			referrer.Email = nil
		}

		// Convert total_rewards from string to decimal.Decimal
		totalRewards, convErr := decimal.NewFromString(totalRewardsStr)
		if convErr != nil {
			return nil, 0, fmt.Errorf("failed to parse total_rewards: %w", convErr)
		}

		referrer.TotalRewards = totalRewards
		result = append(result, referrer)
	}

	return result, totalCount, nil
}

func (s *aggregatorService) GetRewardsStats(req request.GetRewardRequest) ([]response.RewardStats, error) {
	var results []response.RewardStats

	// Parse timezone-aware timestamps if needed
	parseTimestamp := func(ts string) (*time.Time, error) {
		if ts == "" {
			return nil, nil
		}

		// Replace space with 'T' to match RFC3339Nano format
		ts = strings.Replace(ts, " ", "T", 1)

		parsed, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}
		return &parsed, nil
	}

	// Handle date range logic
	if req.PaginationConditions.StartDate == nil || req.PaginationConditions.EndDate == nil {
		var dateRangeStartStr, dateRangeEndStr string

		// Fetch the earliest and latest created_at values from the database
		if err := s.DB.Table("referral_rewards").
			Select("COALESCE(MIN(created_at), '')").
			Row().Scan(&dateRangeStartStr); err != nil {
			return nil, fmt.Errorf("failed to fetch earliest created_at date: %w", err)
		}

		if err := s.DB.Table("referral_rewards").
			Select("COALESCE(MAX(created_at), '')").
			Row().Scan(&dateRangeEndStr); err != nil {
			return nil, fmt.Errorf("failed to fetch latest created_at date: %w", err)
		}

		// Handle case when no records exist
		if dateRangeStartStr == "" || dateRangeEndStr == "" {
			return []response.RewardStats{}, nil // Return an empty result
		}

		// Parse the fetched dates
		if req.PaginationConditions.StartDate == nil {
			parsed, err := parseTimestamp(dateRangeStartStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse earliest created_at date: %w", err)
			}
			req.PaginationConditions.StartDate = parsed
		}
		if req.PaginationConditions.EndDate == nil {
			parsed, err := parseTimestamp(dateRangeEndStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse latest created_at date: %w", err)
			}
			req.PaginationConditions.EndDate = parsed
		}
	}

	// Construct the query
	query := s.DB.Table("referral_rewards").
		Select(`
			CASE 
				WHEN 33 THEN printf('%s %d', 
					substr('JanFebMarAprMayJunJulAugSepOctNovDec', (strftime('%m', created_at) - 1) * 3 + 1, 3),
					CAST(strftime('%d', created_at) AS INTEGER))
				WHEN 190 THEN printf('%s %d',
					substr('JanFebMarAprMayJunJulAugSepOctNovDec',
						(strftime('%m', created_at, 'weekday 1', '-7 days') - 1) * 3 + 1, 3),
					CAST(strftime('%d', created_at, 'weekday 1', '-7 days') AS INTEGER))
				ELSE printf('%s %d',
					substr('JanFebMarAprMayJunJulAugSepOctNovDec', (strftime('%m', created_at) - 1) * 3 + 1, 3),
					cast(strftime('%Y', created_at) as integer))
			END AS date,
			SUM(amount) AS total_rewards,
			COUNT(DISTINCT rewarded_member_reference_id) AS unique_referrers
		`).
		Where(`
			created_at BETWEEN
				COALESCE(?, (SELECT MIN(created_at) FROM referral_rewards)) AND
				COALESCE(?, (SELECT MAX(created_at) FROM referral_rewards))
		`, req.PaginationConditions.StartDate, req.PaginationConditions.EndDate).
		Group("date").
		Order("created_at ASC")

	// Apply filters
	query = request.ApplyGetRewardRequest(req, query)

	// Execute the query
	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch rewards stats: %w", err)
	}

	return results, nil
}
