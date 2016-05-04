package atom

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Payload is the expected body of an http request when tracking
// an event to ironSource.atom.
type Payload struct {
	Stream string `json:"table"`
	Auth   string `json:"auth"`
	Data   string `json:"data"`
	Bulk   bool   `json:"bulk"`
}

// Poster is the interface that wraps the Post method.
type Poster interface {
	Post(path string, body *Payload) (*Response, error)
}

// Response is a minimize version of the http.Response and contains
// the information from ironSource.atom api.
type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Message    string // e.g. "OK"
}

// client is the default implementation of Poster interface.
type client struct {
	base *url.URL
}

func (c *client) Post(path string, body *Payload) (*Response, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(c.base.ResolveReference(u).String(), "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	msg, _ := ioutil.ReadAll(resp.Body)
	return &Response{resp.Status, resp.StatusCode, string(msg)}, nil
}
