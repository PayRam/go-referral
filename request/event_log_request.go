package request

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CreateEventLogRequest struct {
	EventKey    string           `json:"eventKey" binding:"required"`
	ReferenceID string           `json:"referenceID" binding:"required"`
	Amount      *decimal.Decimal `json:"amount"`
	Data        *string          `json:"data"`
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

func ApplyGetEventLogRequest(req GetEventLogRequest, query *gorm.DB) *gorm.DB {
	// Apply filters with table name prepended
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("referral_event_logs.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("referral_event_logs.id = ?", *req.ID)
	}
	if req.EventKey != nil {
		query = query.Where("referral_event_logs.event_key = ?", *req.EventKey)
	}
	if req.ReferenceID != nil {
		query = query.Where("referral_event_logs.member_reference_id = ?", *req.ReferenceID)
	}
	if req.Status != nil {
		query = query.Where("referral_event_logs.status = ?", *req.Status)
	}
	if req.RewardID != nil {
		query = query.Where("referral_event_logs.reward_id = ?", *req.RewardID)
	}
	return query
}
