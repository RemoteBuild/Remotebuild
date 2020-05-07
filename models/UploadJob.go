package models

import (
	"time"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// UploadJob a job which uploads a built package
type UploadJob struct {
	gorm.Model
	State JobState // Upload state

	Type libremotebuild.UploadType

	cancel chan bool `gorm:"-"` // Cancel chan
}

// UploadJobResult result of uploading a binary
type UploadJobResult struct {
	Error error
}

// NewUploadJob create new upload job
func NewUploadJob(db *gorm.DB, uploadJob UploadJob) (*UploadJob, error) {
	uploadJob.State = JobWaiting
	uploadJob.cancel = make(chan bool, 1)

	// Save Job into DB
	err := db.Create(&uploadJob).Error
	if err != nil {
		return nil, err
	}

	return &uploadJob, nil
}

// Run an upload job
func (uploadJob *UploadJob) Run() *UploadJobResult {
	log.Debug("Run UploadJob ", uploadJob.ID)

	// TODO upload the binary
	time.Sleep(8 * time.Second)

	uploadJob.State = JobDone
	return &UploadJobResult{
		Error: nil,
	}
}
