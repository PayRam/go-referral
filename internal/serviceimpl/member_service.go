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

func (s *referrerService) CreateMember(project string, req request.CreateMemberRequest) (*models.Member, error) {
	// Validate email if provided
	if req.Email != nil {
		if *req.Email == "" {
			return nil, fmt.Errorf("email cannot be empty")
		}
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			return nil, fmt.Errorf("invalid email format: %w", err)
		}
	}

	// Initialize `ReferredByMemberID`
	var referredByMemberID *uint

	// 🔹 Step 1: Fetch the existing member by `ReferrerCode`
	if req.ReferrerCode != nil && *req.ReferrerCode != "" {
		var referrerMember models.Member
		if err := s.DB.Where("project = ? AND code = ?", project, *req.ReferrerCode).
			First(&referrerMember).Error; err != nil {
			return nil, fmt.Errorf("invalid referrer code: %w", err)
		}
		referredByMemberID = &referrerMember.ID
	}

	// 🔹 Step 2: Generate a PreferredCode if not provided
	if req.PreferredCode == nil || *req.PreferredCode == "" {
		code, err := utils.CreateReferralCode(7)
		if err != nil {
			return nil, fmt.Errorf("CreateMember: failed to generate referral code: %w", err)
		}
		req.PreferredCode = &code
	}

	// 🔹 Step 3: Create the new member with `ReferredByMemberID`
	member := &models.Member{
		Project:            project,
		Code:               *req.PreferredCode,
		ReferenceID:        req.ReferenceID,
		Email:              req.Email,
		ReferredByMemberID: referredByMemberID, // Assign the referrer
	}

	// 🔹 Step 4: Use a transaction to save the member and associate campaigns
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Save the new member
		if err := tx.Create(member).Error; err != nil {
			return err
		}

		// Associate campaigns if provided
		if len(req.CampaignIDs) > 0 {
			for _, campaignID := range req.CampaignIDs {
				association := &models.MemberCampaign{
					Project:    project,
					MemberID:   member.ID,
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

	// 🔹 Step 5: Reload the member with preloaded campaigns and referrer
	if err := s.DB.Preload("Campaigns").Preload("ReferredByMember").First(member, member.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to preload member data: %w", err)
	}

	return member, nil
}

func (s *referrerService) GetMembers(req request.GetMemberRequest) ([]models.Member, int64, error) {
	var referrers []models.Member
	var count int64

	// Start query
	query := s.DB.Model(&models.Member{})

	query = request.ApplyGetMemberRequest(req, query)

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count referrers: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Preload("Campaigns").Preload("ReferredByMember").Find(&referrers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch referrers: %w", err)
	}

	return referrers, count, nil
}

func (s *referrerService) UpdateMember(project, referenceID string, req request.UpdateMemberRequest) (*models.Member, error) {
	var updatedReferrer *models.Member

	// Use a database transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var referrer models.Member

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
		if err := tx.Unscoped().Where("member_id = ?", referrer.ID).Delete(&models.MemberCampaign{}).Error; err != nil {
			return fmt.Errorf("failed to remove existing campaign associations: %w", err)
		}

		// Add new campaign associations
		for _, campaignID := range req.CampaignIDs {
			association := &models.MemberCampaign{
				Project:    project,
				MemberID:   referrer.ID,
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
		if err := tx.Preload("Campaigns").Preload("ReferredByMember").First(&referrer, referrer.ID).Error; err != nil {
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

func (s *referrerService) UpdateMemberStatus(project, referenceID string, newStatus string) (*models.Member, error) {
	var referrer models.Member

	// Validate newStatus
	if newStatus != "active" && newStatus != "inactive" {
		return nil, fmt.Errorf("invalid new status: must be 'active' or 'inactive'")
	}

	// Use transaction to lock the row
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch the referrer with a row lock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("project = ? AND reference_id = ?", project, referenceID).First(&referrer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("referrer not found")
			}
			return fmt.Errorf("failed to fetch referrer: %w", err)
		}

		// Check if the status is already the desired status
		if referrer.Status == newStatus {
			return fmt.Errorf("referrer is already %s", newStatus)
		}

		// Update status
		referrer.Status = newStatus

		// Save the updated referrer
		if err := tx.Save(&referrer).Error; err != nil {
			return fmt.Errorf("failed to update referrer status: %w", err)
		}

		// Fetch the updated referrer with associated campaigns
		if err := tx.Preload("Campaigns").Preload("ReferredByMember").First(&referrer, referrer.ID).Error; err != nil {
			return fmt.Errorf("failed to preload campaigns for referrer: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &referrer, nil
}

func (s *referrerService) GetTotalMembers(req request.GetMemberRequest) (int64, error) {
	var count int64

	// Build the query
	query := s.DB.Model(&models.Member{})

	query = request.ApplyGetMemberRequest(req, query)

	// Count the records
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count referrers: %w", err)
	}

	return count, nil
}
