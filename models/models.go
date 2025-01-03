package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type Campaign struct {
	gorm.Model
	Name           string           `gorm:"size:255;not null;uniqueIndex"`
	RewardType     *string          `gorm:"size:50;index"` // e.g., "flat_fee", "percentage"
	RewardValue    *float64         `gorm:""`
	MaxOccurrences *uint            `gorm:"default:0"`           // 0 for unlimited
	ValidityDays   *uint            `gorm:"default:0"`           // 0 for no time limit
	Budget         *decimal.Decimal `gorm:"type:decimal(38,18)"` // Pointer to handle nil as unlimited
	Description    string           `gorm:"type:text"`
	StartDate      time.Time        `gorm:"not null;index"`
	EndDate        time.Time        `gorm:"not null;index"`
	IsActive       bool             `gorm:"default:true;index"`
	IsDefault      bool             `gorm:"default:false;index"` // Only one default campaign
	Events         []Event          `gorm:"many2many:referral_campaign_events;"`
}

func (Campaign) TableName() string {
	return "referral_campaigns"
}

// Event represents an action within a campaign that can trigger a reward
type Event struct {
	Key       string     `gorm:"primaryKey;type:varchar(50);not null"` // Custom string primary key (e.g., UUID)
	Name      string     `gorm:"size:255;not null"`                    // Event name (not unique anymore)
	EventType string     `gorm:"size:100;not null;index"`              // e.g., "simple", "payment"
	CreatedAt time.Time  `gorm:"autoCreateTime"`                       // Auto-manage created time
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`                       // Auto-manage updated time
	DeletedAt *time.Time `gorm:"index"`                                // Soft delete support
}

func (Event) TableName() string {
	return "referral_events"
}

type CampaignEvent struct {
	CampaignID uint     `gorm:"not null;index:idx_referral_campaign_event"`
	EventKey   string   `gorm:"not null;type:varchar(50);index:idx_referral_campaign_event"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID"` // Foreign key to Campaign
	Event      Event    `gorm:"foreignKey:EventKey;references:Key"`  // Foreign key to Event
}

func (CampaignEvent) TableName() string {
	return "referral_campaign_events"
}

type Referrer struct {
	gorm.Model
	Code          string     `gorm:"size:50;uniqueIndex;not null"`                         // Unique referral code
	ReferenceID   string     `gorm:"not null;index:idx_referrer_reference_id_type,unique"` // Composite index
	ReferenceType string     `gorm:"size:100;not null;index:idx_referrer_reference_id_type,unique"`
	Campaigns     []Campaign `gorm:"many2many:referral_referrer_campaigns;joinForeignKey:ReferrerID;joinReferences:CampaignID"`
}

func (Referrer) TableName() string {
	return "referral_referrer"
}

type ReferrerCampaign struct {
	ReferrerID uint     `gorm:"not null;index:idx_referral_referrer_campaign"`
	CampaignID uint     `gorm:"not null;index:idx_referral_referrer_campaign"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID"` // Foreign key to Campaign
	Referrer   Referrer `gorm:"foreignKey:ReferrerID;references:ID"` // Foreign key to Referrer
}

func (ReferrerCampaign) TableName() string {
	return "referral_referrer_campaigns"
}

type Referee struct {
	gorm.Model
	ReferrerID    uint      `gorm:"not null"`                                            // ID of the Referrer (Foreign Key to Referrer table)
	ReferenceID   string    `gorm:"not null;index:idx_referee_reference_id_type,unique"` // Composite index
	ReferenceType string    `gorm:"size:100;not null;index:idx_referee_reference_id_type,unique"`
	Referrer      *Referrer `gorm:"foreignKey:ReferrerID" json:"referrer,omitempty"`
}

func (Referee) TableName() string {
	return "referral_referee"
}

type EventLog struct {
	gorm.Model
	EventKey      string           `gorm:"not null;index" foreignKey:"Key" references:"Event"`
	ReferenceID   *string          `gorm:"index"` // Composite index
	ReferenceType *string          `gorm:"index"`
	Amount        *decimal.Decimal `gorm:"type:decimal(38,18);;not null;index"`
	TriggeredAt   time.Time        `gorm:"not null;index"`                     // Timestamp when the event was triggered
	Data          *string          `gorm:"type:json;not null"`                 // Arbitrary data associated with the event (JSON format)
	Status        string           `gorm:"size:50;default:'pending';not null"` // Status of the event processing (e.g., "pending", "processed", "failed")
	FailureReason *string          `gorm:"type:text"`
}

func (EventLog) TableName() string {
	return "referral_event_logs"
}

type Reward struct {
	gorm.Model
	CampaignID    uint            `gorm:"not null;index"` // Foreign key to Campaign
	RefereeID     uint            `gorm:"not null;index"` // Foreign key to Referee
	RefereeType   string          `gorm:"size:100;not null;index"`
	ReferenceID   string          `gorm:"index"`                              // ReferenceID of the entity related to the reward
	ReferenceType string          `gorm:"index"`                              // ReferenceType of the entity related to the reward
	Amount        decimal.Decimal `gorm:"type:decimal(38,18);not null"`       // Calculated reward amount
	Status        string          `gorm:"size:50;default:'pending';not null"` // Reward status (e.g., "pending", "processed", "failed")
	Reason        *string         `gorm:"type:text"`                          // Reason for reward status (optional)
}

func (Reward) TableName() string {
	return "referral_rewards"
}
