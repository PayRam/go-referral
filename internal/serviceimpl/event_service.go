package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
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
func (s *eventService) CreateEvent(project, key, name string, description *string, eventType string) (*models.Event, error) {
	// Check if the event key already exists
	var count int64
	if err := s.DB.Model(&models.Event{}).Where("project = ? AND key = ?", project, key).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check existing event: %w", err)
	}

	if count > 0 {
		return nil, errors.New("event key already exists")
	}

	event := &models.Event{
		Project:     project,
		Key:         key,
		Name:        name,
		Description: description,
		EventType:   eventType,
	}

	if err := s.DB.Create(event).Error; err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	return event, nil
}

// UpdateEvent updates an existing event
func (s *eventService) UpdateEvent(project, key string, req request.UpdateEventRequest) (*models.Event, error) {
	var event models.Event

	// Fetch the event by key
	if err := s.DB.First(&event, "project = ? AND key = ?", project, key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("event not found with key: %s", key)
		}
		return nil, err
	}

	// Prepare updates dynamically based on non-nil fields in the request
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.EventType != nil {
		updates["event_type"] = *req.EventType
	}

	// Apply updates
	if len(updates) > 0 {
		if err := s.DB.Model(&event).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update event: %w", err)
		}
	}

	return &event, nil
}

func (s *eventService) GetAll(project string) ([]models.Event, error) {
	var events []models.Event

	// Fetch all events for the specified project
	if err := s.DB.Where("project = ?", project).Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events for project %s: %w", project, err)
	}

	return events, nil
}

func (s *eventService) GetByKey(project, key string) (*models.Event, error) {
	var event models.Event

	// Fetch the event by key
	if err := s.DB.First(&event, "project = ? AND key = ?", project, key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("event not found with key: %s", key)
		}
		return nil, err
	}

	return &event, nil
}

func (s *eventService) GetByKeys(project string, keys []string) ([]models.Event, error) {
	var events []models.Event

	// Fetch the events by keys
	if err := s.DB.Where("project = ? AND key IN ?", project, keys).Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events for keys %v: %w", keys, err)
	}

	// Check if all keys are found
	if len(events) != len(keys) {
		foundKeys := make(map[string]bool)
		for _, event := range events {
			foundKeys[event.Key] = true
		}

		missingKeys := []string{}
		for _, key := range keys {
			if !foundKeys[key] {
				missingKeys = append(missingKeys, key)
			}
		}

		return nil, fmt.Errorf("some keys were not found: %v", missingKeys)
	}

	return events, nil
}

func (s *eventService) SearchByName(project, name string) ([]models.Event, error) {
	var events []models.Event

	// Fetch events by name using a case-insensitive search with NOCASE
	if err := s.DB.Where("project = ? AND name LIKE ? COLLATE NOCASE", project, "%"+name+"%").Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events by name: %w", err)
	}

	return events, nil
}
