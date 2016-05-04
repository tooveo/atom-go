package atom

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Payload struct {
	Stream string `json:"table"`
	Auth   string `json:"auth"`
	Data   string `json:"data"`
}

type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Message    string // e.g. "OK"
}

type Poster interface {
	Post(path string, body *Payload) (*Response, error)
}

type Client struct {
	base *url.URL
}

func (c *Client) Post(path string, body *Payload) (*Response, error) {
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
