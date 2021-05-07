package user

// BasicInfo represents the basic info about user who invoked the HTTP request
type BasicInfo struct {
	UUID           string
	Name           string
	Surname        string
	OrgDisplayName string
	OrgName        string
}
