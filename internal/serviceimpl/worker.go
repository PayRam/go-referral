package serviceimpl

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/service"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"sort"
	"time"
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
	currentDate := time.Now() // Get the current date and time

	if err := w.DB.
		Preload("Events").
		Where("status = ? AND is_default = ? AND end_date >= ?", "active", true, currentDate).
		Find(&campaigns).Error; err != nil {
		return fmt.Errorf("failed to fetch campaigns: %w", err)
	}

	// Traverse each campaign
	for _, campaign := range campaigns {
		err := w.DB.Transaction(func(tx *gorm.DB) error {
			// Fetch pending EventLogs for this campaign's events

			//var events []models.Event
			//if err := w.DB.Where("project = ?", campaign.Project).Find(&events).Error; err != nil {
			//	return fmt.Errorf("failed to fetch events for project %s: %w", campaign.Project, err)
			//}

			eventKeys := getEventKeys(campaign.Events)
			var eventLogs []models.EventLog
			if err := tx.Where("project = ? AND status = ? AND event_key IN (?)", campaign.Project, "pending", eventKeys).
				Order("id ASC"). // Sort the event logs by ID in ascending order
				Find(&eventLogs).Error; err != nil {
				return fmt.Errorf("failed to fetch pending EventLogs for campaign %d: %w", campaign.ID, err)
			}

			// Group EventLogs by ReferrerReferenceID and ReferenceType
			eventLogGroups := groupEventLogs(eventLogs, eventKeys)

			// Traverse each group of EventLogs
			for _, logs := range eventLogGroups {
				//logs := eventLogGroups[referenceKey]
				project := campaign.Project
				refereeReferenceID := logs[0].ReferenceID
				//project, refereeReferenceID := parseReferenceKey(referenceKey)

				var referee models.Referee
				if err := tx.Preload("Referrer").Where("project = ? AND reference_id = ?", campaign.Project, refereeReferenceID).Find(&referee).Error; err != nil {
					return fmt.Errorf("failed to fetch referee for project %s and reference_id %s: %w", campaign.Project, refereeReferenceID, err)
				}

				// Check if all campaign events are satisfied
				allEventsSatisfied := areAllCampaignEventsSatisfied(campaign.Events, logs)
				if !allEventsSatisfied {
					continue // Skip if not all events are satisfied
				}

				if campaign.CampaignTypePerCustomer == "one_time" {
					// Check if reward for this campaign and referee already exists
					var existingReward models.Reward
					if err := tx.Where("project = ? AND campaign_id = ? AND referrer_reference_id = ?",
						project, campaign.ID, referee.Referrer.ReferenceID).First(&existingReward).Error; err == nil {
						//TODO whether event logs should be updated to invalid status. need to discuss and finalize
						continue // Reward already exists
					}
				}

				// Calculate reward
				rewardAmount, err := calculateReward(tx, campaign, logs)
				if err != nil {
					continue
				}

				if campaign.RewardCap != nil {
					if rewardAmount.GreaterThan(*campaign.RewardCap) {
						rewardAmount = campaign.RewardCap
					}
				}

				totalReward, monthsPassed, rewardsCount, err := w.GetTotalRewardByReferrer(tx, project, campaign.ID, referee.Referrer.ReferenceID)
				if err != nil {
					return fmt.Errorf("failed to calculate total rewards: %w", err)
				}

				if campaign.RewardCapPerCustomer != nil && totalReward.Add(*rewardAmount).GreaterThan(*campaign.RewardCapPerCustomer) {
					continue
				}
				// Check if the validity period is exceeded
				if campaign.ValidityMonthsPerCustomer != nil && monthsPassed >= *campaign.ValidityMonthsPerCustomer {
					continue
				}

				// Check if the max occurrences are exceeded
				if campaign.MaxOccurrencesPerCustomer != nil && rewardsCount >= *campaign.MaxOccurrencesPerCustomer {
					continue
				}

				if campaign.Budget != nil {
					// Get the total reward amount already awarded for this campaign
					var totalReward decimal.Decimal
					err = tx.Model(&models.Reward{}).
						Select("COALESCE(SUM(amount), 0)").
						Where("campaign_id = ?", campaign.ID).
						Scan(&totalReward).Error
					if err != nil {
						continue
					}

					// Check if the total reward (including the current rewardAmount) exceeds the budget
					if totalReward.Add(*rewardAmount).GreaterThan(*campaign.Budget) {
						continue
					}
				}

				// Create the reward
				reward := &models.Reward{
					Project:             project,
					CampaignID:          campaign.ID,
					ReferrerID:          referee.Referrer.ID,
					ReferrerReferenceID: referee.Referrer.ReferenceID,
					ReferrerCode:        referee.Referrer.Code,
					RefereeID:           referee.ID,
					RefereeReferenceID:  referee.ReferenceID,
					Amount:              *rewardAmount,
					Status:              "pending",
				}
				if err := tx.Create(reward).Error; err != nil {
					continue
				}

				// Mark associated EventLogs as processed
				for _, log := range logs {
					log.Status = "processed"
					log.RewardID = &reward.ID
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
			return err
		}
	}

	return nil
}

func (w *worker) GetTotalRewardByReferrer(
	tx *gorm.DB,
	project string,
	campaignID uint,
	referrerReferenceID string,
) (decimal.Decimal, int, int64, error) {
	var totalReward decimal.Decimal
	var firstRewardMonthStr sql.NullString
	var rewardsCount int64

	// Query the rewards table to calculate the total reward and get the first reward month and rewards count
	err := tx.Model(&models.Reward{}).
		Where("project = ? AND campaign_id = ? AND referrer_reference_id = ?", project, campaignID, referrerReferenceID).
		Select("COALESCE(SUM(amount), 0) AS total_reward, MIN(created_at) AS first_reward_month, COUNT(*) AS rewards_count").
		Row().Scan(&totalReward, &firstRewardMonthStr, &rewardsCount)

	if err != nil {
		// Return an error only if it's a database error
		return decimal.Zero, 0, 0, fmt.Errorf("failed to calculate total reward: %w", err)
	}

	// If no rewards exist, return 0 values and no error
	if rewardsCount == 0 {
		return decimal.Zero, 0, 0, nil
	}

	// Parse firstRewardMonth
	var firstRewardMonth *time.Time
	parsedTime, parseErr := time.Parse("2006-01-02 15:04:05-07:00", firstRewardMonthStr.String)
	if parseErr != nil {
		return decimal.Zero, 0, 0, fmt.Errorf("failed to parse first reward month: %w", parseErr)
	}
	firstRewardMonth = &parsedTime

	// Calculate months passed
	currentTime := time.Now()
	years := currentTime.Year() - firstRewardMonth.Year()
	months := int(currentTime.Month() - firstRewardMonth.Month())
	monthsPassed := (years * 12) + months

	return totalReward, monthsPassed, rewardsCount, nil
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

func calculateReward(tx *gorm.DB, campaign models.Campaign, logs []models.EventLog) (*decimal.Decimal, error) {
	if campaign.RewardType == "flat_fee" {
		return &campaign.RewardValue, nil
	}

	if campaign.RewardType == "percentage" {
		// Sum the total amount from event logs for percentage calculation
		var totalAmount decimal.Decimal
		if err := tx.Model(&models.EventLog{}).
			Where("id IN (?)", getEventLogIDs(logs)).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&totalAmount).Error; err != nil {
			return nil, fmt.Errorf("failed to calculate total amount from event logs: %w", err)
		}

		// Calculate the percentage-based reward
		percentage := campaign.RewardValue.Div(decimal.NewFromInt(100))
		reward := totalAmount.Mul(percentage)

		return &reward, nil
	}

	return nil, fmt.Errorf("unknown reward type: %s", campaign.RewardType)
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

func groupEventLogs(eventLogs []models.EventLog, requiredKeys []string) [][]models.EventLog {
	// Initialize a two-dimensional slice
	var eventLogsArray [][]models.EventLog

	// Helper function to check if a group satisfies all required keys
	hasAllKeys := func(group []models.EventLog, requiredKeys []string) bool {
		keyMap := make(map[string]bool)
		for _, log := range group {
			keyMap[log.EventKey] = true
		}
		for _, key := range requiredKeys {
			if !keyMap[key] {
				return false
			}
		}
		return true
	}

	// Traverse the event logs sequentially
	for _, log := range eventLogs {
		added := false
		for i, group := range eventLogsArray {
			if group[0].ReferenceID == log.ReferenceID && !hasAllKeys(group, requiredKeys) {
				keyExists := false
				for _, existingLog := range group {
					if existingLog.EventKey == log.EventKey {
						keyExists = true
						break
					}
				}
				if !keyExists {
					eventLogsArray[i] = append(eventLogsArray[i], log)
					added = true
					break
				}
			}
		}
		if !added {
			eventLogsArray = append(eventLogsArray, []models.EventLog{log})
		}
	}

	// Filter out groups that do not satisfy the required keys
	var result [][]models.EventLog
	for _, group := range eventLogsArray {
		if hasAllKeys(group, requiredKeys) {
			result = append(result, group)
		}
	}

	// Sort groups by maximum CreatedAt in ascending order
	sort.Slice(result, func(i, j int) bool {
		return getMaxCreatedAt(result[i]).Before(getMaxCreatedAt(result[j]))
	})

	return result
}

// Helper function to get the maximum CreatedAt from a group of EventLogs
func getMaxCreatedAt(group []models.EventLog) time.Time {
	var maxTime time.Time
	for _, log := range group {
		if log.CreatedAt.After(maxTime) {
			maxTime = log.CreatedAt
		}
	}
	return maxTime
}
