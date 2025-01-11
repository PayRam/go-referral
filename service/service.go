package service

import (
	"github.com/PayRam/go-referral/internal/db"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/shopspring/decimal"
	"time"
)

// EventService handles operations related to events
type EventService interface {
	CreateEvent(project, key, name string, description *string, eventType string) (*models.Event, error)
	UpdateEvent(project, key string, req request.UpdateEventRequest) (*models.Event, error)
	GetAll(project string) ([]models.Event, error)
	GetByKey(project, key string) (*models.Event, error)
	GetByKeys(project string, keys []string) ([]models.Event, error)
	SearchByName(project, name string) ([]models.Event, error)
}

// CampaignService handles operations related to campaigns
type CampaignService interface {
	CreateCampaign(project, name, description string, startDate, endDate time.Time, events []models.Event, rewardType *string, rewardValue *float64, maxOccurrences *uint, validityDays *uint, budget *decimal.Decimal) (*models.Campaign, error)
	GetCampaigns(project string, conditions []db.QueryCondition, offset, limit int, sort *string) ([]models.Campaign, error)
	UpdateCampaign(project string, id uint, req request.UpdateCampaignRequest) (*models.Campaign, error)
	UpdateCampaignEvents(project string, campaignID uint, events []models.Event) (*models.Campaign, error)
	SetDefaultCampaign(project string, campaignID uint) (*models.Campaign, error)
	GetAll(project string) ([]models.Campaign, error)
	SearchByName(project string, name string) ([]models.Campaign, error)
	PauseCampaign(project string, campaignID uint) (*models.Campaign, error)
	DeleteCampaign(project string, campaignID uint) (bool, error)
}

// ReferrerService handles operations related to referral codes
type ReferrerService interface {
	CreateReferrer(project, referenceID, code string, campaignIDs []uint) (*models.Referrer, error)
	GetReferrerByReference(project, referenceID string) (*models.Referrer, error)
	UpdateCampaigns(project, referenceID string, campaignIDs []uint) (*models.Referrer, error)
}

// RefereeService handles operations related to referral codes
type RefereeService interface {
	CreateReferee(project, code, referenceID string) (*models.Referee, error)
	GetRefereeByReference(project, referenceID string) (*models.Referee, error)
	GetRefereesByReferrer(project string, referrerID uint) ([]models.Referee, error)
}

type EventLogService interface {
	CreateEventLog(project, eventKey string, referenceID string, amount *decimal.Decimal, data *string) (*models.EventLog, error)
	GetEventLogs(project string, conditions []db.QueryCondition, offset, limit *int, sort *string) ([]models.EventLog, error)
}

type RewardCalculator interface {
	CalculateReward(eventLog models.EventLog, campaign models.Campaign, referee models.Referee, referrer models.Referrer) (*models.Reward, error)
}

type Worker interface {
	AddCustomRewardCalculator(eventKey string, calculator RewardCalculator) error
	RemoveCustomRewardCalculator(eventKey string) error
	ProcessPendingEvents() error
}
