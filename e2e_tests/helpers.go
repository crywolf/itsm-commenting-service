package e2e

import (
	"encoding/json"
)

// mapFromJSON converts JSON data into a map
func mapFromJSON(data []byte) map[string]interface{} {
	var result interface{}
	_ = json.Unmarshal(data, &result)
	return result.(map[string]interface{})
}
