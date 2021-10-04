package testutils

import (
	"fmt"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
)

// DatabaseName return database name
func DatabaseName(channelID string, assetType comment.AssetType) string {
	return fmt.Sprintf("p_%s_%s", channelID, pluralize(assetType))
}

func pluralize(assetType comment.AssetType) string {
	return fmt.Sprintf("%ss", assetType)
}
