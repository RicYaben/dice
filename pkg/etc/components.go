// FAQ: Why not using an MQTT broker?
//
// The reason not to use a regular MQTT embedded broker is basically
// lazyness and a bit of performance. We could use the HostID as the root topic,
// then topics and labels as the leaf topics. However, this is rather overkill
// and would force us to handle synchronization differently, and a new strategy
// to handle messages. For example, if we subscribed a node to the labels topic
// of a Host, we would have to add some additional identifier while publishing
// and subscribing, so only the issuer node gets notified on a label creation.
//
// Example:
// labels -> host-id/labels/node-id
// all records -> host-id/records/+
// specific probe records -> host-id/records/mqtt
//
// For the synchronization, take the example of a scanner that outputs multiple
// records. As the node receives messages, it would have to know if it is still
// subscribed before processing messages. Then, the node may create new messages,
// and finally propagate.
// A solution for the synchronization would be to add the messages to a working
// pool, so we can react on each message separately, unsubscribe on demand, and
// forget about the rest of messages in the queue. This way we can still use
// the concept of "events" (just messages), wrapping them properly and handling
// the queue as we see fit.
//
// tl;dr this works just fine
package dice

import (
	"fmt"
	"slices"
	"sync"

	"github.com/rs/zerolog/log"
)

type Emitter interface {
	// adds a component to a list of topics
	subscribe([]string, Component)
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

func (m *eventEmitter) subscribe(events []string, comp Component) {
	for _, e := range events {
		m.subs[e] = append(m.subs[e], comp)
	}
}

func (m *eventEmitter) emit(e Event) error {
	for _, s := range m.subs[string(e.EventType)] {
		return s.publish(e)
	}
	return nil
}

type Component interface {
	ID() uint
	// register an observer
	register(obs Module) error
	// remove an observer
	deregister(obs Module)
	// publish an event
	publish(e Event) error
	// notify a single observer
	notify(event Event, obs ...Module) error
	// notify all observers
	notifyAll(event Event) error
}

type component struct {
	id        uint
	repo      *recordRepo
	observers []Module
	mu        sync.Mutex
}

func (s *component) ID() uint {
	return s.id
}

func (s *component) registered(obs Module) bool {
	for _, ob := range s.observers {
		if ob == obs {
			return true
		}
	}
	return false
}

func (s *component) doAfterRegister(obs Module) error {
	records, err := s.repo.findUnmarkedRecords(s.ID(), obs.ID())
	if err != nil {
		return err
	}

	for _, record := range records {
		event := makeEvent(record.HostID, record.NodeID, record.ID, ON_RECORD)
		if err := obs.update(event); err != nil {
			return err
		}

		// return early if not the observer is no loger registered
		if !s.registered(obs) {
			return nil
		}
	}
	return nil
}

func (s *component) register(obs Module) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.registered(obs) {
		return nil
	}
	s.observers = append(s.observers, obs)
	return s.doAfterRegister(obs)
}

// TODO: remove the subject from
// the monitor at some point when there are no more
// events and no more observers
func (s *component) deregister(obs Module) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sub := range s.observers {
		if sub == obs {
			// move the last element to the current index and return
			// the slice without the last element
			s.observers[i] = s.observers[len(s.observers)-1]
			s.observers = s.observers[:len(s.observers)-1]
			return
		}
	}
}

func (s *component) getObserver(id uint) Module {
	for _, obs := range s.observers {
		if obs.ID() == id {
			return obs
		}
	}
	return nil
}

func (s *component) publish(e Event) error {
	switch e.EventType {
	case "Label":
		obs := s.getObserver(e.NodeID)
		if obs == nil {
			// NOTE: not sure whether this should be returned as an error
			log.Warn().Msgf("failed to find observer %d", e.NodeID)
			return nil
		}
		s.notify(e)
	case "Record":
		s.notifyAll(e)
	default:
		return fmt.Errorf("handler not found Handler(%s)", e.EventType)
	}
	return nil
}

// When notifying, I should wrap the event to include the observer
// the type of event, and the event itself (a label, a record, whatever)
func (s *component) notify(e Event, obs ...Module) error {
	for _, sub := range s.observers {
		if slices.Contains(obs, sub) {
			if err := sub.update(e); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (s *component) notifyAll(e Event) error {
	for _, sub := range s.observers {
		if err := sub.update(e); err != nil {
			return err
		}
	}
	return nil
}

type Module interface {
	// returns the ID of the module
	ID() uint
	// handles an update in a subject
	update(event Event) error
	// returns the children attached to this observer
	getChildren() []Module
	// process an event
	process(e Event) error
	// stops the module
	kill()
}

type module struct {
	// Node attributes, we don't need to hold the whole node
	id       uint
	children []Module
	plugin   *dicePlugin

	handler  eventsHandler
	repo     *cosmosRepo
	eventBus func(Event) error
}

func (o *module) update(e Event) error {
	return o.handler.handle(e)
}

func (o *module) ID() uint {
	return o.id
}

// return the registered observers
// observers are loaded elsewhere, the observer does
// not know how to make new ones.
func (o *module) getChildren() []Module {
	return o.children
}

func (o *module) kill() {
	o.plugin.client.Kill()
}

type scanModule struct {
	module
	scanner ScanPlugin
}

func (o *scanModule) process(e Event) error {
	re := o.repo
	host, err := re.getHost(e.HostID)
	if err != nil {
		return fmt.Errorf("failed to find host %d: %w", e.HostID, err)
	}

	records, err := o.scanner.Scan(host)
	if err != nil {
		return fmt.Errorf("failed to scan host %d: %w", e.HostID, err)
	}

	for _, r := range records {
		if err := re.saveRecord(r); err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}

		event := makeEvent(e.HostID, o.ID(), r.ID, ON_RECORD)
		if err := o.eventBus(event); err != nil {
			return fmt.Errorf("failed to dispatch event: %w", err)
		}
	}
	return nil
}

type classifierModule struct {
	module
	classifier RulePlugin
}

func (o *classifierModule) process(e Event) error {
	re := o.repo
	fps, err := re.getFingerprints(e.ObjectID)
	if err != nil {
		return fmt.Errorf("could not find record %d", e.ObjectID)
	}

	if err := re.saveMark(Mark{FingerprintID: e.ObjectID, NodeID: o.ID()}); err != nil {
		return fmt.Errorf("failed to mark record: %w", err)
	}
	fp := fps[0]

	label, err := o.classifier.Label(fp)
	if err != nil {
		return fmt.Errorf("failed to label record")
	}

	if err := re.saveLabel(label); err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	event := makeEvent(e.HostID, o.ID(), label.ID, ON_LABEL)
	if err := o.eventBus(event); err != nil {
		return fmt.Errorf("failed to dispatch event: %w", err)
	}
	return nil
}
