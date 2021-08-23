package e2e

import (
	"encoding/json"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/KompiTech/go-toolkit/natswatcher"
	"github.com/KompiTech/itsm-commenting-service/pkg/repository/couchdb"
	"github.com/nats-io/stan.go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestE2E initializes test suite
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "End To End tests")
}

var (
	server   *httptest.Server
	storage  *couchdb.DBStorage
	msgQueue *MsgQueue
)

var _ = BeforeSuite(func() {
	var nc *natswatcher.Watcher
	server, storage, nc = StartServer()

	var err error
	msgQueue, err = NewMsgQueue(nc)
	Expect(err).To(BeNil())
})

var _ = AfterSuite(func() {
	destroyTestDatabases(storage)
	_ = msgQueue.nc.Close()
	server.Close()
})

type MsgQueue struct {
	lock     sync.Mutex
	nc       *natswatcher.Watcher
	messages []msgData
}

// NewMsgQueue returns new queue that is subscribed to NATS and collects messages published by the application
func NewMsgQueue(nc *natswatcher.Watcher) (*MsgQueue, error) {
	q := &MsgQueue{nc: nc}
	err := q.nc.Subscribe(natswatcher.Subscription{Subject: "service", Handler: q.msgHandler})
	return q, err
}

// LastEvents returns events array from the last published NATS message and clears the queue
func (q *MsgQueue) LastEvents() []map[string]interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.messages) == 0 {
		return nil
	}

	events := q.messages[len(q.messages)-1].Events
	q.Clear()
	return events
}

// Clear removes all messages from the queue
func (q *MsgQueue) Clear() {
	q.messages = nil
}

type msgData struct {
	Events []map[string]interface{} `json:"events"`
}

// msgHandler appends new NATS message to the queue
func (q *MsgQueue) msgHandler(msg interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	msgJSON, ok := msg.(*stan.Msg)
	Expect(ok).To(BeTrue())
	dataJSON := msgJSON.Data

	var msgObj msgData
	err := json.Unmarshal(dataJSON, &msgObj)
	Expect(err).To(BeNil())

	q.messages = append(q.messages, msgObj)
}
