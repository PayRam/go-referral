package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"regexp"
)

type eventService struct {
	DB *gorm.DB
}

//var _ service.EventService = &eventService{}

func NewEventService(db *gorm.DB) *eventService {
	return &eventService{DB: db}
}

// CreateEvent creates a new event associated with a campaign
func (s *eventService) CreateEvent(project string, request request.CreateEventRequest) (*models.Event, error) {

	var validKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// Inside your method
	if request.Key == "" {
		return nil, errors.New("event key is required")
	}

	if !validKeyRegex.MatchString(request.Key) {
		return nil, errors.New("event key must only contain letters, numbers, underscores (_), or hyphens (-)")
	}

	if request.Name == "" {
		return nil, errors.New("event name is required")
	}

	if request.EventType != "simple" && request.EventType != "payment" {
		return nil, errors.New("event type must be either 'simple' or 'payment'")
	}

	// Check if the event key already exists
	var count int64
	if err := s.DB.Model(&models.Event{}).Where("project = ? AND key = ?", project, request.Key).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check existing event: %w", err)
	}

	if count > 0 {
		return nil, errors.New("event key already exists")
	}

	event := &models.Event{
		Project:     project,
		Key:         request.Key,
		Name:        request.Name,
		Description: request.Description,
		EventType:   request.EventType,
	}

	if err := s.DB.Create(event).Error; err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	return event, nil
}

// UpdateEvent updates an existing event with row-level locking
func (s *eventService) UpdateEvent(project, key string, req request.UpdateEventRequest) (*models.Event, error) {
	if req.Name == nil && req.Description == nil {
		return nil, errors.New("no update fields provided")
	}

	if req.Name != nil && *req.Name == "" {
		return nil, errors.New("event name cannot be empty")
	}

	if req.Description != nil && *req.Description == "" {
		return nil, errors.New("event description cannot be empty")
	}

	var event models.Event

	// Use a transaction to ensure atomicity and apply row-level locking
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the event by key with a row-level lock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&event, "project = ? AND key = ?", project, key).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("event not found with key: %s", key)
			}
			return err
		}

		// Prepare updates dynamically based on non-nil fields in the request
		updates := map[string]interface{}{}
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}

		// Apply updates
		if len(updates) > 0 {
			if err := tx.Model(&event).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &event, nil
}

// GetEvents retrieves events based on dynamic conditions
func (s *eventService) GetEvents(req request.GetEventsRequest) ([]models.Event, int64, error) {
	var events []models.Event
	var count int64

	// Start query
	query := s.DB.Model(&models.Event{})

	// Apply filters
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.Key != nil {
		query = query.Where("key = ?", *req.Key)
	}
	if req.Name != nil {
		query = query.Where("name LIKE ?", "%"+*req.Name+"%")
	}
	if req.EventType != nil {
		query = query.Where("event_type = ?", *req.EventType)
	}

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch events: %w", err)
	}

	return events, count, nil
}
