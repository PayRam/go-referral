package serviceimpl

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/PayRam/go-referral/service/param"
	"gorm.io/gorm"
)

type referralCodeService struct {
	DB *gorm.DB
}

func NewReferralCodeService(db *gorm.DB) param.ReferralCodeService {
	return &referralCodeService{DB: db}
}

func GenerateReferralCode() string {
	b := make([]byte, 8) // 8 bytes = 16 characters
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *referralCodeService) CreateReferralCode(referenceID, referenceType string) (*param.ReferralCode, error) {
	code := GenerateReferralCode()
	referralCode := &param.ReferralCode{
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		Code:          code,
	}
	if err := s.DB.Create(referralCode).Error; err != nil {
		return nil, err
	}
	return referralCode, nil
}
