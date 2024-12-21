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
	}

	// Wrap in a transaction to ensure atomicity
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Create the campaign
		if err := tx.Create(campaign).Error; err != nil {
			return err
		}

		// Add events to the campaign
		for _, event := range events {
			if err := tx.Create(&param.CampaignEvent{
				CampaignID: campaign.ID,
				EventID:    event.ID,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
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

func (s *campaignService) UpdateCampaignEvents(campaignID uint, events []param.Event) error {
	// Wrap in a transaction to ensure atomicity
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// Remove existing event associations
		if err := tx.Where("campaign_id = ?", campaignID).Delete(&param.CampaignEvent{}).Error; err != nil {
			return err
		}

		// Add new event associations
		for _, event := range events {
			if err := tx.Create(&param.CampaignEvent{
				CampaignID: campaignID,
				EventID:    event.ID,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *campaignService) SetDefaultCampaign(campaignID uint) error {
	var existingDefaultCampaign param.Campaign
	if err := s.DB.Where("is_default = ?", true).First(&existingDefaultCampaign).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existingDefaultCampaign.ID == campaignID {
		return nil // Campaign is already the default
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&param.Campaign{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&param.Campaign{}).Where("id = ?", campaignID).Update("is_default", true).Error; err != nil {
			return err
		}
		return nil
	})
}
