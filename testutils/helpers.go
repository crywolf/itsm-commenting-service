package testutils

import "fmt"

// DatabaseName return database name
func DatabaseName(channelID, assetType string) string {
	return fmt.Sprintf("%s_%s", channelID, pluralize(assetType))
}

func pluralize(assetType string) string {
	return fmt.Sprintf("%ss", assetType)
}
