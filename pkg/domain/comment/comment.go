package comment

import "github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"

// Comment object
type Comment struct {
	UUID       string        `json:"uuid,omitempty"`
	Entity     entity.Entity `json:"entity"`
	Text       string        `json:"text,omitempty"`
	ExternalID string        `json:"external_id,omitempty"`
	// TODO ReadBy
	CreatedAt string     `json:"created_at,omitempty"`
	CreatedBy *CreatedBy `json:"created_by,omitempty"`
}

// CreatedBy represents user that created this comment
type CreatedBy struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
