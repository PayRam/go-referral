package migration

import (
	"github.com/PayRam/go-referral/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Initialise = &gormigrate.Migration{
	ID: "202412191749-gr-473842",
	Migrate: func(db *gorm.DB) error {
		return db.AutoMigrate(
			&models.Event{},
			&models.Campaign{},
			&models.CampaignEvent{},
			&models.Member{},
			&models.MemberCampaign{},
			&models.EventLog{},
			&models.CampaignEventLog{},
			&models.Reward{},
		)
	},
	Rollback: func(db *gorm.DB) error {
		return db.Migrator().DropTable(
			&models.Event{},
			&models.Campaign{},
			&models.CampaignEvent{},
			&models.Member{},
			&models.MemberCampaign{},
			&models.EventLog{},
			&models.CampaignEventLog{},
			&models.Reward{},
		)
	},
}
