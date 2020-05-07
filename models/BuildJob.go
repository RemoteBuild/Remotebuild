package models

import "github.com/jinzhu/gorm"

// BuildJob a job which builds a package
type BuildJob struct {
	gorm.Model
	State JobState // Build state
	Type  JobType  // Type of job

	Image   string            // Dockerimage to run
	Args    map[string]string `gorm:"-"` // Envars for Dockerimage
	Argdata string            `grom:"type:jsonb"`
}

// NewBuildJob create new BuildJob
func NewBuildJob(db *gorm.DB, buildJob BuildJob) (*BuildJob, error) {
	buildJob.State = JobWaiting

	// Save Job to Db
	err := db.Create(&buildJob).Error
	if err != nil {
		return nil, err
	}

	return &buildJob, nil
}

// Start a buildjob
func (buildJob *BuildJob) Start() {
	go buildJob.Run()
}

// Run a buildjob (start but await)
func (buildJob *BuildJob) Run() {

}
