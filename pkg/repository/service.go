package repository

import "context"

// Service provides functions to directly work with repository
type Service interface {
	CreateDatabase(ctx context.Context, channelID, assetType string) (alreadyExisted bool, error error)
}
