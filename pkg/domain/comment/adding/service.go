package adding

import (
	"context"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
)

// Service provides comment adding operations
type Service interface {
	AddComment(ctx context.Context, c comment.Comment, channelID, assetType string) (id string, err error)
}

// Repository provides access to comments storage
type Repository interface {
	// AddComment saves a given comment to the repository
	AddComment(ctx context.Context, c comment.Comment, channelID, assetType string) (id string, err error)
}

// NewService creates an adding service
func NewService(r Repository) Service {
	return &service{r}
}

type service struct {
	r Repository
}

// AddComment persists the given comment to storage
func (s *service) AddComment(ctx context.Context, c comment.Comment, channelID, assetType string) (string, error) {
	return s.r.AddComment(ctx, c, channelID, assetType)
}
