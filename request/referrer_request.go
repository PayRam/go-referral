package request

import "gorm.io/gorm"

type CreateReferrerRequest struct {
	ReferenceID string  `json:"referenceID" binding:"required"`
	Code        *string `json:"code"`
	CampaignIDs []uint  `json:"campaignIDs"`
	Email       *string `json:"email"`
}

type UpdateReferrerRequest struct {
	CampaignIDs []uint  `json:"campaignIDs"`
	Email       *string `json:"email"`
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

func ApplyGetReferrerRequest(req GetReferrerRequest, query *gorm.DB) *gorm.DB {
	// Apply filters with explicit table name
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("referral_referrer.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("referral_referrer.id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("referral_referrer.reference_id = ?", *req.ReferenceID)
	}
	if req.Email != nil {
		query = query.Where("referral_referrer.email = ?", *req.Email)
	}
	if req.Code != nil {
		query = query.Where("referral_referrer.code = ?", *req.Code)
	}
	if req.CampaignIDs != nil && len(req.CampaignIDs) > 0 {
		// Join with referral_referrer_campaigns table to filter by CampaignIDs
		query = query.Joins("JOIN referral_referrer_campaigns rc ON rc.referrer_id = referral_referrer.id").
			Where("rc.campaign_id IN (?)", req.CampaignIDs).
			Group("referral_referrer.id") // Avoid duplicates due to the JOIN
	}
	return query
}
