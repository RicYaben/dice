package dice

type Emitter interface {
	// adds a component to a list of topics
	subscribe(Component)
	// dispatch an event to a subject
	// E.g., adding a label or a record
	Emit(e Event) error
}

type eventEmitter struct {
	// Map of event types and its subscribers
	subs map[EventType][]Component
}

func NewEmitter() *eventEmitter {
	return &eventEmitter{subs: make(map[EventType][]Component)}
}

func (m *eventEmitter) subscribe(comp Component) {
	for _, t := range comp.Events {
		m.subs[t] = append(m.subs[t], comp)
	}
}

func (m *eventEmitter) Emit(e Event) error {
	for _, s := range m.subs[e.Type] {
		return s.update(e)
	}
	return nil
}
