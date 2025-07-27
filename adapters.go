// Here are the collection of adapters. An adapter gives a part of DICE access to a bunch
// of interfaces to communicate with storage, databases, etc.
// They include an event bus to which emitters can listen to.
package dice

import (
	"os"
	"slices"

	"github.com/pkg/errors"
)

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

	Find(any, any) error

	// make a copy with a registry
	withRegistry(*registry) ComposerAdapter
}

type NodeAdapter interface {
	AddHost(...*Host) error
	AddFingerprint(...*Fingerprint) error
	AddSource(...*Source) error
	AddScan(...*Scan) error
	AddLabel(...*Label) error

	GetHost(uint) (*Host, error)
	GetFingerprint(uint) (*Fingerprint, error)
	GetSource(uint) (*Source, error)
	GetScan(uint) (*Scan, error)
	GetLabel(uint) (*Label, error)

	Query(string) ([]*Host, error)

	withOriginID(uint) NodeAdapter
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
	FindSources(s, ext []string) ([]*Source, error)

	// Search the cosmos db for some results. Raw queries
	Find(m any, q any) error

	// Search for hosts with criteria
	Query(string) ([]*Host, error)
}

// Adapter to manipulate signatures and modules
type SignatureAdapter interface {
	// Loads a local signature.
	AddSignature(...*Signature) error
	// Loads a local module
	AddModule(...*Module) error

	// Get a registered signature
	GetSignature(uint) (*Signature, error)
	// Get a registered module
	GetModule(uint) (*Module, error)

	// Find signatures in the home directory
	AddMissingSignatures(...string) ([]*Signature, error)
	// Find in the home directory
	AddMissingModules(...string) ([]*Module, error)

	Find(m any, query any) error
	Remove(m any, query any) error
	Update() error
}

type ProjectAdapter interface {
	Find(m any, query any) error
	AddProject(...*Project) error
	AddStudy(...*Study) error
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
	registry *registry
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

// TODO: dont know how to finish this
func (a *composerAdapter) Cosmos(id uint) *cosmosAdapter {
	return &cosmosAdapter{}
}

func (a *composerAdapter) withRegistry(r *registry) ComposerAdapter {
	return &composerAdapter{r, a.repo}
}

type nodeAdapter struct {
	*cosmosAdapter
	originID uint
}

func (a *nodeAdapter) withOriginID(id uint) NodeAdapter {
	cp := *a
	cp.eventAdapter.originID = id
	return &cp
}

type cosmosAdapter struct {
	eventAdapter
	repo    *cosmosRepo
	sources *sourceRepo
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

func (a *cosmosAdapter) FindSources(s, ext []string) ([]*Source, error) {
	return a.sources.findSourceFiles(s, ext)
}

func (a *cosmosAdapter) Find(m any, q any) error {
	return a.repo.find(m, q)
}

func (a *cosmosAdapter) Query(q string) ([]*Host, error) {
	return a.repo.query(q)
}

type signatureAdapter struct {
	registry *registry
	repo     *signatureRepo
}

func (a *signatureAdapter) AddSignature(sig ...*Signature) error {
	return a.repo.addSignature(sig...)
}

func (a *signatureAdapter) AddModule(mod ...*Module) error {
	return a.repo.addModule(mod...)
}

func (a *signatureAdapter) GetSignature(id uint) (*Signature, error) {
	return a.repo.getSignature(id)
}

func (a *signatureAdapter) GetModule(id uint) (*Module, error) {
	return a.repo.getModule(id)
}

func (a *signatureAdapter) AddMissingSignatures(g ...string) ([]*Signature, error) {
	fpaths, err := a.repo.findFiles("signature", g)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find files")
	}

	var names []string
	for _, fpath := range fpaths {
		info, err := os.Stat(fpath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get signature file stats")
		}
		names = append(names, info.Name())
	}

	sigs := []*Signature{}
	q := []string{"name IN ?"}
	if err := a.repo.find(sigs, append(q, names...)); err != nil {
		return nil, errors.Wrap(err, "failed to find signatures")
	}

	var sigNames []string
	for _, m := range sigs {
		sigNames = append(sigNames, m.Name)
	}

	var mSigs []*Signature
	for i, n := range names {
		if slices.Contains(sigNames, n) {
			continue
		}

		sig, err := a.repo.parseSignatureFile(fpaths[i])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse signature")
		}
		mSigs = append(mSigs, sig)
	}

	if err := a.repo.addSignature(mSigs...); err != nil {
		return nil, errors.Wrap(err, "failed to register signature(s)")
	}

	return mSigs, nil
}

