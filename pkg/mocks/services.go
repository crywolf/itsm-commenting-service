package mocks

import (
	"context"
	"net/http"

	"github.com/KompiTech/go-toolkit/natswatcher"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/user"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/http/rest/auth"
	"github.com/stretchr/testify/mock"
)

// ListingMock is a mock of listing service
type ListingMock struct {
	mock.Mock
}

// GetComment returns the comment with given ID
func (l *ListingMock) GetComment(ctx context.Context, id, channelID, assetType string) (comment.Comment, error) {
	args := l.Called(id, channelID, assetType)
	return args.Get(0).(comment.Comment), args.Error(1)
}

// QueryComments finds documents using a declarative JSON querying syntax
func (l *ListingMock) QueryComments(ctx context.Context, query map[string]interface{}, channelID, assetType string) (listing.QueryResult, error) {
	args := l.Called(query, channelID, assetType)
	return args.Get(0).(listing.QueryResult), args.Error(1)
}

// AddingMock is a mock of adding service
type AddingMock struct {
	mock.Mock
}

// AddComment saves a given comment to the repository
func (a *AddingMock) AddComment(ctx context.Context, c comment.Comment, channelID, assetType string, origin string) (*comment.Comment, error) {
	args := a.Called(c, channelID, assetType)
	return &comment.Comment{UUID: args.String(0)}, args.Error(1)
}

// UpdatingMock is a mock of adding service
type UpdatingMock struct {
	mock.Mock
}

// MarkAsReadByUser adds user info to read_by array in the comment in the storage
func (u *UpdatingMock) MarkAsReadByUser(ctx context.Context, id string, readBy comment.ReadBy, channelID, assetType string) (bool, error) {
	args := u.Called(id, readBy, channelID, assetType)
	return args.Bool(0), args.Error(1)
}

// AuthServiceMock is a mock of authentication service
type AuthServiceMock struct {
	mock.Mock
}

// Enforce returns true if action is allowed to be performed on specified asset
func (s *AuthServiceMock) Enforce(assetType string, action auth.Action, channelID, authToken string) (bool, error) {
	args := s.Called(assetType, action, channelID, authToken)
	return args.Bool(0), args.Error(1)
}

// UserServiceMock is a mock of user service
type UserServiceMock struct {
	mock.Mock
}

// UserBasicInfo returns info about user who initiated the request
func (s *UserServiceMock) UserBasicInfo(r *http.Request) (user.BasicInfo, error) {
	args := s.Called(r)
	return args.Get(0).(user.BasicInfo), args.Error(1)
}

// ValidatorMock is a mock of validation service
type ValidatorMock struct {
	mock.Mock
}

// Validate returns error if comment is not valid
func (s *ValidatorMock) Validate(c comment.Comment) error {
	args := s.Called(c)
	return args.Error(0)
}

// PayloadValidatorMock is a mock of payload validation service
type PayloadValidatorMock struct {
	mock.Mock
}

// ValidatePayload returns error if payload is not valid
func (s *PayloadValidatorMock) ValidatePayload(p []byte, schemaFile string) error {
	args := s.Called(p, schemaFile)
	return args.Error(0)
}

// NATSClientMock is a mock of NATS queue client
type NATSClientMock struct {
	mock.Mock
}

// Publish publishes messages to NATS queue
func (c *NATSClientMock) Publish(msgs ...natswatcher.Message) error {
	args := c.Called(msgs)
	return args.Error(0)
}

// EventServiceMock is a mock of event service
type EventServiceMock struct {
	mock.Mock
}

// NewQueue creates new event queue
func (s *EventServiceMock) NewQueue(channelID, orgID event.UUID) (event.Queue, error) {
	args := s.Called(channelID, orgID)
	return args.Get(0).(event.Queue), args.Error(1)
}

// QueueMock is a mock of event queue
type QueueMock struct {
	mock.Mock
}

// AddCreateEvent prepares new event of type CREATE
func (q *QueueMock) AddCreateEvent(c comment.Comment, assetType string, origin string) error {
	args := q.Called(c, assetType)
	return args.Error(0)
}

// PublishEvents publishes all prepared events not published yet
func (q *QueueMock) PublishEvents() error {
	args := q.Called()
	return args.Error(0)
}
