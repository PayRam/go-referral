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
	RewardType         *string          `gorm:"size:50" json:"rewardType"` // e.g., "flat_fee", "percentage"
	RewardValue        *decimal.Decimal `gorm:"" json:"rewardValue"`       // Percentage value or flat fee
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

type Member struct {
	BaseModel
	Project     string  `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference_id" json:"project"`
	ReferenceID string  `gorm:"size:100;not null;uniqueIndex:idx_referrer_project_reference_id" json:"referenceId"`
	Email       *string `gorm:"size:100;" json:"email"`
	Code        string  `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Status      string  `gorm:"size:50;default:'active';index" json:"status"`

	ReferredByMemberID *uint   `gorm:"index" json:"referredByMemberID"` // Nullable, points to another Member
	ReferredByMember   *Member `gorm:"foreignKey:ReferredByMemberID" json:"referredByMember,omitempty"`

	Campaigns []Campaign `gorm:"many2many:referral_member_campaigns;joinForeignKey:MemberID;joinReferences:CampaignID" json:"campaigns"`
}

func (Member) TableName() string {
	return "referral_members"
}

type MemberCampaign struct {
	Project    string   `gorm:"not null;size:100;" json:"project"`
	MemberID   uint     `gorm:"not null;uniqueIndex:idx_referral_referrer_campaign" json:"memberID"`
	CampaignID uint     `gorm:"not null;uniqueIndex:idx_referral_referrer_campaign" json:"campaignID"`
	Campaign   Campaign `gorm:"foreignKey:CampaignID;references:ID" json:"campaign"`
	Member     Member   `gorm:"foreignKey:MemberID;references:ID" json:"member"`
}

func (MemberCampaign) TableName() string {
	return "referral_member_campaigns"
}

type EventLog struct {
	BaseModel
	Project           string           `gorm:"size:100;not null;index" json:"project"`
	EventKey          string           `gorm:"size:100;not null;index" foreignKey:"Key" references:"Event" json:"eventKey"`
	MemberID          uint             `gorm:"not null:index" json:"memberID"`
	MemberReferenceID string           `gorm:"size:100;not null;index" json:"memberReferenceID"`
	Amount            *decimal.Decimal `gorm:"type:decimal(38,18);index" json:"amount"`
	TriggeredAt       time.Time        `gorm:"not null;index" json:"triggeredAt"`
	Data              *string          `gorm:"type:json;" json:"data"`
	Status            string           `gorm:"size:50;default:'pending';not null;index" json:"status"`
	FailureReason     *string          `gorm:"type:text" json:"failureReason"`
	ReferredRewardID  *uint            `gorm:"index" json:"referredRewardID"`
	RefereeRewardID   *uint            `gorm:"index" json:"refereeRewardID"`

	Member         *Member `gorm:"foreignKey:MemberID;references:ID" json:"member"`
	ReferredReward *Reward `gorm:"foreignKey:ReferredRewardID" json:"referredReward"`
	RefereeReward  *Reward `gorm:"foreignKey:RefereeRewardID" json:"refereeReward"`
}

func (EventLog) TableName() string {
	return "referral_event_logs"
}

type Reward struct {
	BaseModel
	Project                   string          `gorm:"size:100;not null;index" json:"project"`
	CampaignID                uint            `gorm:"not null;index" json:"campaignId"`
	CurrencyCode              string          `gorm:"type:varchar(20);not null;index" json:"currencyCode"`
	RewardedMemberID          uint            `gorm:"not null;index" json:"rewardedMemberID"`
	RewardedMemberReferenceID string          `gorm:"size:100;not null;index" json:"rewardedMemberReferenceID"`
	RelatedMemberID           uint            `gorm:"not null;index" json:"relatedMemberID"`
	RelatedMemberReferenceID  string          `gorm:"size:100;not null;index" json:"relatedMemberReferenceID"`
	MemberType                string          `gorm:"size:50;not null;index" json:"memberType"`
	Amount                    decimal.Decimal `gorm:"type:decimal(38,18);not null;index" json:"amount"`
	Status                    string          `gorm:"size:50;default:'pending';not null;index" json:"status"`
	Reason                    *string         `gorm:"type:text" json:"reason"`

	RewardedMember *Member `gorm:"foreignKey:RewardedMemberID;references:ID" json:"rewardedMember,omitempty"`
	RelatedMember  *Member `gorm:"foreignKey:RelatedMemberID;references:ID" json:"relatedMember,omitempty"`
}

func (Reward) TableName() string {
	return "referral_rewards"
}
