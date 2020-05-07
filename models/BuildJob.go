package models

// BuildJob a job which builds a package
type BuildJob struct {
	State JobState // Build state
	Type  JobType  // Type of job

	Image string            // Dockerimage to run
	Args  map[string]string // Envars for Dockerimage
}
