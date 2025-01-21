package request

import "gorm.io/gorm"

type CreateRefereeRequest struct {
	ReferenceID string  `json:"referenceID" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Email       *string `json:"email"`
}

type GetRefereeRequest struct {
	Projects             []string             `form:"projects"`             // Filter by name
	ID                   *uint                `form:"id"`                   // Filter by ID
	ReferenceID          *string              `form:"referenceID"`          // Composite key with Project
	ReferrerReferenceID  *string              `form:"referrerReferenceID"`  // Composite key with Project
	ReferrerID           *uint                `form:"referrerID"`           // ID of the Referrer (Foreign Key to Referrer table)
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

func ApplyGetRefereeRequest(req GetRefereeRequest, query *gorm.DB) *gorm.DB {
	// Apply filters
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("referral_referee.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("referral_referee.id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("referral_referee.reference_id = ?", *req.ReferenceID)
	}
	if req.ReferrerID != nil {
		query = query.Where("referral_referee.referrer_id = ?", *req.ReferrerID)
	}
	if req.ReferrerReferenceID != nil {
		query = query.Where("referral_referee.referrer_reference_id = ?", *req.ReferrerReferenceID)
	}
	return query
}
