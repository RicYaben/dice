package dice

import (
	"github.com/pkg/errors"
)

// A simple registry that holds staged signatures
type registry struct {
	signatures map[uint]Signature
}

func (r *registry) addSignature(s ...Signature) {
	for _, sig := range s {
		if _, ok := r.signatures[sig.ID]; !ok {
			r.signatures[sig.ID] = sig
		}
	}
}

func (r *registry) getOrCreateSignature(id uint) Signature {
	if sig, ok := r.signatures[id]; ok {
		return sig
	}
	sig := Signature{Name: "-"}
	sig.ID = id
	r.signatures[id] = sig
	return sig
}

type composer struct {
	adapter  ComposerAdapter
	registry registry
}

func NewComposer(adapter ComposerAdapter) *composer {
	return &composer{adapter, registry{}}
}

// Preload signatures and modules. This is meant to find
// signatures and register them into a common registry.
func (c *composer) Stage(t string, names []string) error {
	switch t {
	case "signatures":
		sigs, err := c.adapter.FindSignatures(names)
		if err != nil {
			return errors.Wrap(err, "failed to find signatures")
		}
		c.registry.addSignature(sigs...)
		return nil
	case "modules":
		mods, err := c.adapter.FindModules(names)
		if err != nil {
			return errors.Wrap(err, "failed to find modules")
		}
		sig := c.registry.getOrCreateSignature(0)

		// add modules to the signature
		// this converts the modules into nodes
		for _, mod := range mods {
			node := &Node{
				SignatureID: sig.ID,
				Type:        MODULE_NODE,
				ObjectID:    mod.ID,
			}
			sig.Nodes = append(sig.Nodes, node)
		}
		return nil
	}
	return errors.Errorf("unknown type %s", t)
}

// Compose components from signatures and actions. Only load named comps.
func (c *composer) Compose(names []string) ([]*Component, error) {
	// make a component adapter with notion of a bunch of signatures
	ad := c.adapter.withRegistry(c.registry)
	factory := &componentFactory{
		sigs: ad,
		reg:  newGraphRegistry(ad),
	}
	var comps []*Component
	for _, n := range names {
		comp, err := factory.makeComponent(n)
		if err != nil {
			return nil, err
		}
		comps = append(comps, comp)
	}
	return comps, nil
}

type componentFactory struct {
	sigs ComposerAdapter
	reg  *graphRegistry
}

func (f *componentFactory) makeComponent(n string) (comp *Component, err error) {
	switch n {
	case "identifier":
		comp, err = f.makeIdentifier()
	case "classifier":
		comp, err = f.makeClassifier()
	case "scanner":
		comp, err = f.makeScanner()
	default:
		return nil, errors.Errorf("component not found %s", n)
	}
	return
}

func (f *componentFactory) makeIdentifier() (*Component, error) {
	sigs := f.sigs.SearchSingatures(Signature{Type: "identfier"})
	g, err := f.reg.makeGraphs(sigs)
	if err != nil {
		return nil, err
	}

	return &Component{
		Name:   "Identifier",
		Events: []EventType{SOURCE_EVENT},
		Graphs: g,
	}, nil
}

func (f *componentFactory) makeClassifier() (*Component, error) {
	sigs := f.sigs.SearchSingatures(Signature{Type: "classifier"})
	g, err := f.reg.makeGraphs(sigs)
	if err != nil {
		return nil, err
	}

	return &Component{
		Name:   "Classifier",
		Events: []EventType{FINGERPRINT_EVENT, HOST_EVENT},
		Graphs: g,
	}, nil
}

func (f *componentFactory) makeScanner() (*Component, error) {
	sigs := f.sigs.SearchSingatures(Signature{Type: "scanner"})
	g, err := f.reg.makeGraphs(sigs)
	if err != nil {
		return nil, err
	}

	return &Component{
		Name:   "Scanner",
		Events: []EventType{SCAN_EVENT},
		Graphs: g,
	}, nil
}
