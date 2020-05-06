package services

import (
	"fmt"
	"time"

	"github.com/DataManager-Go/DataManagerServer/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// CleanupService cleanupservice cleansup stuff in background from DB
type CleanupService struct {
	db     *gorm.DB
	config *models.Config
}

// NewClienupService create a new cleanupservice
func NewClienupService(config *models.Config, db *gorm.DB) *CleanupService {
	return &CleanupService{
		config: config,
		db:     db,
	}
}

// Start starts the service
func (cs *CleanupService) Start() {
	cs.debug()
	go cs.run()
}

func (cs *CleanupService) run() {
	for {
		cs.deleteUnusedSessions()
		time.Sleep(1 * time.Hour)
	}
}

// Deletes unused sessions after in config specified duration
func (cs *CleanupService) deleteUnusedSessions() {
	// Delete where requests = 0 and creation > specified allowed time
	e := cs.db.Unscoped().
		Where("requests = 0").
		Where(fmt.Sprintf("created_at < now() - interval '%d seconds'",
			int(cs.config.Server.DeleteUnusedSessionsAfter.Seconds()))).
		Delete(&models.LoginSession{})

	// Log error
	if e.Error != nil {
		log.Error(e.Error)
		return
	}

	log.Infof("Deleted %d unused sessions", e.RowsAffected)
}

// just debug things
func (cs *CleanupService) debug() {
	log.Debugf("Deleting unused sessions after %s", cs.config.Server.DeleteUnusedSessionsAfter.String())
}
