package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/service"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type eventLogService struct {
	DB *gorm.DB
}

var _ service.EventLogService = &eventLogService{}

// NewEventLogService initializes the EventLog service
func NewEventLogService(db *gorm.DB) service.EventLogService {
	return &eventLogService{DB: db}
}

// CreateEventLog creates a new event log entry
func (s *eventLogService) CreateEventLog(eventKey string, referenceID, referenceType *string, amount *decimal.Decimal, data *string) (*models.EventLog, error) {
	if amount == nil || amount.IsZero() {
		return nil, errors.New("amount must be greater than 0")
	}

	eventLog := &models.EventLog{
		EventKey:      eventKey,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		Amount:        amount,
		TriggeredAt:   time.Now(),
		Data:          data,
		Status:        "pending",
	}

	if err := s.DB.Create(eventLog).Error; err != nil {
		return nil, err
	}
	return eventLog, nil
}

// GetEventLogs retrieves event logs based on dynamic conditions
func (s *eventLogService) GetEventLogs(conditions map[string]interface{}, offset, limit *int, sort *string) ([]models.EventLog, error) {
	var eventLogs []models.EventLog

	// Start building the query
	query := s.DB.Model(&models.EventLog{})

	// Apply conditions dynamically
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
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
