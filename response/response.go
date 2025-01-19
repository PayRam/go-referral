package response

import (
	"github.com/shopspring/decimal"
	"time"
)

type ReferrerStats struct {
	ID           uint            `json:"id"`
	Project      string          `json:"project"`
	ReferenceID  string          `json:"referenceID"`
	Code         string          `json:"code"`
	RefereeCount int64           `json:"refereeCount"`
	TotalRewards decimal.Decimal `json:"totalRewards"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
	DeletedAt    *time.Time      `json:"deletedAt,omitempty"`
}
