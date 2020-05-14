package services

import (
	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// JobService managing jobs
type JobService struct {
	*gorm.DB
	config *models.Config

	Queue *JobQueue
}

// NewJobService create a new jobservice
func NewJobService(config *models.Config, db *gorm.DB) *JobService {
	return &JobService{
		DB:     db,
		config: config,
		Queue:  NewJobQueue(db),
	}
}

// Start the jobservice
func (js *JobService) Start() {
	// Check for incompatibility
	if !js.check() {
		log.Fatalln("Starting Jobservice failed")
	}

	go js.Run()
}

// Run Start a job and await complete
func (js *JobService) Run() {
	log.Info("Starting JobService")
	// Start Build Queue
	js.Queue.Start()
}

func (js *JobService) check() bool {
	success := true

	if len(js.config.Server.Jobs.Images[libremotebuild.JobAUR.String()]) == 0 {
		log.Error("No Image specified for AUR building!")
		success = false
	}

	return success
}

// Stop the jobservice
func (js *JobService) Stop() {
	js.Queue.stop()

}

// GetOldJobs return n(limit) old jobs
func (js *JobService) GetOldJobs(limit int) ([]models.Job, error) {
	var jobs []models.Job

	if err := js.Model(&models.Job{}).
		Joins("left join build_jobs on build_jobs.id = jobs.build_job_id").
		Joins("left join upload_jobs on upload_jobs.id = jobs.upload_job_id").
		Preload("BuildJob").
		Preload("UploadJob").
		Where("build_jobs.state != 3 AND build_jobs.state != 0").
		Where("upload_jobs.state != 3 AND upload_jobs.state != 0").
		Order("jobs.id DESC").
		Limit(limit).Find(&jobs).Error; err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetOldLogs get old logs for job
func (js *JobService) GetOldLogs(jobID uint) (string, error) {
	var job models.Job

	// Get logs from DB
	err := js.Select("last_logs").Where("id=?", jobID).Find(&job).Error
	if err != nil {
		return "", err
	}

	return job.LastLogs, nil
}
