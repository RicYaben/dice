// Parser includes a simple parsing function to translate signature syntax into
// raw signatures. NOTE: it does NOT
package dice

import (
	"io"
	"slices"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/dice/pkg/ast"
	"github.com/pkg/errors"
)

type Parser interface {
	// Parse signature syntax into a raw signature. NOTE: does not
	// validate the signature, does not load modules, and does not
	// check for validity. It is a simple translating function.
	// Use a dice.SignaturesAdapter instead.
	Parse(string, io.Reader) (*Signature, error)
}

type parser struct {
	parser *participle.Parser[ast.Signature]
}

func NewParser() *parser {
	p := participle.MustBuild[ast.Signature](
		participle.Unquote("String"),
		participle.Union[ast.Value](ast.String{}, ast.Number{}, ast.List{}),
	)

	return &parser{parser: p}
}

func (p *parser) Parse(fname string, r io.Reader) (*Signature, error) {
	s, err := p.parser.Parse(fname, r)
	if err != nil {
		return nil, err
	}

	sig := Signature{
		Name:      strings.TrimSuffix(fname, ".dice"),
		Component: "classifier",
	}

	if err := p.bind(s, &sig); err != nil {
		return nil, errors.Wrapf(err, "failed to bind signature '%s'", fname)
	}
	return &sig, nil
}

// Converts the AST signature to an actual Signature object
func (p *parser) bind(s *ast.Signature, sig *Signature) error {
	// get the type of the signature. ignore the rest
	for _, prop := range s.Properties {
		switch prop.Key {
		case "component":
			if v, ok := prop.Value.(ast.String); ok {
				sig.Component = v.String
			}
		}
	}

	// bind the nodes
	nodes, err := p.collectNodes(s.Nodes)
	if err != nil {
		return err
	}
	sig.Nodes = nodes
	return nil
}

// Convert an ast.Node to a regular Node
// NOTE: this process does NOT validate the nodes!
func (p *parser) collectNodes(nodes []*ast.Node) ([]*Node, error) {
	nList := []*Node{}
	for _, n := range nodes {
		var node Node
		if err := p.bindNode(n, &node, nList); err != nil {
			return nil, err
		}
		nList = append(nList, &node)
	}
	return nList, nil
}

func (p *parser) bindNode(n *ast.Node, node *Node, reg []*Node) error {
	node.name = n.Name

	switch n.Type {
	case "mod":
		node.Type = MODULE_NODE
	case "sig":
		node.Type = SIGNATURE_NODE
	default:
		return errors.Errorf("failed to bind node type '%s'", node.Type)
	}

	if err := p.bindNodeAttributes(n.Attributes, node, reg); err != nil {
		return errors.Wrapf(err, "failed to bind node '%s' attributes", n.Name)
	}

	return nil
}

func (p *parser) bindNodeAttributes(attrs []*ast.Attribute, node *Node, nodes []*Node) error {
	for _, attr := range attrs {
		if val, ok := attr.Value.(ast.List); ok {
			// ignore the rest for now
			if !isNodeType(attr.Key) {
				continue
			}

			t := attrToNodeType(attr.Key)
			if err := p.bindNodeParents(t, val.List, node, nodes); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *parser) bindNodeParents(t NodeType, names []string, n *Node, reg []*Node) error {
	slices.Sort(names)
	filtNodes := Filter(reg, func(n *Node) bool { return n.Type == t })
	slices.SortFunc(filtNodes, func(a *Node, b *Node) int {
		if a.name < b.name {
			return -1
		}
		if a.name > b.name {
			return 1
		}
		return 0
	})

	lastIndex := 0
	for _, name := range names {
		var parent *Node
		for j := lastIndex; lastIndex < len(filtNodes); j++ {
			node := filtNodes[j]
			if node.name != name {
				continue
			}
			parent = node
			lastIndex = j
			break
		}

		if parent != nil {
			parent.Children = append(parent.Children, n)
			continue
		}
		return errors.Errorf("failed to bind node to parent. Parent not found: '%s'", name)
	}
	return nil
}

func isNodeType(k string) bool {
	return slices.Contains([]string{"mod", "sig"}, k)
}

func attrToNodeType(k string) NodeType {
	switch k {
	case "mod":
		return MODULE_NODE
	case "sig":
		return SIGNATURE_NODE
	default:
		panic(errors.Errorf("attribute %s is not a type of node", k))
	}
}
