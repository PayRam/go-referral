package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/service"
	"gorm.io/gorm"
)

type eventService struct {
	DB *gorm.DB
}

var _ service.EventService = &eventService{}

func NewEventService(db *gorm.DB) service.EventService {
	return &eventService{DB: db}
}

// CreateEvent creates a new event associated with a campaign
func (s *eventService) CreateEvent(key, name, eventType, rewardType string, rewardValue float64, maxOccurrences, validityDays uint) (*models.Event, error) {
	if rewardValue <= 0 {
		return nil, errors.New("reward value must be greater than 0")
	}

	// Validate unique key
	var existingEvent models.Event
	if err := s.DB.First(&existingEvent, "key = ?", key).Error; err == nil {
		return nil, errors.New("event key already exists")
	}

	event := &models.Event{
		Key:            key,
		Name:           name,
		EventType:      eventType,
		RewardType:     rewardType,
		RewardValue:    rewardValue,
		MaxOccurrences: maxOccurrences,
		ValidityDays:   validityDays,
	}

	if err := s.DB.Create(event).Error; err != nil {
		return nil, err
	}
	return event, nil
}

// UpdateEvent updates an existing event
func (s *eventService) UpdateEvent(key string, updates map[string]interface{}) (*models.Event, error) {
	var event models.Event
	if err := s.DB.First(&event, "key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("event not found with key: %s", key)
		}
		return nil, err
	}

	// Validate updates
	if rewardValue, ok := updates["rewardValue"]; ok {
		if value, ok := rewardValue.(float64); ok && value <= 0 {
			return nil, errors.New("reward value must be greater than 0")
		}
	}

	// Apply updates
	if err := s.DB.Model(&event).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &event, nil
}
