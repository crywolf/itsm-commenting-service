package couchdb

import "github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"

// Comment object
type Comment struct {
	UUID       string        `json:"uuid"`
	Entity     entity.Entity `json:"entity"`
	Text       string        `json:"text,omitempty"`
	ExternalID string        `json:"external_id,omitempty"`
	// ReadBy
	CreatedAt string `json:"created_at,omitempty"`
	//	CreatedBy   string
}
