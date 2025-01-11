package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"gorm.io/gorm"
)

type referrerService struct {
	DB *gorm.DB
}

//var _ service.ReferrerService = &referrerService{}

func NewReferrerService(db *gorm.DB) *referrerService {
	return &referrerService{DB: db}
}

func (s *referrerService) CreateReferrer(project, referenceID, code string, campaignIDs []uint) (*models.Referrer, error) {
	// Create the referrer
	referrer := &models.Referrer{
		Project:     project,
		Code:        code,
		ReferenceID: referenceID,
	}

	// Use a transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Save the referrer
		if err := tx.Create(referrer).Error; err != nil {
			return err
		}

		// Associate campaigns if provided
		if len(campaignIDs) > 0 {
			for _, campaignID := range campaignIDs {
				association := &models.ReferrerCampaign{
					ReferrerID: referrer.ID,
					CampaignID: campaignID,
				}
				if err := tx.Create(association).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload the referrer with preloaded campaigns
	if err := s.DB.Preload("Campaigns").First(referrer, referrer.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to preload campaigns for referrer: %w", err)
	}

	return referrer, nil
}

func (s *referrerService) GetReferrerByReference(project, referenceID string) (*models.Referrer, error) {
	var referral models.Referrer
	if err := s.DB.Preload("Campaigns").Where("project = ? AND reference_id = ?", project, referenceID).First(&referral).Error; err != nil {
		return nil, err
	}
	return &referral, nil
}

func (s *referrerService) UpdateCampaigns(project, referenceID string, campaignIDs []uint) (*models.Referrer, error) {
	// Use a database transaction to ensure atomicity
	var updatedReferrer *models.Referrer
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var referrer models.Referrer

		// Fetch the referrer for the given reference
		if err := tx.Where("project = ? AND reference_id = ?", project, referenceID).
			First(&referrer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("referrer not found for project=%s and reference_id=%s", project, referenceID)
			}
			return err
		}

		// Remove existing campaign associations
		if err := tx.Unscoped().Where("referrer_id = ?", referrer.ID).Delete(&models.ReferrerCampaign{}).Error; err != nil {
			return fmt.Errorf("failed to remove existing campaign associations: %w", err)
		}

		// Add new campaign associations
		for _, campaignID := range campaignIDs {
			association := &models.ReferrerCampaign{
				ReferrerID: referrer.ID,
				CampaignID: campaignID,
			}
			if err := tx.Create(association).Error; err != nil {
				return fmt.Errorf("failed to associate campaign %d: %w", campaignID, err)
			}
		}

		// Preload campaigns for the updated referrer
		if err := tx.Preload("Campaigns").First(&referrer, referrer.ID).Error; err != nil {
			return fmt.Errorf("failed to preload campaigns for referrer: %w", err)
		}

		updatedReferrer = &referrer
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedReferrer, nil
}
