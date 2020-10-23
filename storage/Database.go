package storage

import (
	"fmt"
	"strings"

	"github.com/RemoteBuild/Remotebuild/models"
	"github.com/RemoteBuild/Remotebuild/services"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//ConnectToDatabase connects to database
func ConnectToDatabase(config *models.Config) (*gorm.DB, error) {
	dbType := strings.ToLower(config.Server.Database.DatabaseType)

	// Use default if notset
	if len(dbType) == 0 {
		dbType = "sqlite"
	}

	var dialector gorm.Dialector
	if dbType == "postgres" {
		sslMode := ""
		if len(config.Server.Database.SSLMode) > 0 {
			sslMode = "sslmode='" + config.Server.Database.SSLMode + "'"
		}

		dsn := fmt.Sprintf("host='%s' port='%d' user='%s' dbname='%s' password='%s' %s", config.Server.Database.Host, config.Server.Database.DatabasePort, config.Server.Database.Username, config.Server.Database.Database, config.Server.Database.Pass, sslMode)
		dialector = postgres.Open(dsn)
	} else if dbType == "sqlite" {
		dbFile := config.Server.Database.DbFile
		if len(config.Server.Database.DbFile) == 0 {
			dbFile = "data.db"
		}
		dialector = sqlite.Open(dbFile)
	} else {
		return nil, fmt.Errorf("%s is not a supported database type!", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
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

	// Don't perform connection tests if sqlite is picked
	if dbType == "sqlite" {
		return db, nil
	}

	connected, err := CheckConnection(db)
	if err != nil {
		return nil, err
	}

	if !connected {
		return nil, fmt.Errorf("Can't connect to DB!")
	}

	// Create default namespace
	return db, nil
}

// CheckConnection return true if connected succesfully
func CheckConnection(db *gorm.DB) (bool, error) {
	err := db.Exec("SELECT version();").Error
	return err == nil, err
}
