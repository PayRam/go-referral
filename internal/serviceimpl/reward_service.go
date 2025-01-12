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
	if request.Project != nil {
		query = query.Where("project = ?", *request.Project)
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
