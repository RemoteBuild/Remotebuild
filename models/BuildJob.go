package models

import (
	"fmt"
	"io"
	"time"

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

	Image     string // Dockerimage to run
	UseCcache bool   // use ccahe to improve build speed

	cancelChan  chan bool `gorm:"-"` // Cancel chan
	ContainerID string    `gorm:"-"`
	Config      *Config   `gorm:"-"`
}

// BuildResult result of a bulid
type BuildResult struct {
	resinfo *ResInfo
	Error   error
}

// NewBuildJob create new BuildJob
func NewBuildJob(db *gorm.DB, config *Config, buildJob BuildJob, image string, useCcache bool) (*BuildJob, error) {
	buildJob.State = libremotebuild.JobWaiting
	buildJob.Image = image
	buildJob.Config = config
	buildJob.UseCcache = useCcache && config.IsCcacheDirValid()
	buildJob.cancelChan = make(chan bool, 1)

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
	if buildJob.cancelChan == nil {
		buildJob.cancelChan = make(chan bool, 1)
	}

	// Connect to docker
	return buildJob.connectDocker()
}

// Run a buildjob (start but await)
func (buildJob *BuildJob) Run(dataDir string, argParser *ArgParser) (*BuildResult, *time.Duration) {
	// Init buildJob
	if err := buildJob.Init(); err != nil {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{Error: err}, nil
	}

	log.Debug("Run BuildJob ", buildJob.ID)
	buildJob.State = libremotebuild.JobRunning

	buildDone := make(chan bool, 1)
	var result *BuildResult
	var duration *time.Duration

	// Run build
	go func() {
		result, duration = buildJob.build(dataDir, argParser)
		buildDone <- true
	}()

	// Await build or cancel
	select {
	case <-buildDone:
		// On done
		return result, duration
	case <-buildJob.cancelChan:
		// On cancel
		buildJob.Stop()
		buildJob.State = libremotebuild.JobCancelled
		return &BuildResult{Error: ErrorJobCancelled}, duration
	}
}

// Save Buildjob
func (buildJob *BuildJob) Save(db *gorm.DB) error {
	return db.Save(buildJob).Error
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

// Build the package
func (buildJob *BuildJob) build(dataDir string, argParser *ArgParser) (*BuildResult, *time.Duration) {
	// Parse args
	envars, err := argParser.ParseEnvars()
	if err != nil {
		return &BuildResult{Error: err}, nil
	}

	// Pull image if neccessary
	if err := buildJob.pullImage(buildJob.Image); err != nil {
		return &BuildResult{Error: err}, nil
	}

	// Create container
	container, err := buildJob.getContainer(dataDir, envars)
	if err != nil {
		return &BuildResult{Error: err}, nil
	}

	start := time.Now()

	// Start container
	if err = buildJob.StartContainer(container.ID, &docker.HostConfig{}); err != nil {
		return &BuildResult{Error: err}, nil
	}

	// Wait until building is done
	n, err := buildJob.WaitContainer(container.ID)
	if err != nil {
		return &BuildResult{Error: err}, nil
	}

	duration := time.Since(start)

	// No Container should be assigned
	// to this job anymore
	buildJob.ContainerID = ""

	// Check container exit code
	if n != 0 {
		return &BuildResult{Error: ErrorNonZeroExit}, &duration
	}

	resInfo, err := ParseResInfo(dataDir, GetResInfoPath(dataDir))
	if err != nil || resInfo == nil {
		return &BuildResult{Error: err}, &duration
	}

	// Set done
	buildJob.State = libremotebuild.JobDone
	return &BuildResult{
		Error:   nil,
		resinfo: resInfo,
	}, &duration
}

func (buildJob *BuildJob) getContainer(dataDir string, env []string) (*docker.Container, error) {
	// Set CCACHE environment variables
	if buildJob.UseCcache {
		env = append(env, "USE_CCACHE=true")
		env = append(env, "CCACHE_DIR=/ccache")
		env = append(env, fmt.Sprintf("CCACHE_MAXSIZE=%dG", buildJob.Config.Server.Ccache.MaxSize))
	}

	// Mount /home/builduser on host /tmp/remotebuild_XXXXXXXXXX
	mounts := []docker.HostMount{{
		Source: dataDir,
		Target: "/home/builduser",
		BindOptions: &docker.BindOptions{
			Propagation: "rprivate",
		},
		ReadOnly: false,
		Type:     "bind",
	}}

	// Monut host ccache dir if ccache is used
	if buildJob.UseCcache {
		mounts = append(mounts, docker.HostMount{
			Source:   buildJob.Config.Server.Ccache.Dir,
			Target:   "/ccache",
			Type:     "bind",
			ReadOnly: false,
			BindOptions: &docker.BindOptions{
				Propagation: "rprivate",
			},
		})
	}

	// Create container
	container, err := buildJob.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: buildJob.Image,
			Env:   env,
		},
		HostConfig: &docker.HostConfig{
			Mounts: mounts,
			// Autodelete container afterwards
			AutoRemove: !buildJob.Config.Server.KeepBuildContainer,
		},
	})

	if err == nil {
		buildJob.ContainerID = container.ID
	}

	return container, err
}

