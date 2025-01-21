package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"gorm.io/gorm"
	"net/mail"
)

type refereeService struct {
	DB *gorm.DB
}

//var _ service.RefereeService = &refereeService{}

// NewRefereeService creates a new instance of the Referee service
func NewRefereeService(db *gorm.DB) *refereeService {
	return &refereeService{DB: db}
}

// CreateReferee creates a mapping between a referee and a referrer
func (s *refereeService) CreateReferee(project string, req request.CreateRefereeRequest) (*models.Referee, error) {
	// Validate email if provided
	if req.Email != nil {
		if *req.Email == "" {
			return nil, fmt.Errorf("email cannot be empty")
		}
		if _, err := mail.ParseAddress(*req.Email); err != nil {
			return nil, fmt.Errorf("invalid email format: %w", err)
		}
	}

	// Validate if the Referrer exists by referral code
	var referrer models.Referrer
	if err := s.DB.Where("code = ?", req.Code).First(&referrer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("referrer not found with code %s", req.Code)
		}
		return nil, err
	}

	// Create the Referee mapping
	referee := &models.Referee{
		Project:             project,
		ReferrerID:          referrer.ID,
		ReferrerReferenceID: referrer.ReferenceID,
		ReferenceID:         req.ReferenceID,
		Email:               req.Email,
	}
	if err := s.DB.Create(referee).Error; err != nil {
		return nil, err
	}
	return referee, nil
}

func (s *refereeService) GetReferees(req request.GetRefereeRequest) ([]models.Referee, int64, error) {
	var referees []models.Referee
	var count int64

	// Start query
	query := s.DB.Model(&models.Referee{})

	// Apply filters
	if req.Project != nil {
		query = query.Where("project = ?", *req.Project)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("reference_id = ?", *req.ReferenceID)
	}
	if req.ReferrerID != nil {
		query = query.Where("referrer_id = ?", *req.ReferrerID)
	}
	if req.ReferrerReferenceID != nil {
		query = query.Where("referrer_reference_id = ?", *req.ReferrerReferenceID)
	}

	// Calculate total count before applying pagination
	countQuery := query
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count referees: %w", err)
	}

	// Apply pagination conditions
	query = request.ApplyPaginationConditions(query, req.PaginationConditions)

	// Fetch records with pagination
	if err := query.Find(&referees).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch referees: %w", err)
	}

	return referees, count, nil
}

func (s *refereeService) GetTotalReferees(req request.GetRefereeRequest) (int64, error) {
	var count int64

	// Build the query
	query := s.DB.Model(&models.Referee{})
	if req.Project != nil {
		query = query.Where("project = ?", *req.Project)
	}
	if req.ID != nil {
		query = query.Where("id = ?", *req.ID)
	}
	if req.ReferenceID != nil {
		query = query.Where("reference_id = ?", *req.ReferenceID)
	}
	if req.ReferrerID != nil {
		query = query.Where("referrer_id = ?", *req.ReferrerID)
	}
	if req.ReferrerReferenceID != nil {
		query = query.Where("referrer_reference_id = ?", *req.ReferrerReferenceID)
	}

	// Count the records
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count referees: %w", err)
	}

	return count, nil
}
