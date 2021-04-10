package couchdb

import "github.com/KompiTech/commenting-service/pkg/domain/entity"

// Comment object
type Comment struct {
	//	UUID       string        `json:"_id"`
	UUID       string        `json:"uuid"`
	Entity     entity.Entity `json:"entity,omitempty"`
	Text       string        `json:"text,omitempty"`
	ExternalID string        `json:"external_id,omitempty"`
	// ReadBy
	CreatedAt string `json:"created_at,omitempty"`
	//	CreatedBy   string
}
