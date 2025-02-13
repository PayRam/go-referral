package serviceimpl

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/service"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	currentDate := time.Now()

	if err := w.DB.
		Preload("Events").
		Where("status = ? AND is_default = ? AND end_date >= ?", "active", true, currentDate).
		Find(&campaigns).Error; err != nil {
		return fmt.Errorf("failed to fetch campaigns: %w", err)
	}

	// Traverse each campaign
	for _, campaign := range campaigns {
		// Fetch pending EventLogs for this campaign's events
		eventKeys := getEventKeys(campaign.Events)
		var eventLogs []models.EventLog

		if err := w.DB.Where("project = ? AND status = ? AND event_key IN (?)", campaign.Project, "pending", eventKeys).
			Order("id ASC").
			Find(&eventLogs).Error; err != nil {
			fmt.Printf("failed to fetch pending EventLogs for campaign %d: %v\n", campaign.ID, err)
			continue
		}

		// Group EventLogs by ReferrerReferenceID and ReferenceType
		eventLogGroups := groupEventLogs(eventLogs, eventKeys)

		if eventLogGroups == nil {
			continue
		}

		// Traverse each group of EventLogs
		for _, logs := range eventLogGroups {
			// Lock each event log row individually
			err := w.DB.Transaction(func(tx *gorm.DB) error {
				eventLogIDs := getEventLogIDs(logs)

				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("id IN (?) AND status = ?", eventLogIDs, "pending").
					Find(&eventLogs).Error; err != nil {
					return fmt.Errorf("failed to lock event logs: %w", err)
				}

				project := campaign.Project
				refereeReferenceID := logs[0].ReferenceID

				var referee models.Referee
				if err := tx.Preload("Referrer").
					Where("project = ? AND reference_id = ?", campaign.Project, refereeReferenceID).
					First(&referee).Error; err != nil {
					fmt.Printf("failed to fetch referee for project %s and reference_id %s: %v\n", campaign.Project, refereeReferenceID, err)
					return err
				}

				if referee.Referrer == nil || referee.Referrer.Status != "active" {
					fmt.Printf("Referrer is either nil or inactive for reference_id %s\n", refereeReferenceID)
					return nil
				}

				// Check if all campaign events are satisfied
				if !areAllCampaignEventsSatisfied(campaign.Events, logs) {
					return nil
				}

				if campaign.CampaignTypePerCustomer == "one_time" {
					var existingReward models.Reward
					if err := tx.Where("project = ? AND campaign_id = ? AND referrer_reference_id = ?",
						project, campaign.ID, referee.Referrer.ReferenceID).First(&existingReward).Error; err == nil {
						return fmt.Errorf("reward already exists for campaign %d and referrer %s", campaign.ID, referee.Referrer.ReferenceID)
					}
				}

				// Calculate reward
				rewardAmount, err := calculateReward(tx, campaign, logs)
				if err != nil {
					fmt.Printf("calculateReward: failed to calculate reward for campaign %d: %v\n", campaign.ID, err)
					return err
				}

				// Apply Reward Cap per Customer
				if campaign.RewardCap != nil && rewardAmount.GreaterThan(*campaign.RewardCap) {
					rewardAmount = campaign.RewardCap
				}

				// Validate limits
				totalReward, monthsPassed, rewardsCount, err := w.GetTotalRewardByReferrer(tx, project, campaign.ID, referee.Referrer.ReferenceID)
				if err != nil {
					fmt.Printf("GetTotalRewardByReferrer: failed to calculate total reward for campaign %d: %v\n", campaign.ID, err)
					return err
				}

				// Reward Cap Per Customer
				if campaign.RewardCapPerCustomer != nil && totalReward.Add(*rewardAmount).GreaterThan(*campaign.RewardCapPerCustomer) {
					return nil
				}

				// Check if the validity period is exceeded
				if campaign.ValidityMonthsPerCustomer != nil && monthsPassed >= *campaign.ValidityMonthsPerCustomer {
					return nil
				}

				// Check if max occurrences are exceeded
				if campaign.MaxOccurrencesPerCustomer != nil && rewardsCount >= *campaign.MaxOccurrencesPerCustomer {
					return nil
				}

				// Budget Limit Check
				if campaign.Budget != nil {
					var totalRewards decimal.Decimal
					err = tx.Model(&models.Reward{}).
						Select("COALESCE(SUM(amount), 0)").
						Where("campaign_id = ?", campaign.ID).
						Scan(&totalRewards).Error
					if err != nil {
						fmt.Printf("failed to calculate total reward for campaign %d: %v\n", campaign.ID, err)
						return err
					}

					// Check if total rewards exceed budget
					if totalRewards.Add(*rewardAmount).GreaterThan(*campaign.Budget) {
						return fmt.Errorf("exceeds budget")
					}
				}

				if rewardAmount.LessThanOrEqual(decimal.NewFromInt(0)) {
					fmt.Printf("Reward amount is zero or negative for campaign %d\n", campaign.ID)
					return fmt.Errorf("Reward amount is zero or negative for campaign %d\n", campaign.ID)
				}

				// Create the reward
				reward := &models.Reward{
					Project:             project,
					CampaignID:          campaign.ID,
					CurrencyCode:        campaign.CurrencyCode,
					ReferrerID:          referee.Referrer.ID,
					ReferrerReferenceID: referee.Referrer.ReferenceID,
					ReferrerCode:        referee.Referrer.Code,
					RefereeID:           referee.ID,
					RefereeReferenceID:  referee.ReferenceID,
					Amount:              *rewardAmount,
					Status:              "pending",
				}
				if err := tx.Create(reward).Error; err != nil {
					fmt.Printf("failed to create reward for campaign %d: %v\n", campaign.ID, err)
					return err
				}

				// Perform bulk update
				if err := tx.Model(&models.EventLog{}).
					Where("id IN (?)", eventLogIDs).
					Updates(map[string]interface{}{
						"status":    "processed",
						"reward_id": reward.ID,
					}).Error; err != nil {
					return fmt.Errorf("failed to bulk update EventLogs: %w", err)
				}

				return nil
			})

			if err != nil {
				// Log the error and continue with other campaigns
				fmt.Printf("Error processing campaign %d: %v\n", campaign.ID, err)
			}
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
