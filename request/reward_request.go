package request

import "gorm.io/gorm"

type GetRewardRequest struct {
	Projects             []string             `form:"projects"`            // Filter by name
	IDs                  []uint               `form:"ids"`                 // Filter by ID
	RefereeID            *uint                `form:"refereeID"`           // Filter by ID
	RefereeReferenceID   *string              `form:"refereeReferenceID"`  // Composite key with Project
	ReferrerID           *uint                `form:"referrerID"`          // Filter by ID
	ReferrerReferenceID  *string              `form:"referrerReferenceID"` // Composite key with Project
	ReferrerCode         *string              `form:"referrerCode"`        // Composite key with Project
	CurrencyCode         *string              `json:"currencyCode"`
	Status               *string              `form:"status"`               // Composite key with Project
	CampaignIDs          []uint               `form:"campaignIDs"`          // Filter by ID
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

func ApplyGetRewardRequest(req GetRewardRequest, query *gorm.DB) *gorm.DB {
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("referral_rewards.project IN (?)", req.Projects)
	}
	if req.IDs != nil && len(req.IDs) > 0 {
		query = query.Where("referral_rewards.id IN (?)", req.IDs)
	}
	if req.CampaignIDs != nil && len(req.CampaignIDs) > 0 {
		query = query.Where("referral_rewards.campaign_id IN (?)", req.CampaignIDs)
	}
	if req.RefereeID != nil {
		query = query.Where("referral_rewards.referee_id = ?", *req.RefereeID)
	}
	if req.RefereeReferenceID != nil {
		query = query.Where("referral_rewards.referee_reference_id = ?", *req.RefereeReferenceID)
	}
	if req.ReferrerID != nil {
		query = query.Where("referral_rewards.referrer_id = ?", *req.ReferrerID)
	}
	if req.ReferrerReferenceID != nil {
		query = query.Where("referral_rewards.referrer_reference_id = ?", *req.ReferrerReferenceID)
	}
	if req.ReferrerCode != nil {
		query = query.Where("referral_rewards.referrer_code = ?", *req.ReferrerCode)
	}
	if req.CurrencyCode != nil {
		query = query.Where("referral_rewards.currency_code = ?", *req.CurrencyCode)
	}
	if req.Status != nil {
		query = query.Where("referral_rewards.status = ?", *req.Status)
	}
	return query
}
