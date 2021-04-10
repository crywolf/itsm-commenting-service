package comment

import "github.com/KompiTech/commenting-service/pkg/domain/entity"

// Comment object
type Comment struct {
	UUID       string        `json:"uuid,omitempty"`
	Entity     entity.Entity `json:"entity"`
	Text       string        `json:"text"`
	ExternalID string        `json:"external_id,omitempty"`
	// ReadBy
	CreatedAt string `json:"created_at,omitempty"`
	//	CreatedBy   string `json:"created_by"`
}
