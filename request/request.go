package request

import (
	"fmt"
	"github.com/shopspring/decimal"
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

type CreateEventRequest struct {
	Key         string  `json:"key" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	EventType   string  `json:"eventType" binding:"required"` // e.g., "simple", "payment"
	Description *string `json:"description"`
}

type UpdateEventRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateCampaignRequest struct {
	Name               string           `json:"name" binding:"required"`
	RewardType         string           `json:"rewardType" binding:"required"` // e.g., "flat_fee", "percentage"
	RewardValue        decimal.Decimal  `json:"rewardValue" binding:"required"`
	RewardCap          *decimal.Decimal `json:"rewardCap"`
	InviteeRewardType  *string          `json:"inviteeRewardType"` // e.g., "flat_fee", "percentage"
	InviteeRewardValue *decimal.Decimal `json:"inviteeRewardValue"`
	InviteeRewardCap   *decimal.Decimal `json:"inviteeRewardCap"` // Cap for invitee reward
	Budget             *decimal.Decimal `json:"budget"`           // Optional budget
	Description        *string          `json:"description"`
	StartDate          *time.Time       `json:"startDate"`
	EndDate            *time.Time       `json:"endDate"`
	IsDefault          bool             `json:"isDefault"` // Only one default campaign

	CampaignTypePerCustomer   string           `json:"campaignTypePerCustomer" binding:"required"` // Campaign type: "one_time", "forever", "months_per_customer", "count_per_customer"
	ValidityMonthsPerCustomer *int             `json:"validityMonthsPerCustomer"`                  // For "months_per_customer"
	MaxOccurrencesPerCustomer *int64           `json:"maxOccurrencesPerCustomer"`                  // For "count_per_customer"
	RewardCapPerCustomer      *decimal.Decimal `json:"rewardCapPerCustomer"`                       // Maximum reward for percentage type

	EventKeys []string `json:"eventKeys"`
}

type UpdateCampaignRequest struct {
	Name               *string          `json:"name"`
	RewardType         *string          `json:"rewardType"` // e.g., "flat_fee", "percentage"
	RewardValue        *decimal.Decimal `json:"rewardValue"`
	RewardCap          *decimal.Decimal `json:"rewardCap"`
	InviteeRewardType  *string          `json:"inviteeRewardType"` // e.g., "flat_fee", "percentage"
	InviteeRewardValue *decimal.Decimal `json:"inviteeRewardValue"`
	InviteeRewardCap   *decimal.Decimal `json:"inviteeRewardCap"` // Cap for invitee reward
	Budget             *decimal.Decimal `json:"budget"`           // Optional budget
	Description        *string          `json:"description"`
	StartDate          *time.Time       `json:"startDate"`
	EndDate            *time.Time       `json:"endDate"`
	Status             *string          `json:"status"`
	IsDefault          *bool            `json:"isDefault"` // Only one default campaign

	CampaignTypePerCustomer   *string          `json:"campaignTypePerCustomer" binding:"required"` // Campaign type: "one_time", "forever", "months_per_customer", "count_per_customer"
	ValidityMonthsPerCustomer *int             `json:"validityMonthsPerCustomer"`                  // For "months_per_customer"
	MaxOccurrencesPerCustomer *int64           `json:"maxOccurrencesPerCustomer"`                  // For "count_per_customer"
	RewardCapPerCustomer      *decimal.Decimal `json:"rewardCapPerCustomer"`                       // Maximum reward for percentage type

	EventKeys []string `json:"eventKeys"`
}

type CreateReferrerRequest struct {
	ReferenceID string  `json:"referenceID" binding:"required"`
	Code        *string `json:"code"`
	CampaignIDs []uint  `json:"campaignIDs"`
}

type CreateEventLogRequest struct {
	EventKey    string           `json:"eventKey" binding:"required"`
	ReferenceID string           `json:"referenceID" binding:"required"`
	Amount      *decimal.Decimal `json:"amount"`
	Data        *string          `json:"data"`
}

type UpdateReferrerRequest struct {
	CampaignIDs []uint `json:"campaignIDs"`
}

type CreateRefereeRequest struct {
	ReferenceID string `json:"referenceID" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type GetEventsRequest struct {
	Project              *string              `json:"project"`              // Filter by name
	ID                   *uint                `json:"id"`                   // Filter by ID
	Key                  *string              `json:"key"`                  // Filter by name
	Name                 *string              `json:"name"`                 // Filter by name
	EventType            *string              `json:"eventType"`            // Filter by name
	PaginationConditions PaginationConditions `json:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetCampaignsRequest struct {
	Project              *string              `json:"project"`   // Filter by name
	ID                   *uint                `json:"id"`        // Filter by ID
	Name                 *string              `json:"name"`      // Filter by name
	Status               *string              `json:"status"`    // Filter by active status
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

type GetEventLogRequest struct {
	Project              *string              `json:"project"` // Filter by name
	ID                   *uint                `json:"id"`      // Filter by ID
	EventKey             *string              `json:"eventKey"`
	ReferenceID          *string              `json:"referenceID"`
	Status               *string              `json:"status"`               // Composite key with Project
	RewardID             *uint                `json:"rewardID"`             // Nullable to allow logs without an associated reward
	PaginationConditions PaginationConditions `json:"paginationConditions"` // Embedded pagination and sorting struct
}
