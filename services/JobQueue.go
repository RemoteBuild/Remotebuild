package services

import (
	"sort"
	"sync"

	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// JobQueue a queue for jobs
type JobQueue struct {
	db *gorm.DB

	jobs []JobQueueItem

	mx sync.RWMutex

	tick chan bool
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

	queue.tick = make(chan bool)

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

	for range jobs {
		go func() {
			jq.tick <- true
		}()
	}

	jq.jobs = jobs
	log.Infof("Loaded %d Jobs from old queue", len(jobs))
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

	go func() {
		jq.tick <- true
	}()

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

	// Remove Job from queue
	jq.removeItem(jqi)

	// Update DB
	if err := jq.db.Save(&jqi).Error; err != nil {
		log.Warn(err)
	}
}

func (jq *JobQueue) nextJob() *JobQueueItem {
	<-jq.tick
	sort.Sort(SortByPosition(jq.jobs))
	return &jq.jobs[0]
}

// Remove item from jobQueue
func (jq *JobQueue) removeItem(item *JobQueueItem) {
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
