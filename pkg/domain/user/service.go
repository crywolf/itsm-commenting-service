package user

import (
	"net/http"
)

// Service provides basic info about user
type Service interface {
	UserData(r *http.Request) (InvokingUserData, error)
}

// NewService creates user service
func NewService() Service {
	return &service{}
}

// InvokingUserData represents the user that invoked the HTTP request
type InvokingUserData struct {
	UUID string
	Name string
}

// service calls external user service that provides basic info about user
type service struct {
	//client
}

// UserData returns info about user who initiated the request
func (s service) UserData(r *http.Request) (InvokingUserData, error) {
	// TODO fetch real user from some user service
	userData := InvokingUserData{
		UUID: "2af4f493-0bd5-4513-b440-6cbb465feadb",
		Name: "Some test user 1",
	}

	return userData, nil
}
