package response

import (
	"github.com/shopspring/decimal"
	"time"
)

type ReferrerStats struct {
	ID           uint            `json:"id"`
	Project      string          `json:"project"`
	Email        *string         `json:"email"`
	ReferenceID  string          `json:"referenceID"`
	Code         string          `json:"code"`
	RefereeCount int64           `json:"refereeCount"`
	TotalRewards decimal.Decimal `json:"totalRewards"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
	IsReferred   bool            `json:"isReferred"`
}

type RewardStats struct {
	Date            string          `json:"date"`
	TotalRewards    decimal.Decimal `json:"totalRewards"`
	UniqueReferrers int64           `json:"uniqueReferrers"`
}
