package comment

import (
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
)

// Comment object
type Comment struct {
	UUID       string        `json:"uuid,omitempty"`
	Entity     entity.Entity `json:"entity"`
	Text       string        `json:"text,omitempty"`
	ExternalID string        `json:"external_id,omitempty"`
	ReadBy     ReadByList    `json:"read_by,omitempty"`
	CreatedAt  string        `json:"created_at,omitempty"`
	CreatedBy  *CreatedBy    `json:"created_by,omitempty"`
}

// ReadByList is the list of users who read this comment
type ReadByList []ReadBy

// ReadBy stores info when some user read this comment
type ReadBy struct {
	Time string   `json:"time,omitempty"`
	User UserInfo `json:"user,omitempty"`
}

// UserInfo represents basic info about user
type UserInfo struct {
	UUID           string `json:"uuid,omitempty"`
	Name           string `json:"name,omitempty"`
	Surname        string `json:"surname,omitempty"`
	OrgDisplayName string `json:"org_display_name"`
	OrgName        string `json:"org_name"`
}

// CreatedBy represents minimalistic info user that created this comment
type CreatedBy struct {
	UUID    string `json:"uuid,omitempty"`
	Name    string `json:"name,omitempty"`
	Surname string `json:"surname,omitempty"`
}
