package services

import (
	"sort"
	"sync"
	"time"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// JobQueue a queue for jobs
type JobQueue struct {
	db      *gorm.DB
	jobs    []JobQueueItem
	mx      sync.RWMutex
	stopped chan bool
	currJob *JobQueueItem
}

// NewJobQueue create a new JobQueue
func NewJobQueue(db *gorm.DB) *JobQueue {
	queue := &JobQueue{
		db:      db,
		stopped: make(chan bool, 1),
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
	var jobs []JobQueueItem

	// Load unfinished jobs
	err := jq.db.Model(&JobQueueItem{}).
		Preload("Job").
		Preload("Job.BuildJob").
		Preload("Job.UploadJob").
		Find(&jobs).Error
	if err != nil {
		return err
	}

	var jobsToUse []JobQueueItem

	for i := range jobs {
		// Init Job
		err := jobs[i].Job.Init(jq.db)
		if err != nil {
			log.Error(err)
			continue
		}

		jobState := jobs[i].Job.GetState()

		// Set running jobs to waiting
		if jobState == libremotebuild.JobRunning {
			jobState = libremotebuild.JobWaiting
			jobs[i].Job.SetState(libremotebuild.JobWaiting)
		}

		// Ignore cancelled/failed/finished jobs
		if jobState == libremotebuild.JobRunning || jobState == libremotebuild.JobWaiting {
			jobsToUse = append(jobsToUse, jobs[i])
		}
	}

	jq.jobs = jobsToUse
	log.Infof("Loaded %d Jobs from old queue", len(jobsToUse))
	return nil
}

// AddJob to a jobqueue
func (jq *JobQueue) AddJob(job *models.Job) (*JobQueueItem, error) {
	item := &JobQueueItem{
		JobID: job.ID,
		Job:   job,
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

	jq.mx.Lock()
	defer jq.mx.Unlock()

	jq.jobs = append(jq.jobs, *item)

	log.Debugf("Job %d added", item.ID)
	return item, nil
}

// Start the queue async
func (jq *JobQueue) Start() {
	go jq.Run()
}

// Run the queue
func (jq *JobQueue) Run() {
	log.Info("Starting JobQueue")

	for {
		job := jq.getNextJob()

		jq.run(job)
		jq.currJob = nil

		// Only continue if not stopped
		select {
		case <-jq.stopped:
			log.Info("Stopped JobQueue")
			return
		default:
		}
	}
}

// Run a QueueItem
func (jq *JobQueue) run(jqi *JobQueueItem) {
	// Job is done after run, whatever
	// state it exited
	defer func() {
		// Delete jqi
		if err := jq.db.Delete(&jqi).Error; err != nil {
			log.Warn(err)
		}

		jqi.Deleted = true
		jq.RemoveJob(jqi.ID)
	}()

	// Get Job
	if err := jqi.Load(jq.db); err != nil {
		log.Error(err)
		return
	}

	jq.currJob = jqi
	jqi.RunningSince = time.Now()

	// Run job and log errors
	if err := jqi.Job.Run(); err != nil {
		if err != models.ErrorJobCancelled {
			log.Warn("Job exited with error: ", err)
		} else {
			log.Info("Job cancelled successfully")
		}
	}
}

func (jq *JobQueue) sortPosition() {
	sort.Sort(SortByPosition(jq.jobs))
}

func (jq *JobQueue) getNextJob() *JobQueueItem {
	for len(jq.jobs) == 0 {
		time.Sleep(1 * time.Second)
	}

	jq.sortPosition()
	return &jq.jobs[0]
}

// FindJob find job in queue
func (jq *JobQueue) FindJob(jobID uint) *JobQueueItem {
	// Find job in Queue slice
	for j := range jq.jobs {
		if jq.jobs[j].JobID == jobID {
			return &jq.jobs[j]
		}
	}

	return nil
}

// RemoveJob remove item from jobQueue
func (jq *JobQueue) RemoveJob(jobID uint) {
	if jq.currJob != nil && jq.currJob.Job.ID == jobID {
		jq.currJob = nil
	}

	i := -1

	// Find job in Queue slice
	for j := range jq.jobs {
		if jq.jobs[j].JobID == jobID {
			i = j
			break
		}
	}

	jq.mx.Lock()
	defer jq.mx.Unlock()

	// Remove job from actual slice
	jq.jobs[len(jq.jobs)-1], jq.jobs[i] = jq.jobs[i], jq.jobs[len(jq.jobs)-1]
	jq.jobs = jq.jobs[:len(jq.jobs)-1]
}

// GetJobQueuePos position of job in the queue
func (jq *JobQueue) GetJobQueuePos(jiq *JobQueueItem) int {
	jq.sortPosition()

	for i := range jq.jobs {
		if jq.jobs[i].ID == jiq.ID {
			return i
		}
	}

	return -1
}

// GetJobs return jobs in queue
func (jq *JobQueue) GetJobs() []JobQueueItem {
	var validJobs []JobQueueItem

	// Build slice with non-deleted jobs
	for i := range jq.jobs {
		if !jq.jobs[i].Deleted {
			validJobs = append(validJobs, jq.jobs[i])
		}
	}

	jq.mx.Lock()
	defer jq.mx.Unlock()

	sort.Sort(SortByPosition(validJobs))

	return validJobs
}

func (jq *JobQueue) stop() {
	jq.stopped <- true

	if jq.currJob != nil {
		jq.currJob.Job.Cancel()
	}
}
