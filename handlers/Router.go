package handlers

import (
	"fmt"
	"net/http"
	"time"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/JojiiOfficial/Remotebuild/services"
	docker "github.com/fsouza/go-dockerclient"

	"github.com/JojiiOfficial/gaw"
	"github.com/jinzhu/gorm"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Route for REST
type Route struct {
	Name        string
	Method      HTTPMethod
	Pattern     libremotebuild.Endpoint
	HandlerFunc RouteFunction
	HandlerType requestType
}

// HTTPMethod http method. GET, POST, DELETE, HEADER, etc...
type HTTPMethod string

// HTTP methods
const (
	GetMethod    HTTPMethod = "GET"
	POSTMethod   HTTPMethod = "POST"
	PUTMethod    HTTPMethod = "PUT"
	DeleteMethod HTTPMethod = "DELETE"
)

type requestType uint8

const (
	defaultRequest requestType = iota
	sessionRequest
	optionalTokenRequest
)

// Routes all REST routes
type Routes []Route

// RouteFunction function for handling a route
type RouteFunction func(HandlerData, http.ResponseWriter, *http.Request)

// Routes
var (
	routes = Routes{
		// Ping
		Route{
			Name:        "ping",
			Pattern:     libremotebuild.EPPing,
			Method:      POSTMethod,
			HandlerFunc: Ping,
			HandlerType: defaultRequest,
		},

		// User
		Route{
			Name:        "login",
			Pattern:     libremotebuild.EPLogin,
			Method:      POSTMethod,
			HandlerFunc: Login,
			HandlerType: defaultRequest,
		},
		Route{
			Name:        "register",
			Pattern:     libremotebuild.EPRegister,
			Method:      POSTMethod,
			HandlerFunc: Register,
			HandlerType: defaultRequest,
		},

		// Job
		Route{
			Name:        "AddJob",
			Pattern:     libremotebuild.EPJobAdd,
			Method:      PUTMethod,
			HandlerFunc: addJob,
			HandlerType: sessionRequest,
		},
	}
)

// NewRouter create new router
func NewRouter(config *models.Config, db *gorm.DB, jobService *services.JobService, dockerClient *docker.Client) *mux.Router {
	handlerData := HandlerData{
		Config:       config,
		Db:           db,
		JobService:   jobService,
		DockerClient: dockerClient,
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(string(route.Method)).
			Path(string(route.Pattern)).
			Name(route.Name).
			Handler(RouteHandler(route.HandlerType, &handlerData, route.HandlerFunc, route.Name))
	}

	return router
}

// RouteHandler logs stuff
func RouteHandler(requestType requestType, handlerData *HandlerData, inner RouteFunction, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := r.Body.Close()
			if err != nil {
				log.Info(err)
			}
		}()

		needDebug := len(name) > 0

		if needDebug {
			log.Infof("[%s] %s\n", r.Method, name)
		}

		start := time.Now()

		if validateHeader(handlerData.Config, w, r) {
			return
		}

		// Validate request by requestType
		if !requestType.validate(handlerData, r, w) {
			return
		}

		// Process request
		inner(*handlerData, w, r)

		// Print duration of processing
		if needDebug {
			printProcessingDuration(start)
		}
	})
}

// Return false on error
func (requestType requestType) validate(handlerData *HandlerData, r *http.Request, w http.ResponseWriter) bool {
	switch requestType {
	case sessionRequest:
		{
			authHandler := NewAuthHandler(r)
			if len(authHandler.GetBearer()) != 64 {
				log.Error("Invalid token len %d", len(authHandler.GetBearer()))
				sendResponse(w, models.ResponseError, "Invalid token", http.StatusUnauthorized)
				return false
			}

			user, err := models.GetUserFromSession(handlerData.Db, authHandler.GetBearer())
			if LogError(err) || user == nil {
				if user == nil && err == nil {
					log.Error("Can't get user")
				}

				sendResponse(w, models.ResponseError, "Invalid token", http.StatusUnauthorized)
				return false
			}

			handlerData.User = user
		}
	}

	return true
}

// Prints the duration of handling the function
func printProcessingDuration(startTime time.Time) {
	dur := time.Since(startTime)

	if dur < 1500*time.Millisecond {
		log.Debugf("Duration: %s\n", dur.String())
	} else if dur > 1500*time.Millisecond {
		log.Warningf("Duration: %s\n", dur.String())
	}
}

// Return true on error
func validateHeader(config *models.Config, w http.ResponseWriter, r *http.Request) bool {
	headerSize := gaw.GetHeaderSize(r.Header)

	// Send error if header are too big. MaxHeaderLength is stored in b
	if headerSize > uint32(config.Webserver.MaxHeaderLength) {
		// Send error response
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		fmt.Fprint(w, "413 request too large")

		log.Warnf("Got request with %db headers. Maximum allowed are %db\n", headerSize, config.Webserver.MaxHeaderLength)
		return true
	}

	return false
}
