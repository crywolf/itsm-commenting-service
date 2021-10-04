package updating

import (
	"context"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
)

// Service provides comment updating operations
type Service interface {
	// MarkAsReadByUser mark comment as read by invoking user
	MarkAsReadByUser(ctx context.Context, id string, readBy comment.ReadBy, channelID string, assetType comment.AssetType) (alreadyMarked bool, error error)
}

// Repository provides access to comments storage
type Repository interface {
	// MarkAsReadByUser adds user info to read_by array
	MarkAsReadByUser(ctx context.Context, id string, readBy comment.ReadBy, channelID string, assetType comment.AssetType) (alreadyMarked bool, error error)
}

// NewService creates an updating service
func NewService(r Repository) Service {
	return &service{r}
}

type service struct {
	r Repository
}

// MarkAsReadByUser adds user info to 'read_by' array in the comment in the storage
// It returns true if comment was already marked before to notify that resource was not changed.
func (s *service) MarkAsReadByUser(ctx context.Context, id string, readBy comment.ReadBy, channelID string, assetType comment.AssetType) (alreadyMarked bool, error error) {
	return s.r.MarkAsReadByUser(ctx, id, readBy, channelID, assetType)
}
