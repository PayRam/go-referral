package serviceimpl

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sort"
	"time"
)

type worker struct {
	DB *gorm.DB
}

var ErrExceedsBudget = errors.New("exceeds budget")

//var _ service.Worker = &worker{}

func NewWorkerService(db *gorm.DB) *worker {
	return &worker{
		DB: db,
	}
}

func (w *worker) ProcessPendingEvents() error {
	// Fetch all active campaigns with preloaded events
	var campaigns []models.Campaign
	currentDate := time.Now()

	if err := w.DB.Model(&models.Campaign{}).
		Where("status = ? AND end_date < ?", "active", currentDate).
		Update("status", "archived").Error; err != nil {
		fmt.Printf("failed to archive expired campaigns: %v\n", err)
	}

	if err := w.DB.
		Preload("Events").
		Where("status = ? AND is_default = ? AND start_date <= ? AND end_date >= ?", "active", true, currentDate, currentDate).
		Find(&campaigns).Error; err != nil {
		return fmt.Errorf("failed to fetch campaigns: %w", err)
	}

	// Traverse each campaign
	for _, campaign := range campaigns {
		// Fetch pending EventLogs for this campaign's events
		eventKeys := getEventKeys(campaign.Events)
		var eventLogs []models.EventLog

		if err := w.DB.Table("referral_event_logs el").
			Select("el.*").
			Joins("LEFT JOIN referral_campaign_event_logs rces ON el.id = rces.event_log_id AND rces.campaign_id = ?", campaign.ID).
			Where("el.project = ? AND el.status = ? AND el.event_key IN (?) AND rces.event_log_id IS NULL",
				campaign.Project, "pending", eventKeys).
			Where("el.created_at > ?", campaign.ConsiderEventsFrom).
			Order("el.id ASC").
			Find(&eventLogs).Error; err != nil {
			fmt.Printf("failed to fetch pending EventLogs for campaign %d: %v\n", campaign.ID, err)
			continue
		}

		// Group EventLogs by ReferredByMemberReferenceID and ReferenceType
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
				refereeReferenceID := logs[0].MemberReferenceID

				var member models.Member
				if err := tx.Preload("ReferredByMember").
					Where("project = ? AND reference_id = ?", campaign.Project, refereeReferenceID).
					First(&member).Error; err != nil {
					fmt.Printf("failed to fetch referee for project %s and reference_id %s: %v\n", campaign.Project, refereeReferenceID, err)
					return err
				}

				if member.ReferredByMember == nil || member.ReferredByMember.Status != "active" {
					fmt.Printf("Member is either nil or inactive for reference_id %s\n", refereeReferenceID)
					return nil
				}

				// Check if all campaign events are satisfied
				if !areAllCampaignEventsSatisfied(campaign.Events, logs) {
					return nil
				}

				if campaign.CampaignTypePerCustomer == "one_time" {
					var existingReward models.Reward
					if err := tx.Where("project = ? AND campaign_id = ? AND rewarded_member_reference_id = ?",
						project, campaign.ID, member.ReferredByMember.ReferenceID).First(&existingReward).Error; err == nil {
						return fmt.Errorf("reward already exists for campaign %d and referrer %s", campaign.ID, member.ReferredByMember.ReferenceID)
					}
				}

				// Calculate reward
				referrerRewardAmount, refereeRewardAmount, err := calculateReward(tx, campaign, logs)
				if err != nil {
					fmt.Printf("calculateReward: failed to calculate reward for campaign %d: %v\n", campaign.ID, err)
					return err
				}

				if referrerRewardAmount != nil {
					// Apply Reward Cap per Customer
					if campaign.RewardCap != nil && referrerRewardAmount.GreaterThan(*campaign.RewardCap) {
						referrerRewardAmount = campaign.RewardCap
					}

					err = w.validateReward(tx, err, project, campaign, member.ReferredByMember.ReferenceID, referrerRewardAmount)
					if err != nil {
						return err
					}
				}
				if refereeRewardAmount != nil {
					if campaign.RewardCap != nil && refereeRewardAmount.GreaterThan(*campaign.InviteeRewardCap) {
						refereeRewardAmount = campaign.RewardCap
					}

					err = w.validateReward(tx, err, project, campaign, member.ReferenceID, refereeRewardAmount)
					if err != nil {
						return err
					}
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

					calculatedTotalRewards := decimal.Zero
					if referrerRewardAmount != nil {
						calculatedTotalRewards = calculatedTotalRewards.Add(*referrerRewardAmount)
					}
					if refereeRewardAmount != nil {
						calculatedTotalRewards = calculatedTotalRewards.Add(*refereeRewardAmount)
					}

					// Check if total rewards exceed budget
					if totalRewards.Add(calculatedTotalRewards).GreaterThan(*campaign.Budget) {
						result := tx.Debug().Model(&models.Campaign{}).
							Where("id = ?", campaign.ID).
							Update("status", "paused")

						if result.Error != nil {
							return fmt.Errorf("failed to pause campaign due to budget overuse: %w", result.Error)
						}

						if err := tx.Commit().Error; err != nil {
							return fmt.Errorf("failed to commit transaction after updating campaign: %w", err)
						}

						return ErrExceedsBudget
					}
				}

				var referrerReward *models.Reward
				var refereeReward *models.Reward
				if referrerRewardAmount != nil && referrerRewardAmount.GreaterThan(decimal.NewFromInt(0)) {
					// Create the reward
					referrerReward = &models.Reward{
						Project:                   project,
						CampaignID:                campaign.ID,
						CurrencyCode:              campaign.CurrencyCode,
						RewardedMemberID:          member.ReferredByMember.ID,
						RewardedMemberReferenceID: member.ReferredByMember.ReferenceID,
						RelatedMemberID:           member.ID,
						RelatedMemberReferenceID:  member.ReferenceID,
						MemberType:                "referrer",
						Amount:                    *referrerRewardAmount,
						Status:                    "pending",
					}
					if err := tx.Create(referrerReward).Error; err != nil {
						fmt.Printf("failed to create reward for campaign %d: %v\n", campaign.ID, err)
						return err
					}
				}

				if refereeRewardAmount != nil && refereeRewardAmount.GreaterThan(decimal.NewFromInt(0)) {
					refereeReward = &models.Reward{
						Project:                   project,
						CampaignID:                campaign.ID,
						CurrencyCode:              campaign.CurrencyCode,
						RewardedMemberID:          member.ID,
						RewardedMemberReferenceID: member.ReferenceID,
						RelatedMemberID:           member.ReferredByMember.ID,
						RelatedMemberReferenceID:  member.ReferredByMember.ReferenceID,
						MemberType:                "referee",
						Amount:                    *refereeRewardAmount,
						Status:                    "pending",
					}
					if err := tx.Create(refereeReward).Error; err != nil {
						fmt.Printf("failed to create referee reward for campaign %d: %v\n", campaign.ID, err)
						return err
					}
				}
				// Prepare bulk insert data for referral_campaign_event_logs
				var campaignEventStatusEntries []models.CampaignEventLog

				for i, eventLogID := range eventLogIDs {
					entry := models.CampaignEventLog{
						Project:           campaign.Project,
						CampaignID:        campaign.ID,
						EventID:           logs[i].ID,                // Assuming you have eventID from previous logic
						MemberID:          logs[i].MemberID,          // Assuming you have memberID from previous logic
						MemberReferenceID: logs[i].MemberReferenceID, // Assuming you have memberReferenceID from previous logic
						Status:            "processed",
						EventLogID:        eventLogID,
					}

					if referrerReward != nil {
						entry.ReferredRewardID = &referrerReward.ID
					}
					if refereeReward != nil {
						entry.RefereeRewardID = &refereeReward.ID
					}

					campaignEventStatusEntries = append(campaignEventStatusEntries, entry)
				}

				// Perform bulk insert
				if err := tx.Create(&campaignEventStatusEntries).Error; err != nil {
					return fmt.Errorf("failed to bulk insert into referral_campaign_event_logs: %w", err)
				}

				return nil
			})

			if err != nil {
				// Log the error and continue with other campaigns
				fmt.Printf("Error processing campaign %d: %v\n", campaign.ID, err)
				if errors.Is(err, ErrExceedsBudget) {
					fmt.Printf("Break Campaign %d exceeds budget\n", campaign.ID)
					break
				}
			}
		}
	}

	return nil
}

