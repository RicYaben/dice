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
	factory := &componentFactory{c.adapter.withRegistry(c.registry)}
	var comps []*Component
	for _, n := range names {
		if cmp := factory.makeComponent(n); cmp != nil {
			comps = append(comps, cmp)
		}
	}
	return comps, nil
}

type componentFactory struct {
	sigs ComposerAdapter
}

func (f *componentFactory) makeComponent(n string) *Component {
	switch n {
	case "identifier":
		return f.makeIdentifier()
	case "classifier":
		return f.makeClassifier()
	case "scanner":
		return f.makeScanner()
	default:
		return nil
	}
}

func (f *componentFactory) makeIdentifier() *Component {
	sigs := f.sigs.SearchSingatures(Signature{Type: "identfier"})
	return &Component{
		Name:       "Identifier",
		Handler:    componentEventHandler,
		Events:     []EventType{SOURCE_EVENT},
		Signatures: sigs,
	}
}

func (f *componentFactory) makeClassifier() *Component {
	sigs := f.sigs.SearchSingatures(Signature{Type: "classifier"})
	return &Component{
		Name:       "Classifier",
		Handler:    componentEventHandler,
		Events:     []EventType{FINGERPRINT_EVENT, HOST_EVENT},
		Signatures: sigs,
	}
}

func (f *componentFactory) makeScanner() *Component {
	sigs := f.sigs.SearchSingatures(Signature{Type: "scanner"})
	return &Component{
		Name:       "Scanner",
		Handler:    componentEventHandler,
		Events:     []EventType{SCAN_EVENT},
		Signatures: sigs,
	}
}
