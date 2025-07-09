// Parser includes a simple parsing function to translate signature syntax into
// raw signatures. NOTE: it does NOT
package dice

import (
	"io"

	"github.com/alecthomas/participle/v2"
	"github.com/dice/pkg/ast"
)

type Parser interface {
	// Parse signature syntax into a raw signature. NOTE: does not
	// validate the signature, does not load modules, and does not
	// check for validity. It is a simple translating function.
	// Use a dice.SignaturesAdapter instead.
	Parse(string, io.Reader) (Signature, error)
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

func (p *parser) Parse(fname string, r io.Reader) (Signature, error) {
	sig, err := p.parser.Parse(fname, r)
	if err != nil {
		return Signature{}, err
	}
	return p.bind(fname, sig)
}

// Converts the AST signature to an actual Signature object
func (p *parser) bind(fname string, s *ast.Signature) (Signature, error) {
	sig := Signature{
		// default to classifier
		Type: "classifier",
		Name: fname,
	}

	// get the type of the signature. ignore the rest
	for _, prop := range s.Properties {
		switch prop.Key {
		case "type":
			if v, ok := prop.Value.(ast.String); ok {
				sig.Type = v.String
			}
		}
	}

	// bind the nodes
	sig.Nodes = p.bindNodes(s.Nodes)
	return sig, nil
}

// Convert an ast.Node to a regular Node
// NOTE: this process does NOT validate the nodes!
func (p *parser) bindNodes(nodes []*ast.Node) []*Node {
	nList := []*Node{}
	for _, node := range nodes {
		n := p.bindNode(node)
		nList = append(nList, n)
	}
	return nList
}

func (p *parser) bindNode(node *ast.Node) *Node {
	n := Node{
		name:     node.Name,
		Children: []*Node{},
	}

	switch n.Type {
	case "mod":
		n.Type = MODULE_NODE
	case "sig":
		n.Type = SIGNATURE_NODE
	}

	child := Node{}
	for _, attr := range node.Attributes {
		if val, ok := attr.Value.(ast.List); ok {
			switch attr.Key {
			case "mod":
				child.Type = MODULE_NODE
			case "sig":
				child.Type = SIGNATURE_NODE
			default:
				continue
			}

			for _, v := range val.List {
				child.name = v
				n.Children = append(n.Children, &child)
			}
		}
	}
	return &n
}
