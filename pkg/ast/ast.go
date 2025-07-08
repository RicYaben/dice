package ast

type Signature struct {
	Properties []*Property `parser:"@@*"`
	Nodes      []*Node     `parser:"@@*"`
}

type Node struct {
	Type       string       `parser:"@('mod' | 'sig')"`
	Name       string       `parser:"@Ident"`
	Attributes []*Attribute `parser:"[ '(' @@ ( ';' @@ )* ')' ]"`
}

type Attribute struct {
	Key   string `parser:"@('mod' | 'sig' | 'args') ':'"`
	Value Value  `parser:"@@"`
}

type Property struct {
	Key   string `parser:"@Ident '='"`
	Value Value  `parser:"@@"`
}

type Value interface{ value() }

type String struct {
	String string `parser:"@String"`
}

func (String) value() {}

type Number struct {
	Number float64 `parser:"@Float | @Int"`
}

func (Number) value() {}

type List struct {
	List []string `parser:"(@Ident(',' @Ident)*)"`
}

func (List) value() {}
