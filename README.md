# ironSource.atom SDK for Go [![Build status][travis-image]][travis-url] [![License][license-image]][license-url] [![GoDoc][godoc-img]][godoc-url]
atom-go is the official [ironSource.atom](http://www.ironsrc.com/data-flow-management) SDK for the Go programming language.

- [Get started](https://atom.ironsrc.com/#/signup)
- [Using the SDK](#using-the-sdk)
  - [Tracking events](#tracking-events)
  - [Batcher](#batcher)

### Using the SDK

#### Tracking events
__Installation__
```sh
$ go get github.com/ironSource/atom-go
```
Now you can start tracking events:
```go
func main() {
  // ...
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
  
  // Put multiple events in one single request
  resp, err := svc.PutEvents(&atom.PutEventsInput{
    StreamName: "STREAM_NAME",
    Events: []*atom.Event{
      {Data: []byte("foo")},
      {Data: []byte("bar")},
      {Data: []byte("baz")},
    },
  })
}
```

### Batcher
Batch producer for ironSource.atom built on top of the SDK.

__Installation__
```sh
$ go get github.com/ironSource/atom-go/batcher
```

```go
func main() {
    log := logrus.New()
    svc := atom.New(&atom.Config{
        ApiKey: "YOUR_API_KEY",
    })
    batch := batcher.New(&batcher.Config{
        StreamName: "STREAM_NAME",
        Client:     svc,
        Logger:     log,
        BatchCount: 100,
    })

    batch.Start()

    // Handle failures
    go func() {
        for ev := range batch.NotifyFailures() {
            log.Error(ev)
        }
    }()

    go func() {
        for i := 0; i < 1000; i++ {
            err := batch.Put([]byte(`{"foo": "bar"}`))
            if err != nil {
                log.WithError(err).Fatal("error producing message")
            }
        }
    }()

    time.Sleep(3 * time.Second)
    batch.Stop()
}
```

### License
MIT

[godoc-url]: https://godoc.org/github.com/ironSource/atom-go
[godoc-img]: https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square
[license-image]: https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square
[license-url]: LICENSE
[travis-image]: https://img.shields.io/travis/ironSource/atom-go.svg?style=flat-square
[travis-url]: https://travis-ci.org/ironSource/atom-go
