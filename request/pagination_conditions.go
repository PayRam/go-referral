package request

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type PaginationConditions struct {
	Limit         *int       `form:"limit"`         // Pagination limit
	Offset        *int       `form:"offset"`        // Pagination offset (optional when using ID-based)
	SortBy        *string    `form:"sortBy"`        // Field to sort by
	Order         *string    `form:"order"`         // ASC or DESC
	GreaterThanID *uint      `form:"greaterThanID"` // For ID-based pagination
	LessThanID    *uint      `form:"lessThanID"`    // For reverse ID-based pagination
	CreatedAfter  *time.Time `form:"createdAfter"`  // Filter records created after this date
	CreatedBefore *time.Time `form:"createdBefore"` // Filter records created before this date
	UpdatedAfter  *time.Time `form:"updatedAfter"`  // Filter records updated after this date
	UpdatedBefore *time.Time `form:"updatedBefore"` // Filter records updated before this date
	StartDate     *time.Time `form:"startDate"`     // Filter records created after this date
	EndDate       *time.Time `form:"endDate"`       // Filter records created after this date
	GroupBy       *string    `form:"groupBy"`
	SelectFields  []string   `form:"selectFields"`
}

func ApplySelectFields(query *gorm.DB, selectFields []string) *gorm.DB {
	if len(selectFields) > 0 {
		query = query.Select(selectFields)
	}
	return query
}

func ApplyGroupBy(query *gorm.DB, groupBy *string) *gorm.DB {
	if groupBy != nil && *groupBy != "" {
		query = query.Group(*groupBy)
	}
	return query
}

func ApplyPaginationConditions(query *gorm.DB, conditions PaginationConditions) *gorm.DB {
	// Count total records (optional based on use case)
	if conditions.Offset != nil && *conditions.Offset > 0 {
		query = query.Offset(*conditions.Offset)
	}

	// Apply ID-based pagination
	if conditions.GreaterThanID != nil {
		query = query.Where("id > ?", *conditions.GreaterThanID)
	}
	if conditions.LessThanID != nil {
		query = query.Where("id < ?", *conditions.LessThanID)
	}

	// Apply date filters
	if conditions.CreatedAfter != nil {
		query = query.Where("created_at > ?", *conditions.CreatedAfter)
	}
	if conditions.CreatedBefore != nil {
		query = query.Where("created_at < ?", *conditions.CreatedBefore)
	}
	if conditions.UpdatedAfter != nil {
		query = query.Where("updated_at > ?", *conditions.UpdatedAfter)
	}
	if conditions.UpdatedBefore != nil {
		query = query.Where("updated_at < ?", *conditions.UpdatedBefore)
	}

	if conditions.StartDate != nil {
		query = query.Where("created_at >= ?", *conditions.StartDate)
	}
	if conditions.EndDate != nil {
		query = query.Where("created_at <= ?", *conditions.EndDate)
	}

	// ✅ Sorting logic
	if conditions.SortBy != nil {
		order := "ASC"
		if conditions.Order != nil {
			order = *conditions.Order
		}
		query = query.Order(fmt.Sprintf("%s %s", *conditions.SortBy, order))
	}

	// Apply limit
	if conditions.Limit != nil && *conditions.Limit > 0 {
		query = query.Limit(*conditions.Limit)
	}

	return query
}
