package dice

type Emitter interface {
	// adds a component to a list of topics
	Subscribe(...*Component)
	// dispatch an event to a subject
	// E.g., adding a label or a record
	Emit(e Event) error
}

type eventEmitter struct {
	// Map of event types and its subscribers
	subs map[EventType][]*Component
}

func NewEmitter() *eventEmitter {
	return &eventEmitter{subs: make(map[EventType][]*Component)}
}

func (m *eventEmitter) Subscribe(comp ...*Component) {
	for _, c := range comp {
		for _, t := range c.Events {
			m.subs[t] = append(m.subs[t], c)
		}
	}
}

func (m *eventEmitter) Emit(e Event) error {
	for _, s := range m.subs[e.Type] {
		return s.update(e)
	}
	return nil
}
