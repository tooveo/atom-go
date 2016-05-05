package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ironSource/atom-go"
	"github.com/ironSource/atom-go/batcher"
)

func main() {
	log := logrus.New()
	svc := atom.New(&atom.Config{
		ApiKey: "YOUR_API_KEY",
	})
	batch := batcher.New(&batcher.Config{
		StreamName: "STREAM_NAME:",
		Logger:     log,
		Client:     svc,
	})

	batch.Start()

	// Handle failures
	go func() {
		for ev := range batch.NotifyFailures() {
			// ev contains 'Data' and 'Error()'
			log.Error(ev)
		}
	}()

	go func() {
		for i := 0; i < 5000; i++ {
			err := batch.Put([]byte(`{"foo": "bar"}`))
			if err != nil {
				log.WithError(err).Fatal("error producing")
			}
		}
	}()

	time.Sleep(3 * time.Second)
	batch.Stop()
}
