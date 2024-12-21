package service

import (
	"github.com/PayRam/go-referral/internal/serviceimpl"
	"github.com/PayRam/go-referral/service/param"
	"gorm.io/gorm"
)

type ReferralService struct {
	param.EventService
	param.CampaignService
	param.ReferrerService
	param.RefereeService
}

// NewReferralService initializes the unified service
func NewReferralService(db *gorm.DB) *ReferralService {
	return &ReferralService{
		EventService:    serviceimpl.NewEventService(db),
		CampaignService: serviceimpl.NewCampaignService(db),
		ReferrerService: serviceimpl.NewReferrerService(db),
		RefereeService:  serviceimpl.NewRefereeService(db),
	}
}
