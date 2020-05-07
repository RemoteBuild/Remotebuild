package models

import (
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// UploadJob a job which uploads a built package
type UploadJob struct {
	gorm.Model
	State JobState // Upload state

	cancel chan bool `gorm:"-"` // Cancel chan
}

// UploadJobResult result of uploading a binary
type UploadJobResult struct {
	Error error
}

// UploadJobType type of uploadJob
type UploadJobType uint8

// ...
const (
	NoUploadType UploadJobType = iota
	DataManagerUploadType
)

// NewUploadJob create new upload job
func NewUploadJob(db *gorm.DB, uploadJob UploadJob) (*UploadJob, error) {
	uploadJob.State = JobWaiting

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
	time.Sleep(1 * time.Second)

	uploadJob.State = JobDone
	return &UploadJobResult{
		Error: nil,
	}
}
