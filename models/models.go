package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id" seeder:"no-update"`
	CreatedAt time.Time      `gorm:"index" json:"createdAt" seeder:"no-update"`
	UpdatedAt time.Time      `gorm:"index" json:"updatedAt" seeder:"no-update"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-" seeder:"no-update"`
}

type Campaign struct {
	BaseModel
	Project            string           `gorm:"size:100;not null;index"`
	Name               string           `gorm:"size:255;not null;index"`
	RewardType         string           `gorm:"size:50"`                        // e.g., "flat_fee", "percentage"
	RewardValue        decimal.Decimal  `gorm:""`                               // Percentage value or flat fee
	RewardCap          *decimal.Decimal `gorm:"type:decimal(38,18)"`            // Maximum reward for percentage type
	InviteeRewardType  *string          `gorm:"size:50"`                        // e.g., "flat_fee", "percentage"
	InviteeRewardValue *decimal.Decimal `gorm:""`                               // Reward for invitee
	InviteeRewardCap   *decimal.Decimal `gorm:"type:decimal(38,18)"`            // Cap for invitee reward
	Budget             *decimal.Decimal `gorm:"type:decimal(38,18)"`            // Budget for the campaign
	Description        *string          `gorm:"type:text"`                      // Optional description
	StartDate          *time.Time       `gorm:"not null;index"`                 // Start date of the campaign
	EndDate            *time.Time       `gorm:"not null;index"`                 // End date of the campaign
	Status             string           `gorm:"size:50;default:'active';index"` // New field to track campaign status (e.g., 'active', 'paused', 'archived')
	IsDefault          bool             `gorm:"default:false;index"`            // Only one default campaign

	CampaignTypePerCustomer   string           `gorm:"size:50;not null;index"` // Campaign type: "one_time", "forever", "months_per_customer", "count_per_customer"
	MaxOccurrencesPerCustomer *int64           `gorm:""`                       // 0 for unlimited
	ValidityMonthsPerCustomer *int             `gorm:""`                       // 0 for no time limit
	RewardCapPerCustomer      *decimal.Decimal `gorm:"type:decimal(38,18)"`    // Maximum reward for percentage type

	Events []Event `gorm:"many2many:referral_campaign_events"` // Associated events
}

func (Campaign) TableName() string {
	return "referral_campaigns"
}

// Event represents an action within a campaign that can trigger a reward
type Event struct {
	BaseModel
	Project     string  `gorm:"size:100;not null;uniqueIndex:idx_event_project_key" seeder:"no-update"`
	Key         string  `gorm:"size:100;not null;uniqueIndex:idx_event_project_key" seeder:"no-update"`
	Name        string  `gorm:"size:255;not null;index" seeder:"no-update"`
	EventType   string  `gorm:"size:100;not null;index" seeder:"no-update"`
	Description *string `gorm:"type:text" seeder:"no-update"`
}

func (Event) TableName() string {
	return "referral_events"
}

type CampaignEvent struct {
	Project    string   `gorm:"not null;size:100;index"`
	CampaignID uint     `gorm:"not null;uniqueIndex:idx_campaign_id_event_id"`
	EventID    uint     `gorm:"not null;uniqueIndex:idx_campaign_id_event_id"`
	EventKey   string   `gorm:"not null;size:100;index"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID"`
	Event      Event    `gorm:"foreignKey:EventID;references:ID"`
}

func (CampaignEvent) TableName() string {
	return "referral_campaign_events"
}

type Referrer struct {
	BaseModel
	Project     string     `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference_id"` // Composite key with ReferenceID
	ReferenceID string     `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference_id"` // Composite key with Project
	Code        string     `gorm:"size:50;uniqueIndex;not null"`                                    // Unique referral code
	Campaigns   []Campaign `gorm:"many2many:referral_referrer_campaigns;joinForeignKey:ReferrerID;joinReferences:CampaignID"`
}

func (Referrer) TableName() string {
	return "referral_referrer"
}

type ReferrerCampaign struct {
	Project    string   `gorm:"not null;size:100;"`
	ReferrerID uint     `gorm:"not null;uniqueIndex:idx_referral_referrer_campaign"`
	CampaignID uint     `gorm:"not null;uniqueIndex:idx_referral_referrer_campaign"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID"` // Foreign key to Campaign
	Referrer   Referrer `gorm:"foreignKey:ReferrerID;references:ID"` // Foreign key to Referrer
}

func (ReferrerCampaign) TableName() string {
	return "referral_referrer_campaigns"
}

type Referee struct {
	BaseModel
	Project             string    `gorm:"size:100;not null;uniqueIndex:idx_referee_project_reference"` // Composite key with ReferenceID
	ReferenceID         string    `gorm:"size:100;not null;uniqueIndex:idx_referee_project_reference"` // Composite key with Project
	ReferrerID          uint      `gorm:"not null;uniqueIndex:idx_referee_project_reference"`          // ID of the Referrer (Foreign Key to Referrer table)
	ReferrerReferenceID string    `gorm:"size:100;not null;uniqueIndex:idx_referee_project_reference"` // ID of the Referrer (Foreign Key to Referrer table)
	Referrer            *Referrer `gorm:"foreignKey:ReferrerID" json:"referrer,omitempty"`
}

func (Referee) TableName() string {
	return "referral_referee"
}

type EventLog struct {
	BaseModel
	Project       string           `gorm:"size:100;not null;index"`
	EventKey      string           `gorm:"size:100;not null;index" foreignKey:"Key" references:"Event"`
	ReferenceID   string           `gorm:"size:100;not null;index"`
	Amount        *decimal.Decimal `gorm:"type:decimal(38,18);index"`
	TriggeredAt   time.Time        `gorm:"not null;index"`                           // Timestamp when the event was triggered
	Data          *string          `gorm:"type:json;"`                               // Arbitrary data associated with the event (JSON format)
	Status        string           `gorm:"size:50;default:'pending';not null;index"` // Status of the event processing (e.g., "pending", "processed", "failed")
	FailureReason *string          `gorm:"type:text"`

	// Foreign key for the reward this log contributes to
	RewardID *uint   `gorm:"index"` // Nullable to allow logs without an associated reward
	Reward   *Reward `gorm:"foreignKey:RewardID"`
}

func (EventLog) TableName() string {
	return "referral_event_logs"
}

type Reward struct {
	BaseModel
	Project             string          `gorm:"size:100;not null;index"`
	CampaignID          uint            `gorm:"not null;index"`                           // Foreign key to Campaign
	RefereeID           uint            `gorm:"not null;index"`                           // Foreign key to Referee
	RefereeReferenceID  string          `gorm:"size:100;not null;index"`                  // Foreign key to Referee
	ReferrerID          uint            `gorm:"not null;index"`                           // Foreign key to Referee
	ReferrerReferenceID string          `gorm:"size:100;index"`                           // ReferrerReferenceID of the entity related to the reward
	ReferrerCode        string          `gorm:"size:50;not null;index"`                   // Unique referral code
	Amount              decimal.Decimal `gorm:"type:decimal(38,18);not null;index"`       // Calculated reward amount
	Status              string          `gorm:"size:50;default:'pending';not null;index"` // Reward status (e.g., "pending", "processed", "failed")
	Reason              *string         `gorm:"type:text"`                                // Reason for reward status (optional)

	// One-to-many relationship with EventLog
	EventLogs []EventLog `gorm:"foreignKey:RewardID"` // Associated EventLogs
}

func (Reward) TableName() string {
	return "referral_rewards"
}
