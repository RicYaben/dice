// Here are the collection of adapters. An adapter gives a part of DICE access to a bunch
// of interfaces to communicate with storage, databases, etc.
// They include an event bus to which emitters can listen to.
package dice

import "github.com/dice/shared"

type Adapters interface {
	Engine() EngineAdapter
	Cosmos() CosmosAdapter
	Composer() ComposerAdapter
	Signatures() SignatureAdapter
	Projects() ProjectAdapter
}

type EngineAdapter interface {
	// Add a source
	AddSource(...*Source) error
	// Find sources in the current study
	FindSources([]string, []string) ([]*Source, error)
}

type ComposerAdapter interface {
	// Get a registered signature
	GetSignature(uint) (*Signature, error)
	// Get a registered module
	GetModule(uint) (*Module, error)
	// Get the roots of a registered signature
	GetRoots(uint) ([]*Node, error)

	// TODO: may need a query to search for required signatures
	// recursively.
	// Raw query for signatures
	Find(any, any) error

	// Find signatures in the home directory
	FindSignatures([]string) ([]*Signature, error)
	// Find in the home directory
	FindModules([]string) ([]*Module, error)

	// make a copy with a registry
	withRegistry(registry) ComposerAdapter
}

type CosmosAdapter interface {
	// Add a new Host (creates a host)
	AddHost(...*Host) error
	// Add a fingerprint to a host (creates a fingerprint)
	AddFingerprint(...*Fingerprint) error
	// Add a source of data (creates a source)
	AddSource(...*Source) error
	// Add a new scan (creates a scan)
	AddScan(...*Scan) error
	// Add a new label (creates a label)
	AddLabel(...*Label) error

	GetHost(uint) (*Host, error)
	GetFingerprint(uint) (*Fingerprint, error)
	GetSource(uint) (*Source, error)
	GetScan(uint) (*Scan, error)
	GetLabel(uint) (*Label, error)

	// Return a list of hooks
	FindHooks(uint) ([]*Hook, error)

	// Find source files by their extension by the name of the source
	// in the current study.
	FindSources(n []string, ext []string) ([]*Source, error)

	// Search the cosmos db for some results. Raw queries
	Search(string) ([]byte, error)

	// Search for hosts with criteria
	Query(string) ([]*Host, error)

	WithOrigin(uint) CosmosAdapter
}

// Adapter to manipulate signatures and modules
type SignatureAdapter interface {
	Find(result any, query []interface{}) error
	Remove(query []interface{}) error
	Locate(model any, query []interface{}) ([]Location, error)
	Update() error

	// Loads a local signature.
	AddSignatures(...*Signature) error
	// Loads a local module
	AddModule(*Module) error
	// Get a registered signature
	GetSignature(uint) (*Signature, error)
	// Get a registered module
	GetModule(uint) (*Module, error)

	// Load a module
	LoadModule(Module) (shared.Module, error)
}

type ProjectAdapter interface {
	Find(result any, query interface{}) error
	AddProject(Project) error
	AddStudy(Study) error
}

// A common intreface for most adapters
// to send their events to.
type eventBus func(Event) error

type eventAdapter struct {
	// Originator of the event
	originID uint
	bus      eventBus
}

func (a *eventAdapter) makeEvent(t EventType, id uint) error {
	ev := Event{
		NodeID:   a.originID,
		Type:     t,
		ObjectID: id,
	}
	return a.bus(ev)
}

type engineAdapter struct {
	eventAdapter
	repo *sourceRepo
}

func (a *engineAdapter) AddSource(s ...*Source) error {
	if err := a.repo.addSource(s...); err != nil {
		return err
	}

	for _, so := range s {
		if err := a.makeEvent(SOURCE_EVENT, so.ID); err != nil {
			return err
		}
	}

	return nil
}

func (a *engineAdapter) FindSources(s, ext []string) ([]*Source, error) {
	return a.repo.findSourceFiles(s, ext)
}

type composerAdapter struct {
	registry registry
	repo     *signatureRepo
}

func (a *composerAdapter) GetSignature(id uint) (*Signature, error) {
	if sig, ok := a.registry.signatures[id]; ok {
		return sig, nil
	}

	sig, err := a.repo.getSignature(id)
	if err != nil {
		return nil, err
	}
	a.registry.signatures[sig.ID] = sig
	return sig, nil
}

func (a *composerAdapter) GetModule(id uint) (*Module, error) {
	return a.repo.getModule(id)
}

func (a *composerAdapter) GetRoots(id uint) ([]*Node, error) {
	return a.repo.getRoots(id)
}

func (a *composerAdapter) Find(m any, q any) error {
	return a.repo.find(m, q)
}

