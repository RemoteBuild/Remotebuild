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
	db   *gorm.DB
	jobs []JobQueueItem
	mx   sync.RWMutex
	tick chan uint
}

// NewJobQueue create a new JobQueue
func NewJobQueue(db *gorm.DB) *JobQueue {
	queue := &JobQueue{
		db: db,
	}

	queue.tick = make(chan uint, 1)

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
		Where("done=false").Find(&jobs).Error
	if err != nil {
		return err
	}

	var jobsToUse []JobQueueItem

	for i := range jobs {
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

	jq.mx.Lock()
	defer jq.mx.Unlock()

	numJobs := uint(len(jobsToUse))

	jq.tick <- numJobs
	jq.jobs = jobsToUse
	log.Infof("Loaded %d Jobs from old queue", numJobs)
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

	jq.mx.Lock()
	defer jq.mx.Unlock()

	jq.jobs = append(jq.jobs, *item)

	i := <-jq.tick
	jq.tick <- i + 1

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
		job := jq.nextJob()
		jq.run(job)
	}
}

// Run a QueueItem
func (jq *JobQueue) run(jqi *JobQueueItem) {
	// Get Job
	err := jqi.Reload(jq.db)
	if err != nil {
		log.Error(err)
		return
	}

	// Run job and log errors
	if err := jqi.Job.Run(); err != nil {
		log.Warn("Job exited with error:", err)
	}

	// Job is done after run, whatever
	// state it exited
	jqi.Done = true

	// Update DB
	if err := jq.db.Save(&jqi).Error; err != nil {
		log.Warn(err)
	}

	// Delete jqi
	if err := jq.db.Delete(&jqi).Error; err != nil {
		log.Warn(err)
	}

	// Remove Job from queue
	jq.RemoveJob(jqi)
}

func (jq *JobQueue) sortPosition() {
	sort.Sort(SortByPosition(jq.jobs))
}

func (jq *JobQueue) nextJob() *JobQueueItem {
	i := <-jq.tick
	jq.mx.Lock()
	jq.tick <- i
	jq.mx.Unlock()

	for i == 0 || len(jq.jobs) == 0 {
		time.Sleep(1 * time.Second)
		i = <-jq.tick

		jq.mx.Lock()
		if len(jq.jobs) == 0 {
			jq.tick <- 0
		} else {
			jq.tick <- i
		}
		jq.mx.Unlock()
	}

	jq.sortPosition()
	return &jq.jobs[0]
}

// RemoveJob remove item from jobQueue
func (jq *JobQueue) RemoveJob(item *JobQueueItem) {
	i := -1

	// Find job in Queue slice
	for j := range jq.jobs {
		if jq.jobs[j].ID == item.JobID {
			i = j
			break
		}
	}

	// exit if pos not found
	if i == -1 {
		return
	}

	// Remove
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
	jq.sortPosition()
	return jq.jobs
}
