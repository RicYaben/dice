package shared

import (
	"context"

	"github.com/dice/pb"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DICE",
	MagicCookieValue: "DICE",
}

var PluginMap = map[string]plugin.Plugin{
	"module": &ModulePlugin{},
}

// This is how modules will interact with DICE to create
// new objects and events.
// Module handlers receive adapters, and use them to send
// requests to DICE. Adapters know which node this request
// is comming from, etc.
// All adapters share a common cache.
type Adapter interface {
	GetHost(uint) (*Host, error)
	GetSource(uint) (*Source, error)
	GetScan(uint) (*Scan, error)

	AddLabel(*Label) error
	AddFingerprint(*Fingerprint) error
	AddScan(*Scan) error
	AddSource(*Source) error

	Query(string) ([]*Host, error)
}

// Modules are how we call DICE plugins. Each module uses a
// specific type of handler to receive a specific type of object.
// This way, we avoid having modules to interact with the adapter
// to query for the object, fill the information, etc
type Module interface {
	// Returns an object with the module properties, i.e., name, help
	// info, description, usage, etc.
	Properties() ([]byte, error)
	// Handle the request
	Handle(e Event, a Adapter, cb func()) error
}

type ModulePlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl Module
}

func (p *ModulePlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterModuleServer(s, &GRPCModuleServer{
		Impl:   p.Impl,
		broker: broker,
	})
	return nil
}

func (p *ModulePlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{
		client: pb.NewModuleClient(c),
		broker: broker,
	}, nil
}
