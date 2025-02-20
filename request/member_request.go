package request

import "gorm.io/gorm"

type CreateMemberRequest struct {
	ReferenceID   string  `json:"referenceID" binding:"required"`
	ReferrerCode  *string `json:"referrerCode"`
	PreferredCode *string `json:"preferredCode"`
	CampaignIDs   []uint  `json:"campaignIDs"`
	Email         *string `json:"email"`
}

type UpdateMemberRequest struct {
	CampaignIDs []uint  `json:"campaignIDs"`
	Email       *string `json:"email"`
}

type GetMemberRequest struct {
	Projects                    []string             `form:"projects"`    // Filter by name
	ID                          *uint                `form:"id"`          // Filter by ID
	ReferenceID                 *string              `form:"referenceID"` // Composite key with Project
	Email                       *string              `form:"email"`       // Composite key with Project
	Code                        *string              `form:"code"`
	CampaignIDs                 []uint               `form:"campaignIDs"`
	IsReferred                  *bool                `form:"isReferrer"`
	ReferredByMemberID          *uint                `form:"referredByMemberID"`
	ReferredByMemberReferenceID *string              `form:"referredByMemberReferenceID"`
	PaginationConditions        PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

func ApplyGetMemberRequest(req GetMemberRequest, query *gorm.DB) *gorm.DB {
	// Apply filters with explicit table name
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("referral_members.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("referral_members.id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("referral_members.reference_id = ?", *req.ReferenceID)
	}
	if req.Email != nil {
		query = query.Where("referral_members.email = ?", *req.Email)
	}
	if req.Code != nil {
		query = query.Where("referral_members.code = ?", *req.Code)
	}
	if req.IsReferred != nil {
		if *req.IsReferred {
			query = query.Where("referral_members.referred_by_member_id IS NOT NULL")
		} else {
			query = query.Where("referral_members.referred_by_member_id IS NULL")
		}
	}
	if req.ReferredByMemberID != nil {
		query = query.Where("referral_members.referred_by_member_id = ?", *req.ReferredByMemberID)
	}
	if req.ReferredByMemberReferenceID != nil {
		query = query.Where("referral_members.referred_by_member_reference_id = ?", *req.ReferredByMemberReferenceID)
	}
	if req.CampaignIDs != nil && len(req.CampaignIDs) > 0 {
		// Join with referral_members_campaigns table to filter by CampaignIDs
		query = query.Joins("JOIN referral_members_campaigns rc ON rc.referrer_id = referral_members.id").
			Where("rc.campaign_id IN (?)", req.CampaignIDs).
			Group("referral_members.id") // Avoid duplicates due to the JOIN
	}
	return query
}
