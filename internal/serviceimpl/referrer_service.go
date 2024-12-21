package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/service"
	"gorm.io/gorm"
)

type referrerService struct {
	DB *gorm.DB
}

var _ service.ReferrerService = &referrerService{}

func NewReferrerService(db *gorm.DB) service.ReferrerService {
	return &referrerService{DB: db}
}

//func GenerateReferralCode() string {
//	b := make([]byte, 8) // 8 bytes = 16 characters
//	_, _ = rand.Read(b)
//	return hex.EncodeToString(b)
//}

func (s *referrerService) CreateReferrer(referenceID, referenceType, code string, campaignID *uint) (*models.Referrer, error) {
	referral := &models.Referrer{
		Code:          code,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		CampaignID:    campaignID, // Can be nil for default campaign
	}
	if err := s.DB.Create(referral).Error; err != nil {
		return nil, err
	}
	return referral, nil
}

func (s *referrerService) GetReferrerByReference(referenceID, referenceType string) (*models.Referrer, error) {
	var referral models.Referrer
	if err := s.DB.Where("reference_id = ? AND reference_type = ?", referenceID, referenceType).First(&referral).Error; err != nil {
		return nil, err
	}
	return &referral, nil
}

func (s *referrerService) AssignCampaign(referenceID, referenceType string, campaignID uint) error {
	// Use a database transaction to ensure atomicity
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var referral models.Referrer

		// Fetch the referral row for the given reference
		if err := tx.Where("reference_id = ? AND reference_type = ?", referenceID, referenceType).
			First(&referral).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("referral not found for reference_id=%s and reference_type=%s", referenceID, referenceType)
			}
			return err
		}

		// Assign the campaign
		referral.CampaignID = &campaignID

		// Save the updated referral
		if err := tx.Save(&referral).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *referrerService) RemoveCampaign(referenceID, referenceType string) error {
	// Use a database transaction to ensure atomicity
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var referral models.Referrer

		// Fetch the referral row for the given reference
		if err := tx.Where("reference_id = ? AND reference_type = ?", referenceID, referenceType).
			First(&referral).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("referral not found for reference_id=%s and reference_type=%s", referenceID, referenceType)
			}
			return err
		}

		// Remove the campaign
		referral.CampaignID = nil

		// Save the updated referral
		if err := tx.Save(&referral).Error; err != nil {
			return err
		}

		return nil
	})
}
