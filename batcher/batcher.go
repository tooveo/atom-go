package batcher

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

// Errors
var (
	ErrorStoppedBatcher = errors.New("Unable to Put event. Batcher is already stopped")
)

// ironSource.atom events batcher.
type Batcher struct {
	sync.Mutex
	*Config
	taskPool *TaskPool
	events   chan []byte
	failure  chan *FailureEvent
	done     chan struct{}

	// state of the Batcher.
	// notify set to true after calling to `NotifyFailures`
	notify bool
	// stopped set to true after `Stop`ing the Batcher.
	// This will prevent from user to `Put` any new data.
	stopped bool
}

// New creates new batcher with the given config.
func New(config *Config) *Batcher {
	config.defaults()
	return &Batcher{
		Config:   config,
		events:   make(chan []byte),
		done:     make(chan struct{}),
		taskPool: newPool(config.MaxConnections),
	}
}

// Start the batcher
func (b *Batcher) Start() {
	b.Logger.WithField("stream", b.StreamName).Info("starting batcher")
	b.taskPool.Start()
	go b.loop()
}

// Stop the batcher gracefully. Flushes any in-flight data.
func (b *Batcher) Stop() {
	b.Lock()
	defer b.Unlock()
	b.stopped = true
	b.Logger.WithField("backlog", len(b.events)).Info("stopping batcher")

	// drain
	b.done <- struct{}{}
	close(b.events)

	// wait
	<-b.done
	b.taskPool.Stop()

	b.Logger.Info("stopped batcher")
}

// Put `data` asynchronously. This method is thread-safe.
//
// Under the covers, the Batchr will automatically re-attempt puts in case of
// transient errors.
// When unrecoverable error has detected(e.g: authentication error or trying
// to put to in a stream that doesn't exist), the message will returned by the
// Batcher.
// Add a listener with `Batcher.NotifyFailures` to handle undeliverable messages.
func (b *Batcher) Put(data []byte) error {
	b.Lock()
	defer b.Unlock()

	if b.stopped {
		return ErrorStoppedBatcher
	}

	b.events <- data
	return nil
}

// Failure event type
type FailureEvent struct {
	error
	Data []byte
}

// NotifyFailures registers and return listener to handle undeliverable messages.
// The incoming struct has a copy of the Data along with some error information
// about why the publishing failed.
func (b *Batcher) NotifyFailures() <-chan *FailureEvent {
	b.Lock()
	defer b.Unlock()
	if !b.notify {
		b.notify = true
		b.failure = make(chan *FailureEvent, b.BacklogCount)
	}
	return b.failure
}

// loop and flush at the configured interval, or when the buffer is exceeded.
func (b *Batcher) loop() {
	size := 0
	drain := false
	buf := make([][]byte, 0, b.BatchCount)
	tick := time.NewTicker(b.FlushInterval)

	flush := func(msg string) {
		batch := buf
		b.taskPool.Put(func() {
			b.flush(batch, msg)
		})
		buf = nil
		size = 0
	}

	defer tick.Stop()
	defer close(b.done)
	for {
		select {
		case event, ok := <-b.events:
			if drain && !ok {
				if size > 0 {
					flush("drain")
				}
				b.Logger.Info("backlog drained")
				return
			}
			esize := len(event)
			if size+esize > b.BatchSize {
				flush("batch size")
			}
			size += esize
			buf = append(buf, event)
			if len(buf) >= b.BatchCount {
				flush("batch length")
			}
		case <-tick.C:
			if size > 0 {
				flush("interval")
			}
		case <-b.done:
			drain = true
		}
	}
}

// flush records and retry failures if necessary.
// for example: when we get "Service Unavailable" response
func (b *Batcher) flush(events [][]byte, reason string) {
	b.Logger.WithField("reason", reason).Infof("flush %v records", len(events))

	resp, err := b.Client.PutEvents(b.StreamName, events...)
	if err != nil {
		b.flushError(events, err)
		return
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
		b.flushError(events, errors.New(resp.Message))
		return
	}

	if resp.StatusCode == http.StatusOK {
		b.Backoff.Reset()
		return
	}

	backoff := b.Backoff.Duration()

	b.Logger.WithFields(logrus.Fields{
		"status":  resp.Status,
		"backoff": backoff,
	}).Warn("flush failed")

	b.flush(events, "retry")
}

func (b *Batcher) flushError(events [][]byte, err error) {
	b.Backoff.Reset()
	b.Logger.WithError(err).Error("flush")
	b.Lock()
	if b.notify {
		for _, ev := range events {
			b.failure <- &FailureEvent{err, ev}
		}
	}
	b.Unlock()
}
