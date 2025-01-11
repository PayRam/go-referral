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

//type ProjectModel struct {
//	Project string `gorm:"size:100;not null;index"`
//}

type Campaign struct {
	BaseModel
	Project            string           `gorm:"size:100;not null;index"`
	Name               string           `gorm:"size:255;not null;index"`
	RewardType         *string          `gorm:"size:50;index"` // e.g., "flat_fee", "percentage"
	RewardValue        *float64         `gorm:""`
	MaxOccurrences     *uint            `gorm:"default:0"`     // 0 for unlimited
	ValidityDays       *uint            `gorm:"default:0"`     // 0 for no time limit
	InviteeRewardType  *string          `gorm:"size:50;index"` // e.g., "flat_fee", "percentage"
	InviteeRewardValue *float64         `gorm:""`
	Budget             *decimal.Decimal `gorm:"type:decimal(38,18)"` // Pointer to handle nil as unlimited
	Description        string           `gorm:"type:text"`
	StartDate          time.Time        `gorm:"not null;index"`
	EndDate            time.Time        `gorm:"not null;index"`
	IsActive           bool             `gorm:"default:true;index"`
	IsDefault          bool             `gorm:"default:false;index"` // Only one default campaign
	Events             []Event          `gorm:"many2many:referral_campaign_events"`
}

func (Campaign) TableName() string {
	return "referral_campaigns"
}

// Event represents an action within a campaign that can trigger a reward
type Event struct {
	Project     string     `gorm:"size:100;primaryKey;not null;"`
	Key         string     `gorm:"size:100;primaryKey;not null;" seeder:"key,no-update"`
	Name        string     `gorm:"size:255;not null"`
	Description *string    `gorm:"type:text"`
	EventType   string     `gorm:"size:100;not null;index"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" seeder:"no-update"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" seeder:"no-update"`
	DeletedAt   *time.Time `gorm:"index" json:"-" seeder:"no-update"`
}

func (Event) TableName() string {
	return "referral_events"
}

type CampaignEvent struct {
	CampaignID uint     `gorm:"not null;uniqueIndex:idx_campaign_event"`
	Project    string   `gorm:"not null;size:100;uniqueIndex:idx_campaign_event"`
	EventKey   string   `gorm:"not null;size:100;uniqueIndex:idx_campaign_event"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID"`
	Event      Event    `gorm:"foreignKey:Project,EventKey;references:Project,Key"`
}

func (CampaignEvent) TableName() string {
	return "referral_campaign_events"
}

type Referrer struct {
	BaseModel
	Project     string     `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference"` // Composite key with ReferenceID
	ReferenceID string     `gorm:"not null;uniqueIndex:idx_referrer_project_reference"`          // Composite key with Project
	Code        string     `gorm:"size:50;uniqueIndex;not null"`                                 // Unique referral code
	Campaigns   []Campaign `gorm:"many2many:referral_referrer_campaigns;joinForeignKey:ReferrerID;joinReferences:CampaignID"`
}

func (Referrer) TableName() string {
	return "referral_referrer"
}

type ReferrerCampaign struct {
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
	Project     string    `gorm:"size:100;not null;uniqueIndex:idx_referee_project_reference"` // Composite key with ReferenceID
	ReferenceID string    `gorm:"not null;uniqueIndex:idx_referee_project_reference"`          // Composite key with Project
	ReferrerID  uint      `gorm:"not null;uniqueIndex:idx_referee_project_reference"`          // ID of the Referrer (Foreign Key to Referrer table)
	Referrer    *Referrer `gorm:"foreignKey:ReferrerID" json:"referrer,omitempty"`
}

func (Referee) TableName() string {
	return "referral_referee"
}

type EventLog struct {
	BaseModel
	Project       string           `gorm:"size:100;not null;index"`
	EventKey      string           `gorm:"not null;index" foreignKey:"Key" references:"Event"`
	ReferenceID   string           `gorm:"index"`
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
	BaseModel
	Project     string          `gorm:"size:100;not null;index"`
	CampaignID  uint            `gorm:"not null;index"`                     // Foreign key to Campaign
	RefereeID   uint            `gorm:"not null;index"`                     // Foreign key to Referee
	ReferenceID string          `gorm:"index"`                              // ReferenceID of the entity related to the reward
	Amount      decimal.Decimal `gorm:"type:decimal(38,18);not null"`       // Calculated reward amount
	Status      string          `gorm:"size:50;default:'pending';not null"` // Reward status (e.g., "pending", "processed", "failed")
	Reason      *string         `gorm:"type:text"`                          // Reason for reward status (optional)
}

func (Reward) TableName() string {
	return "referral_rewards"
}
