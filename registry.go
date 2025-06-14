package dice

import (
	"errors"
	"fmt"
)

var (
	ErrCycle error = errors.New("graph contains a cycle")
)

type WrappedNode interface {
	getNode() *Node
	addChild(WrappedNode)
}

type wrappedNode struct {
	node *Node
}

func (n *wrappedNode) getNode() *Node {
	return n.node
}

func (n *wrappedNode) addChild(c WrappedNode) {
	// append the node to the childs
	n.node.Childs = append(n.node.Childs, c.getNode())
}

// A special type of node. It contains a colored node as the root,
// which is a reference to some graph node, and the embedded graph
// as a child of the root. Adding a child to this type of node links
// it to the exit nodes of the graph.
type wrappedGraphNode struct {
	node  *wrappedNode
	graph *graph
}

func (n *wrappedGraphNode) getNode() *Node {
	return n.node.getNode()
}

func (n *wrappedGraphNode) addChild(c WrappedNode) {
	for _, eNode := range n.graph.leafs {
		eNode.addChild(c)
	}
}

// Node creation options. This is here so we can create multiple
// nodes with the same options.
type nodeOptions struct {
	SignatureID uint
	Type        string
}

// A collection of nodes is a graph
type graph struct {
	registry   *graphRegistry
	signature  *Signature
	registered map[uint]WrappedNode

	roots []WrappedNode
	leafs []WrappedNode
}

func newGraph(sig *Signature, registry *graphRegistry) *graph {
	return &graph{
		signature:  sig,
		registry:   registry,
		registered: make(map[uint]WrappedNode),
		roots:      make([]WrappedNode, 0),
		leafs:      make([]WrappedNode, 0),
	}
}

func (g *graph) init() error {
	options := nodeOptions{
		SignatureID: g.signature.ID,
	}

	if err := g.loadNodes(options); err != nil {
		return err
	}
	return nil
}

func (g *graph) loadNodes(options nodeOptions) error {
	options.Type = "rule"
	for _, rule := range g.signature.Rules {
		if _, err := g.getOrCreateRuleNode(rule, options); err != nil {
			return err
		}
	}

	options.Type = "scan"
	for _, scan := range g.signature.Scans {
		if _, err := g.getOrCreateScanNode(scan, options); err != nil {
			return err
		}
	}

	// Add exit nodes - those without childs
	// For now, we only allow rule nodes to be leafs.
	for _, node := range g.roots {
		n := node.getNode()
		if n.Type == "rule" && len(n.Childs) == 0 {
			g.leafs = append(g.leafs, node)
		}
	}
	return nil
}

func (g *graph) register(n WrappedNode) {
	g.registered[n.getNode().ID] = n
}

func (g *graph) getNode(id uint) WrappedNode {
	return g.registered[id]
}

func (g *graph) makeNode(id uint, options nodeOptions) *wrappedNode {
	return &wrappedNode{
		node: &Node{
			ID:          id,
			SignatureID: options.SignatureID,
			Type:        options.Type,
		},
	}
}

func (g *graph) makeGraphNode(rule *Rule, options nodeOptions) (*wrappedGraphNode, error) {
	node := g.makeNode(rule.ID, options)
	eGraph, err := g.registry.getOrCreateGraph(rule.Source)
	if err != nil {
		return nil, err
	}

	// add the graph nodes as childs for this node
	for _, n := range eGraph.roots {
		node.addChild(n)
	}

	gNode := &wrappedGraphNode{
		node:  node,
		graph: eGraph,
	}
	return gNode, nil
}

func (g *graph) getOrCreateScanNode(scan *Scan, options nodeOptions) (WrappedNode, error) {
	// if already registered, return a copy
	if node := g.getNode(scan.ID); node != nil {
		return node, nil
	}
	node := g.makeNode(scan.ID, options)
	g.register(node)

	// this is a root node
	if len(scan.Rules) == 0 {
		g.roots = append(g.roots, node)
	}

	// check who this node tracks
	for _, rule := range scan.Rules {
		rNode, err := g.getOrCreateRuleNode(rule, options)
		if err != nil {
			return nil, err
		}
		rNode.addChild(node)
	}
	return node, nil
}

func (g *graph) getOrCreateRuleNode(rule *Rule, options nodeOptions) (WrappedNode, error) {
	// if already registered, return a copy
	if node := g.getNode(rule.ID); node != nil {
		return node, nil
	}

	// check is signature
	var node WrappedNode
	switch RuleMode(rule.Mode) {
	case R_MODULE:
		n, err := g.makeGraphNode(rule, options)
		if err != nil {
			return nil, err
		}
		node = n
	case R_SOURCE:
		node = g.makeNode(rule.ID, options)
	}
	g.register(node)

	// this is a root node
	if len(rule.Track) == 0 {
		g.roots = append(g.roots, node)
	}

	for _, tRule := range rule.Track {
		tNode, err := g.getOrCreateRuleNode(tRule, options)
		if err != nil {
			return nil, err
		}
		tNode.addChild(node)
	}
	return node, nil
}

type graphRegistry struct {
	repo    *signatureRepo
	loading map[string]interface{}
	graphs  map[string]*graph
}

func (r *graphRegistry) loadGraph(name string) (*graph, error) {
	if _, ok := r.loading[name]; ok {
		return nil, ErrCycle
	}
	r.loading[name] = struct{}{}
	defer delete(r.loading, name)

	sig, err := r.repo.getSignature(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load signature: %w", err)
	}

	g := newGraph(sig, r)
	if err := g.init(); err != nil {
		return nil, err
	}
	r.graphs[name] = g
	return g, nil
}

func (r *graphRegistry) getGraph(name string) *graph {
	return r.graphs[name]
}

func (r *graphRegistry) getOrCreateGraph(name string) (*graph, error) {
	if sig := r.getGraph(name); sig != nil {
		return sig, nil
	}
	return r.loadGraph(name)
}
