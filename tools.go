package dice

import (
	"encoding/json"
	"os/exec"

	"github.com/dice/shared"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

// Makes a new source from arguments
func MakeTargetArgsSource(args []string) (*Source, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return &Source{
		Name: "targets",
		Type: SourceArgs,
		Args: b,
	}, nil
}

// Filter values from a slice
func Filter[T any](s []T, fn func(T) bool) []T {
	var r []T
	for _, t := range s {
		if fn(t) {
			r = append(r, t)
		}
	}
	return r
}

func LoadModule(name string, fpath string) (shared.Module, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins:         shared.PluginMap,
		Cmd:             exec.Command("sh", "-c", fpath),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC,
		},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load module %s", name)
	}

	raw, err := rpcClient.Dispense("module")
	if err != nil {
		return nil, err
	}

	module := raw.(shared.Module)
	return module, nil
}
