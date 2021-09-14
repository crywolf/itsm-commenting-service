package comment

import (
	"strings"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
)

// Comment object
// swagger:model
type Comment struct {
	// Read Only: true
	// required: true
	// swagger:strfmt uuid
	UUID string `json:"uuid,omitempty"`

	// Entity represents some external entity reference in the form "&lt;entity&gt;:&lt;UUID&gt;"
	// example: incident:f49d5fd5-8da4-4779-b5ba-32e78aa2c444
	// required: true
	// swagger:strfmt string
	Entity entity.Entity `json:"entity"`

	// Content of the comment
	// required: true
	Text string `json:"text,omitempty"`

	// ID in external system
	ExternalID string `json:"external_id,omitempty"`

	// Origin of the request
	Origin string `json:"-"`

	// ReadBy is a list of users who read this comment
	ReadBy ReadByList `json:"read_by,omitempty"`

	// Time when the resource was created
	// required: true
	// swagger:strfmt date-time
	CreatedAt string `json:"created_at,omitempty"`

	// CreatedBy represents user who created this comment
	// required: true
	CreatedBy *UserInfo `json:"created_by,omitempty"`
}

// ReadByList is the list of users who read this comment
type ReadByList []ReadBy

// ReadBy stores info when some user read this comment
type ReadBy struct {
	// required: true
	// swagger:strfmt date-time
	Time string `json:"time,omitempty"`
	// required: true
	User UserInfo `json:"user,omitempty"`
}

// UserInfo represents basic info about user
type UserInfo struct {
	// required: true
	// swagger:strfmt uuid
	UUID string `json:"uuid,omitempty"`
	// required: true
	Name string `json:"name,omitempty"`
	// required: true
	Surname string `json:"surname,omitempty"`
	// required: true
	// example: KompiTech
	OrgDisplayName string `json:"org_display_name,omitempty"`
	// required: true
	// example: a897a407-e41b-4b14-924a-39f5d5a8038f.kompitech.com
	OrgName string `json:"org_name,omitempty"`
}

// OrgID returns org_id based on orgName
func (u *UserInfo) OrgID() string {
	return strings.SplitN(u.OrgName, ".", 2)[0]
}
