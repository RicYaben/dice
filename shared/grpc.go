package shared

import (
	"context"
	// TODO: replace this with "encoding/json/v2" when upgrading to 1.25

	"github.com/dice/pb"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type GRPCModuleClient struct {
	client pb.ModuleClient
	broker *plugin.GRPCBroker
}

func (m *GRPCModuleClient) Handle(e Event, a Adapter, cb func()) error {
	// Setup Adapter server
	adapterServer := &GRPCAdapterServer{Impl: a}
	var aServer *grpc.Server
	aFunc := func(opts []grpc.ServerOption) *grpc.Server {
		aServer = grpc.NewServer(opts...)
		pb.RegisterAdapterServer(aServer, adapterServer)
		return aServer
	}

	aBrokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(aBrokerID, aFunc)

	// Setup Propagation callback server
	pServer := &GRPCPropagationServer{cb: cb}
	var propServer *grpc.Server
	pFunc := func(opts []grpc.ServerOption) *grpc.Server {
		propServer = grpc.NewServer(opts...)
		pb.RegisterPropagateServer(propServer, pServer)
		return propServer
	}

	pBrokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(pBrokerID, pFunc)

	_, err := m.client.Handle(context.Background(), &pb.HandleRequest{
		// adapter server
		AddServer:         aBrokerID,
		PropagationServer: pBrokerID,
		Event: &pb.Event{
			Id:   uint32(e.ID()),
			Type: e.Type(),
		},
	})

	aServer.Stop()
	propServer.Stop()
	return err
}

func (m *GRPCModuleClient) Properties() ([]byte, error) {
	resp, err := m.client.Properties(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.Fields, nil
}

// Implementation of the Module server, i.e., the methods the server
// has access to
type GRPCModuleServer struct {
	pb.UnimplementedModuleServer
	Impl   Module
	broker *plugin.GRPCBroker
}

func (m *GRPCModuleServer) Handle(ctx context.Context, req *pb.HandleRequest) (*pb.Empty, error) {
	conn, err := m.broker.Dial(req.AddServer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial adapter server")
	}
	defer conn.Close()

	adapter := &GRPCAdapterClient{pb.NewAdapterClient(conn)}

	propagateConn, err := m.broker.Dial(req.PropagationServer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial propagation server")
	}
	defer propagateConn.Close()

	propagateClient := pb.NewPropagateClient(propagateConn)
	propagate := func() {
		propagateClient.Trigger(context.Background(), &pb.Empty{})
	}

	event := NewEvent(req.Event.Type, uint(req.Event.Id))
	return &pb.Empty{}, m.Impl.Handle(event, adapter, propagate)
}

func (m *GRPCModuleServer) Properties(ctx context.Context, req *pb.Empty) (*pb.Fields, error) {
	fields, err := m.Impl.Properties()
	if err != nil {
		return nil, err
	}

	return &pb.Fields{Fields: fields}, nil
}

// Server-side (this happens in the Module), sends plugin calls
type GRPCAdapterClient struct{ client pb.AdapterClient }

func (a *GRPCAdapterClient) Query(q string) ([]*Host, error) {
	resp, err := a.client.Query(context.Background(), &pb.QueryRequest{Query: q})
	if err != nil {
		return nil, err
	}

	var hosts []*Host
	for _, h := range resp.Hosts {
		host := FromProtoHost(h)
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func (m *GRPCAdapterClient) GetHost(id uint) (*Host, error) {
	resp, err := m.client.GetHost(context.Background(), &pb.IDRequest{
		Id: uint32(id),
	})

	if err != nil {
		return nil, err
	}

	return FromProtoHost(resp), nil
}

func (m *GRPCAdapterClient) GetSource(id uint) (*Source, error) {
	resp, err := m.client.GetSource(context.Background(), &pb.IDRequest{
		Id: uint32(id),
	})

	if err != nil {
		return nil, err
	}

	return FromProtoSource(resp), nil
}

func (m *GRPCAdapterClient) GetScan(id uint) (*Scan, error) {
	resp, err := m.client.GetScan(context.Background(), &pb.IDRequest{
		Id: uint32(id),
	})

	if err != nil {
		return nil, err
	}

	return FromProtoScan(resp), nil
}

func (m *GRPCAdapterClient) AddLabel(lab *Label) error {
	_, err := m.client.AddLabel(context.Background(), ToProtoLabel(lab))
	return err
}

func (m *GRPCAdapterClient) AddFingerprint(fp *Fingerprint) error {
	_, err := m.client.AddFingerprint(context.Background(), ToProtoFingerprint(fp))
	return err
}

func (m *GRPCAdapterClient) AddSource(src *Source) error {
	_, err := m.client.AddSource(context.Background(), ToProtoSource(src))
	return err
}

func (m *GRPCAdapterClient) AddScan(scn *Scan) error {
	_, err := m.client.AddScan(context.Background(), ToProtoScan(scn))
	return err
}

// Host-side (this happens in DICE), receives plugin calls
type GRPCAdapterServer struct {
	pb.UnimplementedAdapterServer
	Impl Adapter
}

func (s *GRPCAdapterServer) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	resp, err := s.Impl.Query(req.Query)
	if err != nil {
		return nil, err
	}

	var hosts []*pb.Host
	for _, h := range resp {
		host := ToProtoHost(h)
		hosts = append(hosts, host)
	}
	return &pb.QueryResponse{Hosts: hosts}, nil
}

func (s *GRPCAdapterServer) GetHost(ctx context.Context, req *pb.IDRequest) (*pb.Host, error) {
	host, err := s.Impl.GetHost(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return ToProtoHost(host), nil
}

func (s *GRPCAdapterServer) GetSource(ctx context.Context, req *pb.IDRequest) (*pb.Source, error) {
	src, err := s.Impl.GetSource(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return ToProtoSource(src), nil
}

func (s *GRPCAdapterServer) GetScan(ctx context.Context, req *pb.IDRequest) (*pb.Scan, error) {
	scan, err := s.Impl.GetScan(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return ToProtoScan(scan), nil
}

func (m *GRPCAdapterServer) AddLabel(ctx context.Context, req *pb.Label) (*pb.Empty, error) {
	lab := FromProtoLabel(req)
	return &pb.Empty{}, m.Impl.AddLabel(lab)
}

func (m *GRPCAdapterServer) AddFingerprint(ctx context.Context, req *pb.Fingerprint) (*pb.Empty, error) {
	fp := FromProtoFingerprint(req)
	return &pb.Empty{}, m.Impl.AddFingerprint(fp)
}

func (m *GRPCAdapterServer) AddScan(ctx context.Context, req *pb.Scan) (*pb.Empty, error) {
	sc := FromProtoScan(req)
	return &pb.Empty{}, m.Impl.AddScan(sc)
}

func (m *GRPCAdapterServer) AddSource(ctx context.Context, req *pb.Source) (*pb.Empty, error) {
	src := FromProtoSource(req)
	return &pb.Empty{}, m.Impl.AddSource(src)
}

type GRPCPropagationServer struct {
	pb.UnimplementedPropagateServer
	cb func()
}

func (s *GRPCPropagationServer) Trigger(context.Context, *pb.Empty) (*pb.Empty, error) {
	if s.cb != nil {
		s.cb()
	}
	return &pb.Empty{}, nil
}
