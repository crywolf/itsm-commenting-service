package repository

// Service provides functions to directly work with repository
type Service interface {
	CreateDatabase(channelID, assetType string) (alreadyexisted bool, error error)
}
