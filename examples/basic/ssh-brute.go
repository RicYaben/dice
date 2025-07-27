package basic

import (
	"fmt"

	"github.com/dice/pkg/sdk"
	"github.com/dice/shared"
	"github.com/pkg/errors"
)

var labels = sdk.Labels{
	"wac": {
		ShortName:   "weak-access-control",
		LongName:    "Missing SSH authentication policies",
		Description: "SSH service allowing anonymous users",
		Mitigation: `
		Disable anonoymous authentication methods.
		In the SSH service config file ('etc/ssh/sshd_config'), set
		'PermitEmptyPasswords' to 'no', and ensure 'AuthenticationMethods'
		is not set to 'none'.
		`,
	},
}

func handler(e shared.Event, m *sdk.Module, a shared.Adapter, propagate func()) error {
	hosts, err := a.Query(fmt.Sprintf("protocol:ssh auth:true host:%d", e.ID()))
	if err != nil {
		return errors.Wrap(err, "failed to search for fingerprints")
	}

	for _, h := range hosts {
		a.AddLabel(labels.MakeLabel("wac", h.ID))
		propagate()
		return nil
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
	}, handler)
}
