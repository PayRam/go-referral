package request

import (
	"fmt"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type PaginationConditions struct {
	Limit         *int       `form:"limit"`         // Pagination limit
	Offset        *int       `form:"offset"`        // Pagination offset (optional when using ID-based)
	SortBy        *string    `form:"sortBy"`        // Field to sort by
	Order         *string    `form:"order"`         // ASC or DESC
	GreaterThanID *uint      `form:"greaterThanID"` // For ID-based pagination
	LessThanID    *uint      `form:"lessThanID"`    // For reverse ID-based pagination
	CreatedAfter  *time.Time `form:"createdAfter"`  // Filter records created after this date
	CreatedBefore *time.Time `form:"createdBefore"` // Filter records created before this date
	UpdatedAfter  *time.Time `form:"updatedAfter"`  // Filter records updated after this date
	UpdatedBefore *time.Time `form:"updatedBefore"` // Filter records updated before this date
	StartDate     *time.Time `form:"startDate"`     // Filter records created after this date
	EndDate       *time.Time `form:"endDate"`       // Filter records created after this date
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

	// Apply date filters
	if conditions.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *conditions.CreatedAfter)
	}
	if conditions.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *conditions.CreatedBefore)
	}
	if conditions.UpdatedAfter != nil {
		query = query.Where("updated_at >= ?", *conditions.UpdatedAfter)
	}
	if conditions.UpdatedBefore != nil {
		query = query.Where("updated_at <= ?", *conditions.UpdatedBefore)
	}

	if conditions.StartDate != nil {
		query = query.Where("created_at >= ?", *conditions.StartDate)
	}
	if conditions.EndDate != nil {
		query = query.Where("created_at <= ?", *conditions.EndDate)
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
	Email       *string `json:"email"`
}

type CreateEventLogRequest struct {
	EventKey    string           `json:"eventKey" binding:"required"`
	ReferenceID string           `json:"referenceID" binding:"required"`
	Amount      *decimal.Decimal `json:"amount"`
	Data        *string          `json:"data"`
}

type UpdateReferrerRequest struct {
	CampaignIDs []uint  `json:"campaignIDs"`
	Email       *string `json:"email"`
}

type CreateRefereeRequest struct {
	ReferenceID string  `json:"referenceID" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Email       *string `json:"email"`
}

type GetEventsRequest struct {
	Projects             []string             `form:"projects"`             // Filter by name
	ID                   *uint                `form:"id"`                   // Filter by ID
	Key                  *string              `form:"key"`                  // Filter by name
	Name                 *string              `form:"name"`                 // Filter by name
	EventType            *string              `form:"eventType"`            // Filter by name
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetCampaignsRequest struct {
	Projects             []string             `form:"projects"`  // Filter by name
	ID                   *uint                `form:"id"`        // Filter by ID
	Name                 *string              `form:"name"`      // Filter by name
	Status               *string              `form:"status"`    // Filter by active status
	IsDefault            *bool                `form:"isDefault"` // Filter by active status
	StartDateMin         *time.Time           `form:"startDateMin"`
	StartDateMax         *time.Time           `form:"startDateMax"`
	EndDateMin           *time.Time           `form:"endDateMin"`
	EndDateMax           *time.Time           `form:"endDateMax"`
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetReferrerRequest struct {
	Projects             []string             `form:"projects"`    // Filter by name
	ID                   *uint                `form:"id"`          // Filter by ID
	ReferenceID          *string              `form:"referenceID"` // Composite key with Project
	Email                *string              `form:"email"`       // Composite key with Project
	Code                 *string              `form:"code"`
	CampaignIDs          []uint               `form:"campaignIDs"`
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetRefereeRequest struct {
	Projects             []string             `form:"projects"`             // Filter by name
	ID                   *uint                `form:"id"`                   // Filter by ID
	ReferenceID          *string              `form:"referenceID"`          // Composite key with Project
	ReferrerReferenceID  *string              `form:"referrerReferenceID"`  // Composite key with Project
	ReferrerID           *uint                `form:"referrerID"`           // ID of the Referrer (Foreign Key to Referrer table)
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetRewardRequest struct {
	Projects             []string             `form:"projects"`             // Filter by name
	ID                   *uint                `form:"id"`                   // Filter by ID
	RefereeID            *uint                `form:"refereeID"`            // Filter by ID
	RefereeReferenceID   *string              `form:"refereeReferenceID"`   // Composite key with Project
	ReferrerID           *uint                `form:"referrerID"`           // Filter by ID
	ReferrerReferenceID  *string              `form:"referrerReferenceID"`  // Composite key with Project
	ReferrerCode         *string              `form:"referrerCode"`         // Composite key with Project
	Status               *string              `form:"status"`               // Composite key with Project
	CampaignIDs          []uint               `form:"campaignIDs"`          // Filter by ID
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

type GetEventLogRequest struct {
	Projects             []string             `form:"projects"` // Filter by name
	ID                   *uint                `form:"id"`       // Filter by ID
	EventKey             *string              `form:"eventKey"`
	ReferenceID          *string              `form:"referenceID"`
	Status               *string              `form:"status"`               // Composite key with Project
	RewardID             *uint                `form:"rewardID"`             // Nullable to allow logs without an associated reward
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}
