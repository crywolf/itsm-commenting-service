package repository

import (
	"context"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
)

// Service provides functions to directly work with repository
type Service interface {
	CreateDatabase(ctx context.Context, channelID string, assetType comment.AssetType) (alreadyExisted bool, error error)
}
