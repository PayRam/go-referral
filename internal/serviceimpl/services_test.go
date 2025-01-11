package serviceimpl_test

import (
	"fmt"
	go_referral "github.com/PayRam/go-referral"
	db2 "github.com/PayRam/go-referral/internal/db"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/PayRam/go-referral/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"testing"
	"time"
)

var (
	db              *gorm.DB
	referralService *go_referral.ReferralService
)

var project = "reftest"

func TestMain(m *testing.M) {
	// Initialize shared test database
	var err error
	//db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db, err = gorm.Open(sqlite.Open("/Users/sameer/Documents/test1.db"), &gorm.Config{})
	if err != nil {
		panic("failed to initialize test database")
	}

	referralService = go_referral.NewReferralService(db)

	// Run tests
	m.Run()
}

func setupEvents(t *testing.T) {

	event1, err := referralService.Events.CreateEvent(project, "signup-event", "User Signup", nil, "simple")
	assert.NoError(t, err)
	assert.NotNil(t, event1)
	assert.Equal(t, "User Signup", event1.Name)

	event2, err := referralService.Events.CreateEvent(project, "payment-event", "Payment Made", nil, "payment")
	assert.NoError(t, err)
	assert.NotNil(t, event2)
	assert.Equal(t, "Payment Made", event2.Name)

	event3, err := referralService.Events.UpdateEvent(project, "payment-event", request.UpdateEventRequest{
		Name: utils.StringPtr("Payment Done"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, event3)
	assert.Equal(t, "Payment Done", event3.Name)

	event4, err := referralService.Events.SearchByName(project, "Payment")
	if err != nil {
		return
	}
	assert.NoError(t, err)
	assert.NotNil(t, event4[0])
	assert.Equal(t, "Payment Done", event4[0].Name)
}

func setupCampaign(t *testing.T) {
	events, err := referralService.Events.GetAll(project)
	if err != nil {
		return
	}

	// Create a campaign using event keys
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0) // One month from start date
	budget := decimal.NewFromFloat(1000.00)
	campaign, err := referralService.Campaigns.CreateCampaign(
		project,
		"New User Campaign",
		"Campaign for new user signups and payments",
		startDate,
		endDate,
		nil,
		nil, nil, nil, nil,
		&budget,
	)
	assert.NoError(t, err)
	assert.NotNil(t, campaign)
	assert.Equal(t, "New User Campaign", campaign.Name)

	var rewardType = "percentage"
	var rewardValue = 10.0
	var inviteeRewardType = "flat_fee"
	var inviteeRewardValue = 100.0
	var maxOccurrences = uint(0)
	var validityDays = uint(60)

	var updateCampaignRequest = request.UpdateCampaignRequest{
		RewardType:         &rewardType,
		RewardValue:        &rewardValue,
		InviteeRewardType:  &inviteeRewardType,
		InviteeRewardValue: &inviteeRewardValue,
		MaxOccurrences:     &maxOccurrences,
		ValidityDays:       &validityDays,
	}

	campaign, err = referralService.Campaigns.UpdateCampaign(
		project,
		campaign.ID,
		updateCampaignRequest,
	)
	assert.NoError(t, err)
	assert.NotNil(t, campaign)
	assert.Equal(t, "percentage", *campaign.RewardType)

	campaign, err = referralService.Campaigns.UpdateCampaignEvents(
		project,
		campaign.ID,
		events,
	)
	assert.NoError(t, err)
	assert.NotNil(t, campaign)

	// Verify associations
	var campaignEvents []models.CampaignEvent
	err = db.Where("campaign_id = ?", campaign.ID).Find(&campaignEvents).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, len(campaignEvents))

	campaigns, err := referralService.Campaigns.SearchByName(project, "User Campaign")
	if err != nil {
		return
	}
	assert.NoError(t, err)
	assert.NotNil(t, campaigns[0])
	assert.Equal(t, "New User Campaign", campaigns[0].Name)
}

func setupReferrer(t *testing.T) {
	condition := db2.QueryCondition{
		Field:    "id",
		Operator: "=",
		Value:    1,
	}
	campaigns, err := referralService.Campaigns.GetCampaigns(project, []db2.QueryCondition{condition}, 0, 1, nil)
	if err != nil {
		return
	}
	campaign := campaigns[0]
	code := utils.GenerateReferralCode()
	// Create a referrer
	referrer, err := referralService.Referrers.CreateReferrer(
		project,
		"user-123",          // ReferenceID
		code,                // Unique code
		[]uint{campaign.ID}, // CampaignID
	)
	assert.NoError(t, err)
	assert.NotNil(t, referrer)

	// Validate the referrer
	assert.Equal(t, code, referrer.Code)
	assert.Equal(t, project, referrer.Project)
	assert.Equal(t, "user-123", referrer.ReferenceID)
	assert.Equal(t, campaign.ID, referrer.Campaigns[0].ID)

	// Fetch and validate the referrer from the database
	var dbReferrer models.Referrer
	err = db.Preload("Campaigns").Where("id = ?", referrer.ID).First(&dbReferrer).Error
	assert.NoError(t, err)
	assert.Equal(t, code, dbReferrer.Code)
	assert.Equal(t, project, dbReferrer.Project)
	assert.Equal(t, "user-123", dbReferrer.ReferenceID)
	assert.Equal(t, campaign.ID, dbReferrer.Campaigns[0].ID)
}

