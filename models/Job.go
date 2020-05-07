package models

import (
	"os"
	"path/filepath"

	"github.com/JojiiOfficial/gaw"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// Job a job created by a user
type Job struct {
	gorm.Model

	BuildJobID uint      `sql:"index"`
	BuildJob   *BuildJob `gorm:"association_autoupdate:false;association_autocreate:false"`

	UploadJobID uint       `sql:"index"`
	UploadJob   *UploadJob `gorm:"association_autoupdate:false;association_autocreate:false"`

	DataDir string // Shared dir containing build files

	Result string
}

// JobState a state of a job
type JobState uint8

// ...
const (
	JobWaiting JobState = iota
	JobRunning
	JobCancelled
	JobDone
)

// JobType type of job
type JobType uint8

// ...
const (
	JobNoBuild JobType = iota
	JobAUR
)

func (jt JobType) String() string {
	switch jt {
	case JobNoBuild:
		return "NoJob"
	case JobAUR:
		return "buildAUR"
	}

	return ""
}

// ParseJobType parse a jobtype from string
func ParseJobType(inp string) JobType {
	switch inp {
	case "NoJob":
		return JobNoBuild
	case "buildAUR":
		return JobAUR
	}

	return JobNoBuild
}

// NewJob create a new job
func NewJob(db *gorm.DB, buildJob BuildJob, uploadJob UploadJob) (*Job, error) {
	// Create temporary path for storing build data
	path := filepath.Join(os.TempDir(), gaw.RandString(30))
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return nil, err
	}

	job := &Job{DataDir: path}

	// Create BuildJob
	bJob, err := NewBuildJob(db, buildJob)
	if err != nil {
		return nil, err
	}
	job.BuildJobID = bJob.ID

	// Create UploadJob
	upjob, err := NewUploadJob(db, uploadJob)
	if err != nil {
		return nil, err
	}
	job.UploadJobID = upjob.ID

	// Save Job into Db
	err = db.Create(job).Error
	if err != nil {
		return nil, err
	}

	return job, nil
}

// Cancel Job
func (job *Job) Cancel() error {
	job.BuildJob.State = JobCancelled
	job.UploadJob.State = JobCancelled
	job.Result = "Cancelled"

	// TODO
	return nil
}

// GetState get state of a job
func (job *Job) GetState() JobState {
	// If BuildJob is not done yet, use its State
	if job.BuildJob.State != JobDone {
		return job.BuildJob.State
	}

	// Otherwise the Jobs state is the UploadsJob state
	return job.UploadJob.State
}

// Cleanup a job
func (job *Job) cleanup() {
	// Remove Data dir
	err := os.RemoveAll(job.DataDir)
	if err != nil {
		log.Warn(err)
	}

}
