package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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
	RewardType         *string          `gorm:"size:50" json:"rewardType"`              // e.g., "flat_fee", "percentage"
	RewardValue        *decimal.Decimal `gorm:"type:decimal(38,18)" json:"rewardValue"` // Percentage value or flat fee
	CurrencyCode       string           `gorm:"type:varchar(20);default:'USD';index" json:"currencyCode"`
	RewardCap          *decimal.Decimal `gorm:"type:decimal(38,18)" json:"rewardCap"`          // Maximum reward for percentage type
	InviteeRewardType  *string          `gorm:"size:50" json:"inviteeRewardType"`              // e.g., "flat_fee", "percentage"
	InviteeRewardValue *decimal.Decimal `gorm:"type:decimal(38,18)" json:"inviteeRewardValue"` // Reward for invitee
	InviteeRewardCap   *decimal.Decimal `gorm:"type:decimal(38,18)" json:"inviteeRewardCap"`   // Cap for invitee reward
	Budget             *decimal.Decimal `gorm:"type:decimal(38,18)" json:"budget"`             // Budget for the campaign
	Description        *string          `gorm:"type:text" json:"description"`                  // Optional description
	StartDate          *time.Time       `gorm:"not null;index" json:"startDate"`               // Start date of the campaign
	EndDate            *time.Time       `gorm:"not null;index" json:"endDate"`                 // End date of the campaign
	Status             string           `gorm:"size:50;default:'active';index" json:"status"`  // New field to track campaign status (e.g., 'active', 'paused', 'archived')
	IsDefault          bool             `gorm:"default:false;index" json:"isDefault"`          // Only one default campaign

	CampaignTypePerCustomer   string           `gorm:"size:50;not null;index" json:"campaignTypePerCustomer"` // Campaign type: "one_time", "forever", "months_per_customer", "count_per_customer"
	MaxOccurrencesPerCustomer *int64           `gorm:"" json:"maxOccurrencesPerCustomer"`                     // 0 for unlimited
	ValidityMonthsPerCustomer *int             `gorm:"" json:"validityMonthsPerCustomer"`                     // 0 for no time limit
	RewardCapPerCustomer      *decimal.Decimal `gorm:"type:decimal(38,18)" json:"rewardCapPerCustomer"`       // Maximum reward for percentage type

	ConsiderEventsFrom time.Time `gorm:"not null;index" json:"considerEventsFrom"` // Timestamp for event consideration

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

type Customer struct {
	BaseModel
	Project     string  `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference_id" json:"project"`
	ReferenceID string  `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference_id" json:"referenceId"`
	Email       *string `gorm:"size:100;" json:"email"`
	Code        string  `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Status      string  `gorm:"size:50;default:'active';index" json:"status"`

	ReferredByCustomerID          *uint     `gorm:"index" json:"referredByCustomerID"`          // Nullable, points to another Customer
	ReferredByCustomerReferenceID *string   `gorm:"index" json:"referredByCustomerReferenceID"` // Nullable, points to another Customer
	ReferredByCustomer            *Customer `gorm:"foreignKey:ReferredByCustomerID" json:"referredByCustomer,omitempty"`

	Campaigns []Campaign `gorm:"many2many:referral_customer_campaigns;joinForeignKey:CustomerID;joinReferences:CampaignID" json:"campaigns"`
}

func (Customer) TableName() string {
	return "referral_members"
}

type CustomerCampaign struct {
	Project    string   `gorm:"not null;size:100;" json:"project"`
	CustomerID uint     `gorm:"not null;uniqueIndex:idx_referral_referrer_campaign" json:"customerID"`
	CampaignID uint     `gorm:"not null;uniqueIndex:idx_referral_referrer_campaign" json:"campaignID"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID" json:"campaign"`
	Customer   Customer `gorm:"foreignKey:CustomerID;references:ID" json:"customer"`
}

func (CustomerCampaign) TableName() string {
	return "referral_customer_campaigns"
}

