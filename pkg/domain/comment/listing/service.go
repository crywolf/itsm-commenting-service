package listing

import (
	"context"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
)

// Service provides comment listing operations
type Service interface {
	// GetComment returns the comment with given ID from the repository
	GetComment(ctx context.Context, id, channelID string, assetType comment.AssetType) (comment.Comment, error)

	// QueryComments finds documents in the repository using a declarative JSON querying syntax
	QueryComments(ctx context.Context, query map[string]interface{}, channelID string, assetType comment.AssetType) (QueryResult, error)
}

// QueryResult wraps the result returned by querying comments
type QueryResult struct {
	Bookmark string                   `json:"bookmark,omitempty"`
	Result   []map[string]interface{} `json:"result"`
}

// Repository provides reading access to the comments repository
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

func (s *service) GetComment(ctx context.Context, id, channelID string, assetType comment.AssetType) (comment.Comment, error) {
	return s.r.GetComment(ctx, id, channelID, assetType)
}

func (s *service) QueryComments(ctx context.Context, query map[string]interface{}, channelID string, assetType comment.AssetType) (QueryResult, error) {
	return s.r.QueryComments(ctx, query, channelID, assetType)
}
