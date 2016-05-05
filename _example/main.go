package main

import (
	"github.com/ironSource/atom-go"
	"log"
)

func main() {
	svc := atom.New(&atom.Config{
		ApiKey: "YOUR_API_KEY",
	})

	// Put single event
	resp, err := svc.PutEvent(&atom.PutEventInput{
		StreamName: "STREAM_NAME",
		Event: &atom.Event{
			Data: []byte("PAYLOAD"),
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)

	// Put multiple events in one single request
	resp, err = svc.PutEvents(&atom.PutEventsInput{
		StreamName: "STREAM_NAME",
		Events: []*atom.Event{
			{Data: []byte("foo")},
			{Data: []byte("bar")},
			{Data: []byte("baz")},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
