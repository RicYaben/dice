package module

import (
	"fmt"

	"github.com/dice/pkg/database"
)

type classifier struct {
	source string
}

func newClassifier(source string) *classifier {
	return &classifier{
		source: source,
	}
}

func (c *classifier) init() error {
	// TODO: load the module here using `rule.Source`
	panic("not implemented yet")
}

func (c *classifier) Do(h *database.Host, r *database.Record) error {
	panic("not implemented yet")
}

func Classifier(source string) (*classifier, error) {
	// TODO: check in the registry if the module has already been loaded
	// if so, just return that one
	c := newClassifier(source)
	if err := c.init(); err != nil {
		return nil, fmt.Errorf("failed to load classifier")
	}
	return c, nil
}
