package storage

import (
	"fmt"

	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/JojiiOfficial/Remotebuild/services"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

//ConnectToDatabase connects to database
func ConnectToDatabase(config *models.Config) (*gorm.DB, error) {
	sslMode := ""
	if len(config.Server.Database.SSLMode) > 0 {
		sslMode = "sslmode='" + config.Server.Database.SSLMode + "'"
	}

	db, err := gorm.Open("postgres", fmt.Sprintf("host='%s' port='%d' user='%s' dbname='%s' password='%s' %s", config.Server.Database.Host, config.Server.Database.DatabasePort, config.Server.Database.Username, config.Server.Database.Database, config.Server.Database.Pass, sslMode))
	if err != nil {
		return nil, err
	}

	//Automigration
	err = db.AutoMigrate(
		&models.Role{},
		&models.LoginSession{},
		&models.User{},
		&models.BuildJob{},
		&models.UploadJob{},
		&models.Job{},
		&services.JobQueueItem{},
	).Error

	//Return error if automigration fails
	if err != nil {
		return nil, err
	}

	// TODO Create roles
	// createRoles(db, config)

	//Create default namespace
	return db, nil
}

func createRoles(db *gorm.DB, config *models.Config) {
	//Create in config specified roles
	for _, role := range config.Server.Roles.Roles {
		err := db.FirstOrCreate(&role).Error
		if err != nil {
			log.Fatalln(err)
		}
	}
}

//CheckConnection return true if connected succesfully
func CheckConnection(db *gorm.DB) (bool, error) {
	err := db.Exec("SELECT version();").Error
	return err == nil, err
}
