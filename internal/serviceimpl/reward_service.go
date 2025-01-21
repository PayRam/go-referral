package serviceimpl

import (
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type rewardService struct {
	DB *gorm.DB
}

//var _ service.rewardService = &rewardService{}

func NewRewardService(db *gorm.DB) *rewardService {
	return &rewardService{DB: db}
}

func (s *rewardService) GetTotalRewards(request request.GetRewardRequest) (decimal.Decimal, error) {
	var totalAmount decimal.Decimal

	// Build the query
	query := s.DB.Model(&models.Reward{}).Select("SUM(amount)")

	// Apply filters
	if request.Projects != nil && len(request.Projects) > 0 {
		query = query.Where("project IN (?)", request.Projects)
	}
	if request.ID != nil {
		query = query.Where("id = ?", *request.ID)
	}
	if request.CampaignID != nil {
		query = query.Where("campaign_id = ?", *request.CampaignID)
	}
	if request.RefereeID != nil {
		query = query.Where("referee_id = ?", *request.RefereeID)
	}
	if request.RefereeReferenceID != nil {
		query = query.Where("referee_reference_id = ?", *request.RefereeReferenceID)
	}
	if request.ReferrerID != nil {
		query = query.Where("referrer_id = ?", *request.ReferrerID)
	}
	if request.ReferrerReferenceID != nil {
		query = query.Where("referrer_reference_id = ?", *request.ReferrerReferenceID)
	}
	if request.ReferrerCode != nil {
		query = query.Where("referrer_code = ?", *request.ReferrerCode)
	}
	if request.Status != nil {
		query = query.Where("status = ?", *request.Status)
	}

	// Calculate the sum
	if err := query.Scan(&totalAmount).Error; err != nil {
		return decimal.Zero, fmt.Errorf("failed to calculate total rewards: %w", err)
	}

	return totalAmount, nil
}

// GetRewards fetches rewards based on the provided request
func (s *rewardService) GetRewards(req request.GetRewardRequest) ([]models.Reward, int64, error) {
	var rewards []models.Reward
	var count int64

	// Start query
	query := s.DB.Model(&models.Reward{})

	// Apply filters
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.CampaignID != nil {
		query = query.Where("campaign_id = ?", *req.CampaignID)
	}
	if req.RefereeID != nil {
		query = query.Where("referee_id = ?", *req.RefereeID)
	}
	if req.RefereeReferenceID != nil {
		query = query.Where("referee_reference_id = ?", *req.RefereeReferenceID)
	}
	if req.ReferrerID != nil {
		query = query.Where("referrer_id = ?", *req.ReferrerID)
	}
	if req.ReferrerReferenceID != nil {
		query = query.Where("referrer_reference_id = ?", *req.ReferrerReferenceID)
	}
	if req.ReferrerCode != nil {
		query = query.Where("referrer_code = ?", *req.ReferrerCode)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count rewards: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Preload("EventLogs").Find(&rewards).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch rewards: %w", err)
	}

	return rewards, count, nil
}
