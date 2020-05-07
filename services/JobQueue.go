package services

import (
	"github.com/JojiiOfficial/Remotebuild/models"
)

// JobQueue a queue for jobs
type JobQueue struct {
	Jobs []models.Job
}

// NewJobQueue create a new JobQueue
func NewJobQueue() *JobQueue {
	return &JobQueue{}
}
