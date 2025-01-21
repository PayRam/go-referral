package request

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

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

func ApplyGetCampaignRequest(req GetCampaignsRequest, query *gorm.DB) *gorm.DB {
	// Apply filters with table name prepended
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("referral_campaigns.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("referral_campaigns.id = ?", *req.ID)
	}
	if req.Name != nil {
		query = query.Where("referral_campaigns.name LIKE ?", "%"+*req.Name+"%")
	}
	if req.Status != nil {
		query = query.Where("referral_campaigns.status = ?", *req.Status)
	}
	if req.IsDefault != nil {
		query = query.Where("referral_campaigns.is_default = ?", *req.IsDefault)
	}
	if req.StartDateMin != nil {
		query = query.Where("referral_campaigns.start_date >= ?", *req.StartDateMin)
	}
	if req.StartDateMax != nil {
		query = query.Where("referral_campaigns.start_date <= ?", *req.StartDateMax)
	}
	if req.EndDateMin != nil {
		query = query.Where("referral_campaigns.end_date >= ?", *req.EndDateMin)
	}
	if req.EndDateMax != nil {
		query = query.Where("referral_campaigns.end_date <= ?", *req.EndDateMax)
	}
	return query
}
