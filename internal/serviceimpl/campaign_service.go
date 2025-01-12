package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type campaignService struct {
	DB *gorm.DB
}

//var _ service.campaignService = &campaignService{}

func NewCampaignService(db *gorm.DB) *campaignService {
	return &campaignService{DB: db}
}

// CreateCampaign creates a new campaign
func (s *campaignService) CreateCampaign(project, name, description string, startDate, endDate time.Time, events []models.Event, rewardType *string, rewardValue *float64, maxOccurrences *uint, validityDays *uint, budget *decimal.Decimal) (*models.Campaign, error) {
	// Validate start and end dates
	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	// Validate budget (optional)
	if budget != nil && budget.IsNegative() {
		return nil, errors.New("budget cannot be negative")
	}

	// Validate reward value
	if rewardValue != nil && *rewardValue <= 0 {
		return nil, errors.New("reward value must be greater than zero")
	}

	// Ensure at most one event with EventType = "payment"
	paymentCount := 0
	for _, event := range events {
		if event.EventType == "payment" {
			paymentCount++
		}
		if paymentCount > 1 {
			return nil, errors.New("only one event with event type 'payment' is allowed")
		}
	}

	// Create the campaign object
	campaign := &models.Campaign{
		Project:        project,
		Name:           name,
		Description:    description,
		StartDate:      startDate,
		EndDate:        endDate,
		IsActive:       true,
		IsDefault:      true,
		RewardType:     rewardType,
		RewardValue:    rewardValue,
		MaxOccurrences: maxOccurrences,
		ValidityDays:   validityDays,
		Budget:         budget,
	}

	if events != nil && len(events) > 0 {

		// Wrap in a transaction to ensure atomicity
		if err := s.DB.Transaction(func(tx *gorm.DB) error {
			// Save the campaign
			if err := tx.Create(campaign).Error; err != nil {
				return err
			}

			// Link the events to the campaign
			for _, event := range events {
				if err := tx.Create(&models.CampaignEvent{
					CampaignID: campaign.ID,
					EventKey:   event.Key,
				}).Error; err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			return nil, err
		}
	} else {
		if err := s.DB.Create(campaign).Error; err != nil {
			return nil, err
		}
	}

	// Reload the campaign with associated events
	if err := s.DB.Preload("Events").First(campaign, campaign.ID).Error; err != nil {
		return nil, err
	}

	return campaign, nil
}

// GetCampaigns retrieves campaigns based on dynamic conditions
func (s *campaignService) GetCampaigns(req request.GetCampaignsRequest) ([]models.Campaign, int64, error) {
	var campaigns []models.Campaign
	var count int64

	// Start query
	query := s.DB.Model(&models.Campaign{})

	// Apply filters
	if req.Project != nil {
		query = query.Where("project = ?", *req.Project)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.Name != nil {
		query = query.Where("name LIKE ?", "%"+*req.Name+"%")
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}
	if req.IsDefault != nil {
		query = query.Where("is_default = ?", *req.IsActive)
	}
	if req.StartDateMin != nil {
		query = query.Where("start_date >= ?", *req.StartDateMin)
	}
	if req.StartDateMax != nil {
		query = query.Where("start_date <= ?", *req.StartDateMax)
	}
	if req.EndDateMin != nil {
		query = query.Where("end_date >= ?", *req.EndDateMin)
	}
	if req.EndDateMax != nil {
		query = query.Where("end_date <= ?", *req.EndDateMax)
	}

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Find(&campaigns).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch campaigns: %w", err)
	}

	return campaigns, count, nil
}

// UpdateCampaign updates an existing campaign
func (s *campaignService) UpdateCampaign(project string, id uint, req request.UpdateCampaignRequest) (*models.Campaign, error) {
	var campaign models.Campaign

	// Wrap the operation in a transaction
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the campaign using `id` and `project`
		if err := tx.Where("id = ? AND project = ?", id, project).First(&campaign).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("campaign not found for project %s and id %d: %w", project, id, err)
			}
			return err
		}

		// Prepare the updates
		updates := map[string]interface{}{}

		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.RewardType != nil {
			updates["reward_type"] = *req.RewardType
		}
		if req.RewardValue != nil {
			updates["reward_value"] = *req.RewardValue
		}
		if req.InviteeRewardType != nil {
			updates["invitee_reward_type"] = *req.InviteeRewardType
		}
		if req.InviteeRewardValue != nil {
			updates["invitee_reward_value"] = *req.InviteeRewardValue
		}
		if req.MaxOccurrences != nil {
			updates["max_occurrences"] = *req.MaxOccurrences
		}
		if req.ValidityDays != nil {
			updates["validity_days"] = *req.ValidityDays
		}
		if req.Budget != nil {
			updates["budget"] = *req.Budget
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}
		if req.StartDate != nil {
			updates["start_date"] = *req.StartDate
		}
		if req.EndDate != nil {
			updates["end_date"] = *req.EndDate
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}

		// Validate the date range
		if req.StartDate != nil && req.EndDate != nil {
			if req.StartDate.After(*req.EndDate) {
				return fmt.Errorf("start date cannot be after end date")
			}
		}

		// Apply the updates
		if err := tx.Model(&campaign).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update campaign: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload the campaign with associated events after the transaction
	if err := s.DB.Preload("Events").Where("id = ? AND project = ?", id, project).First(&campaign).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated campaign: %w", err)
	}

	return &campaign, nil
}