type EventLog struct {
	BaseModel
	Project             string           `gorm:"size:100;not null;index" json:"project"`
	EventKey            string           `gorm:"size:100;not null;index" foreignKey:"Key" references:"Event" json:"eventKey"`
	CustomerID          uint             `gorm:"not null:index" json:"customerID"`
	CustomerReferenceID string           `gorm:"size:100;not null;index" json:"customerReferenceID"`
	Amount              *decimal.Decimal `gorm:"type:decimal(38,18);index" json:"amount"`
	TriggeredAt         time.Time        `gorm:"not null;index" json:"triggeredAt"`
	Data                *string          `gorm:"type:json;" json:"data"`
	Status              string           `gorm:"size:50;default:'pending';not null;index" json:"status"`
	FailureReason       *string          `gorm:"type:text" json:"failureReason"`

	Customer *Customer `gorm:"foreignKey:CustomerID;references:ID" json:"customer"`
}

func (EventLog) TableName() string {
	return "referral_event_logs"
}

type CampaignEventLog struct {
	BaseModel
	Project             string `gorm:"size:100;not null;index" json:"project"`
	CampaignID          uint   `gorm:"not null;index" json:"campaignID"` // The campaign the event is associated with
	EventID             uint   `gorm:"not null;index" json:"eventID"`    // The event being tracked
	CustomerID          uint   `gorm:"not null;index" json:"customerID"` // The customer who triggered the event
	CustomerReferenceID string `gorm:"size:100;not null;index" json:"customerReferenceID"`
	Status              string `gorm:"size:50;default:'pending';not null;index" json:"status"` // 'pending', 'completed'
	EventLogID          uint   `gorm:"not null;index" json:"eventLogID"`
	ReferredRewardID    *uint  `gorm:"index" json:"referredRewardID"`
	RefereeRewardID     *uint  `gorm:"index" json:"refereeRewardID"` // Reference to the original event log

	Campaign       *Campaign `gorm:"foreignKey:CampaignID" json:"campaign"`
	Event          *Event    `gorm:"foreignKey:EventID" json:"event"`
	Customer       *Customer `gorm:"foreignKey:CustomerID" json:"customer"`
	ReferredReward *Reward   `gorm:"foreignKey:ReferredRewardID" json:"referredReward"`
	RefereeReward  *Reward   `gorm:"foreignKey:RefereeRewardID" json:"refereeReward"`
}

func (CampaignEventLog) TableName() string {
	return "referral_campaign_event_logs"
}

type Reward struct {
	BaseModel
	Project                     string          `gorm:"size:100;not null;index" json:"project"`
	CampaignID                  uint            `gorm:"not null;index" json:"campaignId"`
	CurrencyCode                string          `gorm:"type:varchar(20);not null;index" json:"currencyCode"`
	RewardedCustomerID          uint            `gorm:"not null;index" json:"rewardedCustomerID"`
	RewardedCustomerReferenceID string          `gorm:"size:100;not null;index" json:"rewardedCustomerReferenceID"`
	RelatedCustomerID           uint            `gorm:"not null;index" json:"relatedCustomerID"`
	RelatedCustomerReferenceID  string          `gorm:"size:100;not null;index" json:"relatedCustomerReferenceID"`
	CustomerType                string          `gorm:"size:50;not null;index" json:"customerType"`
	Amount                      decimal.Decimal `gorm:"type:decimal(38,18);not null;index" json:"amount"`
	Status                      string          `gorm:"size:50;default:'pending';not null;index" json:"status"`
	Reason                      *string         `gorm:"type:text" json:"reason"`

	RewardedCustomer *Customer `gorm:"foreignKey:RewardedCustomerID;references:ID" json:"rewardedCustomer,omitempty"`
	RelatedCustomer  *Customer `gorm:"foreignKey:RelatedCustomerID;references:ID" json:"relatedCustomer,omitempty"`
}

func (Reward) TableName() string {
	return "referral_rewards"
}
