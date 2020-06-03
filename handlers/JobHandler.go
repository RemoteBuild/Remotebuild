package handlers

import (
	"net/http"
	"strconv"
	"time"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/JojiiOfficial/Remotebuild/services"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// AddJob add a job
func addJob(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	var request libremotebuild.AddJobRequest

	// Read request
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	// Check input
	if len(request.Type.String()) == 0 {
		sendResponse(w, models.ResponseError, "input missing", nil, http.StatusUnprocessableEntity)
		return
	}

	// Validate request build type
	switch request.Type {
	case libremotebuild.JobAUR:
	default:
		sendResponse(w, models.ResponseError, "build type not supported", "", http.StatusUnprocessableEntity)
		return
	}

	// Add Job to queue
	jqi, err := handlerData.JobService.Queue.AddNewJob(handlerData.Db, request.Type, request.UploadType, request.Args, (!request.DisableCcache))
	if LogError(err) {
		sendServerError(w)
		return
	}

	sendResponse(w, models.ResponseSuccess, "", libremotebuild.AddJobResponse{
		ID:       jqi.ID,
		Position: handlerData.JobService.Queue.GetJobQueuePos(jqi),
	})
}

// listJobs view the queue
func listJobs(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	jobs := handlerData.JobService.Queue.GetJobs()
	jobInfos := make([]libremotebuild.JobInfo, len(jobs))

	// Bulid JobInfos
	for i, jobQueueItem := range jobs {
		jobQueueItem.Load(handlerData.Db, handlerData.Config)
		job := jobQueueItem.Job

		jobInfos[i] = job.ToJobInfo()
		jobInfos[i].Position = jobQueueItem.Position

		if job.GetState() == libremotebuild.JobRunning {
			jobInfos[i].RunningSince = jobQueueItem.RunningSince
		}
	}

	// Build response
	resp := libremotebuild.ListJobsResponse{
		Jobs: jobInfos,
	}

	// Append old jobs
	oldJobs, err := handlerData.JobService.GetOldJobs(2)
	if err != nil {
		log.Warn(err)
	} else {
		for i := range oldJobs {
			resp.Jobs = append(resp.Jobs, oldJobs[i].ToJobInfo())
		}
	}

	// Send list
	sendResponse(w, models.ResponseSuccess, "", resp)
}

// cancelJob cancel a job
func cancelJob(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	var request libremotebuild.JobRequest
	// Read request
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	// Get Job
	job := handlerData.JobService.Queue.FindJob(request.JobID)
	if job == nil {
		sendResponse(w, models.ResponseError, "no such job found", nil, http.StatusNotFound)
		return
	}

	// Cancel job
	job.Job.Cancel()
	job.Deleted = true

	if err := job.Job.Save().Error; err != nil {
		log.Info(err)
	}

	// send success
	sendResponse(w, models.ResponseSuccess, "cancel successful", nil)

	// Remove from Db
	handlerData.Db.Where("job_id=?", request.JobID).Delete(&services.JobQueueItem{})
	log.Info("Cancelled Job ", request.JobID)
}

// get logs of a job
func getLogs(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	var request libremotebuild.JobLogsRequest
	// Read request
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	// Try getting requested runnig job
	if job := handlerData.JobService.Queue.FindJob(request.JobID); job != nil {
		// Check if container is running
		if len(job.Job.BuildJob.ContainerID) == 0 {
			sendResponse(w, models.ResponseError, "No container running for job", nil, http.StatusUnprocessableEntity)
			return
		}

		requestTime := time.Now()

		// If job found, set required header for "success"
		w.Header().Set(models.HeaderStatus, "1")
		w.Header().Set(models.HeaderStatusMessage, strconv.FormatInt(requestTime.Unix(), 10))
		w.WriteHeader(http.StatusOK)

		// Send logs
		if err := job.Job.GetLogs(requestTime, request.Since.Unix(), w, true); err != nil {
			log.Error(err)
		}
	} else {
		// If no runnig job with requested ID was found
		logs, err := handlerData.JobService.GetOldLogs(request.JobID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				sendResponse(w, models.ResponseError, "Job not found", nil, http.StatusNotFound)
				return
			}
		}

		// If job found, set required header for "success"
		w.Header().Set(models.HeaderStatus, "1")
		w.Header().Set(models.HeaderStatusMessage, "-1")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(logs))
	}
}
