package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/gaw"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// Job a job created by a user
type Job struct {
	gorm.Model

	// Buildjob
	BuildJobID uint      `sql:"index"`
	BuildJob   *BuildJob `gorm:"association_autoupdate:false;association_autocreate:false"`

	// UploadJob
	UploadJobID uint       `sql:"index"`
	UploadJob   *UploadJob `gorm:"association_autoupdate:false;association_autocreate:false"`

	DataDir  string // Shared dir containing build files
	Result   string // Message of an exited job
	LastLogs string // Latest logs
	Argdata  string `grom:"type:jsonb"`
	Info     string

	Args           map[string]string `gorm:"-"` // Envars for Dockerimage
	*gorm.DB       `gorm:"-"`
	Cancelled      bool          `gorm:"-"`
	LastSince      int64         `gorm:"-"`
	stopLogUpdater chan struct{} `gorm:"-"`
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
		DataDir:        path,
		Args:           args,
		DB:             db,
		stopLogUpdater: make(chan struct{}, 1),
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

// Init Job
func (job *Job) Init(db *gorm.DB) error {
	// Init channel
	if job.stopLogUpdater == nil {
		job.stopLogUpdater = make(chan struct{}, 1)
	}

	// Set DB
	job.DB = db

	if job.Args == nil {
		// TODO load argdata
	}

	return nil
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
func (job *Job) Cancel() {
	// Cancle actions
	job.stopLogUpdater <- struct{}{}
	job.BuildJob.cancel()
	job.UploadJob.cancel()

	// Update Job data
	job.Cancelled = true
	job.Result = "Cancelled"

	job.cleanup()
}

// SetState set the state of a job
func (job *Job) SetState(newState libremotebuild.JobState) {
	job.BuildJob.State = newState
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

// GetInfo for job
func (job *Job) GetInfo() string {
	if len(job.Info) > 0 {
		return job.Info
	}

	if job.BuildJob == nil {
		return "<noInfo>"
	}

	switch job.BuildJob.Type {
	case libremotebuild.JobAUR:
		job.Info = "AUR: " + job.Args[libremotebuild.AURPackage]
		return job.Info
	}

	return "<noInfo>"
}

// Cleanup a job
func (job *Job) cleanup() {
	// Remove Data dir
	if err := os.RemoveAll(job.DataDir); err != nil && !os.IsNotExist(err) {
		log.Warn(err)
	}

	// Clean Argdata
	job.Argdata = ""

	// Save changes and
	job.Save()
}

// Save job
func (job *Job) Save() (err error) {
	// Save Buildjob
	if err = job.DB.Save(job.BuildJob).Error; err != nil {
		return err
	}

	// Save uploadjob
	if err = job.DB.Save(job.UploadJob).Error; err != nil {
		return err
	}

	// Save actual job
	return job.DB.Save(job).Error
}

// Run a job
func (job *Job) Run() error {
	if job == nil {
		log.Warn("Job is nil")
		return nil
	}
	log.Debug("Run job ", job.ID)

	// Set Jobs info
	job.GetInfo()
	job.Save()

	// Cleanup data at the end
	defer func() {
		job.stopLogUpdater <- struct{}{}
		job.cleanup()
	}()

	go job.runLogUpdater()

	// New argParser
	argParser := NewArgParser(job.Args, job.BuildJob.Type)

	// Run Build
	buildResult := job.BuildJob.Run(job.DataDir, argParser)
	if buildResult.Error != nil {
		if buildResult.Error != ErrorJobCancelled {
			job.SetState(libremotebuild.JobFailed)
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
			job.SetState(libremotebuild.JobFailed)
			log.Info("Upload Failed: ", uploadResult.Error.Error())
		}

		return uploadResult.Error
	}

	log.Infof("Job %d done", job.ID)
	job.Result = "Success"
	return nil
}

// GetLogs for job
func (job *Job) GetLogs(requestTime time.Time, since int64, w io.Writer, checkAmbigious bool) error {
	if checkAmbigious {
		if since > 0 && job.LastSince >= since {
			return nil
		}
		job.LastSince = requestTime.Unix()
	}

	if job.GetState() != libremotebuild.JobRunning {
		return ErrJobNotRunning
	}

	// Get docker container logs, if build is running
	if job.BuildJob.State == libremotebuild.JobRunning {
		return job.BuildJob.GetLogs(since, w, "")
	}

	// If upload job is running, just use "Uploading"
	if job.UploadJob.State == libremotebuild.JobRunning {
		_, err := w.Write([]byte("Uploading"))
		return err
	}

	fmt.Println("no logs")

	return ErrNoLogsFound
}

func (job *Job) runLogUpdater() {
	wasJobRunning := false

	for {
		// Exit on stopLogUpdater
		select {
		case <-job.stopLogUpdater:
			return
		default:
		}

		// Exit if Buildjob is not running/waiting
		if job.BuildJob.State != libremotebuild.JobRunning &&
			job.BuildJob.State != libremotebuild.JobWaiting {
			return
		}

		bb := &bytes.Buffer{}

		// Get latest 20 log entries
		if err := job.BuildJob.GetLogs(0, bb, "20"); err != nil {
			if !wasJobRunning && err == ErrJobNotRunning {
				continue
			}

			log.Error(err)
			return
		}

		wasJobRunning = true

		// Set new lastlog
		if len(bb.Bytes()) > 0 {
			logs := bb.String()

			if len(logs) > 20 && len(job.LastLogs) > 20 {
				job.LastLogs = logs
			} else {
				job.LastLogs += logs
			}

			// Save job to DB
			job.DB.Model(job).Update("last_logs", job.LastLogs)
		}

	}
}

// ToJobInfo return JobInfo by job
func (job Job) ToJobInfo() libremotebuild.JobInfo {
	return libremotebuild.JobInfo{
		ID:         job.ID,
		Info:       job.GetInfo(),
		BuildType:  job.BuildJob.Type,
		Status:     job.GetState(),
		UploadType: job.UploadJob.Type,
	}
}
