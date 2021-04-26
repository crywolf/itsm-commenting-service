package testutils

import "fmt"

func DatabaseName(channelID, kind string) string {
	return fmt.Sprintf("%s_%s", channelID, kind)
}
