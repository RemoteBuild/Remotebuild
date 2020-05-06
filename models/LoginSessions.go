package models

import (
	"github.com/JojiiOfficial/gaw"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

//LoginSession session for loggedin user
type LoginSession struct {
	gorm.Model
	User      *User `gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID    uint
	Token     string
	Requests  int64
	MachineID string
}

// SessionTokenLength length of session token
const SessionTokenLength = 64

//GetUserFromSession return user from session
func GetUserFromSession(db *gorm.DB, token string) (*User, error) {
	var session LoginSession
	err := db.Model(&LoginSession{}).Where(&LoginSession{
		Token: token,
	}).Preload("User").Preload("User.Role").Find(&session).Error

	if err != nil {
		return nil, err
	}

	// Increase request counter
	session.Requests++
	err = db.Save(&session).Error
	if err != nil {
		log.Error(err)
	}

	return session.User, nil
}

// NewSession create new login session
func NewSession(user *User, machineID string) *LoginSession {
	if len(machineID) > 100 {
		machineID = ""
	}

	var token string
	var err error
	tries := 0

	// Try to generate a token
	for tries < 5 {
		token, err = gaw.GenRandString(SessionTokenLength)
		if err != nil {
			log.Error(err)
		} else {
			break
		}
		tries++
	}

	// Return nil if token can't be generated
	if len(token) != 64 {
		return nil
	}

	//Generate session
	return &LoginSession{
		Token:     token,
		UserID:    user.ID,
		User:      user,
		MachineID: machineID,
	}
}
