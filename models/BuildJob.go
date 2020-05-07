package models

// BuildJob a job which builds a package
type BuildJob struct {
	State JobState // Build state

	Image string            // Dockerimage to run
	Args  map[string]string // Envars for Dockerimage
}
