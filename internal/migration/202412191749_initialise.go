package migration

import (
	"github.com/PayRam/go-referral/service/param"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Initialise = &gormigrate.Migration{
	ID: "202412191749-gr-473842",
	Migrate: func(db *gorm.DB) error {
		return db.AutoMigrate(&param.Event{}, &param.Campaign{}, &param.CampaignEvent{}, &param.Referrer{}, &param.Referee{})
	},
	Rollback: func(db *gorm.DB) error {
		return db.Migrator().DropTable(&param.Event{}, &param.Campaign{}, &param.CampaignEvent{}, &param.Referrer{}, &param.Referee{})
	},
}
