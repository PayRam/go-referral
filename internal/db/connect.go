package db

import (
	"github.com/PayRam/go-referral/internal/migration"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

// InitDB initializes and returns the database connection
func InitDB(dbFilePath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbFilePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db = Migrate(db)

	return db
}

func Migrate(db *gorm.DB) *gorm.DB {
	// Run migrations
	if err := migrate(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Log or handle successful database setup as needed
	log.Printf("**** Database initialised and migrations run successfully ****")
	return db
}

func migrate(db *gorm.DB) error {
	// Define migrations
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{ // Integrate the CreateUserTable migration
			ID:       migration.Initialise.ID,
			Migrate:  migration.Initialise.Migrate,
			Rollback: migration.Initialise.Rollback,
		},
	})

	return m.Migrate()
}
