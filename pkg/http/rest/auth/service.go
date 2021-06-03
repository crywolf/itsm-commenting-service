package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

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
	Enforce(assetType string, act Action, channelID, authToken string) (bool, error)
}

// NewService creates an authorization service
func NewService(logger *zap.Logger) Service {
	c := http.DefaultClient
	return &service{
		logger: logger,
		client: c,
	}
}

type service struct {
	logger *zap.Logger
	client *http.Client
}

// Enforce returns true if action is allowed to be performed on specified asset
func (s *service) Enforce(assetType string, act Action, channelID, authToken string) (bool, error) {
	// GET /api/v1/kompiguard/enforce/?obj=/comment/*&act=read
	u := "http://api/v1/kompiguard/enforce/"
	q := url.QueryEscape(fmt.Sprintf("?obj=/%s/*&act=%s", assetType, act))
	fmt.Println(u + q)
	req, err := http.NewRequest(http.MethodGet, u+q, nil)
	if err != nil {
		msg := "could not create authorization service request"
		s.logger.Error(msg, zap.Error(err))
		return false, fmt.Errorf("%s: %v", msg, err)
	}
	req.Header.Set("grpc-metadata-space", channelID)
	req.Header.Set("authorization", authToken)

	resp, err := s.client.Do(req)
	if err != nil {
		msg := "authorization service request failed"
		s.logger.Error(msg, zap.String("url", u+q), zap.Error(err))
		return false, fmt.Errorf("%s: %v", msg, err)
	}

	if resp.StatusCode == http.StatusOK {
		type OKPayload struct {
			Result struct {
				Granted bool   `json:"granted"`
				Reason  string `json:"reason"`
			} `json:"result"`
		}
		var payload OKPayload

		defer func() { _ = resp.Body.Close() }()
		dec := json.NewDecoder(resp.Body)
		err := dec.Decode(&payload)
		if err != nil {
			s.logger.Error("could not decode authorization service Ok response", zap.Error(err))
			return false, err
		}

		if payload.Result.Granted {
			return true, nil
		}
		return false, nil
	}

	// we assume that anything except 200 Ok is an error
	type ErrorPayload struct {
		Error string `json:"error"`
	}

	var payload ErrorPayload
	defer func() { _ = resp.Body.Close() }()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&payload)
	if err != nil {
		msg := "could not decode authorization service non-Ok response"
		s.logger.Error(msg, zap.Error(err))
		return false, fmt.Errorf("%s: %v", msg, payload.Error)
	}

	return false, fmt.Errorf("authorization service returned error: %v", payload.Error)
}
