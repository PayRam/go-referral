package serviceimpl

import (
	"errors"
	"fmt"
	"net/mail"

	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/PayRam/go-referral/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type referrerService struct {
	DB *gorm.DB
}

//var _ service.ReferrerService = &referrerService{}

func NewReferrerService(db *gorm.DB) *referrerService {
	return &referrerService{DB: db}
}

func (s *referrerService) CreateCustomer(project string, req request.CreateCustomerRequest) (*models.Customer, error) {
	// Validate email if provided
	if req.Email != nil {
		if *req.Email == "" {
			return nil, fmt.Errorf("email cannot be empty")
		}
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			return nil, fmt.Errorf("invalid email format: %w", err)
		}
	}

	// Initialize `ReferredByCustomerID`
	var referredByCustomerID *uint
	var referredByCustomerReferenceID *string

	// 🔹 Step 1: Fetch the existing customer by `ReferrerCode`
	if req.ReferrerCode != nil && *req.ReferrerCode != "" {
		var referrerCustomer models.Customer
		if err := s.DB.Where("project = ? AND code = ?", project, *req.ReferrerCode).
			First(&referrerCustomer).Error; err != nil {
			return nil, fmt.Errorf("invalid referrer code: %w", err)
		}
		referredByCustomerID = &referrerCustomer.ID
		referredByCustomerReferenceID = &referrerCustomer.ReferenceID
	}

	// 🔹 Step 2: Generate a PreferredCode if not provided
	if req.PreferredCode == nil || *req.PreferredCode == "" {
		code, err := utils.CreateReferralCode(7)
		if err != nil {
			return nil, fmt.Errorf("CreateCustomer: failed to generate referral code: %w", err)
		}
		req.PreferredCode = &code
	}

	// 🔹 Step 3: Create the new customer with `ReferredByCustomerID`
	customer := &models.Customer{
		Project:                       project,
		Code:                          *req.PreferredCode,
		ReferenceID:                   req.ReferenceID,
		Email:                         req.Email,
		ReferredByCustomerID:          referredByCustomerID, // Assign the referrer
		ReferredByCustomerReferenceID: referredByCustomerReferenceID,
	}

	// 🔹 Step 4: Use a transaction to save the customer and associate campaigns
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Save the new customer
		if err := tx.Create(customer).Error; err != nil {
			return err
		}

		// Associate campaigns if provided
		if len(req.CampaignIDs) > 0 {
			for _, campaignID := range req.CampaignIDs {
				association := &models.CustomerCampaign{
					Project:    project,
					CustomerID: customer.ID,
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

	// 🔹 Step 5: Reload the customer with preloaded campaigns and referrer
	if err := s.DB.Preload("Campaigns").Preload("ReferredByCustomer").First(customer, customer.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to preload customer data: %w", err)
	}

	return customer, nil
}

func (s *referrerService) GetCustomers(req request.GetCustomerRequest) ([]models.Customer, int64, error) {
	var referrers []models.Customer
	var count int64

	// Start query
	query := s.DB.Model(&models.Customer{})

	query = request.ApplyGetCustomerRequest(req, query)

	// Apply Select Fields
	query = request.ApplySelectFields(query, req.PaginationConditions.SelectFields)

	// Apply Group By
	query = request.ApplyGroupBy(query, req.PaginationConditions.GroupBy)

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count referrers: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Preload("Campaigns").Preload("ReferredByCustomer").Find(&referrers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch referrers: %w", err)
	}

	return referrers, count, nil
}

func (s *referrerService) UpdateCustomer(project, referenceID string, req request.UpdateCustomerRequest) (*models.Customer, error) {
	var updatedReferrer *models.Customer

	// Use a database transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var referrer models.Customer

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
		if err := tx.Unscoped().Where("customer_id = ?", referrer.ID).Delete(&models.CustomerCampaign{}).Error; err != nil {
			return fmt.Errorf("failed to remove existing campaign associations: %w", err)
		}

		// Add new campaign associations
		for _, campaignID := range req.CampaignIDs {
			association := &models.CustomerCampaign{
				Project:    project,
				CustomerID: referrer.ID,
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
		if err := tx.Preload("Campaigns").Preload("ReferredByCustomer").First(&referrer, referrer.ID).Error; err != nil {
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

func (s *referrerService) UpdateCustomerStatus(project, referenceID string, newStatus string) (*models.Customer, error) {
	var referrer models.Customer

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
		if err := tx.Preload("Campaigns").Preload("ReferredByCustomer").First(&referrer, referrer.ID).Error; err != nil {
			return fmt.Errorf("failed to preload campaigns for referrer: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &referrer, nil
}

func (s *referrerService) GetTotalCustomers(req request.GetCustomerRequest) (int64, error) {
	var count int64

	// Build the query
	query := s.DB.Model(&models.Customer{})

	query = request.ApplyGetCustomerRequest(req, query)

	// Apply Select Fields
	query = request.ApplySelectFields(query, req.PaginationConditions.SelectFields)

	// Apply Group By
	query = request.ApplyGroupBy(query, req.PaginationConditions.GroupBy)

	// Count the records
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count referrers: %w", err)
	}

	return count, nil
}
