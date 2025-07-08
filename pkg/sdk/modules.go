package sdk

import "github.com/dice"

type ClassifierModule struct {
	Name         string
	Query        string
	Requirements any
	Handler      func(*dice.ComponentModule, dice.Host) error
}

func (c *ClassifierModule) makePlugin() *dice.ClassifierPlugin {
	panic("not implemented yet")
}
