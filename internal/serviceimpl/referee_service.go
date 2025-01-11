package serviceimpl

import (
	"errors"
	"fmt"
	"github.com/PayRam/go-referral/models"
	"gorm.io/gorm"
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
func (s *refereeService) CreateReferee(project, code, referenceID string) (*models.Referee, error) {
	// Validate if the Referrer exists by referral code
	var referrer models.Referrer
	if err := s.DB.Where("code = ?", code).First(&referrer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("referrer not found with code %s", code)
		}
		return nil, err
	}

	// Create the Referee mapping
	referee := &models.Referee{
		Project:     project,
		ReferrerID:  referrer.ID,
		ReferenceID: referenceID,
	}
	if err := s.DB.Create(referee).Error; err != nil {
		return nil, err
	}
	return referee, nil
}

// GetRefereeByReference fetches a referee by reference ID and reference type
func (s *refereeService) GetRefereeByReference(project, referenceID string) (*models.Referee, error) {
	var referee models.Referee
	if err := s.DB.Where("project = ? AND reference_id = ?", project, referenceID).First(&referee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("referee not found for project=%s and reference_id=%s", project, referenceID)
		}
		return nil, err
	}
	return &referee, nil
}

// GetRefereesByReferrer fetches all referees associated with a specific referrer
func (s *refereeService) GetRefereesByReferrer(project string, referrerID uint) ([]models.Referee, error) {
	var referees []models.Referee
	if err := s.DB.Where("project = ? AND referrer_id = ?", project, referrerID).Find(&referees).Error; err != nil {
		return nil, err
	}
	return referees, nil
}
