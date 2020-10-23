package services

import (
	"errors"

	libremotebuild "github.com/RemoteBuild/LibRemotebuild"
	"github.com/RemoteBuild/Remotebuild/models"
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
	// TODO implement auto upgrades for container

	image, has := cs.Config.GetImage(job)
	if !has {
		return "", errors.New("Image not found")
	}

	return image, nil
}
