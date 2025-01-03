package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"gorm.io/gorm"
)

type eventService struct {
	DB *gorm.DB
}

//var _ service.EventService = &eventService{}

func NewEventService(db *gorm.DB) *eventService {
	return &eventService{DB: db}
}

// CreateEvent creates a new event associated with a campaign
func (s *eventService) CreateEvent(key, name, eventType string) (*models.Event, error) {
	// Check if the event key already exists
	var count int64
	if err := s.DB.Model(&models.Event{}).Where("key = ?", key).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check existing event: %w", err)
	}

	if count > 0 {
		return nil, errors.New("event key already exists")
	}

	event := &models.Event{
		Key:       key,
		Name:      name,
		EventType: eventType,
	}

	if err := s.DB.Create(event).Error; err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
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

	// Apply updates
	if err := s.DB.Model(&event).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *eventService) GetAll() ([]models.Event, error) {
	var events []models.Event

	// Fetch all events
	if err := s.DB.Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}