func (w *worker) validateReward(tx *gorm.DB, err error, project string, campaign models.Campaign, referenceID string, rewardAmount *decimal.Decimal) error {
	// Validate limits
	referrerTotalReward, referrerMonthsPassed, referrerRewardsCount, err := w.GetTotalRewardByMember(tx, project, campaign.ID, referenceID)
	if err != nil {
		fmt.Printf("GetTotalRewardByMember: failed to calculate total reward for campaign %d: %v\n", campaign.ID, err)
		return err
	}

	// Reward Cap Per Customer
	if campaign.RewardCapPerCustomer != nil && referrerTotalReward.Add(*rewardAmount).GreaterThan(*campaign.RewardCapPerCustomer) {
		return errors.New("exceeds reward cap per customer")
	}

	// Check if the validity period is exceeded
	if campaign.ValidityMonthsPerCustomer != nil && referrerMonthsPassed >= *campaign.ValidityMonthsPerCustomer {
		return errors.New("exceeds validity period")
	}

	// Check if max occurrences are exceeded
	if campaign.MaxOccurrencesPerCustomer != nil && referrerRewardsCount >= *campaign.MaxOccurrencesPerCustomer {
		return errors.New("exceeds max occurrences per customer")
	}
	return nil
}

func (w *worker) GetTotalRewardByMember(
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
		Where("project = ? AND campaign_id = ? AND rewarded_member_reference_id = ?", project, campaignID, referrerReferenceID).
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

func calculateReward(tx *gorm.DB, campaign models.Campaign, logs []models.EventLog) (*decimal.Decimal, *decimal.Decimal, error) {
	var referrerReward *decimal.Decimal
	var refereeReward *decimal.Decimal
	if campaign.RewardType != nil {
		if *campaign.RewardType == "flat_fee" {
			referrerReward = campaign.RewardValue
		} else if *campaign.RewardType == "percentage" {
			// Sum the total amount from event logs for percentage calculation
			var totalAmount decimal.Decimal
			if err := tx.Model(&models.EventLog{}).
				Where("id IN (?)", getEventLogIDs(logs)).
				Select("COALESCE(SUM(amount), 0)").
				Scan(&totalAmount).Error; err != nil {
				return nil, nil, fmt.Errorf("failed to calculate total amount from event logs: %w", err)
			}

			// Calculate the percentage-based reward
			percentage := campaign.RewardValue.Div(decimal.NewFromInt(100))
			reward := totalAmount.Mul(percentage)
			referrerReward = &reward
			//return &reward, nil
		}
	}
	if campaign.InviteeRewardType != nil {
		if *campaign.InviteeRewardType == "flat_fee" {
			refereeReward = campaign.InviteeRewardValue
		} else if campaign.InviteeRewardType != nil && *campaign.InviteeRewardType == "percentage" {
			// Sum the total amount from event logs for percentage calculation
			var totalAmount decimal.Decimal
			if err := tx.Model(&models.EventLog{}).
				Where("id IN (?)", getEventLogIDs(logs)).
				Select("COALESCE(SUM(amount), 0)").
				Scan(&totalAmount).Error; err != nil {
				return nil, nil, fmt.Errorf("failed to calculate total amount from event logs: %w", err)
			}

			// Calculate the percentage-based reward
			percentage := campaign.InviteeRewardValue.Div(decimal.NewFromInt(100))
			reward := totalAmount.Mul(percentage)
			refereeReward = &reward
		}
	}

	return referrerReward, refereeReward, nil
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
			if group[0].MemberReferenceID == log.MemberReferenceID && !hasAllKeys(group, requiredKeys) {
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
