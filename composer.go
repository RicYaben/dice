package dice

import (
	"slices"

	"github.com/pkg/errors"
)

type StageType string

const (
	STAGE_MODULE    = "module"
	STAGE_SIGNATURE = "signature"
)

type Composer interface {
	// Preload signatures or modules
	Stage(t StageType, name ...string) error
}

// A simple registry that holds staged signatures
type registry struct {
	signatures map[uint]*Signature
}

func (r *registry) addSignature(s ...*Signature) {
	for _, sig := range s {
		if _, ok := r.signatures[sig.ID]; !ok {
			r.signatures[sig.ID] = sig
		}
	}
}

func (r *registry) getOrCreateSignature(id uint) *Signature {
	if sig, ok := r.signatures[id]; ok {
		return sig
	}
	sig := &Signature{Name: "-"}
	sig.ID = id
	r.signatures[id] = sig
	return sig
}

type composer struct {
	adapter  ComposerAdapter
	registry *registry
}

func NewComposer(adapter ComposerAdapter) *composer {
	return &composer{adapter, &registry{}}
}

// Preload signatures and modules. This is meant to find
// signatures and register them into a common registry.
func (c *composer) Stage(t StageType, name ...string) error {
	q := "name LIKE ?"
	switch t {
	case STAGE_SIGNATURE:
		sigs := []*Signature{}
		if err := c.adapter.Find(sigs, q, name); err != nil {
			return errors.Wrap(err, "failed to find signatures")
		}
		c.registry.addSignature(sigs...)
		return nil
	case STAGE_MODULE:
		mods := []*Module{}
		if err := c.adapter.Find(mods, q, name); err != nil {
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
func (c *composer) Compose(names []string, nad NodeAdapter) ([]*Component, error) {
	// make a component adapter with notion of a bunch of signatures
	ad := c.adapter.withRegistry(c.registry)
	factory := &componentFactory{
		compAdapter: ad,
		reg:         NewGraphRegistry(ad, nad),
	}
	var comps []*Component
	for _, n := range names {
		comp, err := factory.MakeComponent(n)
		if err != nil {
			return nil, err
		}
		comps = append(comps, comp)
	}
	return comps, nil
}

type componentFactory struct {
	compAdapter ComposerAdapter
	cosmos      CosmosAdapter
	reg         *graphRegistry
}

func (f *componentFactory) MakeComponent(n string) (comp *Component, err error) {
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
	var sigs []*Signature
	if err := f.compAdapter.Find(sigs, &Signature{Component: "identfier"}); err != nil {
		return nil, err
	}

	eps, err := f.reg.entrypoints(sigs)
	if err != nil {
		return nil, err
	}

	return &Component{
		Name:        "identifier",
		Adapter:     f.cosmos,
		Events:      []EventType{SOURCE_EVENT},
		NodesFilter: f.hookedComponentNodes(eps),
	}, nil
}

func (f *componentFactory) makeClassifier() (*Component, error) {
	var sigs []*Signature
	if err := f.compAdapter.Find(sigs, &Signature{Component: "classifier"}); err != nil {
		return nil, err
	}

	eps, err := f.reg.entrypoints(sigs)
	if err != nil {
		return nil, err
	}

	return &Component{
		Name:        "classifier",
		Adapter:     f.cosmos,
		Events:      []EventType{FINGERPRINT_EVENT, HOST_EVENT},
		NodesFilter: f.hookedComponentNodes(eps),
	}, nil
}

func (f *componentFactory) makeScanner() (*Component, error) {
	var sigs []*Signature
	if err := f.compAdapter.Find(sigs, &Signature{Component: "scanner"}); err != nil {
		return nil, err
	}

	eps, err := f.reg.entrypoints(sigs)
	if err != nil {
		return nil, err
	}

	return &Component{
		Name:        "scanner",
		Adapter:     f.cosmos,
		Events:      []EventType{SCAN_EVENT},
		NodesFilter: f.hookedComponentNodes(eps),
	}, nil
}

func (f *componentFactory) hookedComponentNodes(ep []GraphNode) func(Event) []GraphNode {
	hooks := hookedNodesHandler(f.cosmos, f.reg)
	targets := targetNodesHandler(ep)

	return func(e Event) []GraphNode {
		if n := hooks(e); n != nil {
			return n
		}

		if n := targets(e); n != nil {
			return n
		}

		return ep
	}
}

// Returns the list of hooked nodes
func hookedNodesHandler(c CosmosAdapter, r *graphRegistry) func(Event) []GraphNode {
	return func(e Event) []GraphNode {
		hooks, err := c.FindHooks(e.ObjectID)
		if err != nil {
			// cannot recover from not finding hooks on a host
			panic(err)
		}
		nodes := make([]GraphNode, 0, len(hooks))
		for _, h := range hooks {
			if node, ok := r.nodes[h.NodeID]; ok {
				nodes = append(nodes, node)
			}
		}
		return nodes
	}
}

// Filter a list of nodes based on the event targets
func targetNodesHandler(nodes []GraphNode) func(Event) []GraphNode {
	return func(e Event) []GraphNode {
		if len(e.Targets) == 0 {
			return nodes
		}
		return Filter(nodes, func(s GraphNode) bool {
			return slices.Contains(e.Targets, s.Name())
		})
	}
}
