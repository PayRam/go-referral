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

	if req.CurrencyCode == "" {
		return nil, errors.New("currencyCode is required")
	}

	// Validate required fields
	if req.RewardType != nil || req.RewardValue != nil {
		if req.RewardType == nil || req.RewardValue == nil {
			return nil, errors.New("both rewardType and rewardValue must be provided or omitted")
		}
		if *req.RewardType != "flat_fee" && *req.RewardType != "percentage" {
			return nil, errors.New("rewardType must be either 'flat_fee' or 'percentage'")
		}
		if req.RewardValue.Cmp(decimal.NewFromInt(0)) <= 0 {
			return nil, errors.New("rewardValue must be greater than zero")
		}
		if *req.RewardType == "percentage" {
			if req.RewardValue.Cmp(decimal.NewFromInt(100)) > 0 {
				return nil, errors.New("percentage rewardValue must be between 0 and 100")
			}
			if req.RewardCap != nil && req.RewardCap.Cmp(decimal.NewFromInt(0)) <= 0 {
				return nil, errors.New("rewardCap must be greater than zero")
			}
			if req.RewardCap != nil && req.RewardCapPerCustomer != nil && req.RewardCap.Cmp(*req.RewardCapPerCustomer) > 0 {
				return nil, errors.New("reward cap must be less than or equal to reward cap per customer")
			}
			if req.RewardCapPerCustomer != nil && req.Budget != nil && req.RewardCapPerCustomer.Cmp(*req.Budget) > 0 {
				return nil, errors.New("reward cap per customer must be less than or equal to budget")
			}
			// RewardCap can be nil or set
		} else if *req.RewardType == "flat_fee" {
			if req.RewardCap != nil {
				return nil, errors.New("rewardCap must be nil for flat_fee rewardType")
			}
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
		if req.InviteeRewardValue.Cmp(decimal.NewFromInt(0)) <= 0 {
			return nil, errors.New("inviteeRewardValue must be greater than zero")
		}
		if *req.InviteeRewardType == "percentage" {
			//req.Budget.Cmp(decimal.NewFromInt(0)) <= 0
			if req.InviteeRewardValue.Cmp(decimal.NewFromInt(100)) > 0 {
				return nil, errors.New("percentage inviteeRewardValue must be between 0 and 100")
			}
			if req.InviteeRewardCap != nil && req.InviteeRewardCap.Cmp(decimal.NewFromInt(0)) <= 0 {
				return nil, errors.New("inviteeRewardCap must be greater than zero")
			}
			if req.InviteeRewardCap != nil && req.RewardCapPerCustomer != nil && req.InviteeRewardCap.Cmp(*req.RewardCapPerCustomer) > 0 {
				return nil, errors.New("invitee reward cap must be less than or equal to reward cap per customer")
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
		if req.ValidityMonthsPerCustomer != nil || req.MaxOccurrencesPerCustomer != nil {
			return nil, errors.New("for 'one_time' or 'forever' CampaignTypePerCustomer, ValidityMonthsPerCustomer and MaxOccurrencesPerCustomer must be nil")
		}
	case "months_per_customer":
		if req.ValidityMonthsPerCustomer == nil {
			return nil, errors.New("ValidityMonthsPerCustomer is required for 'months_per_customer' CampaignTypePerCustomer")
		}
		if req.MaxOccurrencesPerCustomer != nil {
			return nil, errors.New("for 'months_per_customer' MaxOccurrencesPerCustomer must be nil")
		}
	case "count_per_customer":
		if req.MaxOccurrencesPerCustomer == nil {
			return nil, errors.New("MaxOccurrencesPerCustomer is required for 'count_per_customer' CampaignTypePerCustomer")
		}
		if req.ValidityMonthsPerCustomer != nil {
			return nil, errors.New("for 'count_per_customer' ValidityMonthsPerCustomer must be nil")
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

	if req.RewardType == nil && req.InviteeRewardType == nil {
		return nil, errors.New("either rewardType or inviteeRewardType must be provided")
	}

	if req.RewardType != nil && *req.RewardType == "percentage" && paymentCount != 1 {
		return nil, errors.New("only one event with event type 'payment' is required for campaigns with 'percentage' reward type")
	}

	if req.InviteeRewardType != nil && *req.InviteeRewardType == "percentage" && paymentCount != 1 {
		return nil, errors.New("only one event with event type 'payment' is required for campaigns with 'percentage' invitee reward type")
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
		CurrencyCode:              req.CurrencyCode,
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
		ConsiderEventsFrom:        time.Now().UTC(),
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

	query = request.ApplyGetCampaignRequest(req, query)

	// Apply Select Fields
	query = request.ApplySelectFields(query, req.PaginationConditions.SelectFields)

	// Apply Group By
	query = request.ApplyGroupBy(query, req.PaginationConditions.GroupBy)

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Preload("Events").Find(&campaigns).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch campaigns: %w", err)
	}

	return campaigns, count, nil
}

func (s *campaignService) GetTotalCampaigns(req request.GetCampaignsRequest) (int64, error) {
	var count int64

	// Build the query
	query := s.DB.Model(&models.Campaign{})

	query = request.ApplyGetCampaignRequest(req, query)

	// Apply Select Fields
	query = request.ApplySelectFields(query, req.PaginationConditions.SelectFields)

	// Apply Group By
	query = request.ApplyGroupBy(query, req.PaginationConditions.GroupBy)

	// Count the records
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count referrers: %w", err)
	}

	return count, nil
}

// UpdateCampaign updates an existing campaign
func (s *campaignService) UpdateCampaign(project string, id uint, req request.UpdateCampaignRequest) (*models.Campaign, error) {
	var campaign models.Campaign

	// Fetch the campaign first
	if err := s.DB.Where("id = ? AND project = ?", id, project).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("campaign not found for project %s and id %d", project, id)
		}
		return nil, err
	}

	currentTime := time.Now()

	isOngoing := campaign.StartDate.Before(currentTime) && campaign.EndDate.After(currentTime)
	isFuture := campaign.StartDate.After(currentTime)

	if !isOngoing && !isFuture {
		return nil, errors.New("cannot update a campaign that has ended")
	}

	// If the campaign is ongoing, restrict the fields that can be updated
	if isOngoing {
		if req.Name == nil && req.Budget == nil && req.Description == nil && req.EndDate == nil {
			return nil, errors.New("only Name, Budget, Description, and EndDate can be updated for ongoing campaigns")
		}
	}

	// If updating the budget, ensure it is not less than the total rewards distributed
	if req.Budget != nil {
		var totalRewards decimal.Decimal
		err := s.DB.Model(&models.Reward{}).
			Where("project = ? AND campaign_id = ?", project, campaign.ID).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&totalRewards).Error

		if err != nil {
			return nil, fmt.Errorf("failed to calculate total rewards: %w", err)
		}

		if req.Budget.Cmp(totalRewards) < 0 {
			return nil, fmt.Errorf("budget cannot be less than the total rewards distributed (%.18s)", totalRewards.String())
		}
	}

	if req.Name != nil && *req.Name == "" {
		return nil, errors.New("name cannot be empty")
	}

	if req.CurrencyCode != nil && *req.CurrencyCode == "" {
		return nil, errors.New("currencyCode cannot be empty")
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
		if req.RewardCap != nil && req.RewardCap.Cmp(decimal.NewFromInt(0)) <= 0 {
			return nil, errors.New("rewardCap must be greater than zero")
		}
		if req.RewardCap != nil && req.RewardCapPerCustomer != nil && req.RewardCap.Cmp(*req.RewardCapPerCustomer) > 0 {
			return nil, errors.New("reward cap must be less than or equal to reward cap per customer")
		}
		if req.RewardCapPerCustomer != nil && req.Budget != nil && req.RewardCapPerCustomer.Cmp(*req.Budget) > 0 {
			return nil, errors.New("reward cap per customer must be less than or equal to budget")
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
			if req.InviteeRewardCap != nil && req.InviteeRewardCap.Cmp(decimal.NewFromInt(0)) <= 0 {
				return nil, errors.New("inviteeRewardCap must be greater than zero")
			}
			if req.InviteeRewardCap != nil && req.RewardCapPerCustomer != nil && req.InviteeRewardCap.Cmp(*req.RewardCapPerCustomer) > 0 {
				return nil, errors.New("invitee reward cap must be less than or equal to reward cap per customer")
			}
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
			if req.ValidityMonthsPerCustomer != nil || req.MaxOccurrencesPerCustomer != nil {
				return nil, errors.New("for 'one_time' or 'forever' CampaignTypePerCustomer, ValidityMonthsPerCustomer and MaxOccurrencesPerCustomer must be nil")
			}
		case "months_per_customer":
			if req.ValidityMonthsPerCustomer == nil {
				return nil, errors.New("ValidityMonthsPerCustomer is required for 'months_per_customer' CampaignTypePerCustomer")
			}
			if req.MaxOccurrencesPerCustomer != nil {
				return nil, errors.New("for 'months_per_customer' MaxOccurrencesPerCustomer must be nil")
			}
		case "count_per_customer":
			if req.MaxOccurrencesPerCustomer == nil {
				return nil, errors.New("MaxOccurrencesPerCustomer is required for 'count_per_customer' CampaignTypePerCustomer")
			}
			if req.ValidityMonthsPerCustomer != nil {
				return nil, errors.New("for 'count_per_customer' ValidityMonthsPerCustomer must be nil")
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

	// Validate the date range
	if req.StartDate != nil && req.EndDate != nil {
		if req.StartDate.After(*req.EndDate) {
			return nil, errors.New("start date cannot be after end date")
		}
	}

	// Prepare the updates
	updates := map[string]interface{}{}

	// If the campaign is future, allow updating all fields
	if isFuture {
		updates = request.UpdateCampaignFields(req, updates)
	} else {
		// Ongoing campaign: Only update allowed fields
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Budget != nil {
			updates["budget"] = req.Budget
		}
		if req.Description != nil {
			updates["description"] = req.Description
		}
		if req.EndDate != nil {
			if req.EndDate.Before(currentTime) {
				return nil, errors.New("end date cannot be in the past")
			}
			updates["end_date"] = req.EndDate
		}
	}

	if req.EventKeys != nil && len(req.EventKeys) > 0 {

		if campaign.RewardType != nil && *campaign.RewardType == "percentage" && paymentCount != 1 {
			return nil, errors.New("at least one event with event type 'payment' is required for campaigns with 'percentage' reward type")
		}

		if campaign.InviteeRewardType != nil && *campaign.InviteeRewardType == "percentage" && paymentCount != 1 {
			return nil, errors.New("at least one event with event type 'payment' is required for campaigns with 'percentage' invitee reward type")
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
			if _, ok := updates["budget"]; ok {
				updates["status"] = "active"
			}
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

func (s *campaignService) SetDefaultCampaign(project string, campaignID uint) (*models.Campaign, error) {
	var updatedCampaign models.Campaign

	// Perform the update in a transaction
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the campaign with a row-level lock
		var campaign models.Campaign
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("project = ? AND id = ?", project, campaignID).
			First(&campaign).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("campaign not found for project %s and ID %d", project, campaignID)
			}
			return fmt.Errorf("failed to fetch campaign with lock: %w", err)
		}

		// Update is_default for the campaign
		if err := tx.Model(&models.Campaign{}).
			Where("project = ? AND id = ?", project, campaignID).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("failed to set campaign %d as default: %w", campaignID, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload the updated campaign with its associations
	if err := s.DB.Preload("Events").First(&updatedCampaign, "project = ? AND id = ?", project, campaignID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated campaign: %w", err)
	}

	return &updatedCampaign, nil
}

func (s *campaignService) RemoveDefaultCampaign(project string, campaignID uint) (*models.Campaign, error) {
	var updatedCampaign models.Campaign
	// Perform the update in a transaction
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Set all campaigns in the project to "is_default = false"
		if err := tx.Model(&models.Campaign{}).
			Where("project = ? AND id = ?", project, campaignID).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to remove default status from campaigns in project %s: %w", project, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload the updated campaign with its associations
	if err := s.DB.Preload("Events").First(&updatedCampaign, "project = ? AND id = ?", project, campaignID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated campaign: %w", err)
	}

	return &updatedCampaign, nil
}

// UpdateCampaignStatus updates the status of an existing campaign
func (s *campaignService) UpdateCampaignStatus(project string, campaignID uint, newStatus string) (*models.Campaign, error) {
	var campaign models.Campaign

	if newStatus != "active" && newStatus != "paused" && newStatus != "archived" {
		return nil, errors.New("status must be either 'active', 'paused', or 'archived'")
	}

	// Use a transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the campaign with a row-level lock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("project = ? AND id = ?", project, campaignID).
			First(&campaign).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("campaign not found for project %s and ID %d: %w", project, campaignID, err)
			}
			return err
		}

		if campaign.Status == "archived" {
			return fmt.Errorf("campaign is archived and cannot be updated")
		}

		if newStatus != "archived" && campaign.EndDate.Before(time.Now()) {
			return fmt.Errorf("campaign has ended and cannot be set to '%s'", newStatus)
		}

		// Check if the status is already the same as newStatus
		if campaign.Status == newStatus {
			return fmt.Errorf("campaign is already in status '%s'", newStatus)
		}

		// Prepare update fields
		updateFields := map[string]interface{}{
			"status": newStatus,
		}

		// If status is changing to active, update ConsiderEventsFrom timestamp
		if newStatus == "active" {
			now := time.Now().UTC()
			updateFields["consider_events_from"] = now
		}

		// Update the campaign status and ConsiderEventsFrom if applicable
		if err := tx.Model(&campaign).Updates(updateFields).Error; err != nil {
			return fmt.Errorf("failed to update the campaign status to '%s': %w", newStatus, err)
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
