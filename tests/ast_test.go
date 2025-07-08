package test

import (
	"reflect"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/dice/pkg/ast"
)

type astTester struct {
	input  string
	result ast.Signature
}

// var diceLexer = lexer.MustSimple([]lexer.SimpleRule{
// 		{`Ident`, `[a-zA-Z][a-zA-Z_\d]*`},
// 		{`String`, `"(?:\\.|[^"])*"`},
// 		{`Float`, `\d+(?:\.\d+)?`},
// 		{`Punct`, `[][=]`},
// 		{"comment", `[#;][^\n]*`},
// 		{"whitespace", `\s+`},
// 	})

func (t *astTester) runTest(test *testing.T) {
	parser := participle.MustBuild[ast.Signature](
		//participle.Lexer(diceLexer),
		participle.Unquote("String"),
		participle.Union[ast.Value](ast.String{}, ast.Number{}, ast.List{}),
	)

	sig, err := parser.ParseString("", t.input)
	if err != nil {
		test.Errorf("failed to parse input: %v", err)
	}

	if reflect.DeepEqual(sig, t.result) {
		test.Errorf("expected %v, got %v", t.result, sig)
		return
	}
}

var astTests = [...]*astTester{
	{
		input: `
	mod mod1
	mod mod2 (mod: mod1; args: "some args")
	sig sig1 (mod: mod2, mod1)
	`,
		result: ast.Signature{
			Nodes: []*ast.Node{
				{
					Type: "mod",
					Name: "mod1",
				},
				{
					Type: "mod",
					Name: "mod2",
					Attributes: []*ast.Attribute{
						{
							Key:   "mod",
							Value: ast.List{List: []string{"mod1"}},
						},
					},
				},
				{
					Type: "sig",
					Name: "sig1",
					Attributes: []*ast.Attribute{
						{
							Key:   "mod",
							Value: ast.List{List: []string{"mod2", "mod1"}},
						},
					},
				},
			},
		},
	},
}

func TestASTParser(t *testing.T) {
	for _, cfg := range astTests {
		cfg.runTest(t)
	}
}
