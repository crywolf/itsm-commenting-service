package repository

// Service provides functions to directly work with repository
type Service interface {
	CreateDatabase(channelID, assetType string) (alreadyExisted bool, error error)
}
