package dice

import "iter"

// Accepted event handlers. Monitors dispatch events
// to their subjects, and these to their subscribers.
// The monitor is in charge of registering the event into the repo
// and dispatching it afterwards
type EventType uint8
type EventIter iter.Seq2[Event, error]

// TODO: clean
const (
	// Fires on record created
	ON_RECORD EventType = "RecordCreated"
	// Fires on label created
	ON_LABEL EventType = "LabelCreated"
)

const (
	SOURCE_EVENT EventType = iota
	LABEL_EVENT
	FINGERPRINT_EVENT
	HOST_EVENT
	SCAN_EVENT
)

type Event struct {
	ObjectID  uint
	NodeID    uint
	EventType EventType
}

type eventsHandler interface {
	addHandler(h eventHandler)
	handle(e Event) error
}

type eventHandler struct {
	Type   EventType
	Handle func(Event) error
}

type eventsHandlerIterator struct {
	observer Module
	handlers []eventHandler
}

func (h *eventsHandlerIterator) handle(e Event) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(e); err != nil {
			return err
		}
	}
	return nil
}

type eventsHandlerMap struct {
	observer Module
	handlers map[EventType]eventHandler
}

func (h *eventsHandlerMap) handle(e Event) error {
	if h, ok := h.handlers[e.EventType]; ok {
		return h.Handle(e)
	}
	// TODO: Handle handler not found??
	return nil
}

// Strategy to handle node propagation
type Strategy string

const (
	// Default strategy. Process, ubsubscribe and propagate
	S_DEFAULT Strategy = "default"
	// Unsubscribe after successfully labelling a single event
	// I.e., after labelling a record, unsubscribe from the host
	// and propagate the subscription to downlinks.
	// This mode is best for gatting records.
	S_GATE Strategy = "gate"
	// Retains the subscription even after labeling
	// results. I.e., after labelling a record, keep receiving
	// parsing and labelling new ones.
	// This mode is best for learning nodes, assertions, and
	// operations that must happen on the majority of records.
	S_HOLD Strategy = "hold"
	// No operation. Unsubscribes directly and lets downstream
	// nodes feed records.
	// NOTE: This strategy does not process records!
	S_NONE Strategy = "none"
)

func propagate(subject Component, observer Module, unsub bool) {
	if unsub {
		subject.deregister(observer)
	}

	for _, child := range observer.getChildren() {
		subject.register(child)
	}
}

func defaultRole(m Emitter, o Module) []eventHandler {
	return []eventHandler{
		{
			Type: ON_RECORD,
			Handle: func(e Event) error {
				if sub := m.getSubject(e.HostID); sub != nil {
					if err := o.process(e); err != nil {
						return err
					}
					propagate(sub, o, true)
				}
				return nil
			},
		},
	}
}

func holdRole(m Emitter, o Module) []eventHandler {
	return []eventHandler{
		{
			Type: ON_RECORD,
			Handle: func(e Event) error {
				if sub := m.getSubject(e.HostID); sub != nil {
					if err := o.process(e); err != nil {
						return err
					}
					propagate(sub, o, false)
				}
				return nil
			},
		},
	}
}

func gateRole(m Emitter, o Module) []eventHandler {
	return []eventHandler{
		{
			Type: ON_LABEL,
			Handle: func(e Event) error {
				if sub := m.getSubject(e.HostID); sub != nil {
					propagate(sub, o, true)
				}
				return nil
			},
		},
		{
			Type: ON_RECORD,
			Handle: func(e Event) error {
				return o.process(e)
			},
		},
	}
}

func noneRole(m Emitter, o Module) []eventHandler {
	return []eventHandler{
		{
			Type: ON_RECORD,
			Handle: func(e Event) error {
				if sub := m.getSubject(e.HostID); sub != nil {
					propagate(sub, o, true)
				}
				return nil
			},
		},
	}
}

type EventsHandlerBuilder struct {
	emitter  Emitter
	module   Module
	strategy Strategy
	handler  eventsHandler
}

func (b *EventsHandlerBuilder) reset() {
	b.module = nil
}

func (b *EventsHandlerBuilder) setMonitor(m Emitter) *EventsHandlerBuilder {
	b.emitter = m
	return b
}

func (b *EventsHandlerBuilder) setObserver(o Module) *EventsHandlerBuilder {
	b.module = o
	return b
}

func (b *EventsHandlerBuilder) setStrategy(s Strategy) *EventsHandlerBuilder {
	b.strategy = s
	return b
}

func (b *EventsHandlerBuilder) Build() eventsHandler {
	defer b.reset()

	var handlers []eventHandler
	switch b.strategy {
	case S_DEFAULT:
		handlers = defaultRole(b.emitter, b.module)
	case S_HOLD:
		handlers = holdRole(b.emitter, b.module)
	case S_GATE:
		handlers = gateRole(b.emitter, b.module)
	case S_NONE:
		handlers = noneRole(b.emitter, b.module)
	}

	h := b.handler
	for _, handler := range handlers {
		h.addHandler(handler)
	}
	return h
}
