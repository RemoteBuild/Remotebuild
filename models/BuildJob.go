package models

import (
	"encoding/json"
	"time"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// BuildJob a job which builds a package
type BuildJob struct {
	gorm.Model
	State libremotebuild.JobState // Build state
	Type  libremotebuild.JobType  // Type of job

	Image   string            // Dockerimage to run
	Args    map[string]string `gorm:"-"` // Envars for Dockerimage
	Argdata string            `grom:"type:jsonb"`

	cancel chan bool `gorm:"-"` // Cancel chan
}

// BuildResult result of a bulid
type BuildResult struct {
	NewBinary string
	Error     error
}

// NewBuildJob create new BuildJob
func NewBuildJob(db *gorm.DB, buildJob BuildJob) (*BuildJob, error) {
	buildJob.State = libremotebuild.JobWaiting
	buildJob.cancel = make(chan bool, 1)

	buildJob.putArgs()

	// Save Job to Db
	err := db.Create(&buildJob).Error
	if err != nil {
		return nil, err
	}

	return &buildJob, nil
}

// Tranlate Args to Argdata
func (buildJob *BuildJob) putArgs() error {
	b, err := json.Marshal(buildJob.Args)
	if err != nil {
		return err
	}

	buildJob.Argdata = string(b)

	return nil
}

// Run a buildjob (start but await)
func (buildJob *BuildJob) Run() *BuildResult {
	log.Debug("Run BuildJob ", buildJob.ID)
	buildJob.State = libremotebuild.JobRunning

	buildDone := make(chan bool, 1)
	var result *BuildResult

	// Run build in goroutine
	go func() {
		result = buildJob.build()
		buildDone <- true
	}()

	// Await build or cancel
	select {
	case <-buildDone:
		// On done
		return result
	case <-buildJob.cancel:
		// On cancel
		buildJob.State = libremotebuild.JobCancelled
		return &BuildResult{
			Error: ErrorJobCancelled,
		}
	}
}

func (buildJob *BuildJob) build() *BuildResult {
	// TODO implement run job
	time.Sleep(5 * time.Second)

	buildJob.State = libremotebuild.JobDone
	return &BuildResult{
		Error: nil,
	}
}