func (s *campaignService) UpdateCampaignEvents(project string, campaignID uint, events []models.Event) (*models.Campaign, error) {
	// Validate at most one event with EventType = "payment"
	paymentCount := 0
	for _, event := range events {
		if event.EventType == "payment" {
			paymentCount++
		}
		if paymentCount > 1 {
			return nil, errors.New("only one event with event type 'payment' is allowed")
		}
	}

	var campaign models.Campaign

	// Wrap the operation in a transaction
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the campaign using `campaignID` and `project`
		if err := tx.Where("id = ? AND project = ?", campaignID, project).First(&campaign).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("campaign not found for project %s and id %d", project, campaignID)
			}
			return err
		}

		// Remove existing event associations
		if err := tx.Unscoped().Where("campaign_id = ?", campaignID).Delete(&models.CampaignEvent{}).Error; err != nil {
			return fmt.Errorf("failed to remove existing event associations: %w", err)
		}

		// Add new event associations
		for _, event := range events {
			if err := tx.Create(&models.CampaignEvent{
				CampaignID: campaignID,
				EventKey:   event.Key,
			}).Error; err != nil {
				return fmt.Errorf("failed to associate event %s with campaign %d: %w", event.Key, campaignID, err)
			}
		}

		// Reload the campaign with associated events
		if err := tx.Preload("Events").Where("id = ? AND project = ?", campaignID, project).First(&campaign).Error; err != nil {
			return fmt.Errorf("failed to reload updated campaign: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &campaign, nil
}

func (s *campaignService) SetDefaultCampaign(project string, campaignID uint) (*models.Campaign, error) {
	var existingDefaultCampaign models.Campaign

	// Check if there is already a default campaign for the project
	if err := s.DB.Where("project = ? AND is_default = ?", project, true).First(&existingDefaultCampaign).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to fetch existing default campaign: %w", err)
	}

	// If the requested campaign is already the default, reload it and return
	if existingDefaultCampaign.ID == campaignID {
		if err := s.DB.Preload("Events").First(&existingDefaultCampaign, "project = ? AND id = ?", project, campaignID).Error; err != nil {
			return nil, fmt.Errorf("failed to reload existing default campaign: %w", err)
		}
		return &existingDefaultCampaign, nil
	}

	var updatedCampaign models.Campaign

	// Perform the update in a transaction
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Unset the existing default campaign for the project
		if err := tx.Model(&models.Campaign{}).
			Where("project = ? AND is_default = ?", project, true).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset existing default campaign: %w", err)
		}

		// Set the new campaign as the default
		if err := tx.Model(&models.Campaign{}).
			Where("project = ? AND id = ?", project, campaignID).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("failed to set campaign %d as default: %w", campaignID, err)
		}

		// Reload the updated campaign with its associations
		if err := tx.Preload("Events").First(&updatedCampaign, "project = ? AND id = ?", project, campaignID).Error; err != nil {
			return fmt.Errorf("failed to reload updated campaign: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &updatedCampaign, nil
}

// PauseCampaign updates an existing campaign to set it as inactive
func (s *campaignService) PauseCampaign(project string, campaignID uint) (*models.Campaign, error) {
	var campaign models.Campaign

	// Use a transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the campaign for the given project and ID
		if err := tx.Where("project = ? AND id = ?", project, campaignID).First(&campaign).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("campaign not found for project %s and ID %d: %w", project, campaignID, err)
			}
			return err
		}

		// Check if the campaign is already paused
		if !campaign.IsActive {
			return fmt.Errorf("campaign is already paused")
		}

		// Update the campaign status
		if err := tx.Model(&campaign).Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to update campaign: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload the campaign with associated events
	if err := s.DB.Preload("Events").Where("project = ? AND id = ?", project, campaignID).First(&campaign).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated campaign: %w", err)
	}

	return &campaign, nil
}

// DeleteCampaign performs a soft delete on an existing campaign
func (s *campaignService) DeleteCampaign(project string, campaignID uint) (bool, error) {
	var campaign models.Campaign

	// Use a transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the campaign with the specified project and ID
		if err := tx.Where("project = ? AND id = ?", project, campaignID).First(&campaign).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("campaign not found for project %s and ID %d: %w", project, campaignID, err)
			}
			return err
		}

		// Perform the soft delete
		if err := tx.Delete(&campaign).Error; err != nil {
			return fmt.Errorf("failed to soft delete campaign: %w", err)
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *campaignService) GetTotalCampaigns(req request.GetCampaignsRequest) (int64, error) {
	var count int64

	// Build the query
	query := s.DB.Model(&models.Campaign{})
	// Apply filters
	if req.Project != nil {
		query = query.Where("project = ?", *req.Project)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.Name != nil {
		query = query.Where("name LIKE ?", "%"+*req.Name+"%")
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}
	if req.IsDefault != nil {
		query = query.Where("is_default = ?", *req.IsActive)
	}
	if req.StartDateMin != nil {
		query = query.Where("start_date >= ?", *req.StartDateMin)
	}
	if req.StartDateMax != nil {
		query = query.Where("start_date <= ?", *req.StartDateMax)
	}
	if req.EndDateMin != nil {
		query = query.Where("end_date >= ?", *req.EndDateMin)
	}
	if req.EndDateMax != nil {
		query = query.Where("end_date <= ?", *req.EndDateMax)
	}

	// Count the records
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count referrers: %w", err)
	}

	return count, nil
}
