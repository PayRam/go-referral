package request

import "gorm.io/gorm"

type GetRewardRequest struct {
	Projects                    []string             `form:"projects"`                    // Filter by name
	IDs                         []uint               `form:"ids"`                         // Filter by ID
	RelatedCustomerID           *uint                `form:"relatedCustomerID"`           // Filter by ID
	RelatedCustomerReferenceID  *string              `form:"relatedCustomerReferenceID"`  // Composite key with Project
	RewardedCustomerID          *uint                `form:"rewardedCustomerID"`          // Filter by ID
	RewardedCustomerReferenceID *string              `form:"rewardedCustomerReferenceID"` // Composite key with Project
	CurrencyCode                *string              `json:"currencyCode"`
	Status                      *string              `form:"status"`               // Composite key with Project
	CampaignIDs                 []uint               `form:"campaignIDs"`          // Filter by ID
	PaginationConditions        PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
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
	if req.RelatedCustomerID != nil {
		query = query.Where("referral_rewards.related_customer_id = ?", *req.RelatedCustomerID)
	}
	if req.RelatedCustomerReferenceID != nil {
		query = query.Where("referral_rewards.related_customer_reference_id = ?", *req.RelatedCustomerReferenceID)
	}
	if req.RewardedCustomerID != nil {
		query = query.Where("referral_rewards.rewarded_customer_id = ?", *req.RewardedCustomerID)
	}
	if req.RewardedCustomerReferenceID != nil {
		query = query.Where("referral_rewards.rewarded_customer_reference_id = ?", *req.RewardedCustomerReferenceID)
	}
	if req.CurrencyCode != nil {
		query = query.Where("referral_rewards.currency_code = ?", *req.CurrencyCode)
	}
	if req.Status != nil {
		query = query.Where("referral_rewards.status = ?", *req.Status)
	}
	return query
}