func setupReferee(t *testing.T) {
	referrer, err := referralService.Referrers.GetReferrerByReference(project, "user-123")
	if err != nil {
		return
	}
	// Create a Referee using the Referrer's code
	referee, err := referralService.Referees.CreateReferee(
		project,
		referrer.Code, // Referrer code
		"user-456",    // ReferenceID
	)
	assert.NoError(t, err)
	assert.NotNil(t, referee)

	// Validate the Referee
	assert.Equal(t, referrer.ID, referee.ReferrerID)
	assert.Equal(t, project, referee.Project)
	assert.Equal(t, "user-456", referee.ReferenceID)

	// Fetch and validate the Referee from the database
	var dbReferee models.Referee
	err = db.Preload("Referrer").Where("id = ?", referee.ID).First(&dbReferee).Error
	assert.NoError(t, err)
	assert.Equal(t, referrer.ID, dbReferee.ReferrerID)
	assert.Equal(t, project, dbReferee.Project)
	assert.Equal(t, "user-456", dbReferee.ReferenceID)
	assert.Equal(t, referrer.Code, dbReferee.Referrer.Code)
}

func TestCreateReferee(t *testing.T) {
	setupEvents(t)   // Ensure events exist
	setupCampaign(t) // Ensure campaign exists
	setupReferrer(t)
	setupReferee(t)

	_, err := triggerSignupEvent(t)
	_, err = triggerPaymentEvent(t)

	err = referralService.Worker.ProcessPendingEvents()
	assert.NoError(t, err)

	// Fetch all rewards
	var rewards []models.Reward
	if err := db.Find(&rewards).Error; err != nil {
		log.Fatalf("failed to fetch rewards: %v", err)
	}

	// Print each reward
	for _, reward := range rewards {
		fmt.Printf("Project: %s\n", reward.Project)
		fmt.Printf("Reward ID: %d\n", reward.ID)
		fmt.Printf("CampaignID: %d\n", reward.CampaignID)
		fmt.Printf("RefereeID: %d\n", reward.RefereeID)
		fmt.Printf("ReferenceID: %s\n", reward.ReferenceID)
		fmt.Printf("Amount: %s\n", reward.Amount.String())
		fmt.Printf("Status: %s\n", reward.Status)
		if reward.Reason != nil {
			fmt.Printf("Reason: %s\n", *reward.Reason)
		} else {
			fmt.Println("Reason: None")
		}
		fmt.Println("--------------------------")
	}

}

func triggerSignupEvent(t *testing.T) (*models.EventLog, error) {
	// Create an EventLog for the Referee
	eventKey := "signup-event"
	amount := decimal.NewFromFloat(100.50)
	data := `{"transactionId": "12345"}`
	user := "user-456"
	eventLog, err := referralService.EventLogs.CreateEventLog(
		project,
		eventKey,
		user,
		&amount,
		&data,
	)
	assert.NoError(t, err)
	assert.NotNil(t, eventLog)

	// Verify the EventLog is created correctly
	assert.Equal(t, project, eventLog.Project)
	assert.Equal(t, eventKey, eventLog.EventKey)
	assert.Equal(t, user, eventLog.ReferenceID)
	assert.Equal(t, data, *eventLog.Data)
	assert.Equal(t, "pending", eventLog.Status)

	// Verify it exists in the database
	var retrievedEventLog models.EventLog
	err = db.First(&retrievedEventLog, eventLog.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, eventLog.ID, retrievedEventLog.ID)

	return eventLog, err
}

func triggerPaymentEvent(t *testing.T) (*models.EventLog, error) {
	// Create an EventLog for the Referee
	eventKey := "payment-event"
	amount := decimal.NewFromFloat(100.50)
	data := `{"transactionId": "12345"}`
	user := "user-456"
	eventLog, err := referralService.EventLogs.CreateEventLog(
		project,
		eventKey,
		user,
		&amount,
		&data,
	)
	assert.NoError(t, err)
	assert.NotNil(t, eventLog)

	// Verify the EventLog is created correctly
	assert.Equal(t, project, eventLog.Project)
	assert.Equal(t, eventKey, eventLog.EventKey)
	assert.Equal(t, user, eventLog.ReferenceID)
	assert.Equal(t, amount, *eventLog.Amount)
	assert.Equal(t, data, *eventLog.Data)
	assert.Equal(t, "pending", eventLog.Status)

	// Verify it exists in the database
	var retrievedEventLog models.EventLog
	err = db.First(&retrievedEventLog, eventLog.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, eventLog.ID, retrievedEventLog.ID)

	return eventLog, err
}
