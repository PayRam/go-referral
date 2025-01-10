package request

import "time"

type UpdateEventRequest struct {
	Name      *string `json:"name"`
	EventType *string `json:"eventType"` // e.g., "simple", "payment"
}

type UpdateCampaignRequest struct {
	Name               *string    `json:"name"`
	RewardType         *string    `json:"rewardType"` // e.g., "flat_fee", "percentage"
	RewardValue        *float64   `json:"rewardValue"`
	MaxOccurrences     *uint      `json:"maxOccurrences"`
	ValidityDays       *uint      `json:"validityDays"`
	InviteeRewardType  *string    `json:"inviteeRewardType"` // e.g., "flat_fee", "percentage"
	InviteeRewardValue *float64   `json:"inviteeRewardValue"`
	Budget             *float64   `json:"budget"` // Optional budget
	Description        *string    `json:"description"`
	StartDate          *time.Time `json:"startDate"`
	EndDate            *time.Time `json:"endDate"`
	IsActive           *bool      `json:"isActive"`
}
