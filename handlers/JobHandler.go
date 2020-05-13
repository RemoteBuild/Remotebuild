package handlers

import (
	"net/http"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/JojiiOfficial/Remotebuild/services"
	docker "github.com/fsouza/go-dockerclient"
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

	// Get image for build job
	image, has := handlerData.Config.GetImage(request.Type)
	if !has {
		sendResponse(w, models.ResponseError, "no image available", nil, http.StatusNotFound)
		return
	}

	// Create new job
	job, err := models.NewJob(handlerData.Db, models.BuildJob{
		Type:  request.Type,
		Image: image,
	}, models.UploadJob{
		Type: request.UploadType,
	}, request.Args)

	if LogError(err) {
		sendServerError(w)
		return
	}

	// Add Job to queue
	jqi, err := handlerData.JobService.Queue.AddJob(job)
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
		jobQueueItem.Reload(handlerData.Db)
		job := jobQueueItem.Job

		jobInfos[i] = libremotebuild.JobInfo{
			ID:         job.ID,
			Info:       job.Info(),
			BuildType:  job.BuildJob.Type,
			Position:   jobQueueItem.Position,
			Status:     job.GetState(),
			UploadType: job.UploadJob.Type,
		}

		if job.GetState() == libremotebuild.JobRunning {
			jobInfos[i].RunningSince = jobQueueItem.RunningSince
		}
	}

	// Send list
	sendResponse(w, models.ResponseSuccess, "", libremotebuild.ListJobsResponse{
		Jobs: jobInfos,
	})
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

	if err := handlerData.Db.Save(job).Error; err != nil {
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

	// find requested job
	job := handlerData.JobService.Queue.FindJob(request.JobID)
	if job == nil {
		sendResponse(w, models.ResponseError, "Job not found", nil, http.StatusNotFound)
		return
	}

	containerID := job.Job.BuildJob.ContainerID

	// Check if container is running
	if len(containerID) == 0 {
		sendResponse(w, models.ResponseError, "No container running for job", nil, http.StatusUnprocessableEntity)
		return
	}

	// If job found, set required header for "success"
	w.Header().Set(models.HeaderStatus, "1")
	w.Header().Set(models.HeaderStatusMessage, "")
	w.WriteHeader(http.StatusOK)

	err := job.Job.BuildJob.Logs(docker.LogsOptions{
		Container:    containerID,
		Stderr:       true,
		Stdout:       true,
		Follow:       false,
		Since:        request.Since.Unix() + 1,
		OutputStream: w,
		ErrorStream:  w,
	})

	if err != nil {
		log.Error(err)
	}
}
