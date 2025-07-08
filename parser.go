package dice

import (
	"io"

	"github.com/alecthomas/participle/v2"
	"github.com/dice/pkg/ast"
)

type Parser interface {
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

func (p *parser) Parse(fname string, r io.Reader) (*Signature, error) {
	sig, err := p.parser.Parse(fname, r)
	if err != nil {
		return nil, err
	}
	return p.bind(fname, sig)
}

// Converts the AST signature to an actual Signature object
func (p *parser) bind(fname string, s *ast.Signature) (*Signature, error) {
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

	// TODO: finish this!
	// for _, node := range s.Nodes {
	// 	n := *Node{}

	// }
	return &sig, nil
}
