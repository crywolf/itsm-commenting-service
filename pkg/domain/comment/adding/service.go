package adding

import "github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"

// Service provides beer adding operations
type Service interface {
	AddComment(c comment.Comment) (id string, err error)
}

// Repository provides access to comments repository
type Repository interface {
	// AddComment saves a given comment to the repository
	AddComment(comment.Comment) (id string, err error)
}

type service struct {
	r Repository
}

// NewService creates an adding service
func NewService(r Repository) Service {
	return &service{r}
}

// AddComment persists the given comment to storage
func (s *service) AddComment(c comment.Comment) (string, error) {
	return s.r.AddComment(c)
}
