package services

import (
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/jinzhu/gorm"
)

// JobService managing jobs
type JobService struct {
	db     *gorm.DB
	config *models.Config
}

// NewJobService create a new jobservice
func NewJobService(config *models.Config, db *gorm.DB) *JobService {
	return &JobService{
		db:     db,
		config: config,
	}
}

// Start the jobservice
func (js *JobService) Start() {
	go js.Run()
}

// Run Start a job and await complete
func (js *JobService) Run() {
	// TODO
}
