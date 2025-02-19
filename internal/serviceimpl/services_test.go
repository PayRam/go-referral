package serviceimpl_test

import (
	"fmt"
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
	//db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db, err = gorm.Open(sqlite.Open("/Users/sameer/Documents/test1.db"), &gorm.Config{})
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
	assert.Equal(t, *req.RewardType, *campaign.RewardType)
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
	utils.AssertEqualIfExpectedNotNil(t, req.RewardType, *campaign.RewardType, "RewardType values should match")
	utils.AssertEqualIfExpectedNotNil(t, req.IsDefault, campaign.IsDefault, "IsDefault values should match")
	utils.AssertEqualIfExpectedNotNil(t, req.CampaignTypePerCustomer, campaign.CampaignTypePerCustomer, "CampaignTypePerCustomer values should match")
	assert.Equal(t, len(req.EventKeys), len(campaign.Events))
	return campaign
}

func createReferrer(t *testing.T, project, referrerUser string, campaignIDs []uint, email *string) *models.Member {
	code, err := utils.CreateReferralCode(7)
	assert.NoError(t, err)
	// Create a referrer
	referrer, err := referralService.Members.CreateMember(
		project,
		request.CreateMemberRequest{
			PreferredCode: &code,
			ReferenceID:   referrerUser,
			CampaignIDs:   campaignIDs,
			Email:         email,
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

func createReferee(t *testing.T, project, code, refereeUser string, email *string) *models.Member {
	req := request.GetMemberRequest{
		Projects: []string{project},
		Code:     &code,
	}

	referrers, _, err := referralService.Members.GetMembers(req)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(referrers), "expected exactly 1 referrer")

	// Create a Referee using the Member's code
	referee, err := referralService.Members.CreateMember(
		project,
		request.CreateMemberRequest{
			ReferrerCode: &code,
			ReferenceID:  refereeUser,
			Email:        email,
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, referee)
	assert.Equal(t, referrers[0].ID, *referee.ReferredByMemberID)
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
	assert.Equal(t, user, eventLog.MemberReferenceID)
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
	referrerEmail := "abc@gmail.com"
	refereeUser := "user-456"
	refereeEmail := "cdd@gmail.com"
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
	var rewardType = "percentage"
	var inviteeRewardType = "percentage"
	var inviteeRewardValue = decimal.NewFromFloat(5.0)
	//var inviteeRewardCap = decimal.NewFromFloat(1000.0)

	campaign := createCampaign(t, project, request.CreateCampaignRequest{
		Name:                    "New User Campaign",
		RewardType:              &rewardType,
		RewardValue:             &rewardValue,
		CurrencyCode:            "USDC",
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

	referrer := createReferrer(t, project, referrerUser, []uint{campaign.ID}, &referrerEmail)

	referee := createReferee(t, project, referrer.Code, refereeUser, &refereeEmail)

	amount := decimal.NewFromFloat(100.50)
	_, err := triggerEvent(t, project, "signup-event", refereeUser, nil, nil)
	_, err = triggerEvent(t, project, "signup-event", refereeUser, nil, nil)
	_, err = triggerEvent(t, project, "payment-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount)
	_, err = triggerEvent(t, project, "payment-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount)

	err = referralService.Worker.ProcessPendingEvents()
	assert.NoError(t, err)

	req := request.GetRewardRequest{
		Projects: []string{project},
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	rewards, count, err := referralService.Reward.GetRewards(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	referredMemberExpectedReward := decimal.NewFromFloat(10.05)
	refereeMemberExpectedReward2 := decimal.NewFromFloat(5.025)

	assert.Equal(t, project, rewards[0].Project)
	assert.Equal(t, campaign.ID, rewards[0].CampaignID)
	assert.Equal(t, referee.ID, rewards[0].RelatedMemberID)
	assert.Equal(t, refereeUser, rewards[0].RelatedMemberReferenceID)
	assert.Equal(t, referrer.ID, rewards[0].RewardedMemberID)
	assert.Equal(t, referrerUser, rewards[0].RewardedMemberReferenceID)
	assert.Equal(t, "referrer", rewards[0].MemberType)
	assert.Equal(t, "pending", rewards[0].Status)
	assert.Equal(t, referredMemberExpectedReward.String(), rewards[0].Amount.String())

	assert.Equal(t, project, rewards[1].Project)
	assert.Equal(t, campaign.ID, rewards[1].CampaignID)
	assert.Equal(t, referrer.ID, rewards[1].RelatedMemberID)
	assert.Equal(t, referrerUser, rewards[1].RelatedMemberReferenceID)
	assert.Equal(t, referee.ID, rewards[1].RewardedMemberID)
	assert.Equal(t, refereeUser, rewards[1].RewardedMemberReferenceID)
	assert.Equal(t, "referee", rewards[1].MemberType)
	assert.Equal(t, "pending", rewards[1].Status)
	assert.Equal(t, refereeMemberExpectedReward2.String(), rewards[1].Amount.String())

	elreg := request.GetCampaignEventLogRequest{
		Projects:           []string{project},
		MemberReferenceIDs: []string{refereeUser},
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}
	fmt.Print(refereeUser)
	campaignEventLogs, count, err := referralService.CampaignEventLog.GetCampaignEventLogs(elreg)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Equal(t, rewards[0].ID, *campaignEventLogs[0].ReferredRewardID)
	assert.Equal(t, rewards[1].ID, *campaignEventLogs[0].RefereeRewardID)
	assert.Equal(t, rewards[0].ID, *campaignEventLogs[1].ReferredRewardID)
	assert.Equal(t, rewards[1].ID, *campaignEventLogs[1].RefereeRewardID)
}

func TestRecurringCampaignWithRewardCapAndLimitedBudget(t *testing.T) {
	project := "recurringproject"
	referrerUser := "user-123"
	referrerEmail := "test@gmail.com"
	refereeUser := "user-456"
	refereeEmail := "trr@gmail.com"
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
	var rewardType = "percentage"
	rewardValue := decimal.NewFromFloat(8.49)
	rewardCap := decimal.NewFromFloat(15.00)
	//var inviteeRewardCap = decimal.NewFromFloat(1000.0)
	maxOccurrencesPerCustomer := int64(10)
	campaign := createCampaign(t, project, request.CreateCampaignRequest{
		Name:                      "New User Campaign",
		RewardType:                &rewardType,
		RewardValue:               &rewardValue,
		CurrencyCode:              "USDT",
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

	referrer := createReferrer(t, project, referrerUser, []uint{campaign.ID}, &referrerEmail)

	referee := createReferee(t, project, referrer.Code, refereeUser, &refereeEmail)

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
		Projects:                  []string{project},
		RewardedMemberReferenceID: &referrerUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	rewards, count, err := referralService.Reward.GetRewards(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	expectedReward := decimal.NewFromFloat(12.77745) //27.77745
	expectedReward2 := decimal.NewFromFloat(15)      //27.77745

	assert.Equal(t, project, rewards[0].Project)
	assert.Equal(t, campaign.ID, rewards[0].CampaignID)
	assert.Equal(t, referee.ID, rewards[0].RelatedMemberID)
	assert.Equal(t, refereeUser, rewards[0].RelatedMemberReferenceID)
	assert.Equal(t, referrer.ID, rewards[0].RewardedMemberID)
	assert.Equal(t, referrerUser, rewards[0].RewardedMemberReferenceID)
	assert.Equal(t, "pending", rewards[0].Status)
	assert.Equal(t, expectedReward.String(), rewards[0].Amount.String())
	assert.Equal(t, expectedReward2.String(), rewards[1].Amount.String())

	//elreq := request.GetEventLogRequest{
	//	Projects:          []string{project},
	//	MemberReferenceID: &refereeUser,
	//	PaginationConditions: request.PaginationConditions{
	//		SortBy: utils.StringPtr("id"),
	//		Order:  utils.StringPtr("asc"),
	//	},
	//}
	//
	//_, count, err = referralService.EventLogs.GetEventLogs(elreq)
	//assert.NoError(t, err)
	//assert.Equal(t, int64(3), count)

	elreg := request.GetCampaignEventLogRequest{
		Projects:           []string{project},
		MemberReferenceIDs: []string{refereeUser},
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}
	fmt.Print(refereeUser)
	campaignEventLogs, count, err := referralService.CampaignEventLog.GetCampaignEventLogs(elreg)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Equal(t, rewards[0].ID, *campaignEventLogs[0].ReferredRewardID)
	assert.Equal(t, rewards[1].ID, *campaignEventLogs[1].ReferredRewardID)

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
	var rewardType = "percentage"
	rewardValue := decimal.NewFromFloat(12.34)
	//rewardCap := decimal.NewFromFloat(15.00)
	//var inviteeRewardCap = decimal.NewFromFloat(1000.0)
	maxOccurrencesPerCustomer := int64(2)
	campaign := createCampaign(t, project, request.CreateCampaignRequest{
		Name:                      "New User Campaign",
		RewardType:                &rewardType,
		RewardValue:               &rewardValue,
		CurrencyCode:              "USDC",
		StartDate:                 &startDate,
		EndDate:                   &endDate,
		Description:               &description,
		Budget:                    &budget,
		IsDefault:                 true,
		CampaignTypePerCustomer:   "count_per_customer",
		MaxOccurrencesPerCustomer: &maxOccurrencesPerCustomer,

		EventKeys: []string{event1.Key},
	})

	referrer := createReferrer(t, project, referrerUser, []uint{campaign.ID}, nil)

	referee := createReferee(t, project, referrer.Code, refereeUser, nil)

	amount1 := decimal.NewFromFloat(250.50)
	amount2 := decimal.NewFromFloat(1510.74565)
	amount3 := decimal.NewFromFloat(1430.346)
	_, err := triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount1)
	_, err = triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount2)
	_, err = triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount3)

	err = referralService.Worker.ProcessPendingEvents()
	assert.NoError(t, err)

	req := request.GetRewardRequest{
		Projects:                  []string{project},
		RewardedMemberReferenceID: &referrerUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	rewards, count, err := referralService.Reward.GetRewards(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	expectedReward := decimal.NewFromFloat(30.9117)       //27.77745
	expectedReward2 := decimal.NewFromFloat(186.42601321) //27.77745

	assert.Equal(t, project, rewards[0].Project)
	assert.Equal(t, campaign.ID, rewards[0].CampaignID)
	assert.Equal(t, referee.ID, rewards[0].RelatedMemberID)
	assert.Equal(t, refereeUser, rewards[0].RelatedMemberReferenceID)
	assert.Equal(t, referrer.ID, rewards[0].RewardedMemberID)
	assert.Equal(t, referrerUser, rewards[0].RewardedMemberReferenceID)
	assert.Equal(t, "pending", rewards[0].Status)
	assert.Equal(t, expectedReward.String(), rewards[0].Amount.String())
	assert.Equal(t, expectedReward2.String(), rewards[1].Amount.String())

	elreg := request.GetCampaignEventLogRequest{
		Projects:           []string{project},
		MemberReferenceIDs: []string{refereeUser},
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}
	fmt.Print(refereeUser)
	campaignEventLogs, count, err := referralService.CampaignEventLog.GetCampaignEventLogs(elreg)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Equal(t, rewards[0].ID, *campaignEventLogs[0].ReferredRewardID)
	assert.Equal(t, rewards[1].ID, *campaignEventLogs[1].ReferredRewardID)
}

func TestEventsSharedByCampaigns(t *testing.T) {
	project := "eventssharedbycampaigns"
	referrerUser := "user-123"
	refereeUser := "user-456"
	event1 := createEvent(t, project, request.CreateEventRequest{
		Key:       "payment-recurring-event",
		Name:      "Payment Recurring Event",
		EventType: "payment",
	})

	// Campaign 1
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0) // One month from start date
	budget := decimal.NewFromFloat(3000.00)
	description := "Campaign for new user signups and payments"
	var rewardType = "percentage"
	rewardValue := decimal.NewFromFloat(12.34)
	var inviteeRewardType = "percentage"
	inviteeRewardValue := decimal.NewFromFloat(3.67)
	campaign1 := createCampaign(t, project, request.CreateCampaignRequest{
		Name:                    "New User Campaign",
		RewardType:              &rewardType,
		RewardValue:             &rewardValue,
		InviteeRewardType:       &inviteeRewardType,
		InviteeRewardValue:      &inviteeRewardValue,
		CurrencyCode:            "USDC",
		StartDate:               &startDate,
		EndDate:                 &endDate,
		Description:             &description,
		Budget:                  &budget,
		IsDefault:               true,
		CampaignTypePerCustomer: "forever",

		EventKeys: []string{event1.Key},
	})

	// Campaign 2
	startDate = time.Now()
	endDate = startDate.AddDate(0, 1, 0) // One month from start date
	budget = decimal.NewFromFloat(3000.00)
	description = "Campaign for new user signups and payments"
	rewardType = "percentage"
	rewardValue = decimal.NewFromFloat(9.76)
	inviteeRewardType = "percentage"
	inviteeRewardValue = decimal.NewFromFloat(2.89)
	campaign2 := createCampaign(t, project, request.CreateCampaignRequest{
		Name:                    "New User Campaign",
		RewardType:              &rewardType,
		RewardValue:             &rewardValue,
		InviteeRewardType:       &inviteeRewardType,
		InviteeRewardValue:      &inviteeRewardValue,
		CurrencyCode:            "USDC",
		StartDate:               &startDate,
		EndDate:                 &endDate,
		Description:             &description,
		Budget:                  &budget,
		IsDefault:               true,
		CampaignTypePerCustomer: "forever",

		EventKeys: []string{event1.Key},
	})

	referrer := createReferrer(t, project, referrerUser, []uint{}, nil)

	_ = createReferee(t, project, referrer.Code, refereeUser, nil)

	amount1 := decimal.NewFromFloat(221.5560)
	_, err := triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount1)

	err = referralService.Worker.ProcessPendingEvents()
	assert.NoError(t, err)

	req := request.GetRewardRequest{
		Projects: []string{project},
		//RewardedMemberReferenceID: &referrerUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	rewards, count, err := referralService.Reward.GetRewards(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), count)

	expectedReward := decimal.NewFromFloat(27.3400104)
	expectedReward2 := decimal.NewFromFloat(8.1311052)
	expectedReward3 := decimal.NewFromFloat(21.6238656)
	expectedReward4 := decimal.NewFromFloat(6.4029684)

	assert.Equal(t, project, rewards[0].Project)
	assert.Equal(t, project, rewards[1].Project)
	assert.Equal(t, project, rewards[2].Project)
	assert.Equal(t, project, rewards[3].Project)

	assert.Equal(t, campaign1.ID, rewards[0].CampaignID)
	assert.Equal(t, campaign1.ID, rewards[1].CampaignID)
	assert.Equal(t, campaign2.ID, rewards[2].CampaignID)
	assert.Equal(t, campaign2.ID, rewards[3].CampaignID)

	assert.Equal(t, referrerUser, rewards[0].RewardedMemberReferenceID)
	assert.Equal(t, refereeUser, rewards[0].RelatedMemberReferenceID)
	assert.Equal(t, refereeUser, rewards[1].RewardedMemberReferenceID)
	assert.Equal(t, referrerUser, rewards[1].RelatedMemberReferenceID)
	assert.Equal(t, referrerUser, rewards[2].RewardedMemberReferenceID)
	assert.Equal(t, refereeUser, rewards[2].RelatedMemberReferenceID)
	assert.Equal(t, refereeUser, rewards[3].RewardedMemberReferenceID)
	assert.Equal(t, referrerUser, rewards[3].RelatedMemberReferenceID)

	assert.Equal(t, expectedReward.String(), rewards[0].Amount.String())
	assert.Equal(t, expectedReward2.String(), rewards[1].Amount.String())
	assert.Equal(t, expectedReward3.String(), rewards[2].Amount.String())
	assert.Equal(t, expectedReward4.String(), rewards[3].Amount.String())

	elreg := request.GetCampaignEventLogRequest{
		Projects: []string{project},
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}
	fmt.Print(refereeUser)
	campaignEventLogs, count, err := referralService.CampaignEventLog.GetCampaignEventLogs(elreg)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Equal(t, rewards[0].ID, *campaignEventLogs[0].ReferredRewardID)
	assert.Equal(t, rewards[1].ID, *campaignEventLogs[0].RefereeRewardID)
	assert.Equal(t, rewards[2].ID, *campaignEventLogs[1].ReferredRewardID)
	assert.Equal(t, rewards[3].ID, *campaignEventLogs[1].RefereeRewardID)
}

func TestFutureCampaign(t *testing.T) {
	project := "futurecampaign"
	referrerUser := "user-123"
	refereeUser := "user-456"
	event1 := createEvent(t, project, request.CreateEventRequest{
		Key:       "payment-recurring-event",
		Name:      "Payment Recurring Event",
		EventType: "payment",
	})

	// Campaign 1
	startDate := time.Now().AddDate(0, 0, 1) // One day from now
	endDate := startDate.AddDate(0, 1, 0)    // One month from start date
	budget := decimal.NewFromFloat(3000.00)
	description := "Campaign for new user signups and payments"
	var rewardType = "percentage"
	rewardValue := decimal.NewFromFloat(12.34)
	var inviteeRewardType = "percentage"
	inviteeRewardValue := decimal.NewFromFloat(3.67)
	createCampaign(t, project, request.CreateCampaignRequest{
		Name:                    "New User Campaign",
		RewardType:              &rewardType,
		RewardValue:             &rewardValue,
		InviteeRewardType:       &inviteeRewardType,
		InviteeRewardValue:      &inviteeRewardValue,
		CurrencyCode:            "USDC",
		StartDate:               &startDate,
		EndDate:                 &endDate,
		Description:             &description,
		Budget:                  &budget,
		IsDefault:               true,
		CampaignTypePerCustomer: "forever",

		EventKeys: []string{event1.Key},
	})

	referrer := createReferrer(t, project, referrerUser, []uint{}, nil)

	_ = createReferee(t, project, referrer.Code, refereeUser, nil)

	amount1 := decimal.NewFromFloat(221.5560)
	_, err := triggerEvent(t, project, "payment-recurring-event", refereeUser, utils.StringPtr(`{"transactionId": "12345"}`), &amount1)

	err = referralService.Worker.ProcessPendingEvents()
	assert.NoError(t, err)

	req := request.GetRewardRequest{
		Projects: []string{project},
		//RewardedMemberReferenceID: &referrerUser,
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}

	rewards, count, err := referralService.Reward.GetRewards(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	assert.Equal(t, 0, len(rewards))

	elreg := request.GetCampaignEventLogRequest{
		Projects: []string{project},
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	}
	fmt.Print(refereeUser)
	campaignEventLogs, count, err := referralService.CampaignEventLog.GetCampaignEventLogs(elreg)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
	assert.Equal(t, 0, len(campaignEventLogs))
}

func TestAggregator(t *testing.T) {
	stats, count, err := referralService.AggregatorService.GetReferrerMembersStats(request.GetMemberRequest{
		PaginationConditions: request.PaginationConditions{
			SortBy: utils.StringPtr("id"),
			Order:  utils.StringPtr("asc"),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)

	assert.Equal(t, "onetimeproject", stats[0].Project)
	assert.Equal(t, int64(1), stats[0].RefereeCount)
	assert.Equal(t, "10.05", stats[0].TotalRewards.String())
	assert.Equal(t, false, stats[0].IsReferred)

	assert.Equal(t, "onetimeproject", stats[1].Project)
	assert.Equal(t, int64(0), stats[1].RefereeCount)
	assert.Equal(t, "5.025", stats[1].TotalRewards.String())
	assert.Equal(t, true, stats[1].IsReferred)

	assert.Equal(t, "recurringproject", stats[2].Project)
	assert.Equal(t, int64(1), stats[2].RefereeCount)
	assert.Equal(t, "27.77745", stats[2].TotalRewards.String())
	assert.Equal(t, false, stats[2].IsReferred)

	assert.Equal(t, "recurringproject", stats[3].Project)
	assert.Equal(t, int64(0), stats[3].RefereeCount)
	assert.Equal(t, "0", stats[3].TotalRewards.String())
	assert.Equal(t, true, stats[3].IsReferred)

	assert.Equal(t, "recumaxoccurrenceproject", stats[4].Project)
	assert.Equal(t, int64(1), stats[4].RefereeCount)
	assert.Equal(t, "217.33771321", stats[4].TotalRewards.String())
	assert.Equal(t, false, stats[4].IsReferred)

	assert.Equal(t, "recumaxoccurrenceproject", stats[5].Project)
	assert.Equal(t, int64(0), stats[5].RefereeCount)
	assert.Equal(t, "0", stats[5].TotalRewards.String())
	assert.Equal(t, true, stats[5].IsReferred)

	rewardsStats, err := referralService.AggregatorService.GetRewardsStats(request.GetRewardRequest{})
	if err != nil {
		log.Fatalf("failed to fetch rewards stats: %v", err)
	}
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rewardsStats))
}

func TestTotalRewardEarned(t *testing.T) {

	totalRewards, err := referralService.Reward.GetTotalRewards(request.GetRewardRequest{
		Projects: []string{"onetimeproject"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "15.075", totalRewards.String())

	totalRewards, err = referralService.Reward.GetTotalRewards(request.GetRewardRequest{
		Projects: []string{"recurringproject"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "27.77745", totalRewards.String())
}
