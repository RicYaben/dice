package dice

import (
	"github.com/pkg/errors"
)

const EngineID uint = 0xD1CE

type engine struct {
	conf     Configuration
	adapters *adapterFactory
	emitter  Emitter
}

func baseEngine() *engine {
	conf := baseConfig()
	return &engine{
		conf:     conf,
		adapters: newAdapterFactory(conf.paths.STATE_HOME),
		emitter:  newEventEmitter(),
	}
}

// Configures an engine with the current setup.
func LoadEngine(eng *engine, conf Configuration, flags EngineSetupFlags) error {
	ad := newAdapterFactory(conf.paths.STATE_HOME)
	composer := newComposer(ad.cosmosAdapter(), ad.signaturesAdapter()) // composer loads components from signatures
	components, err := composer.Compose(flags.Signatures, flags.Actions)
	if err != nil {
		return errors.Wrap(err, "failed to load signatures")
	}

	// register the components into the emitter
	em := newEventEmitter()
	for _, comp := range components {
		em.subscribe(comp)
	}

	eng.emitter = em
	eng.conf = conf
	eng.adapters = ad
	return nil
}

// NOTE: this does not go here, move this to the emitter or the repository registry

// Finds sources within a scan in the current project workspace
// There may be more than one
// func (e *engine) findSources(scan string, names, globs []string) ([]*SourceModel, error) {
// 	var srcs []*SourceModel

// 	// Location of the scan
// 	for _, name := range names {
// 		// Scan source location
// 		sr, err := e.conn.sources.find(scan, name, globs)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "failed to locate source")
// 		}
// 		srcs = append(srcs, sr...)
// 	}
// 	return srcs, nil
// }

// Entrypoint to start the engine.
// Run takes sources into an event emitter,
// and push them into the respective components.
func (e *engine) Run(sources []*SourceModel) error {
	// we make an engine adapter so adding sources becomes a simpler task,
	// with the handler already in place
	adapter := e.adapters.engineAdapter()
	for _, src := range sources {
		if err := adapter.addSource(src); err != nil {
			return errors.Wrapf(err, "failed to consume events from source %v", src)
		}
	}
	return nil
}

// rapidly go through a source to check whether DICE can work with it (a sanity check)
// func (s *Source) check(workers int) error {/* check if the source can be parsed */ return nil}

// These handlers are meant to seat between the module and the plugin as an easy interface to add new things
type Resource interface {
	SourceModel | Label | Fingerprint | Host
}

type adapter[T Resource] struct {
	add func(*T) error
	get func(...uint) ([]*T, error)
}

type connectorCallback[T Resource] func(node uint) adapter[T]
type connectorHandler[T Resource] func(repos *repositoryRegistry, emit func(Event) error) connectorCallback[T]

func sourcesHandler(repos *repositoryRegistry, emit func(Event) error) connectorCallback[SourceModel] {
	repo := repos.Sources()
	return func(node uint) adapter[SourceModel] {
		add := func(s *SourceModel) error {
			if err := repo.addSource(s); err != nil {
				return err
			}

			return emit(Event{
				NodeID:    node,
				ObjectID:  s.ID,
				EventType: SOURCE_EVENT,
			})
		}

		get := func(ids ...uint) ([]*SourceModel, error) {
			return repo.getSources(ids...)
		}
		return adapter[SourceModel]{add, get}
	}
}

func labelsHandler(repos *repositoryRegistry, emit func(Event) error) connectorCallback[Label] {
	repo := repos.Cosmos()
	return func(node uint) adapter[Label] {
		add := func(l *Label) error {
			if err := repo.addLabel(l); err != nil {
				return err
			}
			return emit(Event{
				NodeID:    node,
				ObjectID:  l.ID,
				EventType: LABEL_EVENT,
			})
		}

		get := func(ids ...uint) ([]*Label, error) {
			return repo.getLabels(ids...)
		}
		return adapter[Label]{add, get}
	}
}

func fingerprintsHandler(repos *repositoryRegistry, emit func(Event) error) connectorCallback[Fingerprint] {
	repo := repos.Cosmos()
	return func(node uint) adapter[Fingerprint] {
		add := func(f *Fingerprint) error {
			if err := repo.addFingerprint(f); err != nil {
				return err
			}
			return emit(Event{
				NodeID:    node,
				ObjectID:  f.ID,
				EventType: FINGERPRINT_EVENT,
			})
		}
		get := func(ids ...uint) ([]*Fingerprint, error) {
			return repo.getFingerprints(ids...)
		}
		return adapter[Fingerprint]{add, get}
	}
}

func hostsHandler(repos *repositoryRegistry, emit func(Event) error) connectorCallback[Host] {
	repo := repos.Cosmos()
	return func(node uint) adapter[Host] {
		add := func(h *Host) error {
			if err := repo.addHost(h); err != nil {
				return err
			}
			return emit(Event{
				NodeID:    node,
				ObjectID:  h.ID,
				EventType: HOST_EVENT,
			})
		}
		get := func(ids ...uint) ([]*Host, error) {
			return repo.getHosts(ids...)
		}
		return adapter[Host]{add, get}
	}
}
