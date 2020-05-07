package handlers

import (
	"net/http"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
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
		Args:  request.Args,
	}, models.UploadJob{
		Type: request.UploadType,
	})

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
