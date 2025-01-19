package serviceimpl

import (
	"encoding/json"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/shopspring/decimal"
)

type DefaultRewardCalculator struct{}

//var _ service.RewardCalculator = &DefaultRewardCalculator{}

func NewDefaultRewardCalculator() *DefaultRewardCalculator {
	return &DefaultRewardCalculator{}
}

func (d *DefaultRewardCalculator) CalculateReward(
	eventLog models.EventLog,
	campaign models.Campaign,
	referee models.Referee,
	referrer models.Referrer,
) (*models.Reward, error) {

	if campaign.RewardType != "flat_fee" && campaign.RewardType != "percentage" {
		return nil, fmt.Errorf("unknown reward type: %s", campaign.RewardType)
	}

	// Calculate reward amount based on RewardType
	var rewardAmount decimal.Decimal
	switch campaign.RewardType {
	case "flat_fee":
		// Use event's RewardValue as a flat fee
		rewardAmount = *campaign.RewardValue
	case "percentage":
		// Extract transaction amount from EventLog.Data
		transactionAmount := extractTransactionAmount(eventLog.Data)
		if transactionAmount == nil || transactionAmount.IsZero() {
			return nil, fmt.Errorf("transaction amount missing or invalid in event log")
		}

		// Calculate percentage-based reward
		rewardAmount = transactionAmount.Mul(*campaign.RewardValue).Div(decimal.NewFromInt(100))
	}

	// Construct the Reward object
	reward := &models.Reward{
		Project:             campaign.Project,
		CampaignID:          campaign.ID,
		RefereeID:           referee.ID,
		ReferrerReferenceID: referrer.ReferenceID,
		Amount:              rewardAmount,
		Status:              "pending", // Default status for newly calculated rewards
	}

	return reward, nil
}

func extractTransactionAmount(data *string) *decimal.Decimal {
	if data == nil || *data == "" {
		return nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(*data), &parsed); err != nil {
		return nil
	}

	if amount, ok := parsed["transactionAmount"].(float64); ok {
		v := decimal.NewFromFloat(amount)
		return &v
	}

	return nil
}
