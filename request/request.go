package request

import "time"

type UpdateEventRequest struct {
	Name      *string `json:"name"`
	EventType *string `json:"event_type"` // e.g., "simple", "payment"
}

type UpdateCampaignRequest struct {
	Name           *string    `json:"name"`
	RewardType     *string    `json:"reward_type"` // e.g., "flat_fee", "percentage"
	RewardValue    *float64   `json:"reward_value"`
	MaxOccurrences *uint      `json:"max_occurrences"`
	ValidityDays   *uint      `json:"validity_days"`
	Budget         *float64   `json:"budget"` // Optional budget
	Description    *string    `json:"description"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	IsActive       *bool      `json:"is_active"`
}
