package models

// Job a job created by a user
type Job struct {
	buildjob  *BuildJob
	uploadJob *UploadJob
}

// JobState a state of a job
type JobState uint8

// ...
const (
	JobWaiting JobState = iota
	JobRunning
	JobCancelled
)
