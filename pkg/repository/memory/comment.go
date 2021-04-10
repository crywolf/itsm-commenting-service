package memory

import (
	"github.com/KompiTech/commenting-service/pkg/domain/entity"
)

// Comment object
type Comment struct {
	ID         string
	Entity     entity.Entity `json:"entity"`
	Text       string
	ExternalID string
	// ReadBy
	CreatedAt string
	//	CreatedBy   string
}
