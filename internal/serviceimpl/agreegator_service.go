package serviceimpl

import (
	"fmt"
	"github.com/PayRam/go-referral/request"
	"github.com/PayRam/go-referral/response"
	"gorm.io/gorm"
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
		`).
		Group("r.id, r.project, r.reference_id, r.code, r.created_at, r.updated_at, r.deleted_at")

	// Apply filters from request
	if req.Project != nil {
		query = query.Where("r.project = ?", *req.Project)
	}
	if req.ID != nil {
		query = query.Where("r.id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("r.reference_id = ?", *req.ReferenceID)
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
