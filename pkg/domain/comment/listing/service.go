package listing

import (
	"context"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
)

// Service provides comment listing operations
type Service interface {
	// GetComment returns the comment with given ID
	GetComment(ctx context.Context, id, channelID string, assetType comment.AssetType) (comment.Comment, error)

	// QueryComments finds documents using a declarative JSON querying syntax
	QueryComments(ctx context.Context, query map[string]interface{}, channelID string, assetType comment.AssetType) (QueryResult, error)
}

// QueryResult wraps the result returned by querying comments
type QueryResult struct {
	Bookmark string                   `json:"bookmark,omitempty"`
	Result   []map[string]interface{} `json:"result"`
}

// Repository provides access to the comment storage
type Repository interface {
	// GetComment returns the comment with given ID
	GetComment(ctx context.Context, id, channelID string, assetType comment.AssetType) (comment.Comment, error)

	// QueryComments finds documents using a declarative JSON querying syntax
	QueryComments(ctx context.Context, query map[string]interface{}, channelID string, assetType comment.AssetType) (QueryResult, error)
}

// NewService creates a listing service
func NewService(r Repository) Service {
	return &service{r}
}

type service struct {
	r Repository
}

// GetComment returns the comment with given ID
func (s *service) GetComment(ctx context.Context, id, channelID string, assetType comment.AssetType) (comment.Comment, error) {
	return s.r.GetComment(ctx, id, channelID, assetType)
}

// QueryComments finds documents using a declarative JSON querying syntax
func (s *service) QueryComments(ctx context.Context, query map[string]interface{}, channelID string, assetType comment.AssetType) (QueryResult, error) {
	return s.r.QueryComments(ctx, query, channelID, assetType)
}
