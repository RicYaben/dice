package dice

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

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

// Make a project. Typically followed by InitProject()
func MakeProject(p string, conf *Configuration) (*Project, error) {
	switch p {
	// no project
	case "-":
		return &Project{Name: "-"}, nil
	// current directory
	case ".", "":
		// get current directory project
		dir, err := os.Getwd()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get current directory")
		}

		return &Project{
			Name: filepath.Dir(dir),
			Path: dir,
		}, nil
	// another location
	default:
		return &Project{
			Name: filepath.Dir(p),
			Path: p,
		}, nil
	}
}
