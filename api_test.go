package atom

import (
	"testing"
)

func TestPutEvent(t *testing.T) {
	client := newMockPoster()
	api := New(&Config{
		ApiKey: "API_KEY",
		Client: client,
	})
	api.PutEvent("foo.bar", []byte("hello"))
	if _, ok := client.stash["bulk"]; ok {
		t.Error("should track event to '/' path")
	}
	body := client.stash[""][0]
	expected := &Payload{Stream: "foo.bar", Data: "hello", Bulk: false}
	if !equal(body, expected) {
		t.Errorf("PutEvent: got\n\t%+v\nexpected\n\t%+v", expected, body)
	}
}

func TestPutEvents(t *testing.T) {
	client := newMockPoster()
	api := New(&Config{
		ApiKey: "API_KEY",
		Client: client,
	})
	if _, err := api.PutEvents("stream"); err == nil {
		t.Error("should throw when tracking empty batch")
	}
	api.PutEvents("foo.bar", []byte(`{"foo":"bar"}`), []byte(`{"foo":"baz"}`))
	if _, ok := client.stash[""]; ok {
		t.Error("should track event to '/bulk' path")
	}
	body := client.stash["bulk"][0]
	expected := &Payload{Stream: "foo.bar", Data: `["{\"foo\":\"bar\"}","{\"foo\":\"baz\"}"]`, Bulk: true}
	if !equal(body, expected) {
		t.Errorf("PutEvents: got\n\t%+v\nexpected\n\t%+v", expected, body)
	}
}

func equal(p1, p2 *Payload) bool {
	return p1.Stream == p2.Stream && p1.Data == p2.Data && p1.Bulk == p2.Bulk
}

// Mocks
type posterMock struct {
	stash map[string][]*Payload
}

func newMockPoster() *posterMock {
	return &posterMock{make(map[string][]*Payload)}
}

func (p *posterMock) Post(path string, body *Payload) (*Response, error) {
	p.stash[path] = append(p.stash[path], body)
	return &Response{"200 OK", 200, "OK"}, nil
}
