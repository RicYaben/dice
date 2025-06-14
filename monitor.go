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

type Monitor interface {
	// adds a subject
	addSubject(Subject) error
	// returns a subject
	getSubject(id uint) Subject
	// delete a subject
	deleteSubject(id uint)
	// dispatch an event to a subject
	// E.g., adding a label or a record
	dispatch(event hostEvent) error
}

type hostMonitor struct {
	repo     *eventRepo
	subjects map[uint]Subject
}

func (m *hostMonitor) addSubject(id uint, s Subject) {
	m.subjects[id] = s
}

func (m *hostMonitor) getSubject(id uint) Subject {
	return m.subjects[id]
}

func (m *hostMonitor) deleteSubject(id uint) {
	delete(m.subjects, id)
}

func (m *hostMonitor) dispatch(e hostEvent) error {
	if err := m.repo.addEvent(e); err != nil {
		return err
	}

	if sub, ok := m.subjects[e.HostID]; ok {
		return sub.publish(e)
	}
	return fmt.Errorf("subject not registered %d", e.HostID)
}

type Subject interface {
	ID() uint
	// register an observer
	register(obs Observer) error
	// remove an observer
	deregister(obs Observer)
	// publish an event
	publish(e hostEvent) error
	// notify a single observer
	notify(event hostEvent, obs ...Observer) error
	// notify all observers
	notifyAll(event hostEvent) error
}

type hostSubject struct {
	id        uint
	repo      *recordRepo
	observers []Observer
	mu        sync.Mutex
}

func (s *hostSubject) ID() uint {
	return s.id
}

func (s *hostSubject) registered(obs Observer) bool {
	for _, ob := range s.observers {
		if ob == obs {
			return true
		}
	}
	return false
}

func (s *hostSubject) doAfterRegister(obs Observer) error {
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

func (s *hostSubject) register(obs Observer) error {
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
func (s *hostSubject) deregister(obs Observer) {
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

func (s *hostSubject) getObserver(id uint) Observer {
	for _, obs := range s.observers {
		if obs.ID() == id {
			return obs
		}
	}
	return nil
}

func (s *hostSubject) publish(e hostEvent) error {
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
func (s *hostSubject) notify(e hostEvent, obs ...Observer) error {
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

func (s *hostSubject) notifyAll(e hostEvent) error {
	for _, sub := range s.observers {
		if err := sub.update(e); err != nil {
			return err
		}
	}
	return nil
}

type Observer interface {
	// returns the ID of the observer
	ID() uint
	// handles an update in a subject
	update(event hostEvent) error
	// returns the children attached to this observer
	getChildren() []Observer
	// process an event
	process(e hostEvent) error
	// stops the module
	kill()
}

type nodeObserver struct {
	// Node attributes, we don't need to hold the whole node
	id       uint
	children []Observer
	module   *dicePlugin

	handler  eventsHandler
	repo     *hostRepo
	eventBus func(hostEvent) error
}

func (o *nodeObserver) update(e hostEvent) error {
	return o.handler.handle(e)
}

func (o *nodeObserver) ID() uint {
	return o.id
}

// return the registered observers
// observers are loaded elsewhere, the observer does
// not know how to make new ones.
func (o *nodeObserver) getChildren() []Observer {
	return o.children
}

func (o *nodeObserver) kill() {
	o.module.client.Kill()
}

type scanObserver struct {
	nodeObserver
	scanner ScanPlugin
}

func (o *scanObserver) process(e hostEvent) error {
	return o.repo.withTransaction(func(repo Repository) error {
		re := repo.(*hostRepo)
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
	})
}

type ruleObserver struct {
	nodeObserver
	classifier RulePlugin
}

func (o *ruleObserver) process(e hostEvent) error {
	return o.repo.withTransaction(func(repo Repository) error {
		re := repo.(*hostRepo)
		record, err := re.getRecord(e.ObjectID)
		if err != nil {
			return fmt.Errorf("could not find record %d", e.ObjectID)
		}

		if err := re.saveMark(Mark{RecordID: e.ObjectID, NodeID: o.ID()}); err != nil {
			return fmt.Errorf("failed to mark record: %w", err)
		}

		label, err := o.classifier.Label(record)
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
	})
}
