package request

import "gorm.io/gorm"

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

type GetEventsRequest struct {
	Projects             []string             `form:"projects"`             // Filter by name
	ID                   *uint                `form:"id"`                   // Filter by ID
	Key                  *string              `form:"key"`                  // Filter by name
	Name                 *string              `form:"name"`                 // Filter by name
	EventType            *string              `form:"eventType"`            // Filter by name
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

func ApplyGetEventRequest(req GetEventsRequest, query *gorm.DB) *gorm.DB {
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("referral_events.project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("referral_events.id = ?", *req.ID)
	}
	if req.Key != nil {
		query = query.Where("referral_events.key = ?", *req.Key)
	}
	if req.Name != nil {
		query = query.Where("referral_events.name LIKE ?", "%"+*req.Name+"%")
	}
	if req.EventType != nil {
		query = query.Where("referral_events.event_type = ?", *req.EventType)
	}
	return query
}
