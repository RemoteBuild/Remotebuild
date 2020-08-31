package storage

import (
	"fmt"

	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/JojiiOfficial/Remotebuild/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//ConnectToDatabase connects to database
func ConnectToDatabase(config *models.Config) (*gorm.DB, error) {
	sslMode := ""
	if len(config.Server.Database.SSLMode) > 0 {
		sslMode = "sslmode='" + config.Server.Database.SSLMode + "'"
	}

	dsn := fmt.Sprintf("host='%s' port='%d' user='%s' dbname='%s' password='%s' %s", config.Server.Database.Host, config.Server.Database.DatabasePort, config.Server.Database.Username, config.Server.Database.Database, config.Server.Database.Pass, sslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Automigration
	db.AutoMigrate(
		&models.LoginSession{},
		&models.User{},
		&models.BuildJob{},
		&models.UploadJob{},
		&models.Job{},
		&services.JobQueueItem{},
	)

	// Create default namespace
	return db, nil
}

// CheckConnection return true if connected succesfully
func CheckConnection(db *gorm.DB) (bool, error) {
	err := db.Exec("SELECT version();").Error
	return err == nil, err
}
