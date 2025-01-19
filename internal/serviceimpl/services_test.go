package serviceimpl_test

import (
	go_referral "github.com/PayRam/go-referral"
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

func TestMain(m *testing.M) {
	// Initialize shared test database
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	//db, err = gorm.Open(sqlite.Open("/Users/sameer/Documents/test1.db"), &gorm.Config{})
	if err != nil {
		panic("failed to initialize test database")
	}

	referralService = go_referral.NewReferralService(db)

	// Run tests
	m.Run()
}

func createEvent(t *testing.T, project string, req request.CreateEventRequest) *models.Event {
	event, err := referralService.Events.CreateEvent(project, req)
	assert.NoError(t, err, "failed to create event")
	assert.NotNil(t, event)
	assert.Equal(t, req.Key, event.Key)
	assert.Equal(t, req.Name, event.Name)
	assert.Equal(t, req.EventType, event.EventType)
	utils.AssertEqualNilable(t, req.Description, event.Description, "Description values should match")
	return event
}

func updateEvent(t *testing.T, project, key string, req request.UpdateEventRequest) *models.Event {
	event, err := referralService.Events.UpdateEvent(project, key, req)
	assert.NoError(t, err, "failed to update event")
	assert.NotNil(t, event)
	if req.Name != nil {
		assert.Equal(t, *req.Name, event.Name)
	}
	if req.Description != nil {
		utils.AssertEqualNilable(t, req.Description, event.Description, "Description values should match")
	}
	return event
}

func createCampaign(t *testing.T, project string, req request.CreateCampaignRequest) *models.Campaign {
	campaign, err := referralService.Campaigns.CreateCampaign(project, req)
	assert.NoError(t, err, "failed to create campaign")
	assert.NotNil(t, campaign)
	assert.Equal(t, req.Name, campaign.Name)
	assert.Equal(t, req.RewardType, campaign.RewardType)
	assert.Equal(t, req.IsDefault, campaign.IsDefault)
	assert.Equal(t, req.CampaignTypePerCustomer, campaign.CampaignTypePerCustomer)
	assert.Equal(t, len(req.EventKeys), len(campaign.Events))
	return campaign
}

func updateCampaign(t *testing.T, project string, campaignID uint, req request.UpdateCampaignRequest) *models.Campaign {
	campaign, err := referralService.Campaigns.UpdateCampaign(project, campaignID, req)
	if err != nil {
		log.Printf("failed to create campaign: %v", err)
	}
	assert.NoError(t, err)
	assert.NotNil(t, campaign)
	utils.AssertEqualIfExpectedNotNil(t, req.Name, campaign.Name, "Name values should match")
	utils.AssertEqualIfExpectedNotNil(t, req.RewardType, campaign.RewardType, "RewardType values should match")
	utils.AssertEqualIfExpectedNotNil(t, req.IsDefault, campaign.IsDefault, "IsDefault values should match")
	utils.AssertEqualIfExpectedNotNil(t, req.CampaignTypePerCustomer, campaign.CampaignTypePerCustomer, "CampaignTypePerCustomer values should match")
	assert.Equal(t, len(req.EventKeys), len(campaign.Events))
	return campaign
}

func createReferrer(t *testing.T, project, referrerUser string, campaignIDs []uint) *models.Referrer {
	code, err := utils.CreateReferralCode(7)
	assert.NoError(t, err)
	// Create a referrer
	referrer, err := referralService.Referrers.CreateReferrer(
		project,
		request.CreateReferrerRequest{
			Code:        &code,
			ReferenceID: referrerUser,
			CampaignIDs: campaignIDs,
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, referrer)

	// Validate the referrer
	assert.Equal(t, code, referrer.Code)
	assert.Equal(t, project, referrer.Project)
	assert.Equal(t, referrerUser, referrer.ReferenceID)
	assert.Equal(t, len(campaignIDs), len(referrer.Campaigns))
	return referrer
}

func createReferee(t *testing.T, project, code, refereeUser string) *models.Referee {
	req := request.GetReferrerRequest{
		Project: &project,
		Code:    &code,
	}

	referrers, _, err := referralService.Referrers.GetReferrers(req)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(referrers), "expected exactly 1 referrer")

	// Create a Referee using the Referrer's code
	referee, err := referralService.Referees.CreateReferee(
		project,
		request.CreateRefereeRequest{
			Code:        code,
			ReferenceID: refereeUser,
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, referee)
	assert.Equal(t, referrers[0].ID, referee.ReferrerID)
	assert.Equal(t, project, referee.Project)
	assert.Equal(t, refereeUser, referee.ReferenceID)
	return referee
}

func triggerEvent(t *testing.T, project, eventKey, user string, data *string, amount *decimal.Decimal) (*models.EventLog, error) {
	// Create an EventLog for the Referee
	eventLog, err := referralService.EventLogs.CreateEventLog(
		project,
		request.CreateEventLogRequest{
			EventKey:    eventKey,
			ReferenceID: user,
			Amount:      amount,
			Data:        data,
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, eventLog)

	// Verify the EventLog is created correctly
	assert.Equal(t, project, eventLog.Project)
	assert.Equal(t, eventKey, eventLog.EventKey)
	assert.Equal(t, user, eventLog.ReferenceID)
	if data == nil && eventLog.Data == nil {
		assert.True(t, true, "Both data and eventLog.Data are nil")
	} else if data != nil && eventLog.Data != nil {
		assert.Equal(t, *data, *eventLog.Data, "Data values should match")
	} else {
		t.Errorf("One of data or eventLog.Data is nil while the other is not")
	}

	if amount == nil && eventLog.Amount == nil {
		assert.True(t, true, "Both amount and eventLog.Amount are nil")
	} else if amount != nil && eventLog.Amount != nil {
		assert.Equal(t, *amount, *eventLog.Amount, "Amount values should match")
	} else {
		t.Errorf("One of amount or eventLog.Amount is nil while the other is not")
	}
	assert.Equal(t, "pending", eventLog.Status)

	// Verify it exists in the database
	var retrievedEventLog models.EventLog
	err = db.First(&retrievedEventLog, eventLog.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, eventLog.ID, retrievedEventLog.ID)

	return eventLog, err
}

func TestOneTimeCampaign(t *testing.T) {
	project := "onetimeproject"
	referrerUser := "user-123"
	refereeUser := "user-456"
	event := createEvent(t, project, request.CreateEventRequest{
		Key:       "test-event",
		Name:      "Test Event",
		EventType: "payment",
	})

	event = updateEvent(t, project, "test-event", request.UpdateEventRequest{
		Name: utils.StringPtr("Test Done"),
	})

	// Create a campaign using event keys
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0) // One month from start date
	budget := decimal.NewFromFloat(100.00)
	description := "Campaign for new user signups and payments"
	rewardValue := decimal.NewFromFloat(10.0)
	var inviteeRewardType = "flat_fee"
	var inviteeRewardValue = decimal.NewFromFloat(40.0)
	//var inviteeRewardCap = decimal.NewFromFloat(1000.0)

	campaign := createCampaign(t, project, request.CreateCampaignRequest{
		Name:                    "New User Campaign",
		RewardType:              "percentage",
		RewardValue:             rewardValue,
		StartDate:               &startDate,
		EndDate:                 &endDate,
		Description:             &description,
		Budget:                  &budget,
		IsDefault:               true,
		CampaignTypePerCustomer: "one_time",
		InviteeRewardType:       &inviteeRewardType,
		InviteeRewardValue:      &inviteeRewardValue,

		EventKeys: []string{event.Key},
	})

	event1 := createEvent(t, project, request.CreateEventRequest{
		Key:       "signup-event",
		Name:      "User Signup",
		EventType: "simple",
	})
	event2 := createEvent(t, project, request.CreateEventRequest{
		Key:       "payment-event",
		Name:      "Payment Made",
		EventType: "payment",
	})

	campaign = updateCampaign(t, project, campaign.ID, request.UpdateCampaignRequest{
		Name:      utils.StringPtr("New User Campaign Updated"),
		EventKeys: []string{event1.Key, event2.Key},
	})

	referrer := createReferrer(t, project, referrerUser, []uint{campaign.ID})

	referee := createReferee(t, project, referrer.Code, refereeUser)

	amount := decimal.NewFromFloat(100.50)
	_, err := triggerEvent(t, project, "signup-event", refereeUser, nil, nil)
	_, err = triggerEvent(t, project, "signup-event", refereeUser, nil, nil)
	_, err = triggerEvent(t, project, "payment-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount)
	_, err = triggerEvent(t, project, "payment-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount)

	err = referralService.Worker.ProcessPendingEvents()
	if err != nil {
		log.Fatalf("****************failed to process pending events: %v", err)
	}
	assert.NoError(t, err)

	req := request.GetRewardRequest{
		Project:             &project,
		ReferrerReferenceID: &referrerUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	rewards, count, err := referralService.RewardService.GetRewards(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	expectedReward := decimal.NewFromFloat(10.05)

	assert.Equal(t, project, rewards[0].Project)
	assert.Equal(t, campaign.ID, rewards[0].CampaignID)
	assert.Equal(t, referee.ID, rewards[0].RefereeID)
	assert.Equal(t, refereeUser, rewards[0].RefereeReferenceID)
	assert.Equal(t, referrer.ID, rewards[0].ReferrerID)
	assert.Equal(t, referrerUser, rewards[0].ReferrerReferenceID)
	assert.Equal(t, "pending", rewards[0].Status)
	assert.Equal(t, expectedReward.String(), rewards[0].Amount.String())

	elreq := request.GetEventLogRequest{
		Project:     &project,
		ReferenceID: &refereeUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	eventLogs, count, err := referralService.EventLogs.GetEventLogs(elreq)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), count)

	assert.Equal(t, rewards[0].ID, eventLogs[0].RewardID)
	assert.Equal(t, uint(0), eventLogs[1].RewardID)
	assert.Equal(t, rewards[0].ID, eventLogs[2].RewardID)
	assert.Equal(t, uint(0), eventLogs[3].RewardID)

}

func TestRecurringCampaignWithRewardCapAndLimitedBudget(t *testing.T) {
	project := "recurringproject"
	referrerUser := "user-123"
	refereeUser := "user-456"
	event1 := createEvent(t, project, request.CreateEventRequest{
		Key:       "payment-recurring-event",
		Name:      "Payment Recurring Event",
		EventType: "payment",
	})

	// Create a campaign using event keys
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0) // One month from start date
	budget := decimal.NewFromFloat(30.00)
	description := "Campaign for new user signups and payments"
	rewardValue := decimal.NewFromFloat(8.49)
	rewardCap := decimal.NewFromFloat(15.00)
	//var inviteeRewardCap = decimal.NewFromFloat(1000.0)
	maxOccurrencesPerCustomer := int64(10)
	campaign := createCampaign(t, project, request.CreateCampaignRequest{
		Name:                      "New User Campaign",
		RewardType:                "percentage",
		RewardValue:               rewardValue,
		RewardCap:                 &rewardCap,
		StartDate:                 &startDate,
		EndDate:                   &endDate,
		Description:               &description,
		Budget:                    &budget,
		IsDefault:                 true,
		CampaignTypePerCustomer:   "count_per_customer",
		MaxOccurrencesPerCustomer: &maxOccurrencesPerCustomer,

		EventKeys: []string{event1.Key},
	})

	referrer := createReferrer(t, project, referrerUser, []uint{campaign.ID})

	referee := createReferee(t, project, referrer.Code, refereeUser)

	amount1 := decimal.NewFromFloat(150.50)
	amount2 := decimal.NewFromFloat(330.50)
	amount3 := decimal.NewFromFloat(430.50)
	//_, err := triggerEvent(t, project, "signup-event", refereeUser, nil, nil)
	//_, err = triggerEvent(t, project, "signup-event", refereeUser, nil, nil)
	_, err := triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount1)
	_, err = triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount2)
	_, err = triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount3)

	err = referralService.Worker.ProcessPendingEvents()
	assert.NoError(t, err)

	req := request.GetRewardRequest{
		Project:             &project,
		ReferrerReferenceID: &referrerUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	rewards, count, err := referralService.RewardService.GetRewards(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	expectedReward := decimal.NewFromFloat(12.77745) //27.77745
	expectedReward2 := decimal.NewFromFloat(15)      //27.77745

	assert.Equal(t, project, rewards[0].Project)
	assert.Equal(t, campaign.ID, rewards[0].CampaignID)
	assert.Equal(t, referee.ID, rewards[0].RefereeID)
	assert.Equal(t, refereeUser, rewards[0].RefereeReferenceID)
	assert.Equal(t, referrer.ID, rewards[0].ReferrerID)
	assert.Equal(t, referrerUser, rewards[0].ReferrerReferenceID)
	assert.Equal(t, "pending", rewards[0].Status)
	assert.Equal(t, expectedReward.String(), rewards[0].Amount.String())
	assert.Equal(t, expectedReward2.String(), rewards[1].Amount.String())

	elreq := request.GetEventLogRequest{
		Project:     &project,
		ReferenceID: &refereeUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	eventLogs, count, err := referralService.EventLogs.GetEventLogs(elreq)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	assert.Equal(t, rewards[0].ID, eventLogs[0].RewardID)
	assert.Equal(t, rewards[1].ID, eventLogs[1].RewardID)
}

func TestRecurringCampaignWithMaxOccurrencesPerCustomer(t *testing.T) {
	project := "recumaxoccurrenceproject"
	referrerUser := "user-123"
	refereeUser := "user-456"
	event1 := createEvent(t, project, request.CreateEventRequest{
		Key:       "payment-recurring-event",
		Name:      "Payment Recurring Event",
		EventType: "payment",
	})

	// Create a campaign using event keys
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0) // One month from start date
	budget := decimal.NewFromFloat(3000.00)
	description := "Campaign for new user signups and payments"
	rewardValue := decimal.NewFromFloat(12.34)
	//rewardCap := decimal.NewFromFloat(15.00)
	//var inviteeRewardCap = decimal.NewFromFloat(1000.0)
	maxOccurrencesPerCustomer := int64(2)
	campaign := createCampaign(t, project, request.CreateCampaignRequest{
		Name:        "New User Campaign",
		RewardType:  "percentage",
		RewardValue: rewardValue,
		//RewardCap:                 &rewardCap,
		StartDate:                 &startDate,
		EndDate:                   &endDate,
		Description:               &description,
		Budget:                    &budget,
		IsDefault:                 true,
		CampaignTypePerCustomer:   "count_per_customer",
		MaxOccurrencesPerCustomer: &maxOccurrencesPerCustomer,

		EventKeys: []string{event1.Key},
	})

	referrer := createReferrer(t, project, referrerUser, []uint{campaign.ID})

	referee := createReferee(t, project, referrer.Code, refereeUser)

	amount1 := decimal.NewFromFloat(250.50)
	amount2 := decimal.NewFromFloat(1510.74565)
	amount3 := decimal.NewFromFloat(1430.346)
	_, err := triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount1)
	_, err = triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount2)
	_, err = triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount3)

	err = referralService.Worker.ProcessPendingEvents()
	assert.NoError(t, err)

	req := request.GetRewardRequest{
		Project:             &project,
		ReferrerReferenceID: &referrerUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	rewards, count, err := referralService.RewardService.GetRewards(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	expectedReward := decimal.NewFromFloat(30.9117)       //27.77745
	expectedReward2 := decimal.NewFromFloat(186.42601321) //27.77745

	assert.Equal(t, project, rewards[0].Project)
	assert.Equal(t, campaign.ID, rewards[0].CampaignID)
	assert.Equal(t, referee.ID, rewards[0].RefereeID)
	assert.Equal(t, refereeUser, rewards[0].RefereeReferenceID)
	assert.Equal(t, referrer.ID, rewards[0].ReferrerID)
	assert.Equal(t, referrerUser, rewards[0].ReferrerReferenceID)
	assert.Equal(t, "pending", rewards[0].Status)
	assert.Equal(t, expectedReward.String(), rewards[0].Amount.String())
	assert.Equal(t, expectedReward2.String(), rewards[1].Amount.String())

	elreq := request.GetEventLogRequest{
		Project:     &project,
		ReferenceID: &refereeUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	eventLogs, count, err := referralService.EventLogs.GetEventLogs(elreq)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	assert.Equal(t, rewards[0].ID, eventLogs[0].RewardID)
	assert.Equal(t, rewards[1].ID, eventLogs[1].RewardID)
	//assert.Equal(t, rewards[1].ID, eventLogs[1].RewardID)
}

func TestAggregator(t *testing.T) {
	stats, count, err := referralService.AggregatorService.GetReferrersWithStats(request.GetReferrerRequest{
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	assert.Equal(t, "onetimeproject", stats[0].Project)
	assert.Equal(t, int64(1), stats[0].RefereeCount)
	assert.Equal(t, "10.05", stats[0].TotalRewards.String())

	assert.Equal(t, "recurringproject", stats[1].Project)
	assert.Equal(t, int64(1), stats[1].RefereeCount)
	assert.Equal(t, "27.77745", stats[1].TotalRewards.String())

	assert.Equal(t, "recumaxoccurrenceproject", stats[2].Project)
	assert.Equal(t, int64(1), stats[2].RefereeCount)
	assert.Equal(t, "217.33771321", stats[2].TotalRewards.String())
}
