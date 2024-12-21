package serviceimpl

import (
	"encoding/json"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/service"
	"github.com/shopspring/decimal"
)

type DefaultRewardCalculator struct{}

// Ensure AccountAddressServiceImpl implements AddressPoolService.
var _ service.RewardCalculator = &DefaultRewardCalculator{}

func (d *DefaultRewardCalculator) CalculateReward(
	eventLog models.EventLog,
	event models.Event,
	campaign models.Campaign,
	referee models.Referee,
	referrer models.Referrer,
) (*models.Reward, error) {
	// Validate RewardType
	if event.RewardType != "flat_fee" && event.RewardType != "percentage" {
		return nil, fmt.Errorf("unknown reward type: %s", event.RewardType)
	}

	// Calculate reward amount based on RewardType
	var rewardAmount decimal.Decimal
	switch event.RewardType {
	case "flat_fee":
		// Use event's RewardValue as a flat fee
		rewardAmount = decimal.NewFromFloat(event.RewardValue)
	case "percentage":
		// Extract transaction amount from EventLog.Data
		transactionAmount := extractTransactionAmount(eventLog.Data)
		if transactionAmount == nil || transactionAmount.IsZero() {
			return nil, fmt.Errorf("transaction amount missing or invalid in event log")
		}

		// Calculate percentage-based reward
		rewardAmount = transactionAmount.Mul(decimal.NewFromFloat(event.RewardValue)).Div(decimal.NewFromInt(100))
	}

	// Construct the Reward object
	reward := &models.Reward{
		EventLogID:    eventLog.ID,
		EventKey:      event.Key,
		CampaignID:    campaign.ID,
		RefereeID:     referee.ID,
		RefereeType:   referee.ReferenceType,
		ReferenceID:   referrer.ReferenceID,
		ReferenceType: referrer.ReferenceType,
		Amount:        rewardAmount,
		Status:        "pending", // Default status for newly calculated rewards
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
