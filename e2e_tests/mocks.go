package e2e

import (
	"net/http"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
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


// UserServiceStub to simulate of user service
type UserServiceStub struct{}

// UserBasicInfo returns info about user who initiated the request
func (s *UserServiceStub) UserBasicInfo(r *http.Request) (user.BasicInfo, error) {
	if r.Header.Get("authorization") == "" {
		return user.BasicInfo{}, status.Error(codes.Unauthenticated, "authorization failed")
	}
	return mockUserData, nil
}

// EventServiceStub to simulate Event service
type EventServiceStub struct{}

// EventQueueStub to simulate Event queue
type EventQueueStub struct{}

// NewQueue creates new event queue
func (s *EventServiceStub) NewQueue(_, _ event.UUID) (event.Queue, error) {
	return &EventQueueStub{}, nil
}

// AddCreateEvent prepares new event of type CREATE
func (q *EventQueueStub) AddCreateEvent(_ comment.Comment, _ string) error {
	return nil
}

// PublishEvents publishes all prepared events not published yet
func (q *EventQueueStub) PublishEvents() error {
	return nil
}
