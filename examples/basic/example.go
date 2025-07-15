package basic

import (
	"github.com/dice/pkg/sdk"
	"github.com/dice/shared"
	"github.com/pkg/errors"
)

func handler(e shared.Event, m *sdk.Module, a shared.Adapter) error {
	resp, err := a.Query([]any{e.Host(), "protocol:ssh auth:true"})
	if err != nil {
		return errors.Wrap(err, "failed to search for fingerprints")
	}

	fps := resp.([]shared.Fingerprint)
	if len(fps) > 0 {
		a.Label("weak-access-control", e.ID())
		m.Propagate()
	}

	return nil
}

func main() {
	sdk.Serve(&sdk.Module{
		Name:        "shh-brute",
		Type:        "classifier",
		Help:        "dice -M ssh-brute",
		Description: "Whether the SSH service has weak auth",
		Query:       "protocol:ssh",
		Requirements: shared.Scan{
			Module: "scn",
			Flags: map[string]any{
				"ports":  22,
				"probe":  "synack",
				"module": "ssh-brute",
				"flags": []string{
					"dictionary=root:root,admin:admin",
				},
			},
		},
	}, handler)
}
