package serviceimpl

import (
	"github.com/PayRam/go-referral/service/param"
	"gorm.io/gorm"
)

type userCampaignService struct {
	DB *gorm.DB
}

func NewUserCampaignService(db *gorm.DB) param.UserCampaignService {
	return &userCampaignService{DB: db}
}

func (s *userCampaignService) AssignUserToCampaign(userID, campaignID uint) error {
	mapping := &param.UserCampaignMapping{
		UserID:     userID,
		CampaignID: campaignID,
	}
	if err := s.DB.Create(mapping).Error; err != nil {
		return err
	}
	return nil
}
