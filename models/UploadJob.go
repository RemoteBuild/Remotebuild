package models

import (
	"github.com/jinzhu/gorm"
)

// UploadJob a job which uploads a built package
type UploadJob struct {
	gorm.Model
	State JobState // Upload state
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
	// TODO upload the binary

	uploadJob.State = JobDone
	return nil
}
