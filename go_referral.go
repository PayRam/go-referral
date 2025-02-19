package go_referral

import (
	db2 "github.com/PayRam/go-referral/internal/db"
	"github.com/PayRam/go-referral/internal/serviceimpl"
	"github.com/PayRam/go-referral/service"
	"gorm.io/gorm"
)

type ReferralService struct {
	Events            service.EventService
	Campaigns         service.CampaignService
	Members           service.MemberService
	EventLogs         service.EventLogService
	CampaignEventLog  service.CampaignEventLogService
	Reward            service.RewardService
	AggregatorService service.AggregatorService
	Worker            service.Worker
}

func NewReferralService(db *gorm.DB) *ReferralService {
	db2.Migrate(db)
	return &ReferralService{
		Events:            serviceimpl.NewEventService(db),
		Campaigns:         serviceimpl.NewCampaignService(db),
		Members:           serviceimpl.NewReferrerService(db),
		EventLogs:         serviceimpl.NewEventLogService(db),
		CampaignEventLog:  serviceimpl.NewCampaignEventLogService(db),
		Reward:            serviceimpl.NewRewardService(db),
		AggregatorService: serviceimpl.NewAggregatorService(db),
		Worker:            serviceimpl.NewWorkerService(db),
	}
}
