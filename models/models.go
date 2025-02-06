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
	Project            string           `gorm:"size:100;not null;index" json:"project"`
	Name               string           `gorm:"size:255;not null;index" json:"name"`
	RewardType         string           `gorm:"size:50" json:"rewardType"` // e.g., "flat_fee", "percentage"
	RewardValue        decimal.Decimal  `gorm:"" json:"rewardValue"`       // Percentage value or flat fee
	CurrencyCode       string           `gorm:"type:varchar(20);default:'USD';index" json:"currencyCode"`
	RewardCap          *decimal.Decimal `gorm:"type:decimal(38,18)" json:"rewardCap"`         // Maximum reward for percentage type
	InviteeRewardType  *string          `gorm:"size:50" json:"inviteeRewardType"`             // e.g., "flat_fee", "percentage"
	InviteeRewardValue *decimal.Decimal `gorm:"" json:"inviteeRewardValue"`                   // Reward for invitee
	InviteeRewardCap   *decimal.Decimal `gorm:"type:decimal(38,18)" json:"inviteeRewardCap"`  // Cap for invitee reward
	Budget             *decimal.Decimal `gorm:"type:decimal(38,18)" json:"budget"`            // Budget for the campaign
	Description        *string          `gorm:"type:text" json:"description"`                 // Optional description
	StartDate          *time.Time       `gorm:"not null;index" json:"startDate"`              // Start date of the campaign
	EndDate            *time.Time       `gorm:"not null;index" json:"endDate"`                // End date of the campaign
	Status             string           `gorm:"size:50;default:'active';index" json:"status"` // New field to track campaign status (e.g., 'active', 'paused', 'archived')
	IsDefault          bool             `gorm:"default:false;index" json:"isDefault"`         // Only one default campaign

	CampaignTypePerCustomer   string           `gorm:"size:50;not null;index" json:"campaignTypePerCustomer"` // Campaign type: "one_time", "forever", "months_per_customer", "count_per_customer"
	MaxOccurrencesPerCustomer *int64           `gorm:"" json:"maxOccurrencesPerCustomer"`                     // 0 for unlimited
	ValidityMonthsPerCustomer *int             `gorm:"" json:"validityMonthsPerCustomer"`                     // 0 for no time limit
	RewardCapPerCustomer      *decimal.Decimal `gorm:"type:decimal(38,18)" json:"rewardCapPerCustomer"`       // Maximum reward for percentage type

	Events []Event `gorm:"many2many:referral_campaign_events" json:"events"` // Associated events
}

func (Campaign) TableName() string {
	return "referral_campaigns"
}

// Event represents an action within a campaign that can trigger a reward
type Event struct {
	BaseModel
	Project     string  `gorm:"size:100;not null;uniqueIndex:idx_event_project_key" seeder:"no-update" json:"project"`
	Key         string  `gorm:"size:100;not null;uniqueIndex:idx_event_project_key" seeder:"no-update" json:"key"`
	Name        string  `gorm:"size:255;not null;index" seeder:"no-update" json:"name"`
	EventType   string  `gorm:"size:100;not null;index" seeder:"no-update" json:"eventType"`
	Description *string `gorm:"type:text" seeder:"no-update" json:"description"`
}

func (Event) TableName() string {
	return "referral_events"
}

type CampaignEvent struct {
	Project    string   `gorm:"not null;size:100;index" json:"project"`
	CampaignID uint     `gorm:"not null;uniqueIndex:idx_campaign_id_event_id" json:"campaignId"`
	EventID    uint     `gorm:"not null;uniqueIndex:idx_campaign_id_event_id" json:"eventId"`
	EventKey   string   `gorm:"not null;size:100;index" json:"eventKey"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID" json:"campaign"`
	Event      Event    `gorm:"foreignKey:EventID;references:ID" json:"event"`
}

func (CampaignEvent) TableName() string {
	return "referral_campaign_events"
}

