package auth

// Action represents type of action to be performed on asset
type Action int

// Action values
const (
	ReadAction Action = iota
	UpdateAction
	DeleteAction
)

func (a Action) String() string {
	return [...]string{"read", "update", "delete"}[a]
}

// Service provides ACL functionality
type Service interface {
	Enforce(assetType string, act Action, authToken string) (bool, error)
}

// NewService creates an authorization service
func NewService() Service {
	return &service{}
}

type service struct{}

// Enforce returns true if action is allowed to be performed on specified asset
func (s *service) Enforce(assetType string, act Action, authToken string) (bool, error) {
	// TODO - implement calling KompiGuard service
	return true, nil
}
