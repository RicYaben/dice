package dice

import (
	"slices"
)

type Emitter interface {
	// adds a component to a list of topics
	subscribe(Component)
	// dispatch an event to a subject
	// E.g., adding a label or a record
	emit(e Event) error
}

type eventEmitter struct {
	// Map of event types and its subscribers
	subs map[EventType][]Component
}

func newEventEmitter() *eventEmitter {
	return &eventEmitter{subs: make(map[EventType][]Component)}
}

func (m *eventEmitter) subscribe(comp Component) {
	for _, t := range comp.Events {
		m.subs[t] = append(m.subs[t], comp)
	}
}

func (m *eventEmitter) emit(e Event) error {
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
	// A handler to interact with the event and update registered modules
	Handler func(comp *Component, e Event) error
	// Types of events the component listens to
	Events []EventType
	// List of signatures the component handles
	Signatures []Signature
}

// Sends the event to the modules to handle.
// If the event points to some object with hooks,
// the event is only pushed to the hookers
func (c *Component) update(e Event) error {
	return c.Handler(c, e)
}

// Wrapper around a plugin to handle events and propagation
type ComponentModule struct {
	Module  Module
	Adapter CosmosAdapter
	Handler func(mod *ComponentModule, e Event) error

	// this is the only important thing not to duplicate
	Plugin   *dicePlugin
	Children []ComponentModule
}

// TODO: make wrappers for events and modules
// Module wrappers are basically pre-component modules. Closely
// related to nodes, b.c., we dont have scanners or identifiers
// that are nodes in the signature graph
func (m *ComponentModule) update(e Event) error {
	return m.Handler(m, e)
}

func (m *ComponentModule) Propagate(e Event) error {
	for _, ch := range m.Children {
		if err := ch.update(e); err != nil {
			return err
		}
	}
	return nil
}

func componentEventHandler(comp *Component, e Event) error {
	sigs := comp.Signatures

	// Filter the targetted signatures
	if len(e.Targets) > 0 {
		sigs = Filter(sigs, func(s Signature) bool {
			return slices.Contains(e.Targets, s.Name)
		})
	}

	// Send the event to the remaining signatures
	// Each should handle the event.
	for _, sig := range sigs {
		if err := sig.update(e); err != nil {
			return err
		}
	}
	return nil
}