func (buildJob *BuildJob) hasImage(image string) (bool, error) {
	// Get all images
	images, err := buildJob.ListImages(docker.ListImagesOptions{All: false})
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

func (buildJob *BuildJob) pullImage(image string) error {
	// Check if image is present
	hasImage, err := buildJob.hasImage(image)
	if err != nil || hasImage {
		return err
	}

	log.Debug("Pulling Image ", image)

	// Actually pull the image
	err = buildJob.PullImage(docker.PullImageOptions{
		Registry:   "docker.io",
		Repository: image,
	}, docker.AuthConfiguration{
		ServerAddress: "docker.io",
	})

	if err == nil {
		log.Debug("Successfully pulled Image ", image)
	}

	return err
}

// Pause buildjob
func (buildJob *BuildJob) Pause() error {
	if len(buildJob.ContainerID) > 0 && buildJob.State != libremotebuild.JobPaused {
		err := buildJob.PauseContainer(buildJob.ContainerID)
		if err != nil {
			return err
		}

		buildJob.State = libremotebuild.JobPaused
	}

	return nil
}

// Resume buildjob
func (buildJob *BuildJob) Resume() error {
	if len(buildJob.ContainerID) > 0 && buildJob.State == libremotebuild.JobPaused {
		err := buildJob.Client.UnpauseContainer(buildJob.ContainerID)
		if err != nil {
			return err
		}

		buildJob.State = libremotebuild.JobRunning
	}

	return nil
}

// Stop building
func (buildJob *BuildJob) Stop() {
	if len(buildJob.ContainerID) > 0 && buildJob.Client != nil {
		log.Info("Stopping container ", buildJob.ContainerID)
		buildJob.StopContainer(buildJob.ContainerID, 1)
	}
}

// Cancel a buildJob
func (buildJob *BuildJob) cancel() {
	if buildJob.State == libremotebuild.JobRunning {
		buildJob.cancelChan <- true
		buildJob.Stop()
	}

	buildJob.State = libremotebuild.JobCancelled
}

// GetLogs of Buildjob
func (buildJob *BuildJob) GetLogs(since int64, w io.Writer, tail string) error {
	// Check build is running
	if buildJob.State != libremotebuild.JobRunning || len(buildJob.ContainerID) == 0 {
		return ErrJobNotRunning
	}

	// Build options
	logOptions := docker.LogsOptions{
		Container:    buildJob.ContainerID,
		Stderr:       true,
		Stdout:       true,
		Follow:       false,
		Since:        since,
		OutputStream: w,
		ErrorStream:  w,
	}

	if len(tail) > 0 {
		logOptions.Tail = tail
	}

	// Get container logs
	return buildJob.Logs(logOptions)
}
