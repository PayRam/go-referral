package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/service"
	"gorm.io/gorm"
)

type worker struct {
	DB                      *gorm.DB
	DefaultRewardCalculator service.RewardCalculator
	CustomCalculators       map[string]service.RewardCalculator // Map of EventID to custom calculators
}

var _ service.Worker = &worker{}

func NewWorkerService(db *gorm.DB) service.Worker {
	return &worker{DB: db}
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
	// Fetch all pending EventLog entries
	var pendingEventLogs []models.EventLog
	if err := w.DB.Where("status = ?", "pending").Find(&pendingEventLogs).Error; err != nil {
		return fmt.Errorf("failed to fetch pending EventLogs: %w", err)
	}

	// Process each pending EventLog
	for _, eventLog := range pendingEventLogs {
		err := w.DB.Transaction(func(tx *gorm.DB) error {
			// Check if a reward already exists for this EventLog
			var existingReward models.Reward
			if err := tx.Where("event_log_id = ?", eventLog.ID).First(&existingReward).Error; err == nil {
				// Reward already exists, mark EventLog as processed
				eventLog.Status = "processed"
				eventLog.FailureReason = nil
				return tx.Save(&eventLog).Error
			}

			// Fetch referee using ReferenceID and ReferenceType
			var referee models.Referee
			if err := tx.Where("reference_id = ? AND reference_type = ?", eventLog.ReferenceID, eventLog.ReferenceType).
				First(&referee).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					eventLog.Status = "failed"
					reason := "Referee not found"
					eventLog.FailureReason = &reason
					return tx.Save(&eventLog).Error
				}
				return fmt.Errorf("failed to fetch referee for EventLog %d: %w", eventLog.ID, err)
			}

			// Fetch referrer and associated campaign
			var referrer models.Referrer
			if err := tx.First(&referrer, referee.ReferrerID).Error; err != nil {
				eventLog.Status = "failed"
				reason := "Referrer not found"
				eventLog.FailureReason = &reason
				return tx.Save(&eventLog).Error
			}

			var campaign models.Campaign
			if err := tx.First(&campaign, referrer.CampaignID).Error; err != nil {
				eventLog.Status = "failed"
				reason := "Campaign not found"
				eventLog.FailureReason = &reason
				return tx.Save(&eventLog).Error
			}

			// Ensure the EventKey is associated with the campaign
			var campaignEvent models.CampaignEvent
			if err := tx.Where("campaign_id = ? AND event_key = ?", campaign.ID, eventLog.EventKey).
				First(&campaignEvent).Error; err != nil {
				eventLog.Status = "failed"
				reason := "EventKey not associated with campaign"
				eventLog.FailureReason = &reason
				return tx.Save(&eventLog).Error
			}

			// Fetch the corresponding event rule
			var event models.Event
			if err := tx.First(&event, "key = ?", eventLog.EventKey).Error; err != nil {
				eventLog.Status = "failed"
				reason := "Event rule not found"
				eventLog.FailureReason = &reason
				return tx.Save(&eventLog).Error
			}

			// Select the appropriate calculator
			calculator := w.DefaultRewardCalculator
			if customCalculator, exists := w.CustomCalculators[event.Key]; exists {
				calculator = customCalculator
			}

			// Calculate the reward
			reward, err := calculator.CalculateReward(eventLog, event, campaign, referee, referrer)
			if err != nil {
				eventLog.Status = "failed"
				eventLog.FailureReason = ptr(fmt.Sprintf("Reward calculation failed: %v", err))
				return tx.Save(&eventLog).Error
			}

			// Create the reward
			if err := tx.Create(reward).Error; err != nil {
				eventLog.Status = "failed"
				eventLog.FailureReason = ptr(fmt.Sprintf("Failed to create reward: %v", err))
				return tx.Save(&eventLog).Error
			}

			// Mark the EventLog as processed
			eventLog.Status = "processed"
			eventLog.FailureReason = nil
			return tx.Save(&eventLog).Error
		})

		if err != nil {
			// Log the error and continue processing other EventLogs
			fmt.Printf("Error processing EventLog %d: %v\n", eventLog.ID, err)
		}
	}

	return nil
}

func ptr(s string) *string {
	return &s
}
