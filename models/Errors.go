package models

import "errors"

//ErrorUserAlreadyExists error if user exists
var ErrorUserAlreadyExists = errors.New("user already exists")
