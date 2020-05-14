package models

import (
	"encoding/json"
	"os"
	"path/filepath"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
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

	DataDir   string // Shared dir containing build files
	Result    string
	Cancelled bool `gorm:"-"`

	Args    map[string]string `gorm:"-"` // Envars for Dockerimage
	Argdata string            `grom:"type:jsonb"`
}

// NewJob create a new job
func NewJob(db *gorm.DB, buildJob BuildJob, uploadJob UploadJob, args map[string]string) (*Job, error) {
	// Create temporary path for storing build data
	path := filepath.Join(os.TempDir(), "remotebuild_"+gaw.RandString(30))
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return nil, err
	}

	job := &Job{
		DataDir: path,
		Args:    args,
	}

	job.putArgs()

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

// Tranlate Args to Argdata
func (job *Job) putArgs() error {
	b, err := json.Marshal(job.Args)
	if err != nil {
		return err
	}

	job.Argdata = string(b)

	return nil
}

// Cancel Job
func (job *Job) Cancel(db *gorm.DB) {
	// Cancle actions
	job.BuildJob.cancel()
	job.UploadJob.cancel()

	// Update Job data
	job.Cancelled = true
	job.Result = "Cancelled"
	job.Argdata = ""

	// Save changes and
	// delete job
	db.Save(job)
	db.Delete(job)
}

// SetState set the state of a job
func (job *Job) SetState(newState libremotebuild.JobState) {
	// If build job is not done, set its job
	if job.BuildJob.State != libremotebuild.JobDone {
		job.BuildJob.State = newState
		return
	}

	job.UploadJob.State = newState
}

// GetState get state of a job
func (job *Job) GetState() libremotebuild.JobState {
	// If BuildJob is not done yet, use its State
	if job.BuildJob.State != libremotebuild.JobDone {
		return job.BuildJob.State
	}

	// Otherwise the Jobs state is the UploadsJob state
	return job.UploadJob.State
}

// Info for job
func (job *Job) Info() string {
	if job.BuildJob == nil {
		return "<noInfo>"
	}

	switch job.BuildJob.Type {
	case libremotebuild.JobAUR:
		return "AUR: " + job.Args[libremotebuild.AURPackage]
	}

	return "<noInfo>"
}

// Cleanup a job
func (job *Job) cleanup() {
	// Remove Data dir
	err := os.RemoveAll(job.DataDir)
	if err != nil {
		log.Warn(err)
	}
}

// Run a job
func (job *Job) Run() error {
	if job == nil {
		log.Warn("Job is nil")
		return nil
	}
	log.Debug("Run job ", job.ID)

	// Cleanup data at the end
	defer job.cleanup()

	// New argParser
	argParser := NewArgParser(job.Args, job.BuildJob.Type)

	// Run Build
	buildResult := job.BuildJob.Run(job.DataDir, argParser)
	if buildResult.Error != nil {
		if buildResult.Error != ErrorJobCancelled {
			job.BuildJob.State = libremotebuild.JobFailed
			log.Info("Build Failed: ", buildResult.Error.Error())
		}

		return buildResult.Error
	}

	if job.Cancelled {
		return ErrorJobCancelled
	}

	// Run upload
	uploadResult := job.UploadJob.Run(*buildResult, argParser)
	if uploadResult != nil && uploadResult.Error != nil {
		if uploadResult.Error != ErrorJobCancelled {
			job.UploadJob.State = libremotebuild.JobFailed
			log.Info("Upload Failed: ", uploadResult.Error.Error())
		}
		return uploadResult.Error
	}

	log.Infof("Job %d done", job.ID)
	return nil
}
