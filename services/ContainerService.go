package services

import (
	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	"github.com/JojiiOfficial/Remotebuild/models"
)

// ContainerService provides management for
// container versions and availability
type ContainerService struct {
	Config *models.Config
}

// NewContainerService create a new container Service
func NewContainerService(config *models.Config) *ContainerService {
	return &ContainerService{
		Config: config,
	}
}

// GetContainer for a job
func (cs *ContainerService) GetContainer(job libremotebuild.JobType) (string, error) {

	return "", nil
}
