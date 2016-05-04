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

// PutEvent writes a single data event into ironSource.atom delivery stream.
//
// To write multiple data records into a delivery stream, use PutEvents.
func (a *Atom) PutEvent(streamName string, event []byte) (*Response, error) {
	return a.put("", streamName, event)
}

// PutEvents Writes multiple data events into a delivery stream in a single request.
// Each data event must be a byte-array, that represent a stringified JSON object.
//
// To write single data event into a delivery stream, use PutEvent.
func (a *Atom) PutEvents(streamName string, events ...[]byte) (*Response, error) {
	if len(events) == 0 {
		return nil, ErrorEmptyBatch
	}
	batch := make([]string, 0, len(events))
	for _, ev := range events {
		batch = append(batch, string(ev))
	}
	buf, err := json.Marshal(batch)
	if err != nil {
		return nil, err
	}
	return a.put("bulk", streamName, buf)
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
