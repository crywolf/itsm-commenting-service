package listing

import "github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"

// Service provides comment listing operations
type Service interface {
	GetComment(id string) (comment.Comment, error)
}

// Repository provides access to the comment storage.
type Repository interface {
	// GetComment returns the comment with given ID
	GetComment(id string) (comment.Comment, error)

	// ListComments enables to call rich queries
	//ListComments(query string) ([]comment.Comment, error)
}

// NewService creates a listing service
func NewService(r Repository) Service {
	return &service{r}
}

type service struct {
	r Repository
}

// GetComment returns the comment with given ID
func (s *service) GetComment(id string) (comment.Comment, error) {
	return s.r.GetComment(id)
}
