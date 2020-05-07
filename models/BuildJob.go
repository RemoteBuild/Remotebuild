package models

import (
	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// BuildJob a job which builds a package
type BuildJob struct {
	*docker.Client `gorm:"-"`
	gorm.Model
	State libremotebuild.JobState // Build state
	Type  libremotebuild.JobType  // Type of job

	Image string // Dockerimage to run

	cancel      chan bool `gorm:"-"` // Cancel chan
	containerID string    `gorm:"-"`
}

// BuildResult result of a bulid
type BuildResult struct {
	NewBinary string
	Error     error
}

// NewBuildJob create new BuildJob
func NewBuildJob(db *gorm.DB, buildJob BuildJob) (*BuildJob, error) {
	buildJob.State = libremotebuild.JobWaiting
	buildJob.cancel = make(chan bool, 1)

	// Connect to docker
	if err := buildJob.connectDocker(); err != nil {
		return nil, err
	}

	// Save Job to Db
	err := db.Create(&buildJob).Error
	if err != nil {
		return nil, err
	}

	return &buildJob, nil
}

// Init buildJob
func (buildJob *BuildJob) Init() error {
	// Init channel
	if buildJob.cancel == nil {
		buildJob.cancel = make(chan bool, 1)
	}

	// Connect to docker
	return buildJob.connectDocker()
}

// Run a buildjob (start but await)
func (buildJob *BuildJob) Run(dataDir string, args map[string]string) *BuildResult {
	// Init buildJob
	if err := buildJob.Init(); err != nil {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{
			Error: err,
		}
	}

	log.Debug("Run BuildJob ", buildJob.ID)
	buildJob.State = libremotebuild.JobRunning

	buildDone := make(chan bool, 1)
	var result *BuildResult

	// Run build in goroutine
	go func() {
		result = buildJob.build(dataDir, args)
		buildDone <- true
	}()

	// Await build or cancel
	select {
	case <-buildDone:
		// On done
		return result
	case <-buildJob.cancel:
		// On cancel
		buildJob.Stop()
		buildJob.State = libremotebuild.JobCancelled
		return &BuildResult{
			Error: ErrorJobCancelled,
		}
	}
}

// Connect to dockerClient
func (buildJob *BuildJob) connectDocker() error {
	// Skip if already connected
	if buildJob.Client != nil {
		return nil
	}

	// Connect
	var err error
	buildJob.Client, err = docker.NewClientFromEnv()
	return err
}

func (buildJob *BuildJob) build(dataDir string, args map[string]string) *BuildResult {
	// Pull image if neccessary
	if err := buildJob.pullImageIfNeeded(buildJob.Image); err != nil {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{
			Error: err,
		}
	}

	// Create container
	container, err := buildJob.getContainer(dataDir, args)
	if err != nil {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{
			Error: err,
		}
	}

	// Start container
	if err = buildJob.StartContainer(container.ID, &docker.HostConfig{}); err != nil {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{
			Error: err,
		}
	}

	// Wait until building is done
	n, err := buildJob.WaitContainer(container.ID)
	if err != nil {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{
			Error: err,
		}
	}

	// Check container exit code
	if n != 0 {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{
			Error: ErrorNonZeroExit,
		}
	}

	// Set done
	buildJob.State = libremotebuild.JobDone
	return &BuildResult{
		Error: nil,
	}
}

// Create build container
func (buildJob *BuildJob) getContainer(dataDir string, args map[string]string) (*docker.Container, error) {
	// Create container
	container, err := buildJob.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: buildJob.Image,
			Env:   argsToEnvs(args),
		},
		HostConfig: &docker.HostConfig{
			// Mount /home/builduser on host /tmp/remotebuild_XXXXXXXXXX
			Mounts: []docker.HostMount{{
				Source: dataDir,
				BindOptions: &docker.BindOptions{
					Propagation: "rprivate",
				},
				ReadOnly: false,
				Type:     "bind",
				Target:   "/home/builduser",
			}},
			// Autodelete container afterwards
			AutoRemove: true,
		},
	})

	if err == nil {
		buildJob.containerID = container.ID
	}

	return container, err
}

func (buildJob *BuildJob) hasImage(image string) (bool, error) {
	// Get all images
	images, err := buildJob.Client.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		return false, err
	}

	// Search tag in available images/tags
	for i := range images {
		for _, tag := range images[i].RepoTags {
			if image == tag {
				return true, nil
			}
		}
	}

	return false, nil
}

func (buildJob *BuildJob) pullImageIfNeeded(image string) error {
	// Check if image is present
	hasImage, err := buildJob.hasImage(image)
	if err != nil || hasImage {
		return err
	}

	log.Debug("Pulling Image ", image)

	// Pull image
	err = buildJob.PullImage(docker.PullImageOptions{
		Registry:   "docker.io",
		Repository: image,
	}, docker.AuthConfiguration{
		ServerAddress: "docker.io",
	})

	if err == nil {
		log.Debug("Successful pulled Image ", image)
	}

	return err
}

// Stop building
func (buildJob *BuildJob) Stop() {
	if len(buildJob.containerID) > 0 && buildJob.Client != nil {
		log.Info("Stopping container ", buildJob.containerID)
		buildJob.StopContainer(buildJob.containerID, 1)
	}
}
