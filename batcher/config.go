package batcher

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ironSource/atom-go"
	"github.com/jpillora/backoff"
)

const (
	maxBatchSize          = 5 << 20
	maxBatchCount         = 500
	defaultMaxConnections = 24
	defaultFlushInterval  = time.Second
)

// Atom api interface
type Atomface interface {
	PutEvent(streamName string, event []byte) (*atom.Response, error)
	PutEvents(streamName string, events ...[]byte) (*atom.Response, error)
}

type Config struct {
	// StreamName is the ironSource.atom stream.
	StreamName string

	// FlushInterval is a regular interval for flushing the buffer. Defaults to 5s.
	FlushInterval time.Duration

	// BatchCount determine the maximum number of items to pack into batch.
	// Must not exceed length. Defaults to 500.
	BatchCount int

	// BatchSize determine the maximum number of bytes to send with a PutRecords request.
	// Must not exceed 5MiB; Default to 5MiB.
	BatchSize int

	// BacklogCount determines the channel capacity before Put() will begin blocking. Default to `BatchCount`.
	BacklogCount int

	// Maximum number of connections to open to the backend.
	// HTTP requests are sent in parallel over multiple connections.
	// Default to 24.
	MaxConnections int

	// Backoff determines the backoff strategy for record failures.
	Backoff backoff.Backoff

	// Logger is the logger used. Default to logrus.Log.
	Logger *logrus.Logger

	// Client is the Putter interface implementation.
	Client Atomface
}

// defaults for configuration
func (c *Config) defaults() {
	if c.Logger == nil {
		c.Logger = logrus.New()
	}
	if c.BatchCount == 0 {
		c.BatchCount = maxBatchCount
	}
	falseOrPanic(c.BatchCount > maxBatchCount, "atom: BatchCount exceeds 500")
	if c.BatchSize == 0 {
		c.BatchSize = maxBatchSize
	}
	falseOrPanic(c.BatchSize > maxBatchSize, "atom: BatchSize exceeds 5MiB")
	if c.BacklogCount == 0 {
		c.BacklogCount = maxBatchCount
	}
	if c.MaxConnections == 0 {
		c.MaxConnections = defaultMaxConnections
	}
	falseOrPanic(c.MaxConnections < 1 || c.MaxConnections > 256, "atom: MaxConnections must be between 1 and 256")
	if c.FlushInterval == 0 {
		c.FlushInterval = time.Second * 5
	}
}

func falseOrPanic(p bool, msg string) {
	if p {
		panic(msg)
	}
}
