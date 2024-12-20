package serviceimpl

import (
	"errors"
	"github.com/PayRam/go-referral/internal/db"
	"github.com/PayRam/go-referral/service/param"
	"gorm.io/gorm"
	"time"
)

type service struct {
	DB *gorm.DB
}

func NewReferralServiceWithDB(db *gorm.DB) param.ReferralService {
	return &service{DB: db}
}

func NewReferralService(dbPath string) param.ReferralService {
	db := db.InitDB(dbPath)
	return &service{DB: db}
}

// CreateEvent creates a new event associated with a campaign
func (s *service) CreateEvent(name, eventType, rewardType string, rewardValue float64, maxOccurrences, validityDays uint) (*param.Event, error) {
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
func (s *service) UpdateEvent(id uint, updates map[string]interface{}) (*param.Event, error) {
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

// CreateCampaign creates a new campaign
func (s *service) CreateCampaign(name, description string, startDate, endDate time.Time, isActive bool, events []param.Event) (*param.Campaign, error) {
	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	campaign := &param.Campaign{
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    isActive,
		Events:      events,
	}

	if err := s.DB.Create(campaign).Error; err != nil {
		return nil, err
	}
	return campaign, nil
}

// UpdateCampaign updates an existing campaign
func (s *service) UpdateCampaign(id uint, updates map[string]interface{}) (*param.Campaign, error) {
	var campaign param.Campaign
	if err := s.DB.First(&campaign, id).Error; err != nil {
		return nil, err
	}

	// Apply updates
	if err := s.DB.Model(&campaign).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}
