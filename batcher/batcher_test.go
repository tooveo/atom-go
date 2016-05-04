package batcher

import (
	"github.com/ironSource/atom-go"
	"net/http"
	"reflect"
	"sync"
	"testing"
)

type testCase struct {
	// configuration
	name   string      // test name
	config *Config     // test config
	events []string    // all outgoing events
	client *clientMock // mocked client

	// expectations
	outgoing map[int][]string // [call number][data]
}

var testCases = []testCase{
	{"one event with batch count 1",
		&Config{BatchCount: 1},
		[]string{`{"foo": 1}`},
		&clientMock{
			incoming: make(map[int][]string),
			responses: []*atom.Response{
				{StatusCode: http.StatusOK},
			},
		},
		map[int][]string{
			0: []string{`{"foo": 1}`},
		},
	},
	{"two event with batch count 1",
		&Config{BatchCount: 1},
		[]string{`{"foo": 1}`, `{"foo": 1}`},
		&clientMock{
			incoming: make(map[int][]string),
			responses: []*atom.Response{
				{StatusCode: http.StatusOK},
				{StatusCode: http.StatusOK},
			},
		},
		map[int][]string{
			0: []string{`{"foo": 1}`},
			1: []string{`{"foo": 1}`},
		},
	},
	{"two events with batch count 2. simulating retries",
		&Config{BatchCount: 2},
		[]string{`{"foo": 1}`, `{"foo": 1}`},
		&clientMock{
			incoming: make(map[int][]string),
			responses: []*atom.Response{
				{StatusCode: http.StatusInternalServerError, Status: "Internal Server Error"},
				{StatusCode: http.StatusOK},
			},
		},
		map[int][]string{
			1: []string{`{"foo": 1}`, `{"foo": 1}`},
		},
	},
}

func TestBatcher(t *testing.T) {
	for _, test := range testCases {
		test.config.MaxConnections = 1
		test.config.Client = test.client
		test.config.StreamName = test.name
		b := New(test.config)
		b.Start()
		var wg sync.WaitGroup
		wg.Add(len(test.events))
		for _, r := range test.events {
			go func(s string) {
				b.Put([]byte(s))
				wg.Done()

			}(r)
		}
		wg.Wait()
		b.Stop()
		for k, v := range test.client.incoming {
			if !reflect.DeepEqual(v, test.outgoing[k]) {
				t.Errorf("failed test: %s\n\texcpeted:%v\n\tactual:%v", test.name, test.outgoing[k], v)
			}
		}
	}
}

func TestNotify(t *testing.T) {
	b := New(&Config{
		StreamName:     "STREAM_NAME",
		MaxConnections: 1,
		BatchCount:     10,
		Client: &clientMock{
			incoming: make(map[int][]string),
			responses: []*atom.Response{
				{Status: "401 Unauthorized", StatusCode: http.StatusUnauthorized},
			},
		},
	})
	b.Start()
	var wg sync.WaitGroup
	n := 10
	wg.Add(n)
	failed := 0
	go func() {
		for _ = range b.NotifyFailures() {
			failed++
			wg.Done()
		}
	}()
	for i := 0; i < n; i++ {
		b.Put([]byte("{}"))
	}
	wg.Wait()
	b.Stop()
	if failed != n {
		t.Errorf("NotifyFailure\n\texcpeted:%v\n\tactual:%v", failed, n)
	}
}

type clientMock struct {
	calls     int
	incoming  map[int][]string
	responses []*atom.Response
}

func (c *clientMock) PutEvents(streamName string, events ...[]byte) (*atom.Response, error) {
	res := c.responses[c.calls]
	if res.StatusCode == http.StatusOK {
		for _, ev := range events {
			c.incoming[c.calls] = append(c.incoming[c.calls], string(ev))
		}
	}
	c.calls++
	return res, nil
}
