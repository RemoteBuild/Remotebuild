package models

import (
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// BuildJob a job which builds a package
type BuildJob struct {
	gorm.Model
	State JobState // Build state
	Type  JobType  // Type of job

	Image   string            // Dockerimage to run
	Args    map[string]string `gorm:"-"` // Envars for Dockerimage
	Argdata string            `grom:"type:jsonb"`
}

// BuildResult result of a bulid
type BuildResult struct {
	NewBinary string
	Error     error
}

// NewBuildJob create new BuildJob
func NewBuildJob(db *gorm.DB, buildJob BuildJob) (*BuildJob, error) {
	buildJob.State = JobWaiting

	// Save Job to Db
	err := db.Create(&buildJob).Error
	if err != nil {
		return nil, err
	}

	return &buildJob, nil
}

// Run a buildjob (start but await)
func (buildJob *BuildJob) Run() *BuildResult {
	log.Debug("Run BuildJob ", buildJob.ID)
	// TODO implement run job
	time.Sleep(2 * time.Second)

	buildJob.State = JobDone
	return &BuildResult{
		Error: nil,
	}
}