func (a *composerAdapter) FindSignatures(globs []string) ([]*Signature, error) {
	return a.repo.findSignatureFiles(globs)
}

func (a *composerAdapter) FindModules(globs []string) ([]*Module, error) {
	return a.repo.findModuleFiles(globs)
}

// TODO: dont know how to finish this
func (a *composerAdapter) Cosmos(id uint) *cosmosAdapter {
	return &cosmosAdapter{}
}

func (a *composerAdapter) withRegistry(r registry) ComposerAdapter {
	return &composerAdapter{r, a.repo}
}

type cosmosAdapter struct {
	eventAdapter
	repo *cosmosRepo
}

func (a *cosmosAdapter) WithOrigin(id uint) CosmosAdapter {
	cp := *a
	cp.eventAdapter.originID = id
	return &cp
}

func (a *cosmosAdapter) AddHost(h ...*Host) error {
	if err := a.repo.addHost(h...); err != nil {
		return err
	}

	for _, host := range h {
		if err := a.makeEvent(HOST_EVENT, host.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *cosmosAdapter) AddFingerprint(f ...*Fingerprint) error {
	if err := a.repo.addFingerprint(f...); err != nil {
		return err
	}

	for _, fp := range f {
		if err := a.makeEvent(FINGERPRINT_EVENT, fp.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *cosmosAdapter) AddSource(s ...*Source) error {
	if err := a.repo.addSource(s...); err != nil {
		return err
	}

	for _, src := range s {
		if err := a.makeEvent(SOURCE_EVENT, src.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *cosmosAdapter) AddLabel(l ...*Label) error {
	if err := a.repo.addLabel(l...); err != nil {
		return err
	}

	for _, lab := range l {
		if err := a.makeEvent(LABEL_EVENT, lab.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *cosmosAdapter) AddScan(s ...*Scan) error {
	if err := a.repo.addScan(s...); err != nil {
		return err
	}

	for _, sc := range s {
		if err := a.makeEvent(SCAN_EVENT, sc.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *cosmosAdapter) GetHost(id uint) (*Host, error) {
	return a.repo.getHost(id)
}

func (a *cosmosAdapter) GetFingerprint(id uint) (*Fingerprint, error) {
	return a.repo.getFingerprint(id)
}

func (a *cosmosAdapter) GetSource(id uint) (*Source, error) {
	return a.repo.getSource(id)
}

func (a *cosmosAdapter) GetLabel(id uint) (*Label, error) {
	return a.repo.getLabel(id)
}

func (a *cosmosAdapter) GetScan(id uint) (*Scan, error) {
	return a.repo.getScan(id)
}

func (a *cosmosAdapter) FindHooks(id uint) ([]*Hook, error) {
	return a.repo.getHooks(id)
}

func (a *cosmosAdapter) FindSources(n []string, ext []string) ([]*Source, error) {
	panic("not implemented yet")
}

func (a *cosmosAdapter) Search(q string) ([]byte, error) {
	return a.repo.search(q)
}

func (a *cosmosAdapter) Query(q string) ([]*Host, error) {
	return a.repo.query(q)
}

type signatureAdapter struct {
	repo *signatureRepo
}

func (a *signatureAdapter) AddSignature(sig Signature) error {
	panic("not implemented yet")
}

func (a *signatureAdapter) AddModule(mod Module) error {
	panic("not implemented yet")
}

func (a *signatureAdapter) GetSignature(id uint) (Signature, error) {
	panic("not implemented yet")
}

func (a *signatureAdapter) GetModule(id uint) (Module, error) {
	panic("not implemented yet")
}

// Adapters factory
type adapterFactory struct {
	eventBus
	repos *repositoryRegistry
}

func MakeAdapters(bus eventBus, conf *Configuration) *adapterFactory {
	return &adapterFactory{
		bus,
		newRepositoryFactory(conf),
	}
}

func (f *adapterFactory) SetConfig(conf *Configuration) *adapterFactory {
	f.repos.conf = conf
	return f
}

func (f *adapterFactory) Cosmos() CosmosAdapter {
	return &cosmosAdapter{
		eventAdapter: eventAdapter{bus: f.eventBus},
		repo:         f.repos.Cosmos(),
	}
}

func (f *adapterFactory) Engine() EngineAdapter {
	return &engineAdapter{
		eventAdapter: eventAdapter{originID: 0xD1CE, bus: f.eventBus},
		repo:         f.repos.Sources(),
	}
}

func (f *adapterFactory) Composer() ComposerAdapter {
	return &composerAdapter{
		registry: registry{},
		repo:     f.repos.Signatures(),
	}
}

func (f *adapterFactory) Signatures() SignatureAdapter {
	panic("not implemented yet")
}

func (f *adapterFactory) Projects() ProjectAdapter {
	panic("not implemented yet")
}
