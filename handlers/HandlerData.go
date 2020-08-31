package handlers

import (
	"github.com/JojiiOfficial/Remotebuild/models"
	"github.com/JojiiOfficial/Remotebuild/services"
	docker "github.com/fsouza/go-dockerclient"
	"gorm.io/gorm"
)

//HandlerData handlerData for web
type HandlerData struct {
	Config       *models.Config
	Db           *gorm.DB
	User         *models.User
	JobService   *services.JobService
	DockerClient *docker.Client
}
