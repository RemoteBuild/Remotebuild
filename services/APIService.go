package services

import (
	"net/http"
	"time"

	"github.com/JojiiOfficial/Remotebuild/models"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

//APIService the service handling the API
type APIService struct {
	router        *mux.Router
	config        *models.Config
	HTTPServer    *http.Server
	HTTPTLSServer *http.Server
}

//NewAPIService create new API service
func NewAPIService(config *models.Config, getRouter func() *mux.Router) *APIService {
	router := getRouter()

	var httpServer, httpsServer *http.Server

	//Init http server
	if config.Webserver.HTTP.Enabled {
		httpServer = &http.Server{
			Handler:           router,
			Addr:              config.Webserver.HTTP.ListenAddress,
			ReadHeaderTimeout: config.Webserver.ReadTimeout,
			WriteTimeout:      config.Webserver.WriteTimeout,
			IdleTimeout:       0 * time.Second,
		}
	}

	//Init https server
	if config.Webserver.HTTPS.Enabled {
		httpsServer = &http.Server{
			Handler:      router,
			Addr:         config.Webserver.HTTPS.ListenAddress,
			ReadTimeout:  config.Webserver.ReadTimeout,
			WriteTimeout: config.Webserver.WriteTimeout,
			IdleTimeout:  0 * time.Second,
		}
	}

	apiService := &APIService{
		config:        config,
		router:        router,
		HTTPServer:    httpServer,
		HTTPTLSServer: httpsServer,
	}

	return apiService
}

//Start the API service
func (service *APIService) Start() {
	//Start HTTPS if enabled
	if service.HTTPTLSServer != nil {
		log.Infof("Server started TLS on port (%s)\n", service.config.Webserver.HTTPS.ListenAddress)
		go (func() {
			if err := service.HTTPTLSServer.ListenAndServeTLS(service.config.Webserver.HTTPS.CertFile, service.config.Webserver.HTTPS.KeyFile); err != nil {
				if err != http.ErrServerClosed {
					log.Fatal(err)
				}
			}
		})()
	}

	//Start HTTP if enabled
	if service.HTTPServer != nil {
		log.Infof("Server started HTTP on port (%s)\n", service.config.Webserver.HTTP.ListenAddress)
		go (func() {
			if err := service.HTTPServer.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					log.Fatal(err)
				}
			}
		})()
	}
}
