package services

import (
	"time"

	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/jinzhu/gorm"
)

// JobQueueItem Item in JobQueue
type JobQueueItem struct {
	gorm.Model

	JobID uint        `sql:"index"`
	Job   *models.Job `gorm:"association_autoupdate:false;association_autocreate:false"`

	Position uint // The position in the Queue

	RunningSince time.Time `gorm:"-"`
	Deleted      bool      `gorm:"-"`
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

// Load JobQueueItem
func (jqi *JobQueueItem) Load(db *gorm.DB, config *models.Config) error {
	var queueItem JobQueueItem
	err := db.Model(&JobQueueItem{}).
		Preload("Job.BuildJob").
		Preload("Job.UploadJob").
		Where("id=?", jqi.ID).First(&queueItem).Error

	if err != nil {
		return err
	}

	if jqi.Job.BuildJob == nil {
		jqi.Job.BuildJob = queueItem.Job.BuildJob
		jqi.Job.BuildJob.Config = config
	}

	if jqi.Job.UploadJob == nil {
		jqi.Job.UploadJob = queueItem.Job.UploadJob
	}

	return nil
}
