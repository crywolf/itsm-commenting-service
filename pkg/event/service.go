package event

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/pkg/errors"
)

// Service provides event publishing operations
type Service interface {
	// NewQueue creates new event queue
	NewQueue(channelID, orgID UUID) (Queue, error)
}

// Queue provides event publishing operations
type Queue interface {
	// AddCreateEvent prepares new event of type CREATE
	AddCreateEvent(c comment.Comment, assetType string) error
	// PublishEvents publishes all prepared events not published yet
	PublishEvents() error
}

// UUID represents UUID value
type UUID string

var uuidRegex = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")

func (u UUID) isValid() bool {
	return uuidRegex.MatchString(string(u))
}

// NewService creates an adding service
func NewService() Service {
	return &service{}
}

type service struct{}

const eventCreated = "CREATED"

// NewQueue creates new event queue
func (s *service) NewQueue(channelID, orgID UUID) (Queue, error) {
	if channelID == "" || !channelID.isValid() {
		return nil, errors.New("empty or invalid channelID param")
	}

	if orgID == "" || !orgID.isValid() {
		return nil, errors.New("empty or invalid orgID param")
	}

	return &queue{
		channelID: channelID,
		orgID:     orgID,
	}, nil
}

// AddCreateEvent prepares new event of type CREATE
func (q *queue) AddCreateEvent(c comment.Comment, assetType string) error {
	e := event{
		DocType:   assetType,
		UUID:      UUID(c.UUID),
		EventType: eventCreated,
		Entity:    c.Entity,
		Text:      c.Text,
	}

	q.events = append(q.events, e)

	return nil
}

// PublishEvents publishes all prepared events not published yet
func (q *queue) PublishEvents() error {
	type finalEvent struct {
		Events  []event `json:"events"`
		Source  string  `json:"source"`
		SpaceID UUID    `json:"space_id"`
		OrgID   UUID    `json:"org_id"`
	}

	fEvent := finalEvent{
		SpaceID: q.channelID,
		Events:  q.events,
		Source:  "itsm",
		OrgID:   q.orgID,
	}

	mEvents, err := json.Marshal(fEvent)
	if err != nil {
		return errors.Wrap(err, "unable to marshal event")
	}

	fmt.Printf("\n===> %s\n", mEvents)

	// TODO implement events publishing to NATS

	// clear the events queue
	q.events = nil

	return nil
}

type queue struct {
	channelID UUID
	orgID     UUID
	events    []event
}

type event struct {
	DocType   string        `json:"docType"`
	UUID      UUID          `json:"uuid"`
	EventType string        `json:"event"`
	Entity    entity.Entity `json:"entity"`
	Text      string        `json:"text"`
}
