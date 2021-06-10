package e2e

import (
	"errors"
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var mockUserData = user.BasicInfo{
	UUID:           "2af4f493-0bd5-4513-b440-6cbb465feadb",
	Name:           "Alfred",
	Surname:        "Koletschko",
	OrgName:        "cc4c7533-4e34-4890-a79c-c1fda3c1be1e.kompitech.com",
	OrgDisplayName: "KompiTech",
}

var expectedMockUserJSON = `{
	"name": "Alfred",
	"surname": "Koletschko",
	"uuid": "2af4f493-0bd5-4513-b440-6cbb465feadb",
	"org_display_name": "KompiTech",
	"org_name": "cc4c7533-4e34-4890-a79c-c1fda3c1be1e.kompitech.com"
}`

// AuthServiceStub to simulate authorization service
type AuthServiceStub struct{}

// Enforce returns true if action is allowed to be performed on specified asset
func (s *AuthServiceStub) Enforce(assetType string, act auth.Action, channelID, authToken string) (bool, error) {
	if authToken == "" {
		return false, errors.New("authorization service failed - missing authorization token")
	}
	return true, nil
}

// UserServiceStub to simulate user service
type UserServiceStub struct{}

// UserBasicInfo returns info about user who initiated the request
func (s *UserServiceStub) UserBasicInfo(r *http.Request) (user.BasicInfo, error) {
	if r.Header.Get("authorization") == "" {
		return user.BasicInfo{}, status.Error(codes.Unauthenticated, "user service failed - missing authorization token")
	}
	return mockUserData, nil
}
