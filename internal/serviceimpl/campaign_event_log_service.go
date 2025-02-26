package serviceimpl

import (
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"gorm.io/gorm"
)

type campaignEventLogService struct {
	DB *gorm.DB
}

//var _ service.campaignEventLogService = &campaignEventLogService{}

// NewCampaignEventLogService initializes the EventLog service
func NewCampaignEventLogService(db *gorm.DB) *campaignEventLogService {
	return &campaignEventLogService{DB: db}
}

// GetCampaignEventLogs retrieves event logs based on dynamic conditions
func (s *campaignEventLogService) GetCampaignEventLogs(req request.GetCampaignEventLogRequest) ([]models.CampaignEventLog, int64, error) {
	var campaignEventLogs []models.CampaignEventLog
	var count int64

	// Start query
	query := s.DB.Model(&models.CampaignEventLog{})

	query = request.ApplyGetCampaignEventLogRequest(req, query)

	// Apply Select Fields
	query = request.ApplySelectFields(query, req.PaginationConditions.SelectFields)

	// Apply Group By
	query = request.ApplyGroupBy(query, req.PaginationConditions.GroupBy)

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count campaignEventLogs: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Preload("Campaign").Preload("Event").Preload("Member").Preload("ReferredReward").Preload("RefereeReward").Find(&campaignEventLogs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch campaignEventLogs: %w", err)
	}

	return campaignEventLogs, count, nil
}
