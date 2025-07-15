package sdk

import (
	"errors"
	"fmt"
	"slices"

	"github.com/dice/shared"
	"github.com/hashicorp/go-plugin"
)

type Module struct {
	// Name of the module
	Name string `json:"name"`
	// Type of module: identifier, classifier, scanner...
	Type string `json:"type"`
	// How to use the module
	Help string `json:"help"`
	// What this classifier does
	Description string `json:"description"`
	// How to use
	Usage string `json:"usage"`
	// Pre-load data ahead of time.
	// Classifiers request fingerprints!
	Query string `json:"query"`
	// If the query returns no results, this module
	// will attempt to fire a scan event
	Requirements any
}

func (m *Module) Propagate() error { return nil }
func (m *Module) Properties() (map[string]string, error)

type Handler func(e shared.Event, m *Module, a shared.Adapter) error

// A simple wrapper to handle event requests
type moduleImpl struct {
	*Module
	handler Handler
}

func (impl *moduleImpl) Handle(e shared.Event, a shared.Adapter) error {
	switch impl.Type {
	case "classifier":
		if !slices.Contains([]string{"fingerprint", "host"}, e.Type()) {
			return errors.New("classifier cannot handle event type")
		}
		host, err := a.GetHost(e.ID())
		if err != nil {
			return err
		}
		return impl.handler(e.WithHost(host), impl.Module, a)

	case "identifier":
		if e.Type() != "source" {
			return errors.New("identifier cannot handle event type")
		}
		src, err := a.GetSource(e.ID())
		if err != nil {
			return err
		}
		return impl.handler(e.WithSource(src), impl.Module, a)

	case "scanner":
		if e.Type() != "source" {
			return errors.New("scanner cannot handle event type")
		}
		scan, err := a.GetScan(e.ID())
		if err != nil {
			return err
		}
		return impl.handler(e.WithScan(scan), impl.Module, a)

	default:
		return fmt.Errorf("unknown module type: %s", impl.Type)
	}
}

func Serve(mod *Module, handler Handler) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			mod.Type: &shared.ModulePlugin{
				Impl: &moduleImpl{
					Module:  mod,
					handler: handler,
				},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
