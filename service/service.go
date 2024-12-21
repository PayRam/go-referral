package service

import (
	"github.com/PayRam/go-referral/internal/serviceimpl"
	"github.com/PayRam/go-referral/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

// EventService handles operations related to events
type EventService interface {
	CreateEvent(key, name, eventType, rewardType string, rewardValue float64, maxOccurrences, validityDays uint) (*models.Event, error)
	UpdateEvent(key string, updates map[string]interface{}) (*models.Event, error)
}

// CampaignService handles operations related to campaigns
type CampaignService interface {
	CreateCampaign(name, description string, startDate, endDate time.Time, isActive bool, events []models.Event) (*models.Campaign, error)
	UpdateCampaign(id uint, updates map[string]interface{}) (*models.Campaign, error)
	UpdateCampaignEvents(campaignID uint, events []models.Event) error
	SetDefaultCampaign(campaignID uint) error
}

// ReferrerService handles operations related to referral codes
type ReferrerService interface {
	CreateReferrer(referenceID, referenceType, code string, campaignID *uint) (*models.Referrer, error)
	GetReferrerByReference(referenceID, referenceType string) (*models.Referrer, error)
	AssignCampaign(referenceID, referenceType string, campaignID uint) error
	RemoveCampaign(referenceID, referenceType string) error
}

// RefereeService handles operations related to referral codes
type RefereeService interface {
	CreateRefereeByCode(code, referenceID, referenceType string) (*models.Referee, error)
	GetRefereeByReference(referenceID, referenceType string) (*models.Referee, error)
	GetRefereesByReferrer(referrerID uint) ([]models.Referee, error)
}

type EventLogService interface {
	CreateEventLog(eventKey string, referenceID, referenceType *string, amount *decimal.Decimal, data *string) (*models.EventLog, error)
	GetEventLogs(conditions map[string]interface{}, offset, limit *int, sort *string) ([]models.EventLog, error)
}

type RewardCalculator interface {
	CalculateReward(eventLog models.EventLog, event models.Event, campaign models.Campaign, referee models.Referee, referrer models.Referrer) (*models.Reward, error)
}

type Worker interface {
	AddCustomRewardCalculator(eventKey string, calculator RewardCalculator) error
	RemoveCustomRewardCalculator(eventKey string) error
	ProcessPendingEvents() error
}

type ReferralService struct {
	EventService
	CampaignService
	ReferrerService
	RefereeService
	EventLogService
	Worker
}

// NewReferralService initializes the unified service
func NewReferralService(db *gorm.DB) *ReferralService {
	return &ReferralService{
		EventService:    serviceimpl.NewEventService(db),
		CampaignService: serviceimpl.NewCampaignService(db),
		ReferrerService: serviceimpl.NewReferrerService(db),
		RefereeService:  serviceimpl.NewRefereeService(db),
		EventLogService: serviceimpl.NewEventLogService(db),
		Worker:          serviceimpl.NewWorkerService(db),
	}
}
