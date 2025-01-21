package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/PayRam/go-referral/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/mail"
)

type referrerService struct {
	DB *gorm.DB
}

//var _ service.ReferrerService = &referrerService{}

func NewReferrerService(db *gorm.DB) *referrerService {
	return &referrerService{DB: db}
}

func (s *referrerService) CreateReferrer(project string, req request.CreateReferrerRequest) (*models.Referrer, error) {
	// Validate email if provided
	if req.Email != nil {
		if *req.Email == "" {
			return nil, fmt.Errorf("email cannot be empty")
		}
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			return nil, fmt.Errorf("invalid email format: %w", err)
		}
	}

	// Create the referrer
	if req.Code == nil || *req.Code == "" {
		code, err := utils.CreateReferralCode(7)
		if err != nil {
			return nil, fmt.Errorf("CreateReferrer: failed to generate referral code: %w", err)
		}
		req.Code = &code
	}

	referrer := &models.Referrer{
		Project:     project,
		Code:        *req.Code,
		ReferenceID: req.ReferenceID,
		Email:       req.Email,
	}

	// Use a transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Save the referrer
		if err := tx.Create(referrer).Error; err != nil {
			return err
		}

		// Associate campaigns if provided
		if len(req.CampaignIDs) > 0 {
			for _, campaignID := range req.CampaignIDs {
				association := &models.ReferrerCampaign{
					Project:    project,
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
func (s *referrerService) GetReferrers(req request.GetReferrerRequest) ([]models.Referrer, int64, error) {
	var referrers []models.Referrer
	var count int64

	// Start query
	query := s.DB.Model(&models.Referrer{})

	// Apply filters
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("reference_id = ?", *req.ReferenceID)
	}
	if req.Email != nil {
		query = query.Where("email = ?", *req.Email)
	}
	if req.Code != nil {
		query = query.Where("code = ?", *req.Code)
	}
	if req.CampaignIDs != nil && len(req.CampaignIDs) > 0 {
		// Join with referral_referrer_campaigns table to filter by CampaignIDs
		query = query.Joins("JOIN referral_referrer_campaigns rc ON rc.referrer_id = referral_referrer.id").
			Where("rc.campaign_id IN (?)", req.CampaignIDs).
			Group("referral_referrer.id") // Avoid duplicates due to the JOIN
	}

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count referrers: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Preload("Campaigns").Find(&referrers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch referrers: %w", err)
	}

	return referrers, count, nil
}

func (s *referrerService) UpdateReferrer(project, referenceID string, req request.UpdateReferrerRequest) (*models.Referrer, error) {
	var updatedReferrer *models.Referrer

	// Use a database transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var referrer models.Referrer

		// Fetch the referrer for the given reference with a row-level lock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("project = ? AND reference_id = ?", project, referenceID).
			First(&referrer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("referrer not found for project=%s and reference_id=%s", project, referenceID)
			}
			return err
		}

		// Validate email if provided
		if req.Email != nil {
			if *req.Email == "" {
				return fmt.Errorf("email cannot be empty")
			}
			if _, err := mail.ParseAddress(*req.Email); err != nil {
				return fmt.Errorf("invalid email format: %w", err)
			}
			referrer.Email = req.Email // Update email
		}

		// Remove existing campaign associations
		if err := tx.Unscoped().Where("referrer_id = ?", referrer.ID).Delete(&models.ReferrerCampaign{}).Error; err != nil {
			return fmt.Errorf("failed to remove existing campaign associations: %w", err)
		}

		// Add new campaign associations
		for _, campaignID := range req.CampaignIDs {
			association := &models.ReferrerCampaign{
				Project:    project,
				ReferrerID: referrer.ID,
				CampaignID: campaignID,
			}
			if err := tx.Create(association).Error; err != nil {
				return fmt.Errorf("failed to associate campaign %d: %w", campaignID, err)
			}
		}

		// Save the updated referrer details
		if err := tx.Save(&referrer).Error; err != nil {
			return fmt.Errorf("failed to save referrer updates: %w", err)
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

func (s *referrerService) GetTotalReferrers(req request.GetReferrerRequest) (int64, error) {
	var count int64

	// Build the query
	query := s.DB.Model(&models.Referrer{})
	// Apply filters
	if req.Projects != nil && len(req.Projects) > 0 {
		query = query.Where("project IN (?)", req.Projects)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("reference_id = ?", *req.ReferenceID)
	}
	if req.Code != nil {
		query = query.Where("code = ?", *req.Code)
	}

	// Count the records
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count referrers: %w", err)
	}

	return count, nil
}
