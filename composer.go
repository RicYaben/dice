package dice

import "github.com/pkg/errors"

type composer struct {
	adapter *composerAdapter
}

func newComposer(adapter *composerAdapter) *composer {
	return &composer{adapter}
}

// Compose components from signatures and actions. Only load components from actions.
// FAQ: can I use DICE with my parameters to scan for something in particular without classifying?
// You can create a basic signature that says exactly that, on new targets added, scan for x,y,z
/*
// dummy.dice
cls dummy
// dummy module `dummy.go`
*/
func (c *composer) Compose(sigs []string, actions EngineActions) ([]Component, error) {
	// Load all signatures in the adapter. We dont need to know the graph or anything else here
	if err := c.adapter.loadSignatures(sigs); err != nil {
		return nil, errors.Wrap(err, "failed to make graph")
	}

	var comps []Component
	if actions.Classify {
		clsComp := c.newClassifier().addModule(c.adapter.Classifiers())
		comps = append(comps, clsComp)
	}

	if actions.Identify {
		// identifiers are loaded at run-time or are static
		// the adapter should know the static ones
		idComp := c.newIdentifier().addModule(c.adapter.Identifiers())
		comps = append(comps, idComp)
	}

	if actions.Scan {
		scnComp := c.newScanner().addModule(c.adapter.Scanners())
		comps = append(comps, scnComp)
	}

	return comps, nil
}

func (c *composer) newClassifier() *component {
	// TODO: add event handler
	// A component adapter is an adapter that can create module adapters with handlers, etc.
	// is a middle interface
	handleSource := func(comp *component, host *Host) error {
		// the adapter contains a registry of signatures and modules loaded
		// so we can get access to those modules directly
		mods, err := c.adapter.getClassifiers(host.Hooks...)
		if err != nil {
			return errors.Wrap(err, "failed to find modules")
		}

		for _, mod := range mods {
			if err := mod.update(host); err != nil {
				return err
			}
		}
		return nil
	}

	handleHost := func(comp *component, host *Host) error {
		for _, mod := range comp.modules {
			if err := mod.update(host); err != nil {
				return err
			}
		}
		return nil
	}

	return &component{
		Name:    "classifier",
		Adapter: c.adapter.componentAdapter(),
		EventHandler: func(comp *component, e Event) error {
			host, err := comp.adapter.getHost(e.ObjectID)
			if err != nil {
				return errors.Wrapf(err, "failed to find host %d", e.ObjectID)
			}

			switch e.EventType {
			case SOURCE_EVENT:
				return handleSource(comp, host)
			case HOST_EVENT:
				return handleHost(comp, host)
			}
			return errors.Errorf("unable to handle event type %v", e.EventType)
		},
		Events: []EventType{FINGERPRINT_EVENT, HOST_EVENT},
	}
}

func (c *composer) newIdentifier() *component {
	return &component{
		Name:    "identifier",
		Adapter: c.adapter.componentAdapter(),
		EventHandler: func(comp *component, e Event) error {
			source, err := comp.adapter.getSource(e.ObjectID)
			if err != nil {
				return errors.Wrap(err, "failed to find source")
			}

			// we only send the source to the module that can handle it
			for _, mod := range comp.getModules(source.Name) {
				if err := mod.update(source); err != nil {
					return errors.Wrap(err, "failed to identify source")
				}
			}
			return nil
		},
		Events: []EventType{SOURCE_EVENT},
	}
}

func (c *composer) newScanner() *component {
	return &component{
		Name:    "scanner",
		Adapter: c.adapter.componentAdapter(),
		EventHandler: func(comp *componenet, e event) error {
			scn, err := comp.adapter.getScan(e.ObjectID)
			if err != nil {
				return errors.Wrap(err, "failed to find source")
			}
			// we only send the source to the module that can handle it
			for _, mod := range comp.getModules(scn.Name) {
				if err := mod.update(scn); err != nil {
					return errors.Wrap(err, "failed to identify source")
				}
			}
			return nil
		},
		Events: []EventType{SCAN_EVENT},
	}
}
