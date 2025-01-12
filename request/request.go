package request

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type PaginationConditions struct {
	Limit         *int    `json:"limit"`           // Pagination limit
	Offset        *int    `json:"offset"`          // Pagination offset (optional when using ID-based)
	SortBy        *string `json:"sort_by"`         // Field to sort by
	Order         *string `json:"order"`           // ASC or DESC
	GreaterThanID *uint   `json:"greater_than_id"` // For ID-based pagination
	LessThanID    *uint   `json:"less_than_id"`    // For reverse ID-based pagination
}

func ApplyPaginationConditions(query *gorm.DB, conditions PaginationConditions) *gorm.DB {
	// Count total records (optional based on use case)
	if conditions.Offset != nil && *conditions.Offset > 0 {
		query = query.Offset(*conditions.Offset)
	}

	// Apply ID-based pagination
	if conditions.GreaterThanID != nil {
		query = query.Where("id > ?", *conditions.GreaterThanID)
	}
	if conditions.LessThanID != nil {
		query = query.Where("id < ?", *conditions.LessThanID)
	}

	// Apply sorting
	sortBy := "id"
	if conditions.SortBy != nil {
		sortBy = *conditions.SortBy
	}
	order := "DESC"
	if conditions.Order != nil {
		order = *conditions.Order
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, order))

	// Apply limit
	if conditions.Limit != nil && *conditions.Limit > 0 {
		query = query.Limit(*conditions.Limit)
	}

	return query
}

type UpdateEventRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	EventType   *string `json:"eventType"` // e.g., "simple", "payment"
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

type GetCampaignsRequest struct {
	Project              *string              `json:"project"`   // Filter by name
	ID                   *uint                `json:"id"`        // Filter by ID
	Name                 *string              `json:"name"`      // Filter by name
	IsActive             *bool                `json:"isActive"`  // Filter by active status
	IsDefault            *bool                `json:"isDefault"` // Filter by active status
	StartDateMin         *time.Time           `json:"startDateMin"`
	StartDateMax         *time.Time           `json:"startDateMax"`
	EndDateMin           *time.Time           `json:"endDateMin"`
	EndDateMax           *time.Time           `json:"endDateMax"`
	PaginationConditions PaginationConditions `json:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetReferrerRequest struct {
	Project              *string              `json:"project"`     // Filter by name
	ID                   *uint                `json:"id"`          // Filter by ID
	ReferenceID          *string              `json:"referenceID"` // Composite key with Project
	Code                 *string              `json:"code"`
	PaginationConditions PaginationConditions `json:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetRefereeRequest struct {
	Project              *string              `json:"project"`              // Filter by name
	ID                   *uint                `json:"id"`                   // Filter by ID
	ReferenceID          *string              `json:"referenceID"`          // Composite key with Project
	ReferrerReferenceID  *string              `json:"referrerReferenceID"`  // Composite key with Project
	ReferrerID           *uint                `json:"referrerID"`           // ID of the Referrer (Foreign Key to Referrer table)
	PaginationConditions PaginationConditions `json:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetRewardRequest struct {
	Project              *string              `json:"project"`              // Filter by name
	ID                   *uint                `json:"id"`                   // Filter by ID
	CampaignID           *uint                `json:"campaignID"`           // Filter by ID
	RefereeID            *uint                `json:"refereeID"`            // Filter by ID
	RefereeReferenceID   *string              `json:"refereeReferenceID"`   // Composite key with Project
	ReferrerID           *uint                `json:"referrerID"`           // Filter by ID
	ReferrerReferenceID  *string              `json:"referrerReferenceID"`  // Composite key with Project
	ReferrerCode         *string              `json:"referrerCode"`         // Composite key with Project
	Status               *string              `json:"status"`               // Composite key with Project
	PaginationConditions PaginationConditions `json:"paginationConditions"` // Embedded pagination and sorting struct
}

//Amount              decimal.Decimal `gorm:"type:decimal(38,18);not null"`       // Calculated reward amount
//Reason              *string         `gorm:"type:text"`
