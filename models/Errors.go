package models

import "errors"

// ErrorUserAlreadyExists error if user exists
var ErrorUserAlreadyExists = errors.New("user already exists")

// ErrorJobCancelled error if user exists
var ErrorJobCancelled = errors.New("job cancelled")

// ErrorNonZeroExit error if user exists
var ErrorNonZeroExit = errors.New("Non zero exit code from container")
