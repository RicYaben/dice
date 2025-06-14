package dice

// Accepted event handlers. Monitors dispatch events
// to their subjects, and these to their subscribers.
// The monitor is in charge of registering the event into the repo
// and dispatching it afterwards
type EventType string

const (
	// Fires on record created
	ON_RECORD EventType = "RecordCreated"
	// Fires on label created
	ON_LABEL EventType = "LabelCreated"
)

type hostEvent struct {
	// Host attributes and the object event triggering the
	// event, e.g., a new record or label
	HostID   uint
	ObjectID uint

	// Event producer. It is always a node
	NodeID    uint
	EventType EventType
}

func makeEvent(hostID, nodeID, objectID uint, eType EventType) hostEvent {
	return hostEvent{
		HostID:    hostID,
		NodeID:    nodeID,
		ObjectID:  objectID,
		EventType: eType,
	}
}

type eventsHandler interface {
	addHandler(h eventHandler)
	handle(e hostEvent) error
}

type eventHandler struct {
	Type   EventType
	Handle func(hostEvent) error
}

type eventsHandlerIterator struct {
	observer Observer
	handlers []eventHandler
}

func (h *eventsHandlerIterator) handle(e hostEvent) error {
	for _, handler := range h.handlers {
		if err := handler.Handle(e); err != nil {
			return err
		}
	}
	return nil
}

type eventsHandlerMap struct {
	observer Observer
	handlers map[EventType]eventHandler
}

func (h *eventsHandlerMap) handle(e hostEvent) error {
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

func propagate(subject Subject, observer Observer, unsub bool) {
	if unsub {
		subject.deregister(observer)
	}

	for _, child := range observer.getChildren() {
		subject.register(child)
	}
}

func defaultRole(m Monitor, o Observer) []eventHandler {
	return []eventHandler{
		{
			Type: ON_RECORD,
			Handle: func(e hostEvent) error {
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

func holdRole(m Monitor, o Observer) []eventHandler {
	return []eventHandler{
		{
			Type: ON_RECORD,
			Handle: func(e hostEvent) error {
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

func gateRole(m Monitor, o Observer) []eventHandler {
	return []eventHandler{
		{
			Type: ON_LABEL,
			Handle: func(e hostEvent) error {
				if sub := m.getSubject(e.HostID); sub != nil {
					propagate(sub, o, true)
				}
				return nil
			},
		},
		{
			Type: ON_RECORD,
			Handle: func(e hostEvent) error {
				return o.process(e)
			},
		},
	}
}

func noneRole(m Monitor, o Observer) []eventHandler {
	return []eventHandler{
		{
			Type: ON_RECORD,
			Handle: func(e hostEvent) error {
				if sub := m.getSubject(e.HostID); sub != nil {
					propagate(sub, o, true)
				}
				return nil
			},
		},
	}
}

type EventsHandlerBuilder struct {
	monitor  Monitor
	observer Observer
	strategy Strategy
	handler  eventsHandler
}

func (b *EventsHandlerBuilder) reset() {
	b.observer = nil
}

func (b *EventsHandlerBuilder) setMonitor(m Monitor) *EventsHandlerBuilder {
	b.monitor = m
	return b
}

func (b *EventsHandlerBuilder) setObserver(o Observer) *EventsHandlerBuilder {
	b.observer = o
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
		handlers = defaultRole(b.monitor, b.observer)
	case S_HOLD:
		handlers = holdRole(b.monitor, b.observer)
	case S_GATE:
		handlers = gateRole(b.monitor, b.observer)
	case S_NONE:
		handlers = noneRole(b.monitor, b.observer)
	}

	h := b.handler
	for _, handler := range handlers {
		h.addHandler(handler)
	}
	return h
}
