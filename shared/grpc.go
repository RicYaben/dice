package shared

import (
	"context"

	"github.com/dice"
	"github.com/hashicorp/go-plugin"
)

// TODO: we need a different client for each?
type GRPCClient struct {
	client proto.ModuleClient
	broker *plugin.GRPCBroker
}

func (m *GRPCClient) Handle(a Adapter, e Event) error {
	adapterServer := &GRPCAdapterServer{Impl: a}

	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOptions) *grpc.Server {
		s = grpc.NewServer(opts...)
		proto.RegisterAdapterServer(s, adapterServer)
		return s
	}

	brokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(brokerID, serverFunc)

	_, err := m.client.Handle(context.Background(), &proto.HandleRequest{
		// adapter server
		AddServer: brokerID,
		EventType: e.Type(),
		ObjectID:  e.ID(),
	})

	s.Stop()
	return err
}

func (m *GRPCClient) Properties() (dice.Properties, error) {
	resp, err := m.client.Properties(context.Background(), &proto.PropertiesRequest{})
	if err != nil {
		return dice.Properties{}, err
	}

	return dice.Properties{
		Name:        resp.Name,
		Help:        resp.Help,
		Description: resp.Description,
		Query:       resp.Query,
	}, nil
}

// Implementation of the Module server, i.e., the methods the server
// has access to
type GRPCModuleServer struct {
	Impl   Module
	broker *plugin.GRPCBroker
}

func (m *GRPCModuleServer) Propagate(ctx context.Context, req *proto.PropagateRequest) error {
	return m.Impl.Propagate()
}

type GRPCAdapterServer struct {
	Impl Adapter
}

// TODO: convert into their types!
func (m *GRPCAdapterServer) Label(ctx context.Context, req *proto.LabelRequest) (*proto.Result, error) {
	if err := m.Impl.Label(req.Label); err != nil {
		return nil, err
	}
	return &proto.Result{nil, nil}, nil
}

func (m *GRPCAdapterServer) Fingerprint(ctx context.Context, req *proto.FingerprintRequest) (*proto.Result, error) {
	if err := m.Impl.Fingerprint(req.Fingerprint); err != nil {
		return nil, err
	}
	return &proto.Result{nil, nil}, nil
}

func (m *GRPCAdapterServer) Scan(ctx context.Context, req *proto.ScanRequest) (*proto.Result, error) {
	if err := m.Impl.Scan(req.Scan); err != nil {
		return nil, err
	}
	return &proto.Result{nil, nil}, nil
}

func (m *GRPCAdapterServer) Source(ctx context.Context, req *proto.SourceRequest) (*proto.Result, error) {
	if err := m.Impl.Source(req.Source); err != nil {
		return nil, err
	}
	return &proto.Result{nil, nil}, nil
}
