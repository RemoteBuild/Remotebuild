package models

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/DataManager-Go/libdatamanager"
	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrNoManagerDataAvailable if no datamanager data is available but required
	ErrNoManagerDataAvailable = errors.New("No DManager data available")

	// ErrNoVaildUploadMetodPassed if no uploadmethod/data was passed
	ErrNoVaildUploadMetodPassed = errors.New("No vaild upolad method passed")
)

// UploadJob a job which uploads a built package
type UploadJob struct {
	gorm.Model
	State libremotebuild.JobState // Upload state

	Type libremotebuild.UploadType

	cancel chan bool `gorm:"-"` // Cancel chan
}

// UploadJobResult result of uploading a binary
type UploadJobResult struct {
	Error error
}

// NewUploadJob create new upload job
func NewUploadJob(db *gorm.DB, uploadJob UploadJob) (*UploadJob, error) {
	uploadJob.State = libremotebuild.JobWaiting
	uploadJob.cancel = make(chan bool, 1)

	// Save Job into DB
	err := db.Create(&uploadJob).Error
	if err != nil {
		return nil, err
	}

	return &uploadJob, nil
}

// Init the uploadJob
func (uploadJob *UploadJob) Init() {
	if uploadJob.cancel == nil {
		uploadJob.cancel = make(chan bool, 1)
	}
}

// Run an upload job
func (uploadJob *UploadJob) Run(buildResult BuildResult, argParser *ArgParser) *UploadJobResult {
	uploadJob.Init()

	log.Debug("Run UploadJob ", uploadJob.ID)
	uploadJob.State = libremotebuild.JobRunning

	// Verify Dmanager data
	if uploadJob.Type == libremotebuild.DataManagerUploadType && !argParser.HasDataManagerArgs() {
		return &UploadJobResult{
			Error: ErrNoManagerDataAvailable,
		}
	}

	return uploadJob.upload(buildResult, argParser)
}

func (uploadJob *UploadJob) upload(buildResult BuildResult, argParser *ArgParser) *UploadJobResult {
	// Pick correct upload method
	switch uploadJob.Type {
	case libremotebuild.DataManagerUploadType:
		return uploadJob.uploadDmanager(buildResult, argParser)
	}

	// If no uploadtype was set, return error
	uploadJob.State = libremotebuild.JobFailed
	return &UploadJobResult{
		Error: ErrNoVaildUploadMetodPassed,
	}
}

// Upload to datamanager
// See https://github.com/DataManager-Go/DataManagerServer
func (uploadJob *UploadJob) uploadDmanager(buildResult BuildResult, argParser *ArgParser) *UploadJobResult {
	dmanagerData := argParser.GetDManagerData()

	// Decode base64 encoded token
	unencodedToken, err := base64Decode(dmanagerData.Token)
	if err != nil {
		uploadJob.State = libremotebuild.JobFailed
		return &UploadJobResult{
			Error: err,
		}
	}

	_, filename := filepath.Split(buildResult.Archive)

	// Create uploadrequest
	uploadRequest := libdatamanager.NewLibDM(&libdatamanager.RequestConfig{
		URL:          dmanagerData.Host,
		Username:     dmanagerData.Username,
		SessionToken: unencodedToken,
	}).NewUploadRequest(filename, libdatamanager.FileAttributes{
		Groups: []string{"AURpackage"},
	})

	// Open file
	f, err := os.Open(buildResult.Archive)
	if err != nil {
		uploadJob.State = libremotebuild.JobFailed
		return &UploadJobResult{
			Error: err,
		}
	}

	// Upload file
	_, err = uploadRequest.UploadFile(f, nil, uploadJob.cancel)
	if err != nil {
		uploadJob.State = libremotebuild.JobFailed
		return &UploadJobResult{
			Error: err,
		}
	}

	uploadJob.State = libremotebuild.JobDone
	return nil
}
