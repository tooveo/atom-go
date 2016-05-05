package atom

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

// Errors
var (
	ErrorInvalidStream = errors.New("Invalid stream name. length must be at least 1")
	ErrorEmptyBatch    = errors.New("events length must be at least 1")
)

// ironSource.atom api object.
type Atom struct {
	*Config
}

// New creates new Atom object with the given config.
// ApiKey is required.
func New(config *Config) *Atom {
	config.defaults()
	return &Atom{
		Config: config,
	}
}

// The unit of data in a delivery stream.
type Event struct {
	Data []byte
}

// PutEventInput used as a PutEvent parameter.
// both `StreamName` and `Event` are required.
type PutEventInput struct {
	StreamName string // name of the delivery stream.
	Event      *Event // event payload.
}

// PutEvent writes a single data event into ironSource.atom delivery stream.
//
// To write multiple data records into a delivery stream, use PutEvents.
func (a *Atom) PutEvent(input *PutEventInput) (*Response, error) {
	return a.put("", input.StreamName, input.Event.Data)
}

// PutEventsInput used as a PutEvents parameter.
// `StreamName` is required and `Events` length must be at least 1.
type PutEventsInput struct {
	StreamName string   // name of the delivery stream.
	Events     []*Event // batch of payloads.
}

// PutEvents Writes multiple data events into a delivery stream in a single request.
//
// To write single data event into a delivery stream, use PutEvent.
func (a *Atom) PutEvents(input *PutEventsInput) (*Response, error) {
	if len(input.Events) == 0 {
		return nil, ErrorEmptyBatch
	}
	batch := make([]string, 0, len(input.Events))
	for _, ev := range input.Events {
		batch = append(batch, string(ev.Data))
	}
	buf, err := json.Marshal(batch)
	if err != nil {
		return nil, err
	}
	return a.put("bulk", input.StreamName, buf)
}

func (a *Atom) put(path, streamName string, data []byte) (*Response, error) {
	if len(streamName) == 0 {
		return nil, ErrorInvalidStream
	}
	return a.Client.Post(path, &Payload{
		Stream: streamName,
		Data:   string(data),
		Auth:   a.hmac(data),
		Bulk:   path == "bulk",
	})
}

func (a *Atom) hmac(data []byte) string {
	h := hmac.New(sha256.New, a.apikey)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
