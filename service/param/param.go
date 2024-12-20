package param

import (
	"gorm.io/gorm"
	"time"
)

type ReferralService interface {
	CreateEvent(name, eventType, rewardType string, rewardValue float64, maxOccurrences, validityDays uint) (*Event, error)
	UpdateEvent(id uint, updates map[string]interface{}) (*Event, error)
	CreateCampaign(name, description string, startDate, endDate time.Time, isActive bool, events []Event) (*Campaign, error)
	UpdateCampaign(id uint, updates map[string]interface{}) (*Campaign, error)
}

// Campaign represents a referral campaign
type Campaign struct {
	gorm.Model
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	StartDate   time.Time `gorm:"not null"`
	EndDate     time.Time `gorm:"not null"`
	IsActive    bool      `gorm:"default:true"`
	Events      []Event   `gorm:"foreignKey:CampaignID"`
}

// TableName sets the table name for the Campaign model
func (Campaign) TableName() string {
	return "referral_campaigns"
}

// Event represents an action within a campaign that can trigger a reward
type Event struct {
	gorm.Model
	Name           string  `gorm:"size:255;not null"`
	EventType      string  `gorm:"size:100;not null"` // e.g., "signup", "payment"
	RewardType     string  `gorm:"size:50;not null"`  // e.g., "flat_fee", "percentage"
	RewardValue    float64 `gorm:"not null"`
	MaxOccurrences uint    `gorm:"default:0"` // 0 for unlimited
	ValidityDays   uint    `gorm:"default:0"` // 0 for no time limit
}

// TableName sets the table name for the Event model
func (Event) TableName() string {
	return "referral_events"
}
