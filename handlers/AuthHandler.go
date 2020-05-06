package handlers

import (
	"errors"
	"net/http"
	"strings"
)

var (
	//ErrorTokenInvalid error if token is invalid
	ErrorTokenInvalid error = errors.New("Token invalid")
	//ErrorTokenEmpty error if token is empty
	ErrorTokenEmpty error = errors.New("Token empty")
)

//AuthHandler handler for http auth
type AuthHandler struct {
	Request *http.Request
}

//NewAuthHandler returns a new AuthHandler
func NewAuthHandler(request *http.Request) *AuthHandler {
	return &AuthHandler{
		Request: request,
	}
}

//GetBearer return the bearer token
func (authHandler AuthHandler) GetBearer() string {
	authHeader, has := authHandler.Request.Header["Authorization"]
	//Validate bearer token
	if !has || len(authHeader) == 0 || !strings.HasPrefix(authHeader[0], "Bearer") {
		return ""
	}
	return tokenFromBearerHeader(authHeader[0])
}

func tokenFromBearerHeader(header string) string {
	return strings.TrimSpace(strings.ReplaceAll(header, "Bearer", ""))
}

//IsInvalid return true if err is invalid
func (authHandler AuthHandler) IsInvalid(err error) bool {
	return err == ErrorTokenInvalid
}
