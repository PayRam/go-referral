package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"gorm.io/gorm"
	"time"
)

type eventLogService struct {
	DB *gorm.DB
}

//var _ service.EventLogService = &eventLogService{}

// NewEventLogService initializes the EventLog service
func NewEventLogService(db *gorm.DB) *eventLogService {
	return &eventLogService{DB: db}
}

// CreateEventLog creates a new event log entry
func (s *eventLogService) CreateEventLog(project string, req request.CreateEventLogRequest) (*models.EventLog, error) {
	// Fetch the event by project and eventKey
	var event models.Event
	if err := s.DB.Where("project = ? AND key = ?", project, req.EventKey).First(&event).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch event with key '%s' for project '%s': %w", req.EventKey, project, err)
	}

	// Validate amount based on event type
	if event.EventType == "payment" {
		if req.Amount == nil || req.Amount.IsZero() {
			return nil, errors.New("amount must be greater than 0 for payment events")
		}
	} else {
		if req.Amount != nil {
			return nil, errors.New("amount must be nil for non-payment events")
		}
	}

	// Create the event log
	eventLog := &models.EventLog{
		Project:     project,
		EventKey:    req.EventKey,
		ReferenceID: req.ReferenceID,
		Amount:      req.Amount,
		TriggeredAt: time.Now(),
		Data:        req.Data,
		Status:      "pending",
	}

	if err := s.DB.Create(eventLog).Error; err != nil {
		return nil, fmt.Errorf("failed to create event log: %w", err)
	}

	return eventLog, nil
}

// GetEventLogs retrieves event logs based on dynamic conditions
func (s *eventLogService) GetEventLogs(req request.GetEventLogRequest) ([]models.EventLog, int64, error) {
	var eventLogs []models.EventLog
	var count int64

	// Start query
	query := s.DB.Model(&models.EventLog{})

	// Apply filters
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.EventKey != nil {
		query = query.Where("event_key = ?", *req.EventKey)
	}
	if req.ReferenceID != nil {
		query = query.Where("reference_id = ?", *req.ReferenceID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.RewardID != nil {
		query = query.Where("reward_id = ?", *req.RewardID)
	}

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count eventLogs: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Preload("Reward").Find(&eventLogs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch eventLogs: %w", err)
	}

	return eventLogs, count, nil
}
