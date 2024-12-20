package serviceimpl

import (
	"errors"
	"github.com/PayRam/go-referral/service/param"
	"gorm.io/gorm"
	"time"
)

type campaignService struct {
	DB *gorm.DB
}

func NewCampaignService(db *gorm.DB) param.CampaignService {
	return &campaignService{DB: db}
}

// CreateCampaign creates a new campaign
func (s *campaignService) CreateCampaign(name, description string, startDate, endDate time.Time, isActive bool, events []param.Event) (*param.Campaign, error) {
	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	campaign := &param.Campaign{
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    isActive,
		Events:      events,
	}

	if err := s.DB.Create(campaign).Error; err != nil {
		return nil, err
	}
	return campaign, nil
}

// UpdateCampaign updates an existing campaign
func (s *campaignService) UpdateCampaign(id uint, updates map[string]interface{}) (*param.Campaign, error) {
	var campaign param.Campaign
	if err := s.DB.First(&campaign, id).Error; err != nil {
		return nil, err
	}

	// Apply updates
	if err := s.DB.Model(&campaign).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *campaignService) SetDefaultCampaign(campaignID uint) error {
	// Reset existing default campaign
	if err := s.DB.Model(&param.Campaign{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return err
	}
	// Set the new default campaign
	if err := s.DB.Model(&param.Campaign{}).Where("id = ?", campaignID).Update("is_default", true).Error; err != nil {
		return err
	}
	return nil
}
