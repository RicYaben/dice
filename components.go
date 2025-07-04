package dice

import (
	"sync"

	"github.com/pkg/errors"
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
	subs map[string][]Component
}

func newEventEmitter() *eventEmitter {
	return &eventEmitter{subs: make(map[string][]Component)}
}

func (m *eventEmitter) subscribe(comp Component) {
	for _, t := range comp.tags() {
		m.subs[t] = append(m.subs[t], comp)
	}
}

func (m *eventEmitter) emit(e Event) error {
	for _, s := range m.subs[string(e.EventType)] {
		return s.update(e)
	}
	return nil
}

type Component interface {
	name() string
	// register a module
	register(Module) error
	// downstream an event into the registered modules
	update(Event) error
	// type of events this component is subscribed to
	tags() []string
}

type component struct {
	name        string
	modules     map[uint]Module
	entrypoints []Module
	tags        []string
	adapter     *cosmosAdapter
	handler     func(*component, Event) error

	mu sync.Mutex
}

func newComponent(name string, adapter *cosmosAdapter) *component {
	return &component{
		mu:          sync.Mutex{},
		modules:     make(map[uint]Module),
		entrypoints: []Module{},
		name:        name,
		adapter:     adapter,
	}
}
func (c *component) register(m Module) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.modules[m.ID()]; !ok {
		return errors.Errorf("module %s already registered", m.ID())
	}
	c.modules[m.ID()] = m
	return nil
}

// Sends the event to the modules to handle.
// If the event points to some object with hooks,
// the event is only pushed to the hookers
func (c *component) update(e Event) error {
	return c.handler(c, e)
	/*
		hooks, err := c.adapter.getHooks(e.ObjectID)
		if err != nil {
			return err
		}

		if len(hooks) > 0 {
			for _, h := range hooks {
				if mod, ok := c.modules[h]; ok {
					if err := mod.update(e); err != nil {
						return err
					}
				}
			}
		}

		for _, m := range c.entrypoints {
			if err := m.update(e); err != nil {
				return err
			}
		}
		return nil
	*/
}

type Module interface {
	// returns the ID of the module
	ID() uint
	// handles an update in a subject
	update(Event) error
	// returns the children attached to this observer
	getChildren() []Module
	// propagate to the children
	propagate(Event) error
	// stops the module
	kill()
}

type module struct {
	id       uint
	children []Module
	plugin   *dicePlugin
}

func (m *module) ID() uint {
	return m.id
}

func (m *module) getChildren() []Module {
	return m.children
}

func (m *module) propagate(e Event) error {
	for _, ch := range m.children {
		if err := ch.update(e); err != nil {
			return err
		}
	}
	return nil
}

func (m *module) kill() {
	m.plugin.client.Kill()
}

type classifierModule struct {
	*module
	classifier ClassifierPlugin
	adapter    *cosmosAdapter
}

func (m *classifierModule) update(e Event) error {
	// find the host owning the object
	host, err := m.adapter.getHost(e.ObjectID)
	if err != nil {
		return err
	}

	// the classifier has access to the adapter so it cant get
	// whatever fingerprint or other info it needs, create a new
	// scan if needed, etc.
	if err := m.classifier.Label(host); err != nil {
		return err
	}
	return m.propagate(e)
}
