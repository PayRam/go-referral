package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
func (s *campaignService) CreateCampaign(project string, req request.CreateCampaignRequest) (*models.Campaign, error) {

	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	// Validate required fields
	if req.RewardType != "flat_fee" && req.RewardType != "percentage" {
		return nil, errors.New("rewardType must be either 'flat_fee' or 'percentage'")
	}
	if req.RewardValue.Cmp(decimal.NewFromInt(0)) <= 0 {
		return nil, errors.New("rewardValue must be greater than zero")
	}
	if req.RewardType == "percentage" {
		if req.RewardValue.Cmp(decimal.NewFromInt(100)) > 0 {
			return nil, errors.New("percentage rewardValue must be between 0 and 100")
		}
		// RewardCap can be nil or set
	} else if req.RewardType == "flat_fee" {
		if req.RewardCap != nil {
			return nil, errors.New("rewardCap must be nil for flat_fee rewardType")
		}
	}

	// Validate InviteeRewardType and InviteeRewardValue
	if req.InviteeRewardType != nil || req.InviteeRewardValue != nil {
		if req.InviteeRewardType == nil || req.InviteeRewardValue == nil {
			return nil, errors.New("both inviteeRewardType and inviteeRewardValue must be provided or omitted")
		}
		if *req.InviteeRewardType != "flat_fee" && *req.InviteeRewardType != "percentage" {
			return nil, errors.New("inviteeRewardType must be either 'flat_fee' or 'percentage'")
		}
		if *req.InviteeRewardType == "percentage" {
			//req.Budget.Cmp(decimal.NewFromInt(0)) <= 0
			if req.InviteeRewardValue.Cmp(decimal.NewFromInt(100)) > 0 || req.InviteeRewardValue.Cmp(decimal.NewFromInt(0)) < 0 {
				return nil, errors.New("percentage inviteeRewardValue must be between 0 and 100")
			}
			// InviteeRewardCap can be nil or set
		} else if *req.InviteeRewardType == "flat_fee" {
			if req.InviteeRewardCap != nil {
				return nil, errors.New("inviteeRewardCap must be nil for flat_fee inviteeRewardType")
			}
		}
	}

	// Validate CampaignTypePerCustomer
	switch req.CampaignTypePerCustomer {
	case "one_time", "forever":
		if req.ValidityMonthsPerCustomer != nil || req.MaxOccurrencesPerCustomer != nil || req.RewardCapPerCustomer != nil {
			return nil, errors.New("for 'one_time' or 'forever' CampaignTypePerCustomer, ValidityMonthsPerCustomer, MaxOccurrencesPerCustomer, and RewardCapPerCustomer must be nil")
		}
	case "months_per_customer":
		if req.ValidityMonthsPerCustomer == nil {
			return nil, errors.New("ValidityMonthsPerCustomer is required for 'months_per_customer' CampaignTypePerCustomer")
		}
		if req.MaxOccurrencesPerCustomer != nil || req.RewardCapPerCustomer != nil {
			return nil, errors.New("for 'months_per_customer' CampaignTypePerCustomer, MaxOccurrencesPerCustomer and RewardCapPerCustomer must be nil")
		}
	case "count_per_customer":
		if req.MaxOccurrencesPerCustomer == nil {
			return nil, errors.New("MaxOccurrencesPerCustomer is required for 'count_per_customer' CampaignTypePerCustomer")
		}
		if req.ValidityMonthsPerCustomer != nil || req.RewardCapPerCustomer != nil {
			return nil, errors.New("for 'count_per_customer' CampaignTypePerCustomer, ValidityMonthsPerCustomer and RewardCapPerCustomer must be nil")
		}
	default:
		return nil, errors.New("invalid CampaignTypePerCustomer; must be 'one_time', 'forever', 'months_per_customer', or 'count_per_customer'")
	}

	if req.Budget != nil && req.Budget.Cmp(decimal.NewFromInt(0)) <= 0 {
		return nil, errors.New("budget must be greater than zero")
	}

	if req.StartDate != nil && req.EndDate == nil {
		return nil, errors.New("end date is required if start date is provided")
	}

	if req.EndDate != nil && req.StartDate == nil {
		return nil, errors.New("start date is required if end date is provided")
	}

	if req.StartDate != nil && req.EndDate != nil {
		if req.StartDate.After(*req.EndDate) {
			return nil, errors.New("start date cannot be after end date")
		}
		if req.EndDate.Before(time.Now()) {
			return nil, errors.New("end date cannot be in the past")
		}
	}

	if req.EventKeys == nil || len(req.EventKeys) == 0 {
		return nil, errors.New("eventKeys must be provided")
	}

	// fetch events using event keys
	var events []models.Event

	// Fetch the events by keys
	if err := s.DB.Where("project = ? AND key IN ?", project, req.EventKeys).Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events for keys %v: %w", req.EventKeys, err)
	}

	if len(events) != len(req.EventKeys) {
		return nil, errors.New("not all event keys were found")
	}

	paymentCount := 0
	for _, event := range events {
		if event.EventType == "payment" {
			paymentCount++
		}
		if paymentCount > 1 {
			return nil, errors.New("only one event with event type 'payment' is allowed")
		}
	}

	if req.CampaignTypePerCustomer != "one_time" && len(events) > 1 {
		return nil, errors.New("only one event is allowed for campaigns with other than 'one_time' campaign type")
	}

	if req.RewardType == "percentage" && paymentCount == 0 {
		return nil, errors.New("at least one event with event type 'payment' is required for campaigns with 'percentage' reward type")
	}

	if req.RewardType == "flat_fee" && paymentCount > 0 {
		return nil, errors.New("no event with event type 'payment' is allowed for campaigns with 'flat_fee' reward type")
	}

	// Create the campaign object
	campaign := &models.Campaign{
		Project:                   project,
		Name:                      req.Name,
		Description:               req.Description,
		StartDate:                 req.StartDate,
		EndDate:                   req.EndDate,
		IsDefault:                 req.IsDefault,
		RewardType:                req.RewardType,
		RewardValue:               req.RewardValue,
		RewardCap:                 req.RewardCap,
		Budget:                    req.Budget,
		InviteeRewardType:         req.InviteeRewardType,
		InviteeRewardValue:        req.InviteeRewardValue,
		InviteeRewardCap:          req.InviteeRewardCap,
		CampaignTypePerCustomer:   req.CampaignTypePerCustomer,
		ValidityMonthsPerCustomer: req.ValidityMonthsPerCustomer,
		MaxOccurrencesPerCustomer: req.MaxOccurrencesPerCustomer,
		RewardCapPerCustomer:      req.RewardCapPerCustomer,
		Status:                    "active",
	}

	// Wrap the operation in a transaction
	if err := s.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(campaign).Error; err != nil {
			return fmt.Errorf("failed to create campaign: %w", err)
		}

		// Add new event associations
		for _, event := range events {
			if err := tx.Create(&models.CampaignEvent{
				Project:    project,
				CampaignID: campaign.ID,
				EventID:    event.ID,
				EventKey:   event.Key,
			}).Error; err != nil {
				return fmt.Errorf("failed to associate event %s with campaign %d: %w", event.Key, campaign.ID, err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Reload the campaign with associated events after the transaction
	if err := s.DB.Preload("Events").Where("id = ? AND project = ?", campaign.ID, project).First(&campaign).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated campaign: %w", err)
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
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.IsDefault != nil {
		query = query.Where("is_default = ?", *req.IsDefault)
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

	if req.Name != nil && *req.Name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if req.Status != nil && *req.Status != "active" && *req.Status != "paused" && *req.Status != "archived" {
		return nil, errors.New("status must be either 'active', 'paused', or 'archived'")
	}

	if req.RewardType != nil && *req.RewardType != "flat_fee" && *req.RewardType != "percentage" {
		return nil, errors.New("rewardType must be either 'flat_fee' or 'percentage'")
	}

	if req.RewardValue != nil && req.RewardValue.Cmp(decimal.NewFromInt(0)) <= 0 {
		return nil, errors.New("rewardValue must be greater than zero")
	}

	if req.RewardType != nil && *req.RewardType == "percentage" {
		if req.RewardValue != nil && req.RewardValue.Cmp(decimal.NewFromInt(100)) > 0 {
			return nil, errors.New("percentage rewardValue must be between 0 and 100")
		}
		// RewardCap can be nil or set
	} else if req.RewardType != nil && *req.RewardType == "flat_fee" {
		if req.RewardCap != nil {
			return nil, errors.New("rewardCap must be nil for flat_fee rewardType")
		}
	}

	// Validate InviteeRewardType and InviteeRewardValue
	if req.InviteeRewardType != nil || req.InviteeRewardValue != nil {
		if req.InviteeRewardType == nil || req.InviteeRewardValue == nil {
			return nil, errors.New("both inviteeRewardType and inviteeRewardValue must be provided or omitted")
		}
		if *req.InviteeRewardType != "flat_fee" && *req.InviteeRewardType != "percentage" {
			return nil, errors.New("inviteeRewardType must be either 'flat_fee' or 'percentage'")
		}
		if *req.InviteeRewardType == "percentage" {
			if req.InviteeRewardValue.Cmp(decimal.NewFromInt(100)) > 0 || req.InviteeRewardValue.Cmp(decimal.NewFromInt(0)) < 0 {
				return nil, errors.New("percentage inviteeRewardValue must be between 0 and 100")
			}
			// InviteeRewardCap can be nil or set
		} else if *req.InviteeRewardType == "flat_fee" {
			if req.InviteeRewardCap != nil {
				return nil, errors.New("inviteeRewardCap must be nil for flat_fee inviteeRewardType")
			}
		}
	}

	// Validate CampaignTypePerCustomer
	if req.CampaignTypePerCustomer != nil {
		switch *req.CampaignTypePerCustomer {
		case "one_time", "forever":
			if req.ValidityMonthsPerCustomer != nil || req.MaxOccurrencesPerCustomer != nil || req.RewardCapPerCustomer != nil {
				return nil, errors.New("for 'one_time' or 'forever' CampaignTypePerCustomer, ValidityMonthsPerCustomer, MaxOccurrencesPerCustomer, and RewardCapPerCustomer must be nil")
			}
		case "months_per_customer":
			if req.ValidityMonthsPerCustomer == nil {
				return nil, errors.New("ValidityMonthsPerCustomer is required for 'months_per_customer' CampaignTypePerCustomer")
			}
			if req.MaxOccurrencesPerCustomer != nil || req.RewardCapPerCustomer != nil {
				return nil, errors.New("for 'months_per_customer' CampaignTypePerCustomer, MaxOccurrencesPerCustomer and RewardCapPerCustomer must be nil")
			}
		case "count_per_customer":
			if req.MaxOccurrencesPerCustomer == nil {
				return nil, errors.New("MaxOccurrencesPerCustomer is required for 'count_per_customer' CampaignTypePerCustomer")
			}
			if req.ValidityMonthsPerCustomer != nil || req.RewardCapPerCustomer != nil {
				return nil, errors.New("for 'count_per_customer' CampaignTypePerCustomer, ValidityMonthsPerCustomer and RewardCapPerCustomer must be nil")
			}
		default:
			return nil, errors.New("invalid CampaignTypePerCustomer; must be 'one_time', 'forever', 'months_per_customer', or 'count_per_customer'")
		}
	}

	if req.Budget != nil && req.Budget.Cmp(decimal.NewFromInt(0)) <= 0 {
		return nil, errors.New("budget must be greater than zero")
	}

	if req.StartDate != nil && req.EndDate == nil {
		return nil, errors.New("end date is required if start date is provided")
	}

	if req.EndDate != nil && req.StartDate == nil {
		return nil, errors.New("start date is required if end date is provided")
	}

	if req.StartDate != nil && req.EndDate != nil {
		if req.StartDate.After(*req.EndDate) {
			return nil, errors.New("start date cannot be after end date")
		}
		if req.EndDate.Before(time.Now()) {
			return nil, errors.New("end date cannot be in the past")
		}
	}

	// fetch events using event keys
	var events []models.Event
	paymentCount := 0

	if req.EventKeys != nil && len(req.EventKeys) > 0 {

		// Fetch the events by keys
		if err := s.DB.Where("project = ? AND key IN ?", project, req.EventKeys).Find(&events).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch events for keys %v: %w", req.EventKeys, err)
		}

		if len(events) != len(req.EventKeys) {
			return nil, errors.New("not all event keys were found")
		}

		for _, event := range events {
			if event.EventType == "payment" {
				paymentCount++
			}
			if paymentCount > 1 {
				return nil, errors.New("only one event with event type 'payment' is allowed")
			}
		}
	}

	// Prepare the updates
	updates := map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.CampaignTypePerCustomer != nil {
		updates["campaign_type_per_customer"] = *req.CampaignTypePerCustomer
	}
	if req.ValidityMonthsPerCustomer != nil {
		updates["validity_months_per_customer"] = *req.ValidityMonthsPerCustomer
	}
	if req.MaxOccurrencesPerCustomer != nil {
		updates["max_occurrences_per_customer"] = *req.MaxOccurrencesPerCustomer
	}
	if req.RewardCapPerCustomer != nil {
		updates["reward_cap_per_customer"] = *req.RewardCapPerCustomer
	}
	if req.RewardType != nil {
		updates["reward_type"] = *req.RewardType
	}
	if req.RewardValue != nil {
		updates["reward_value"] = *req.RewardValue
	}
	if req.RewardCap != nil {
		updates["reward_cap"] = *req.RewardCap
	}
	if req.InviteeRewardType != nil {
		updates["invitee_reward_type"] = *req.InviteeRewardType
	}
	if req.InviteeRewardValue != nil {
		updates["invitee_reward_value"] = *req.InviteeRewardValue
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
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	// Validate the date range
	if req.StartDate != nil && req.EndDate != nil {
		if req.StartDate.After(*req.EndDate) {
			return nil, errors.New("start date cannot be after end date")
		}
	}

	var campaign models.Campaign

	// Fetch the campaign first
	if err := s.DB.Where("id = ? AND project = ?", id, project).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("campaign not found for project %s and id %d", project, id)
		}
		return nil, err
	}

	if req.EventKeys != nil && len(req.EventKeys) > 0 {
		if campaign.CampaignTypePerCustomer != "one_time" && len(events) > 1 {
			return nil, errors.New("only one event is allowed for campaigns with other than 'one_time' campaign type")
		}

		if campaign.RewardType == "percentage" && paymentCount == 0 {
			return nil, errors.New("at least one event with event type 'payment' is required for campaigns with 'percentage' reward type")
		}

		if campaign.RewardType == "flat_fee" && paymentCount > 0 {
			return nil, errors.New("no event with event type 'payment' is allowed for campaigns with 'flat_fee' reward type")
		}
	}

	// Wrap the operation in a transaction
	err := s.DB.Transaction(func(tx *gorm.DB) error {

		// Fetch the campaign with a record-level lock
		if err := tx.Where("id = ? AND project = ?", id, project).
			Clauses(clause.Locking{Strength: "UPDATE"}). // Add record-level lock
			First(&campaign).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("campaign not found for project %s and id %d", project, id)
			}
			return err
		}

		if len(updates) > 0 {
			// Apply the updates
			if err := tx.Model(&campaign).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update campaign: %w", err)
			}
		}

		if req.EventKeys != nil && len(req.EventKeys) > 0 {
			// Remove existing event associations
			if err := tx.Unscoped().Where("campaign_id = ?", campaign.ID).Delete(&models.CampaignEvent{}).Error; err != nil {
				return fmt.Errorf("failed to remove existing event associations: %w", err)
			}

			// Add new event associations
			for _, event := range events {
				if err := tx.Create(&models.CampaignEvent{
					Project:    project,
					CampaignID: campaign.ID,
					EventID:    event.ID,
					EventKey:   event.Key,
				}).Error; err != nil {
					return fmt.Errorf("failed to associate event %s with campaign %d: %w", event.Key, campaign.ID, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Assign events to avoid redundant reloading
	campaign.Events = events

	return &campaign, nil
}