type Referrer struct {
	BaseModel
	Project     string     `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference_id" json:"project"`
	ReferenceID string     `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference_id" json:"referenceId"`
	Email       *string    `gorm:"size:100;" json:"email"`
	Code        string     `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Status      string     `gorm:"size:50;default:'active';index" json:"status"`
	Campaigns   []Campaign `gorm:"many2many:referral_referrer_campaigns;joinForeignKey:ReferrerID;joinReferences:CampaignID" json:"campaigns"`
}

func (Referrer) TableName() string {
	return "referral_referrer"
}

type ReferrerCampaign struct {
	Project    string   `gorm:"not null;size:100;" json:"project"`
	ReferrerID uint     `gorm:"not null;uniqueIndex:idx_referral_referrer_campaign" json:"referrerId"`
	CampaignID uint     `gorm:"not null;uniqueIndex:idx_referral_referrer_campaign" json:"campaignId"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID" json:"campaign"`
	Referrer   Referrer `gorm:"foreignKey:ReferrerID;references:ID" json:"referrer"`
}

func (ReferrerCampaign) TableName() string {
	return "referral_referrer_campaigns"
}

type Referee struct {
	BaseModel
	Project             string    `gorm:"size:100;not null;uniqueIndex:idx_referee_project_reference" json:"project"`
	ReferenceID         string    `gorm:"size:100;not null;uniqueIndex:idx_referee_project_reference" json:"referenceId"`
	ReferrerID          uint      `gorm:"not null;uniqueIndex:idx_referee_project_reference" json:"referrerId"`
	ReferrerReferenceID string    `gorm:"size:100;not null;uniqueIndex:idx_referee_project_reference" json:"referrerReferenceId"`
	Email               *string   `gorm:"size:100;" json:"email"`
	Referrer            *Referrer `gorm:"foreignKey:ReferrerID" json:"referrer"`
}

func (Referee) TableName() string {
	return "referral_referee"
}

type EventLog struct {
	BaseModel
	Project       string           `gorm:"size:100;not null;index" json:"project"`
	EventKey      string           `gorm:"size:100;not null;index" foreignKey:"Key" references:"Event" json:"eventKey"`
	ReferenceID   string           `gorm:"size:100;not null;index" json:"referenceId"`
	Amount        *decimal.Decimal `gorm:"type:decimal(38,18);index" json:"amount"`
	TriggeredAt   time.Time        `gorm:"not null;index" json:"triggeredAt"`
	Data          *string          `gorm:"type:json;" json:"data"`
	Status        string           `gorm:"size:50;default:'pending';not null;index" json:"status"`
	FailureReason *string          `gorm:"type:text" json:"failureReason"`
	RewardID      *uint            `gorm:"index" json:"rewardId"`
	Reward        *Reward          `gorm:"foreignKey:RewardID" json:"reward"`
}

func (EventLog) TableName() string {
	return "referral_event_logs"
}

type Reward struct {
	BaseModel
	Project             string          `gorm:"size:100;not null;index" json:"project"`
	CampaignID          uint            `gorm:"not null;index" json:"campaignId"`
	CurrencyCode        string          `gorm:"type:varchar(20);not null;index" json:"currencyCode"`
	RefereeID           uint            `gorm:"not null;index" json:"refereeId"`
	RefereeReferenceID  string          `gorm:"size:100;not null;index" json:"refereeReferenceId"`
	ReferrerID          uint            `gorm:"not null;index" json:"referrerId"`
	ReferrerReferenceID string          `gorm:"size:100;index" json:"referrerReferenceId"`
	ReferrerCode        string          `gorm:"size:50;not null;index" json:"referrerCode"`
	Amount              decimal.Decimal `gorm:"type:decimal(38,18);not null;index" json:"amount"`
	Status              string          `gorm:"size:50;default:'pending';not null;index" json:"status"`
	Reason              *string         `gorm:"type:text" json:"reason"`
	EventLogs           []EventLog      `gorm:"foreignKey:RewardID" json:"eventLogs"`
}

func (Reward) TableName() string {
	return "referral_rewards"
}
