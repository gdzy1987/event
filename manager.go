package event

import (
	"regexp"
	"strings"
	"sync"
)

// Wildcard event name
const Wildcard = "*"

// Manager event manager definition. for manage events and listeners
type Manager struct {
	name string
	pool sync.Pool
	// storage user custom Event instance. you can pre-define some Event instances.
	events map[string]Event
	// storage all event name and ListenerQueue map
	listeners map[string]*ListenerQueue
	// storage all event names by listened
	listenedNames map[string]int
}

var goodNameReg = regexp.MustCompile(`^[a-zA-Z][\w-.*]*$`)

// NewManager create event manager
func NewManager(name string) *Manager {
	em := &Manager{
		name:          name,
		events:        make(map[string]Event),
		listeners:     make(map[string]*ListenerQueue),
		listenedNames: make(map[string]int),
	}

	// set pool creator
	em.pool.New = func() interface{} {
		return &BasicEvent{}
	}

	return em
}

/*************************************************************
 * Listener Manager
 *************************************************************/

// On register a event handler/listener
func (em *Manager) On(name string, listener Listener, priority int) {
	name = goodName(name)

	if listener == nil {
		panic("event: the event '" + name + "' listener cannot be empty")
	}

	li := &ListenerItem{priority, listener}

	if lq, ok := em.listeners[name]; ok {
		em.listenedNames[name]++
		em.listeners[name] = lq.Push(li)
	} else { // first add.
		em.listenedNames[name] = 1
		em.listeners[name] = (&ListenerQueue{}).Push(li)
	}
}

// Fire event by name
func (em *Manager) Fire(name string, args ...interface{}) (err error) {
	name = goodName(name)

	// call listeners use defined Event
	if e, ok := em.events[name]; ok {
		return em.FireEvent(e)
	}

	// create a basic event instance
	e := em.pool.Get().(*BasicEvent)
	e.SetName(name)
	e.Fill(nil, args...)

	// call listeners
	err = em.FireEvent(e)

	e.reset()
	em.pool.Put(e)
	return
}

// MustFire fire event by name. will panic on error
func (em *Manager) MustFire(name string, args ...interface{}) {
	err := em.Fire(name, args...)
	if err != nil {
		panic(err)
	}
}

// FireEvent fire event by given BasicEvent instance
func (em *Manager) FireEvent(e Event) (err error) {
	// find matched listeners
	name := e.Name()
	lq, ok := em.listeners[name]
	if !ok {
		return
	}

	// sort by priority before call.
	for _, li := range lq.Sort().Items() {
		err = li.listener.Handle(e)
		if err != nil || e.Aborted() {
			return
		}
	}

	// has group listeners. "app.*" "app.db.*"
	// eg: "app.run" will trigger listeners on the "app.*"
	pos := strings.LastIndexByte(name, '.')
	if pos > 0 && pos < len(name) {
		groupName := name[:pos] + Wildcard // "app.*"

		if lq, ok := em.listeners[groupName]; ok {
			for _, li := range lq.Sort().Items() {
				err = li.listener.Handle(e)
				if err != nil || e.Aborted() {
					return
				}
			}
		}
	}

	// has wildcard event listeners
	if lq, ok := em.listeners[Wildcard]; ok {
		for _, li := range lq.Sort().Items() {
			err = li.listener.Handle(e)
			if err != nil || e.Aborted() {
				return
			}
		}
	}
	return
}

// HasListeners has listeners for the event name.
func (em *Manager) HasListeners(name string) bool {
	_, ok := em.listenedNames[name]
	return ok
}

// ClearListeners by name
func (em *Manager) ClearListeners(name string) {
	_, ok := em.listenedNames[name]
	if ok {
		delete(em.listenedNames, name)
		delete(em.listeners, name)
	}
}

/*************************************************************
 * Event Manager
 *************************************************************/

// AddEvent add a defined event instance to manager.
func (em *Manager) AddEvent(e Event) {
	name := goodName(e.Name())
	em.events[name] = e
}

// GetEvent get a defined event instance by name
func (em *Manager) GetEvent(name string) (e Event, ok bool) {
	e, ok = em.events[name]
	return
}

// HasEvent has event check
func (em *Manager) HasEvent(name string) bool {
	_, ok := em.events[name]
	return ok
}

// DelEvent delete Event by name
func (em *Manager) DelEvent(name string) {
	if _, ok := em.events[name]; ok {
		delete(em.events, name)
	}
}

// ClearEvents clear all events
func (em *Manager) ClearEvents() {
	em.events = map[string]Event{}
}

/*************************************************************
 * Helper Methods
 *************************************************************/

// Clear all data
func (em *Manager) Clear() {
	em.name = ""
	em.events = make(map[string]Event)
	em.listeners = make(map[string]*ListenerQueue)
	em.listenedNames = make(map[string]int)
}

func (em *Manager) callListeners(e Event, listeners []Listener) (err error) {
	for _, listener := range listeners {
		err = listener.Handle(e)
		if err != nil || e.Aborted() {
			return
		}
	}
	return
}

func goodName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("event: the event name cannot be empty")
	}

	if !goodNameReg.MatchString(name) {
		panic(`event: the event name is invalid, must match regex '^[a-zA-Z][\w-.]*$'`)
	}

	return name
}
