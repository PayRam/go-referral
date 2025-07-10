package request

import "gorm.io/gorm"

type CreateCustomerRequest struct {
	ReferenceID   string  `json:"referenceID" binding:"required"`
	ReferrerCode  *string `json:"referrerCode"`
	PreferredCode *string `json:"preferredCode"`
	CampaignIDs   []uint  `json:"campaignIDs"`
	Email         *string `json:"email"`
}

type UpdateCustomerRequest struct {
	CampaignIDs []uint  `json:"campaignIDs"`
	Email       *string `json:"email"`
}

type GetCustomerRequest struct {
	Projects                      []string             `form:"projects"`    // Filter by name
	ID                            *uint                `form:"id"`          // Filter by ID
	ReferenceID                   *string              `form:"referenceID"` // Composite key with Project
	Email                         *string              `form:"email"`       // Composite key with Project
	Code                          *string              `form:"code"`
	CampaignIDs                   []uint               `form:"campaignIDs"`
	IsReferred                    *bool                `form:"isReferrer"`
	ReferredByCustomerID          *uint                `form:"referredByCustomerID"`
	ReferredByCustomerReferenceID *string              `form:"referredByCustomerReferenceID"`
	PaginationConditions          PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

func ApplyGetCustomerRequest(req GetCustomerRequest, query *gorm.DB) *gorm.DB {
	// Apply filters with explicit table name
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("referral_customers.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("referral_customers.id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("referral_customers.reference_id = ?", *req.ReferenceID)
	}
	if req.Email != nil {
		query = query.Where("referral_customers.email = ?", *req.Email)
	}
	if req.Code != nil {
		query = query.Where("referral_customers.code = ?", *req.Code)
	}
	if req.IsReferred != nil {
		if *req.IsReferred {
			query = query.Where("referral_customers.referred_by_customer_id IS NOT NULL")
		} else {
			query = query.Where("referral_customers.referred_by_customer_id IS NULL")
		}
	}
	if req.ReferredByCustomerID != nil {
		query = query.Where("referral_customers.referred_by_customer_id = ?", *req.ReferredByCustomerID)
	}
	if req.ReferredByCustomerReferenceID != nil {
		query = query.Where("referral_customers.referred_by_customer_reference_id = ?", *req.ReferredByCustomerReferenceID)
	}
	if req.CampaignIDs != nil && len(req.CampaignIDs) > 0 {
		// Join with referral_customers_campaigns table to filter by CampaignIDs
		query = query.Joins("JOIN referral_customers_campaigns rc ON rc.referrer_id = referral_customers.id").
			Where("rc.campaign_id IN (?)", req.CampaignIDs).
			Group("referral_customers.id") // Avoid duplicates due to the JOIN
	}
	return query
}
