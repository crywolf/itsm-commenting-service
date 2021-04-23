package adding

import "github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"

// Service provides comment adding operations
type Service interface {
	AddComment(c comment.Comment, channelID string) (id string, err error)
}

// Repository provides access to comments repository
type Repository interface {
	// AddComment saves a given comment to the repository
	AddComment(c comment.Comment, channelID string) (id string, err error)
}

// NewService creates an adding service
func NewService(r Repository) Service {
	return &service{r}
}

type service struct {
	r Repository
}

// AddComment persists the given comment to storage
func (s *service) AddComment(c comment.Comment, channelID string) (string, error) {
	return s.r.AddComment(c, channelID)
}
