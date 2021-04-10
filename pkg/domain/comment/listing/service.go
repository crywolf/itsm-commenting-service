package listing

import "github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"

// Repository provides access to the comment storage.
type Repository interface {
	// GetComment returns the comment with given ID
	GetComment(id string) (comment.Comment, error)

	// QueryComments enables to call rich queries
	//QueryComments(query string) ([]comment.Comment, error)
}

// Service provides comment listing operations
type Service interface {
	GetComment(id string) (comment.Comment, error)
}

type service struct {
	r Repository
}

// NewService creates a listing service
func NewService(r Repository) Service {
	return &service{r}
}

// GetComment returns the comment with given ID
func (s *service) GetComment(id string) (comment.Comment, error) {
	return s.r.GetComment(id)
}
