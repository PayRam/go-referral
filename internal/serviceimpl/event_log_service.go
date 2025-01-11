package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/internal/db"
	"github.com/PayRam/go-referral/models"
	"github.com/shopspring/decimal"
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
func (s *eventLogService) CreateEventLog(project, eventKey string, referenceID string, amount *decimal.Decimal, data *string) (*models.EventLog, error) {
	if amount == nil || amount.IsZero() {
		return nil, errors.New("amount must be greater than 0")
	}

	eventLog := &models.EventLog{
		Project:     project,
		EventKey:    eventKey,
		ReferenceID: referenceID,
		Amount:      amount,
		TriggeredAt: time.Now(),
		Data:        data,
		Status:      "pending",
	}

	if err := s.DB.Create(eventLog).Error; err != nil {
		return nil, err
	}
	return eventLog, nil
}

// GetEventLogs retrieves event logs based on dynamic conditions
func (s *eventLogService) GetEventLogs(project string, conditions []db.QueryCondition, offset, limit *int, sort *string) ([]models.EventLog, error) {
	var eventLogs []models.EventLog

	// Start building the query
	query := s.DB.Model(&models.EventLog{})

	query = query.Where("project = ?", project)

	// Apply conditions dynamically
	for _, condition := range conditions {
		// Build the query with operator
		query = query.Where(fmt.Sprintf("%s %s ?", condition.Field, condition.Operator), condition.Value)
	}

	// Apply offset and limit
	if offset != nil {
		query = query.Offset(*offset)
	}
	if limit != nil {
		query = query.Limit(*limit)
	}

	// Apply sorting
	if sort != nil {
		query = query.Order(*sort)
	}

	// Execute the query
	if err := query.Find(&eventLogs).Error; err != nil {
		return nil, err
	}

	return eventLogs, nil
}
