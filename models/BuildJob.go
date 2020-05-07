package models

import (
	"io/ioutil"
	"path/filepath"
	"strings"

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
	Archive string
	Error   error
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
func (buildJob *BuildJob) Run(dataDir string, argParser *ArgParser) *BuildResult {
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

	// Run build
	go func() {
		result = buildJob.build(dataDir, argParser)
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

func (buildJob *BuildJob) build(dataDir string, argParser *ArgParser) *BuildResult {
	// Parse args
	envars, err := argParser.ParseEnvars()
	if err != nil {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{
			Error: err,
		}
	}

	// Pull image if neccessary
	if err := buildJob.pullImageIfNeeded(buildJob.Image); err != nil {
		buildJob.State = libremotebuild.JobFailed
		return &BuildResult{
			Error: err,
		}
	}

	// Create container
	container, err := buildJob.getContainer(dataDir, envars)
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

	// Get Archive
	archive := buildJob.getArchive(dataDir, argParser)

	// If not found serach for it
	if len(archive) == 0 {
		log.Debug("Archive not found. Searching...")

		archive, err = buildJob.findBuiltPackage(dataDir)
		if err != nil {
			buildJob.State = libremotebuild.JobFailed
			return &BuildResult{
				Error: err,
			}
		}
	}

	log.Info("archive: ", archive)

	// Set done
	buildJob.State = libremotebuild.JobDone
	return &BuildResult{
		Error:   nil,
		Archive: archive,
	}
}

func (buildJob *BuildJob) getArchive(dir string, argParser *ArgParser) string {
	var fileName string

	switch buildJob.Type {
	case libremotebuild.JobAUR:
		fileName = argParser.getAURRepoName() + ".pkg.tar.xz"
	}

	if len(fileName) == 0 {
		return ""
	}

	return filepath.Join(dir, fileName)
}

// Return built archive
func (buildJob *BuildJob) findBuiltPackage(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, fileinfo := range files {
		if strings.HasSuffix(fileinfo.Name(), "pkg.tar.xz") {
			return fileinfo.Name(), nil
		}
	}

	return "", nil
}

// Create build container
func (buildJob *BuildJob) getContainer(dataDir string, env []string) (*docker.Container, error) {
	// Create container
	container, err := buildJob.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: buildJob.Image,
			Env:   env,
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
