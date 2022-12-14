package hypermedia

import (
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
)

// AllowedLinksForComment returns allowed hypermedia link names based on Comment state
func AllowedLinksForComment(_ comment.Comment, assetType comment.AssetType) []string {
	// here might be some logic based on business state of the resource, in this case there is none

	var links []string

	name := "MarkCommentAsReadByUser"

	if assetType == comment.AssetTypeWorknote {
		name = "MarkWorknoteAsReadByUser"
	}

	links = append(links, name)

	return links
}
