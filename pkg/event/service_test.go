package event_test

import (
	"testing"

	"github.com/KompiTech/go-toolkit/natswatcher"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/comment"
	"github.com/KompiTech/itsm-commenting-service/pkg/domain/entity"
	"github.com/KompiTech/itsm-commenting-service/pkg/event"
	"github.com/KompiTech/itsm-commenting-service/pkg/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Events_Publishing(t *testing.T) {
	client := new(mocks.NATSClientMock)
	expectedData := []byte(`
		{
			"events":[
				{
					"docType":"worknote",
					"entity":"incident:7e0d38d1-e5f5-4211-b2aa-3b142e4da80e",
					"event":"CREATED",
					"text":"Test comment 1",
					"uuid":"8de32c9d-8578-45a9-ab4b-32dd5c3008c7",
					"origin":""
				}
			],
			"source":"itsm",
			"space_id":"97671694-c01a-4294-8852-3500e6e5553e",
			"org_id":"23d1ddf9-107d-4555-a740-87ec5dd78234"
		}`)

	// client should publish events from the queue only once even if the queue.PublishEvents is called repeatedly
	client.On("Publish", mock.AnythingOfType("[]natswatcher.Message")).Return(nil).Run(func(args mock.Arguments) {
		msgs := args.Get(0).([]natswatcher.Message)
		msg := msgs[0]

		// test JSON event message data
		assert.Equalf(t, "service", msg.Subject, "event queue message subject is not correct")
		assert.JSONEqf(t, string(expectedData), string(msg.Data), "event queue message data is not correct")
	}).Once()

	es := event.NewService(client)

	channelID := "97671694-c01a-4294-8852-3500e6e5553e"
	orgID := "23d1ddf9-107d-4555-a740-87ec5dd78234"

	q, err := es.NewQueue(event.UUID(channelID), event.UUID(orgID))
	require.NoError(t, err)

	c := comment.Comment{
		UUID:   "8de32c9d-8578-45a9-ab4b-32dd5c3008c7",
		Text:   "Test comment 1",
		Entity: entity.NewEntity("incident", "7e0d38d1-e5f5-4211-b2aa-3b142e4da80e"),
		// the rest is omitted
	}

	err = q.AddCreateEvent(c, "worknote", "")
	require.NoError(t, err)

	err = q.PublishEvents()
	require.NoError(t, err)

	client.AssertExpectations(t)

	// try calling again - it should not call the client
	err = q.PublishEvents()
	require.NoError(t, err)

	client.AssertExpectations(t)
}
