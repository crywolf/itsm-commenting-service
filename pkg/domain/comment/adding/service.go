package adding

import (
	"context"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
)

// Service provides comment adding operations
type Service interface {
	// AddComment adds the given comment to the repository
	AddComment(ctx context.Context, c comment.Comment, channelID string, assetType comment.AssetType) (comment *comment.Comment, err error)
}

// Repository provides adding functionality to the comments repository
type Repository interface {
	// AddComment persists the given comment to the repository
	AddComment(ctx context.Context, c comment.Comment, channelID string, assetType comment.AssetType) (comment *comment.Comment, err error)
}

// NewService creates an adding service
func NewService(r Repository) Service {
	return &service{r}
}

type service struct {
	r Repository
}

func (s *service) AddComment(ctx context.Context, c comment.Comment, channelID string, assetType comment.AssetType) (*comment.Comment, error) {
	return s.r.AddComment(ctx, c, channelID, assetType)
}
