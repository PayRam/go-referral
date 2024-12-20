package param

import (
	"github.com/PayRam/go-referral/internal/serviceimpl"
	"gorm.io/gorm"
	"time"
)

// EventService handles operations related to events
type EventService interface {
	CreateEvent(name, eventType, rewardType string, rewardValue float64, maxOccurrences, validityDays uint) (*Event, error)
	UpdateEvent(id uint, updates map[string]interface{}) (*Event, error)
}

// CampaignService handles operations related to campaigns
type CampaignService interface {
	CreateCampaign(name, description string, startDate, endDate time.Time, isActive bool, events []Event) (*Campaign, error)
	UpdateCampaign(id uint, updates map[string]interface{}) (*Campaign, error)
	SetDefaultCampaign(campaignID uint) error
}

// ReferralCodeService handles operations related to referral codes
type ReferralCodeService interface {
	CreateReferralCode(referenceID, referenceType string) (*ReferralCode, error)
}

// UserCampaignService handles the association between users and campaigns
type UserCampaignService interface {
	AssignUserToCampaign(userID, campaignID uint) error
}

type ReferralService struct {
	EventService
	CampaignService
	ReferralCodeService
	UserCampaignService
}

// NewReferralService initializes the unified service
func NewReferralService(db *gorm.DB) *ReferralService {
	return &ReferralService{
		EventService:        serviceimpl.NewEventService(db),
		CampaignService:     serviceimpl.NewCampaignService(db),
		ReferralCodeService: serviceimpl.NewReferralCodeService(db),
		UserCampaignService: serviceimpl.NewUserCampaignService(db),
	}
}

// Campaign represents a referral campaign
type Campaign struct {
	gorm.Model
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	StartDate   time.Time `gorm:"not null"`
	EndDate     time.Time `gorm:"not null"`
	IsActive    bool      `gorm:"default:true"`
	IsDefault   bool      `gorm:"default:false"` // Only one default campaign
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

// UserCampaignMapping represents the association between users and campaigns
type UserCampaignMapping struct {
	gorm.Model
	UserID     uint `gorm:"not null;unique"` // Each user can only be in one campaign
	CampaignID uint `gorm:"not null"`        // Associated campaign
}

// TableName sets the table name for the UserCampaignMapping model
func (UserCampaignMapping) TableName() string {
	return "referral_user_campaign_mappings"
}

// ReferralCode represents a unique referral code for a user
type ReferralCode struct {
	gorm.Model
	Code          string `gorm:"size:50;unique;not null"` // Unique referral code
	ReferenceID   string `gorm:"not null;unique"`         // ID of the associated entity
	ReferenceType string `gorm:"size:100;not null"`       // Type of the associated entity (e.g., "User", "Campaign")
}

// TableName sets the table name for the ReferralCode model
func (ReferralCode) TableName() string {
	return "referral_codes"
}
