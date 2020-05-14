package models

import "errors"

// ErrorUserAlreadyExists error if user exists
var ErrorUserAlreadyExists = errors.New("user already exists")

// ErrorJobCancelled error if user exists
var ErrorJobCancelled = errors.New("job cancelled")

// ErrorNonZeroExit error if user exists
var ErrorNonZeroExit = errors.New("Non zero exit code from container")

// ErrJobNotRunning if job is not running
var ErrJobNotRunning = errors.New("Job not running")

// ErrNoLogsFound if no logs were found
var ErrNoLogsFound = errors.New("No logs found")
