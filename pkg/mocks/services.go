package mocks

import (
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment/listing"
	"github.com/stretchr/testify/mock"
)

// ListingMock is a mock of listing service
type ListingMock struct {
	mock.Mock
}

// GetComment returns the comment with given ID
func (l *ListingMock) GetComment(id string) (comment.Comment, error) {
	args := l.Called(id)
	return args.Get(0).(comment.Comment), args.Error(1)
}

// QueryComments finds documents using a declarative JSON querying syntax
func (l *ListingMock) QueryComments(query map[string]interface{}) (listing.QueryResult, error) {
	args := l.Called(query)
	return args.Get(0).(listing.QueryResult), args.Error(1)
}

// AddingMock is a mock of adding service
type AddingMock struct {
	mock.Mock
}

// AddComment saves a given comment to the repository
func (a *AddingMock) AddComment(c comment.Comment) (string, error) {
	args := a.Called(c)
	return args.String(0), args.Error(1)
}