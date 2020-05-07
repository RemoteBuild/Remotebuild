package services

import (
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// JobQueueItem Item in JobQueue
type JobQueueItem struct {
	gorm.Model

	JobID uint `sql:"index"`
	Job   *models.Job

	Position uint // The position in the Queue

	Done bool // Wether the Job is already done or not
}

// TableName use "JobQueue" as tablename
func (jqi JobQueueItem) TableName() string {
	return "job_queue"
}

// JobQueue a queue for jobs
type JobQueue struct {
	db   *gorm.DB
	Jobs []*JobQueueItem
}

// NewJobQueue create a new JobQueue
func NewJobQueue(db *gorm.DB) *JobQueue {
	queue := &JobQueue{
		db: db,
	}

	// Load Queue
	err := queue.Load()
	if err != nil {
		log.Fatalln(err)
	}

	return queue
}

// Load queue from Db
func (jq *JobQueue) Load() error {
	return nil
}

// AddJob to a jobqueue
func (jq *JobQueue) AddJob(job *models.Job) (*JobQueueItem, error) {
	item := &JobQueueItem{
		JobID: job.ID,
	}

	// Insert Item
	err := jq.db.Create(item).Error
	if err != nil {
		return nil, err
	}

	// Use ID as Position
	item.Position = item.ID
	err = jq.db.Save(item).Error
	if err != nil {
		return nil, err
	}

	jq.Jobs = append(jq.Jobs, item)
	return item, nil
}
