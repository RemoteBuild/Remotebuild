package handlers

import (
	"net/http"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
)

//Ping handles ping request
func Ping(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	var request libremotebuild.PingRequest
	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	payload := "pong"

	auth := NewAuthHandler(r)
	if len(auth.GetBearer()) > 0 {
		payload = "Authorized pong"
	}

	response := models.StringResponse{
		String: payload,
	}
	sendResponse(w, models.ResponseSuccess, "", response)
}
