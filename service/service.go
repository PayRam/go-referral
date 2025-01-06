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
	CreateEvent(key, name, eventType string) (*models.Event, error)
	UpdateEvent(key string, updates map[string]interface{}) (*models.Event, error)
	GetAll() ([]models.Event, error)
}

// CampaignService handles operations related to campaigns
type CampaignService interface {
	CreateCampaign(name, description string, startDate, endDate time.Time, events []models.Event, rewardType *string, rewardValue *float64, maxOccurrences *uint, validityDays *uint, budget *decimal.Decimal) (*models.Campaign, error)
	GetCampaigns(conditions []db.QueryCondition, offset, limit int, sort *string) ([]models.Campaign, error)
	UpdateCampaign(id uint, req request.UpdateCampaignRequest) (*models.Campaign, error)
	UpdateCampaignEvents(campaignID uint, events []models.Event) error
	SetDefaultCampaign(campaignID uint) error
}

// ReferrerService handles operations related to referral codes
type ReferrerService interface {
	CreateReferrer(referenceID, referenceType, code string, campaignIDs []uint) (*models.Referrer, error)
	GetReferrerByReference(referenceID, referenceType string) (*models.Referrer, error)
	UpdateCampaigns(referenceID, referenceType string, campaignIDs []uint) error
}

// RefereeService handles operations related to referral codes
type RefereeService interface {
	CreateRefereeByCode(code, referenceID, referenceType string) (*models.Referee, error)
	GetRefereeByReference(referenceID, referenceType string) (*models.Referee, error)
	GetRefereesByReferrer(referrerID uint) ([]models.Referee, error)
}

type EventLogService interface {
	CreateEventLog(eventKey string, referenceID, referenceType *string, amount *decimal.Decimal, data *string) (*models.EventLog, error)
	GetEventLogs(conditions []db.QueryCondition, offset, limit *int, sort *string) ([]models.EventLog, error)
}

type RewardCalculator interface {
	CalculateReward(eventLog models.EventLog, campaign models.Campaign, referee models.Referee, referrer models.Referrer) (*models.Reward, error)
}

type Worker interface {
	AddCustomRewardCalculator(eventKey string, calculator RewardCalculator) error
	RemoveCustomRewardCalculator(eventKey string) error
	ProcessPendingEvents() error
}
