package services

import (
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// JobService managing jobs
type JobService struct {
	db     *gorm.DB
	config *models.Config

	Queue *JobQueue
}

// NewJobService create a new jobservice
func NewJobService(config *models.Config, db *gorm.DB) *JobService {
	return &JobService{
		db:     db,
		config: config,
		Queue:  NewJobQueue(db),
	}
}

// Start the jobservice
func (js *JobService) Start() {
	// Check for incompatibility
	if !js.check() {
		log.Fatalln("Starting Jobservice failed")
	}

	go js.Run()
}

// Run Start a job and await complete
func (js *JobService) Run() {
	// TODO
}

func (js *JobService) check() bool {
	success := true

	if len(js.config.Server.Jobs.Images[models.JobAUR.String()]) == 0 {
		log.Error("No Image specified for AUR building!")
		success = false
	}

	return success
}
