# Event 

[![GoDoc](https://godoc.org/github.com/gookit/event?status.svg)](https://godoc.org/github.com/gookit/event)
[![Build Status](https://travis-ci.org/gookit/event.svg?branch=master)](https://travis-ci.org/gookit/event)
[![Coverage Status](https://coveralls.io/repos/github/gookit/event/badge.svg?branch=master)](https://coveralls.io/github/gookit/event?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gookit/event)](https://goreportcard.com/report/github.com/gookit/event)

> **[中文说明](README_cn.md)**

Lightweight event management, dispatch tool library implemented by Go

- Support for custom definition event objects
- Support for adding multiple listeners to an event
- Supports setting the priority of the event listener. The higher the priority, the higher the trigger.
- Support for a set of event listeners based on the event name prefix `PREFIX.*`.
  - add `app.*` event listen, trigger `app.run` `app.end`, Both will trigger the `app.*` event at the same time
- Support for using the wildcard `*` to listen for triggers for all events
- Complete unit testing, unit coverage `> 95%`

## GoDoc

- [godoc for github](https://godoc.org/github.com/gookit/event)

## Main method

- `On(name string, listener Listener, priority ...int)` Register event listener
- `AddSubscriber(sbr Subscriber)`  Subscribe to support registration of multiple event listeners
- `Fire(name string, params M) (error, Event)` Trigger event
- `MustFire(name string, params M) Event`   Trigger event, there will be panic if there is an error
- `FireEvent(e Event) (err error)`    Trigger an event based on a given event instance
- `FireBatch(es ...interface{}) (ers []error)` Trigger multiple events at once

## Quick start

```go
package main

import (
	"fmt"
	
	"github.com/gookit/event"
)

func main() {
	// Register event listener
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.Normal)
	
	// Register multiple listeners
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
        fmt.Printf("handle event: %s\n", e.Name())
        return nil
    }), event.High)
	
	// ... ...
	
	// Trigger event
	// Note: The second listener has a higher priority, so it will be executed first.
	event.MustFire("evt1", event.M{"arg0": "val0", "arg1": "val1"})
}
```

## Write event listeners

### Using anonymous functions

```go
package mypgk

import (
	"fmt"
	
	"github.com/gookit/event"
)

var fnHandler = func(e event.Event) error {
	fmt.Printf("handle event: %s\n", e.Name())
    return nil
}

func Run() {
    // register
    event.On("evt1", event.ListenerFunc(fnHandler), event.High)
}
```

### Using the structure method

> Implementation interface `event.Listener`

```go
package mypgk

import (
	"fmt"
	"github.com/gookit/event"
)

type MyListener struct {
	// userData string
}

func (l *MyListener) Handle(e event.Event) error {
	e.Set("result", "OK")
	return nil
}
```

## Register multiple event listeners

> Implementation interface `event.Subscriber`

```go
package mypgk

import (
	"fmt"
	
	"github.com/gookit/event"
)

type MySubscriber struct {
	// ooo
}

func (s *MySubscriber) SubscribedEvents() map[string]interface{} {
	return map[string]interface{}{
		"e1": event.ListenerFunc(s.e1Handler),
		"e2": event.ListenerItem{
			Priority: event.AboveNormal,
			Listener: event.ListenerFunc(func(e Event) error {
				return fmt.Errorf("an error")
			}),
		},
		"e3": &MyListener{},
	}
}

func (s *MySubscriber) e1Handler(e event.Event) error {
	e.Set("e1-key", "val1")
	return nil
}
```

## Write custom events

```go
package mypgk 

import (
	"fmt"
	
	"github.com/gookit/event"
)

type MyEvent struct{
	event.BasicEvent
	customData string
}

func (e *MyEvent) CustomData() string {
    return e.customData
}
```

Usage:

```go
e := &MyEvent{customData: "hello"}
e.SetName("e1")
event.AddEvent(e)

// add listener
event.On("e1", event.ListenerFunc(func(e event.Event) error {
   fmt.Printf("custom Data: %s\n", e.(*MyEvent).CustomData())
   return nil
}))

// trigger
event.Fire("e1", nil)
// OR
// event.FireEvent(e)
```

## LICENSE

**[MIT](LICENSE)**
