package go_referral

import (
	"github.com/PayRam/go-referral/internal/serviceimpl"
	"github.com/PayRam/go-referral/service"
	"gorm.io/gorm"
)

type ReferralService struct {
	Events    service.EventService
	Campaigns service.CampaignService
	Referrers service.ReferrerService
	Referees  service.RefereeService
	EventLogs service.EventLogService
	Worker    service.Worker
}

func NewReferralService(db *gorm.DB) *ReferralService {
	return &ReferralService{
		Events:    serviceimpl.NewEventService(db),
		Campaigns: serviceimpl.NewCampaignService(db),
		Referrers: serviceimpl.NewReferrerService(db),
		Referees:  serviceimpl.NewRefereeService(db),
		EventLogs: serviceimpl.NewEventLogService(db),
		Worker:    serviceimpl.NewWorkerService(db),
	}
}
