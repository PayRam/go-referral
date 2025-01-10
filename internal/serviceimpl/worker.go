package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/service"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
)

type worker struct {
	DB                      *gorm.DB
	DefaultRewardCalculator service.RewardCalculator
	CustomCalculators       map[string]service.RewardCalculator // Map of EventID to custom calculators
}

//var _ service.Worker = &worker{}

func NewWorkerService(db *gorm.DB) *worker {
	return &worker{
		DefaultRewardCalculator: NewDefaultRewardCalculator(),
		DB:                      db,
	}
}

func (w *worker) AddCustomRewardCalculator(eventKey string, calculator service.RewardCalculator) error {
	// Validate if the event key exists in the database
	var event models.Event
	if err := w.DB.Where("key = ?", eventKey).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("event key '%s' does not exist", eventKey)
		}
		return fmt.Errorf("failed to validate event key '%s': %w", eventKey, err)
	}

	// Add the calculator to the map
	w.CustomCalculators[eventKey] = calculator
	return nil
}

func (w *worker) RemoveCustomRewardCalculator(eventKey string) error {
	// Check if a custom calculator exists for the event key
	if _, exists := w.CustomCalculators[eventKey]; !exists {
		return fmt.Errorf("no custom calculator found for event key '%s'", eventKey)
	}

	// Remove the calculator from the map
	delete(w.CustomCalculators, eventKey)
	return nil
}

func (w *worker) ProcessPendingEvents() error {
	// Fetch all active campaigns with preloaded events
	var campaigns []models.Campaign
	if err := w.DB.Preload("Events").Where("is_active = ?", true).Find(&campaigns).Error; err != nil {
		return fmt.Errorf("failed to fetch campaigns: %w", err)
	}

	// Traverse each campaign
	for _, campaign := range campaigns {
		err := w.DB.Transaction(func(tx *gorm.DB) error {
			// Fetch pending EventLogs for this campaign's events
			eventKeys := getEventKeys(campaign.Events)
			var eventLogs []models.EventLog
			if err := tx.Where("status = ? AND event_key IN (?)", "pending", eventKeys).Find(&eventLogs).Error; err != nil {
				return fmt.Errorf("failed to fetch pending EventLogs for campaign %d: %w", campaign.ID, err)
			}

			// Group EventLogs by ReferenceID and ReferenceType
			eventLogGroups := groupEventLogs(eventLogs)

			// Traverse each group of EventLogs
			for referenceKey, logs := range eventLogGroups {
				referenceID, referenceType := parseReferenceKey(referenceKey)

				// Check if all campaign events are satisfied
				allEventsSatisfied := areAllCampaignEventsSatisfied(campaign.Events, logs)
				if !allEventsSatisfied {
					continue // Skip if not all events are satisfied
				}

				// Check if reward for this campaign and referee already exists
				var existingReward models.Reward
				if err := tx.Where("campaign_id = ? AND reference_id = ? AND reference_type = ?",
					campaign.ID, referenceID, referenceType).First(&existingReward).Error; err == nil {
					continue // Reward already exists
				}

				// Calculate reward
				rewardAmount, err := calculateReward(tx, campaign, logs)
				if err != nil {
					return fmt.Errorf("failed to calculate reward for campaign %d: %w", campaign.ID, err)
				}

				// Create the reward
				reward := &models.Reward{
					CampaignID:    campaign.ID,
					ReferenceID:   referenceID,
					ReferenceType: referenceType,
					Amount:        decimal.NewFromFloat(rewardAmount),
					Status:        "pending",
				}
				if err := tx.Create(reward).Error; err != nil {
					return fmt.Errorf("failed to create reward for campaign %d: %w", campaign.ID, err)
				}

				// Mark associated EventLogs as processed
				for _, log := range logs {
					log.Status = "processed"
					log.FailureReason = nil
					if err := tx.Save(&log).Error; err != nil {
						return fmt.Errorf("failed to update EventLog status: %w", err)
					}
				}
			}
			return nil
		})

		if err != nil {
			// Log the error and continue with other campaigns
			fmt.Printf("Error processing campaign %d: %v\n", campaign.ID, err)
		}
	}

	return nil
}

func areAllCampaignEventsSatisfied(events []models.Event, logs []models.EventLog) bool {
	eventKeys := make(map[string]bool)
	for _, log := range logs {
		eventKeys[log.EventKey] = true
	}

	for _, event := range events {
		if !eventKeys[event.Key] {
			return false
		}
	}
	return true
}

func calculateReward(tx *gorm.DB, campaign models.Campaign, logs []models.EventLog) (float64, error) {
	if *campaign.RewardType == "flat_fee" {
		return *campaign.RewardValue, nil
	}

	if *campaign.RewardType == "percentage" {
		// Sum the total amount from event logs for percentage calculation
		var totalAmount decimal.Decimal
		if err := tx.Model(&models.EventLog{}).
			Where("id IN (?)", getEventLogIDs(logs)).
			Select("SUM(amount)").
			Scan(&totalAmount).Error; err != nil {
			return 0, err
		}
		return (totalAmount.InexactFloat64() * *campaign.RewardValue) / 100, nil
	}

	return 0, fmt.Errorf("unknown reward type: %s", *campaign.RewardType)
}

func getEventLogIDs(logs []models.EventLog) []uint {
	var ids []uint
	for _, log := range logs {
		ids = append(ids, log.ID)
	}
	return ids
}

func getEventKeys(events []models.Event) []string {
	keys := make([]string, len(events))
	for i, event := range events {
		keys[i] = event.Key
	}
	return keys
}

func groupEventLogs(eventLogs []models.EventLog) map[string][]models.EventLog {
	groupedLogs := make(map[string][]models.EventLog)
	for _, log := range eventLogs {
		//if log.ReferenceID == nil || log.ReferenceType == nil {
		//	continue
		//}
		key := generateReferenceKey(log.ReferenceID, log.ReferenceType)
		groupedLogs[key] = append(groupedLogs[key], log)
	}
	return groupedLogs
}

func generateReferenceKey(referenceID, referenceType string) string {
	return fmt.Sprintf("%s|%s", referenceID, referenceType)
}

func parseReferenceKey(key string) (string, string) {
	parts := strings.Split(key, "|")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func ptr(s string) *string {
	return &s
}
