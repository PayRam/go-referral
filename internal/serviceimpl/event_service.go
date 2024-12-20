package serviceimpl

import (
	"errors"
	"github.com/PayRam/go-referral/service/param"
	"gorm.io/gorm"
)

type eventService struct {
	DB *gorm.DB
}

func NewEventService(db *gorm.DB) param.EventService {
	return &eventService{DB: db}
}

// CreateEvent creates a new event associated with a campaign
func (s *eventService) CreateEvent(name, eventType, rewardType string, rewardValue float64, maxOccurrences, validityDays uint) (*param.Event, error) {
	if rewardValue <= 0 {
		return nil, errors.New("reward value must be greater than 0")
	}

	event := &param.Event{
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
func (s *eventService) UpdateEvent(id uint, updates map[string]interface{}) (*param.Event, error) {
	var event param.Event
	if err := s.DB.First(&event, id).Error; err != nil {
		return nil, err
	}

	// Apply updates
	if err := s.DB.Model(&event).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &event, nil
}
