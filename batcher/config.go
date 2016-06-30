package batcher

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ironSource/atom-go"
)

const (
	maxBatchSize          = 5 << 20
	maxBatchCount         = 500
	defaultMaxConnections = 24
	defaultFlushInterval  = time.Second
)

// Putter is the interface that wraps the AtomAPI.PutEvents method.
type Putter interface {
	PutEvents(*atom.PutEventsInput) (*atom.Response, error)
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

	// Number of requests to sent concurrently. Default to 24.
	MaxConnections int

	// Logger is the logger used. Default to logrus.Log.
	Logger *logrus.Logger

	// Client is the Putter interface implementation.
	Client Putter
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
	falseOrPanic(len(c.StreamName) == 0, "atom: StreamName length must be at least 1")
}

func falseOrPanic(p bool, msg string) {
	if p {
		panic(msg)
	}
}
