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
	ReadBy     ReadByList
	CreatedAt  string
	CreatedBy  CreatedBy
}

// ReadByList is the list of users who read this comment
type ReadByList []ReadBy

// ReadBy stores info when some user read this comment
type ReadBy struct {
	Time string
	User UserInfo
}

// UserInfo represents basic info about user
type UserInfo struct {
	UUID           string
	Name           string
	Surname        string
	OrgDisplayName string
	OrgName        string
}

// CreatedBy represents minimalistic info user that created this comment
type CreatedBy struct {
	UUID    string
	Name    string
	Surname string
}
