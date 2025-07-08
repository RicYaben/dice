package basic

import (
	"github.com/dice"
	"github.com/dice/pkg/sdk"
	"github.com/pkg/errors"
)

// This handler will be wrapped every time is called
func handle(cls *dice.Connector, h dice.Host) error {
	fps, err := cls.Fingerprints("protocol:ssh auth:true", h)
	if err != nil {
		return errors.Wrap(err, "failed to search for fingerprints")
	}

	if len(fps) > 0 {
		cls.Label("weak-access-control")
		return cls.Propagate()
	}

	return nil
}

func Main() *sdk.ClassifierModule {
	return &sdk.ClassifierModule{
		// Name of the module
		Name: "shh-brute",
		// How to use the module
		Help: "dice -M ssh-brute",
		// What this classifier does
		Description: "Whether the SSH service has weak auth",
		// Pre-load data ahead of time.
		// Classifiers request fingerprints!
		Query: "protocol:ssh",
		// If the query returns no results, this module
		// will attempt to fire a scan event
		Requirements: dice.Scan{
			Module: "scn",
			Args: map[string]any{
				"ports":  22,
				"probe":  "synack",
				"module": "ssh-brute",
				"flags": []string{
					"dictionary=root:root,admin:admin",
				},
			},
		},
		// The event handler
		Handler: handle,
	}
}
