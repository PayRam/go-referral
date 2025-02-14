package request

import "gorm.io/gorm"

type GetRewardRequest struct {
	Projects                  []string             `form:"projects"`                  // Filter by name
	IDs                       []uint               `form:"ids"`                       // Filter by ID
	RelatedMemberID           *uint                `form:"relatedMemberID"`           // Filter by ID
	RelatedMemberReferenceID  *string              `form:"relatedMemberReferenceID"`  // Composite key with Project
	RewardedMemberID          *uint                `form:"rewardedMemberID"`          // Filter by ID
	RewardedMemberReferenceID *string              `form:"rewardedMemberReferenceID"` // Composite key with Project
	CurrencyCode              *string              `json:"currencyCode"`
	Status                    *string              `form:"status"`               // Composite key with Project
	CampaignIDs               []uint               `form:"campaignIDs"`          // Filter by ID
	PaginationConditions      PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
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
	if req.RelatedMemberID != nil {
		query = query.Where("referral_rewards.related_member_id = ?", *req.RelatedMemberID)
	}
	if req.RelatedMemberReferenceID != nil {
		query = query.Where("referral_rewards.related_member_reference_id = ?", *req.RelatedMemberReferenceID)
	}
	if req.RewardedMemberID != nil {
		query = query.Where("referral_rewards.rewarded_member_id = ?", *req.RewardedMemberID)
	}
	if req.RewardedMemberReferenceID != nil {
		query = query.Where("referral_rewards.rewarded_member_reference_id = ?", *req.RewardedMemberReferenceID)
	}
	if req.CurrencyCode != nil {
		query = query.Where("referral_rewards.currency_code = ?", *req.CurrencyCode)
	}
	if req.Status != nil {
		query = query.Where("referral_rewards.status = ?", *req.Status)
	}
	return query
}
