package dice

import (
	"slices"
)

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

type Component struct {
	// Name of the component
	Name string
	// An adapter to call operations on the cosmos database
	Adapter CosmosAdapter
	// Types of events the component listens to
	Events []EventType
	// List of signatures the component handles
	Graphs []*graph
}

// Sends the event to the modules to handle.
// If the event points to some object with hooks,
// the event is only pushed to the hookers
func (c *Component) update(e Event) error {
	g := c.Graphs

	// Filter the targetted signatures
	if len(e.Targets) > 0 {
		g = Filter(g, func(s *graph) bool {
			return slices.Contains(e.Targets, s.signature.Name)
		})
	}

	// Send the event to the remaining signatures
	// Each should handle the event.
	for _, sig := range g {
		if err := sig.Update(e); err != nil {
			return err
		}
	}
	return nil
}
