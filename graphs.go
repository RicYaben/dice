package dice

import (
	"slices"

	"github.com/pkg/errors"
)

// A graph is the result of loading and linking a signature and its nodes
type Graph interface {
	// Receives an event and sends it to its nodes
	Update(Event) error
}

// Implementation of linked and prepared signature nodes.
// The difference with Nodes stores in the database is that these
// nodes are already loaded, linked, and prepared to receive events.
type GraphNode interface {
	// Handle an event
	Update(Event) error
	// add a dependant children
	addChild(GraphNode)
}

// Implementation of Graph
type graph struct {
	signature Signature

	leafs []GraphNode
	nodes []GraphNode
}

// TODO: continue here
func (g *graph) Update(e Event) error {
	panic("not implemented yet")
}

type graphNode struct {
	ID       uint
	children []GraphNode
	module   Module
	// TODO: this can be a wrapped callback to some connector,
	// come back to this!
	// Adapter CosmosAdapter
	handler func(node *graphNode, e Event) error
	plugin  *dicePlugin
}

func (n *graphNode) Update(e Event) error {
	return n.handler(n, e)
}

func (n *graphNode) propagate(e Event) error {
	for _, ch := range n.children {
		if err := ch.Update(e); err != nil {
			return err
		}
	}
	return nil
}

func (n *graphNode) addChild(gnode GraphNode) {
	// no need for duplicate nodes.
	// That would create double edges between nodes
	if !slices.Contains(n.children, gnode) {
		n.children = append(n.children, gnode)
	}
}

type embeddedGraphNode struct {
	ID    uint
	graph *graph
}

func (n *embeddedGraphNode) Update(e Event) error {
	return n.graph.Update(e)
}

func (n *embeddedGraphNode) addChild(gnode GraphNode) {
	for _, leaf := range n.graph.leafs {
		leaf.addChild(gnode)
	}
}

// A registry to hold all nodes, loading signatures, and loaded graphs
type graphRegistry struct {
	adapter ComposerAdapter

	// currently loading graphs and nodes
	loadingSigs  map[uint]struct{}
	loadingNodes map[uint]struct{}

	// Graphs already loaded, including embedded
	// This is used to avoid loading the same graph twice
	graphs map[uint]*graph
	// Nodes loaded
	nodes map[uint]GraphNode
}

func newGraphRegistry(ad ComposerAdapter) *graphRegistry {
	return &graphRegistry{
		adapter:      ad,
		loadingSigs:  make(map[uint]struct{}),
		loadingNodes: make(map[uint]struct{}),
		graphs:       make(map[uint]*graph),
		nodes:        make(map[uint]GraphNode),
	}
}

func (r *graphRegistry) makeGraphs(sigs []Signature) ([]*graph, error) {
	var graphs []*graph
	for _, sig := range sigs {
		g, err := r.makeGrpah(sig)
		if err != nil {
			return nil, err
		}
		graphs = append(graphs, g)
	}
	return graphs, nil
}

func (r *graphRegistry) makeGrpah(sig Signature) (*graph, error) {
	// check if we are still loading the signature, i.e., on the path,
	// but not yet registered
	if _, ok := r.loadingSigs[sig.ID]; ok {
		return nil, errors.Errorf("failed to construct graph. Signature %s cotnains a cycle", sig.Name)
	}
	defer delete(r.loadingSigs, sig.ID)
	r.loadingSigs[sig.ID] = struct{}{}

	// check if the graph is already registered
	if g := r.graphs[sig.ID]; g != nil {
		return g, nil
	}

	// create the graph
	g := graph{signature: sig}

	// only the roots
	roots, err := r.adapter.GetRoots(sig.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find signature roots")
	}

	for _, node := range roots {
		gnode, err := r.makeNode(node)
		if err != nil {
			return nil, errors.Wrap(err, "failed to make node")
		}
		g.nodes = append(g.nodes, gnode)
	}

	// find the leafs and add them to the signature
	// leafs are needed to link children of this node
	for _, node := range sig.Nodes {
		if len(node.Children) == 0 {
			leaf := r.nodes[node.ID]
			if leaf == nil {
				return nil, errors.Errorf("leaf node never loaded %d in signature %s", node.ID, sig.Name)
			}
			g.leafs = append(g.leafs, leaf)
		}
	}

	ret := &g
	r.graphs[sig.ID] = ret
	return ret, nil
}

// Creates a graph node from a regular signature node
func (r *graphRegistry) makeNode(node *Node) (GraphNode, error) {
	// check if we are still loading this node. This means there is a cycle
	if _, ok := r.loadingNodes[node.ID]; ok {
		return nil, errors.Errorf("failed to make node. Found a cycle in signature %d, node %d", node.SignatureID, node.ID)
	}
	defer delete(r.loadingNodes, node.ID)
	r.loadingNodes[node.ID] = struct{}{}

	// check if the node is already registered
	if gnode := r.nodes[node.ID]; gnode != nil {
		return gnode, nil
	}

	// make the node
	var (
		gnode GraphNode
		err   error
	)

	switch node.Type {
	case MODULE_NODE:
		gnode, err = r.makeGraphNode(node)
	case SIGNATURE_NODE:
		gnode, err = r.makeEmbeddedGraphNode(node)
	default:
		err = errors.Errorf("failed to make graph node from node %v", node)
	}

	if err != nil {
		return nil, err
	}

	// Iterate its children and do the same
	for _, ch := range node.Children {
		chNode, err := r.makeNode(ch)
		if err != nil {
			return nil, errors.Wrap(err, "failed to make child node")
		}
		gnode.addChild(chNode)
	}

	// NOTE: register both embedded graphs and other nodes.
	// both can be exit or root nodes!
	// Example signature:
	// sig root				<-- root
	// mod mod1 (sig: root)
	// sig exit (mod: mod1)	<-- exit
	r.nodes[node.ID] = gnode
	return gnode, nil
}

// makes a new graph module node
func (r *graphRegistry) makeGraphNode(node *Node) (*graphNode, error) {
	nmod, err := r.adapter.GetModule(node.ObjectID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find module %d", node.ObjectID)
	}

	gnode := &graphNode{
		ID:     node.ID,
		module: nmod,
	}

	return gnode, nil
}

// makes a new embedded graph node
func (r *graphRegistry) makeEmbeddedGraphNode(node *Node) (*embeddedGraphNode, error) {
	nsig, err := r.adapter.GetSignature(node.ObjectID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find signature %d", node.ObjectID)
	}

	egraph, err := r.makeGrpah(nsig)
	if err != nil {
		return nil, err
	}

	gnode := &embeddedGraphNode{
		ID:    node.ID,
		graph: egraph,
	}

	return gnode, nil
}