func (a *signatureAdapter) AddMissingModules(g ...string) ([]*Module, error) {

	fpaths, err := a.repo.findFiles("module", g)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find files")
	}

	var names []string
	for _, fpath := range fpaths {
		info, err := os.Stat(fpath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get module file stats")
		}
		names = append(names, info.Name())
	}

	mods := []*Module{}
	q := []string{"name IN ?"}
	if err := a.repo.find(mods, append(q, names...)); err != nil {
		return nil, errors.Wrap(err, "failed to find modules")
	}

	var modNames []string
	for _, m := range mods {
		modNames = append(modNames, m.Name)
	}

	var mMods []*Module
	for i, n := range names {
		if slices.Contains(modNames, n) {
			continue
		}

		m := Module{
			Name:     n,
			Location: fpaths[i],
		}

		mod, err := LoadModule(m.Name, m.Location)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load module")
		}

		banner, err := mod.Properties()
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve module banner")
		}
		m.Properties = banner
		mMods = append(mMods, &m)
	}

	if err := a.repo.addModule(mMods...); err != nil {
		return nil, errors.Wrap(err, "failed to add module(s)")
	}

	return mMods, nil
}

func (a *signatureAdapter) Remove(m any, q any) error {
	return a.repo.remove(m, q)
}

func (a *signatureAdapter) Find(m any, q any) error {
	return a.repo.find(m, q)
}

func (a *signatureAdapter) Update() error {
	if err := a.repo.deleteAll(); err != nil {
		return errors.Wrap(err, "failed to clear database")
	}

	if _, err := a.AddMissingModules("*"); err != nil {
		return errors.Wrap(err, "failed to add missing modules")
	}

	if _, err := a.AddMissingSignatures("*"); err != nil {
		return errors.Wrap(err, "failed to add missing signatures")
	}

	return nil
}

type projectAdapter struct {
	repo *projectRepo
}

func (a *projectAdapter) AddProject(p ...*Project) error {
	return a.repo.addProject(p...)
}

func (a *projectAdapter) AddStudy(s ...*Study) error {
	return a.repo.addStudy(s...)
}

func (a *projectAdapter) Find(m any, q any) error {
	return a.repo.find(m, q)
}

// Adapters factory
type adapterFactory struct {
	eventBus
	registry *registry
	repos    *repositoryRegistry
}

func MakeAdapters(bus eventBus, conf *Configuration) *adapterFactory {
	return &adapterFactory{
		bus,
		&registry{},
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
		sources:      f.repos.Sources(),
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
		registry: f.registry,
		repo:     f.repos.Signatures(),
	}
}

func (f *adapterFactory) Signatures() SignatureAdapter {
	return &signatureAdapter{
		registry: f.registry,
		repo:     f.repos.Signatures(),
	}
}

func (f *adapterFactory) Projects() ProjectAdapter {
	return &projectAdapter{
		repo: f.repos.Projects(),
	}
}

func (f *adapterFactory) Nodes() NodeAdapter {
	return &nodeAdapter{
		cosmosAdapter: &cosmosAdapter{
			eventAdapter: eventAdapter{bus: f.eventBus},
			repo:         f.repos.Cosmos(),
			sources:      f.repos.Sources(),
		},
	}
}
