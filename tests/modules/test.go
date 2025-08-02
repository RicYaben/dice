package main

import (
	"fmt"

	"github.com/dice/pkg/sdk"
	"github.com/dice/shared"
	"github.com/pkg/errors"
)

var labels = sdk.Labels{
	"test": {
		ShortName:   "test",
		LongName:    "Testing label",
		Description: "Test",
		Mitigation:  "test",
	},
}

func handler(e shared.Event, m *sdk.Module, a shared.Adapter, propagate func()) error {
	hosts, err := a.Query(fmt.Sprintf("host:%d", e.ID()))
	if err != nil {
		return errors.Wrap(err, "failed to query hosts")
	}

	for _, h := range hosts {
		a.AddLabel(labels.MakeLabel("test", h.ID))
		propagate()
		return nil
	}
	return nil
}

func main() {
	sdk.Serve(&sdk.Module{
		Name:        "test",
		Type:        "classifier",
		Help:        "test",
		Description: "Test classifier",
		Query:       "protocol:ssh",
	}, handler)
}
