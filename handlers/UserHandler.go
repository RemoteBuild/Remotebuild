package handlers

import (
	"net/http"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/JojiiOfficial/gaw"
)

//Login login handler
//-> /user/login
func Login(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	var request libremotebuild.CredentialsRequest

	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	if len(request.Password) == 0 || len(request.Username) == 0 {
		sendResponse(w, models.ResponseError, "input missing", nil, http.StatusUnprocessableEntity)
		return
	}

	user := models.User{
		Username: request.Username,
		Password: gaw.SHA512(request.Username + request.Password),
	}

	session, err := user.Login(handlerData.Db, request.MachineID)
	if err != nil {
		sendResponse(w, models.ResponseError, "Invalid credentials", nil)
		return
	}

	if session != nil {
		sendResponse(w, models.ResponseSuccess, "", models.LoginResponse{
			Token: session.Token,
		})
	} else {
		sendResponse(w, models.ResponseError, "Error logging in", nil, http.StatusUnauthorized)
	}
}

//Register register handler
//-> /user/create
func Register(handlerData HandlerData, w http.ResponseWriter, r *http.Request) {
	if !handlerData.Config.Server.AllowRegistration {
		sendResponse(w, models.ResponseError, "Server doesn't accept registrations", nil, http.StatusForbidden)
		return
	}

	var request libremotebuild.CredentialsRequest

	if !readRequestLimited(w, r, &request, handlerData.Config.Webserver.MaxRequestBodyLength) {
		return
	}

	if len(request.Password) == 0 || len(request.Username) == 0 {
		sendResponse(w, models.ResponseError, "input missing", nil, http.StatusUnprocessableEntity)
		return
	}

	user := models.User{
		Username: request.Username,
		Password: request.Password,
	}

	err := user.Register(handlerData.Db, handlerData.Config)
	if err == models.ErrorUserAlreadyExists {
		sendResponse(w, models.ResponseError, "User already exists", nil, http.StatusForbidden)
		return
	} else if err != nil {
		return
	}

	sendResponse(w, models.ResponseSuccess, "success", nil, http.StatusOK)
}
