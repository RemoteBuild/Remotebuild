package models

import (
	"github.com/jinzhu/gorm"
)

// UploadJob a job which uploads a built package
type UploadJob struct {
	gorm.Model
	State JobState // Upload state
}

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
