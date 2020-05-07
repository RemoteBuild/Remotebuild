package services

import (
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/jinzhu/gorm"
)

// JobQueueItem Item in JobQueue
type JobQueueItem struct {
	gorm.Model

	JobID uint        `sql:"index"`
	Job   *models.Job `gorm:"association_autoupdate:false;association_autocreate:false"`

	Position uint // The position in the Queue

	Done bool // Wether the Job is already done or not
}

// SortByPosition sort by JobQueueItem position
type SortByPosition []JobQueueItem

func (a SortByPosition) Len() int           { return len(a) }
func (a SortByPosition) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByPosition) Less(i, j int) bool { return a[i].Position < a[j].Position }

// TableName use "JobQueue" as tablename
func (jqi JobQueueItem) TableName() string {
	return "job_queue"
}

// GetJob returns job for JobQueueitem
func (jqi *JobQueueItem) GetJob(db *gorm.DB) (*models.Job, error) {
	var job JobQueueItem

	err := db.Model(&JobQueueItem{}).
		Preload("Job").
		Preload("Job.BuildJob").
		Preload("Job.UploadJob").
		Where("id=?", jqi.ID).First(&job).Error

	jqi.Job = job.Job
	return job.Job, err
}
