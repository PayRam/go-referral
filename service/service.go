package service

import (
	db2 "github.com/PayRam/go-referral/internal/db"
	"github.com/PayRam/go-referral/internal/serviceimpl"
	"github.com/PayRam/go-referral/service/param"
	"gorm.io/gorm"
)

func NewReferralServiceWithDB(db *gorm.DB) param.ReferralService {
	return serviceimpl.NewReferralServiceWithDB(db2.Migrate(db))
}

func NewReferralService(dbPath string) param.ReferralService {
	return serviceimpl.NewReferralServiceWithDB(db2.InitDB(dbPath))
}
