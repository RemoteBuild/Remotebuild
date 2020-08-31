package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/handlers"
	"github.com/JojiiOfficial/Remotebuild/services"
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

// Services
var (
	apiService       *services.APIService       // Handle endpoints
	jobService       *services.JobService       // Handle Jobs
	cleanupService   *services.CleanupService   // Cleanup db stuff
	containerService *services.ContainerService // Managing containers
)

func startAPI() {
	log.Info("Starting version " + version)

	// Create new container service
	containerService = services.NewContainerService(config)

	// Create and start the jobservice
	jobService = services.NewJobService(config, db, func(jobType libremotebuild.JobType) (string, error) {
		return containerService.GetContainer(jobType)
	})
	jobService.Start()

	// Create and start required services
	apiService = services.NewAPIService(config, func() *mux.Router {
		return handlers.NewRouter(config, db, jobService)
	})
	apiService.Start()

	// Create cleanup service
	cleanupService = services.NewClienupService(config, db)
	cleanupService.Start()

	// Startup done
	log.Info("Startup completed")

	awaitExit(apiService, db)
}

// Shutdown server gracefully
func awaitExit(httpServer *services.APIService, db *gorm.DB) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	// await os signal
	<-signalChan

	// Stop all jobs
	jobService.Stop()

	// Create a deadline for the await
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	log.Info("Shutting down server")

	if httpServer.HTTPServer != nil {
		err := httpServer.HTTPServer.Shutdown(ctx)
		if err != nil {
			log.Warn(err)
		}

		log.Info("HTTP server shutdown complete")
	}

	if httpServer.HTTPTLSServer != nil {
		err := httpServer.HTTPTLSServer.Shutdown(ctx)
		if err != nil {
			log.Warn(err)
		}

		log.Info("HTTPs server shutdown complete")
	}

	log.Info("Shutting down complete")
	os.Exit(0)
}
