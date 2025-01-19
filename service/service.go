package service

import (
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/PayRam/go-referral/response"
	"github.com/shopspring/decimal"
)

// EventService handles operations related to events
type EventService interface {
	CreateEvent(project string, request request.CreateEventRequest) (*models.Event, error)
	GetEvents(req request.GetEventsRequest) ([]models.Event, int64, error)
	UpdateEvent(project, key string, req request.UpdateEventRequest) (*models.Event, error)
}

// CampaignService handles operations related to campaigns
type CampaignService interface {
	CreateCampaign(project string, req request.CreateCampaignRequest) (*models.Campaign, error)
	GetCampaigns(req request.GetCampaignsRequest) ([]models.Campaign, int64, error)
	GetTotalCampaigns(req request.GetCampaignsRequest) (int64, error)
	UpdateCampaign(project string, id uint, req request.UpdateCampaignRequest) (*models.Campaign, error)
	SetDefaultCampaign(project string, campaignID uint) (*models.Campaign, error)
	RemoveDefaultCampaign(project string, campaignID uint) (*models.Campaign, error)
	UpdateCampaignStatus(project string, campaignID uint, newStatus string) (*models.Campaign, error)
}

// ReferrerService handles operations related to referral codes
type ReferrerService interface {
	CreateReferrer(project string, req request.CreateReferrerRequest) (*models.Referrer, error)
	GetReferrers(req request.GetReferrerRequest) ([]models.Referrer, int64, error)
	GetTotalReferrers(req request.GetReferrerRequest) (int64, error)
	UpdateReferrer(project, referenceID string, request request.UpdateReferrerRequest) (*models.Referrer, error)
}

// RefereeService handles operations related to referral codes
type RefereeService interface {
	CreateReferee(project string, req request.CreateRefereeRequest) (*models.Referee, error)
	GetReferees(req request.GetRefereeRequest) ([]models.Referee, int64, error)
	GetTotalReferees(req request.GetRefereeRequest) (int64, error)
}

type EventLogService interface {
	CreateEventLog(project string, req request.CreateEventLogRequest) (*models.EventLog, error)
	GetEventLogs(req request.GetEventLogRequest) ([]models.EventLog, int64, error)
}

type RewardService interface {
	GetTotalRewards(request request.GetRewardRequest) (decimal.Decimal, error)
	GetRewards(req request.GetRewardRequest) ([]models.Reward, int64, error)
}

type AggregatorService interface {
	GetReferrersWithStats(req request.GetReferrerRequest) ([]response.ReferrerStats, int64, error)
}

type RewardCalculator interface {
	CalculateReward(eventLog models.EventLog, campaign models.Campaign, referee models.Referee, referrer models.Referrer) (*models.Reward, error)
}

type Worker interface {
	AddCustomRewardCalculator(eventKey string, calculator RewardCalculator) error
	RemoveCustomRewardCalculator(eventKey string) error
	ProcessPendingEvents() error
}
