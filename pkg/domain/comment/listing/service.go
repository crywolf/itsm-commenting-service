package listing

import "github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"

// Service provides comment listing operations
type Service interface {
	// GetComment returns the comment with given ID
	GetComment(id string) (comment.Comment, error)

	// QueryComments finds documents using a declarative JSON querying syntax
	QueryComments(query map[string]interface{}) (QueryResult, error)
}

// QueryResult wraps the result returned by querying comments
type QueryResult struct {
	Bookmark string                   `json:"bookmark,omitempty"`
	Result   []map[string]interface{} `json:"result"`
}

// Repository provides access to the comment storage.
type Repository interface {
	// GetComment returns the comment with given ID
	GetComment(id string) (comment.Comment, error)

	// QueryComments finds documents using a declarative JSON querying syntax
	QueryComments(query map[string]interface{}) (QueryResult, error)
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

func (s *service) QueryComments(query map[string]interface{}) (QueryResult, error) {
	return s.r.QueryComments(query)
}
