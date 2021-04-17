package memory

import (
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
)

// Comment object
type Comment struct {
	ID         string
	Entity     entity.Entity
	Text       string
	ExternalID string
	// ReadBy
	CreatedAt string
	CreatedBy CreatedBy
}

// CreatedBy represents user that created this comment
type CreatedBy struct {
	UUID string
	Name string
}
