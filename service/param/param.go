package param

import (
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
	UpdateCampaignEvents(campaignID uint, events []Event) error
	SetDefaultCampaign(campaignID uint) error
}

// ReferrerService handles operations related to referral codes
type ReferrerService interface {
	CreateReferrer(referenceID, referenceType, code string, campaignID *uint) (*Referrer, error)
	GetReferrerByReference(referenceID, referenceType string) (*Referrer, error)
	AssignCampaign(referenceID, referenceType string, campaignID uint) error
	RemoveCampaign(referenceID, referenceType string) error
}

// RefereeService handles operations related to referral codes
type RefereeService interface {
	CreateRefereeByCode(code, referenceID, referenceType string) (*Referee, error)
	GetRefereeByReference(referenceID, referenceType string) (*Referee, error)
	GetRefereesByReferrer(referrerID uint) ([]Referee, error)
}

type Campaign struct {
	gorm.Model
	Name        string    `gorm:"size:255;not null;uniqueIndex"`
	Description string    `gorm:"type:text"`
	StartDate   time.Time `gorm:"not null;index"`
	EndDate     time.Time `gorm:"not null;index"`
	IsActive    bool      `gorm:"default:true;index"`
	IsDefault   bool      `gorm:"default:false;index"` // Only one default campaign
	Events      []Event   `gorm:"many2many:referral_campaign_events;joinForeignKey:CampaignID;joinReferences:EventID"`
}

func (Campaign) TableName() string {
	return "referral_campaigns"
}

// Event represents an action within a campaign that can trigger a reward
type Event struct {
	gorm.Model
	Name           string  `gorm:"size:255;not null;uniqueIndex"` // Unique name across events
	EventType      string  `gorm:"size:100;not null;index"`       // e.g., "signup", "payment"
	RewardType     string  `gorm:"size:50;not null;index"`        // e.g., "flat_fee", "percentage"
	RewardValue    float64 `gorm:"not null"`
	MaxOccurrences uint    `gorm:"default:0"` // 0 for unlimited
	ValidityDays   uint    `gorm:"default:0"` // 0 for no time limit
}

func (Event) TableName() string {
	return "referral_events"
}

type CampaignEvent struct {
	CampaignID uint `gorm:"not null;index:idx_referral_campaign_event,unique"`
	EventID    uint `gorm:"not null;index:idx_referral_campaign_event,unique"`
}

func (CampaignEvent) TableName() string {
	return "referral_campaign_events"
}

type Referrer struct {
	gorm.Model
	Code          string `gorm:"size:50;uniqueIndex;not null"`        // Unique referral code
	ReferenceID   string `gorm:"not null;index:idx_reference,unique"` // Composite index
	ReferenceType string `gorm:"size:100;not null;index:idx_reference,unique"`
	CampaignID    *uint  `gorm:"default:null"`
}

func (Referrer) TableName() string {
	return "referral_referrer"
}

type Referee struct {
	gorm.Model
	ReferrerID    uint      `gorm:"not null"`                            // ID of the Referrer (Foreign Key to Referrer table)
	ReferenceID   string    `gorm:"not null;index:idx_reference,unique"` // Composite index
	ReferenceType string    `gorm:"size:100;not null;index:idx_reference,unique"`
	Referrer      *Referrer `gorm:"foreignKey:ReferrerID" json:"referrer,omitempty"`
}

func (Referee) TableName() string {
	return "referral_referee"
}
