package shared

import (
	"context"
	// TODO: replace this with "encoding/json/v2" when upgrading to 1.25
	"encoding/json"

	"github.com/dice/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// TODO: we need a different client for each?
type GRPCClient struct {
	client proto.ModuleClient
	broker *plugin.GRPCBroker
}

func (m *GRPCClient) Handle(a Adapter, e Event) error {
	adapterServer := &GRPCAdapterServer{Impl: a}

	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		proto.RegisterAdapterServer(s, adapterServer)
		return s
	}

	brokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(brokerID, serverFunc)

	_, err := m.client.Handle(context.Background(), &proto.HandleRequest{
		// adapter server
		AddServer: brokerID,
		Event: &proto.Event{
			Id:   uint32(e.ID()),
			Type: e.Type(),
		},
	})

	s.Stop()
	return err
}

func (m *GRPCClient) Properties() (map[string]any, error) {
	resp, err := m.client.Properties(context.Background(), &proto.Empty{})
	if err != nil {
		return nil, err
	}

	fields := make(map[string]any)
	if err := json.Unmarshal(resp.Fields, fields); err != nil {
		return nil, err
	}

	return fields, nil
}

// Implementation of the Module server, i.e., the methods the server
// has access to
type GRPCModuleServer struct {
	proto.UnimplementedModuleServer
	Impl   Module
	broker *plugin.GRPCBroker
}

func (m *GRPCModuleServer) Handle(ctx context.Context, req *proto.HandleRequest) (*proto.Empty, error) {
	conn, err := m.broker.Dial(req.AddServer)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	a := &GRPCAdapterClient{proto.NewAdapterClient(conn)}
	e := NewEvent(req.Event.Type, uint(req.Event.Id))
	return &proto.Empty{}, m.Impl.Handle(e, a)
}

func (m *GRPCModuleServer) Propagate(ctx context.Context, req *proto.Empty) (*proto.Empty, error) {
	return &proto.Empty{}, m.Impl.Propagate()
}

func (m *GRPCModuleServer) Properties(ctx context.Context, req *proto.Empty) (*proto.Fields, error) {
	fields, err := m.Impl.Properties()
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}

	return &proto.Fields{Fields: b}, nil
}

// Server-side (this happens in the Module), sends plugin calls
type GRPCAdapterClient struct{ client proto.AdapterClient }

func (a *GRPCAdapterClient) Query(q string) ([]Host, error) {
	resp, err := a.client.Query(context.Background(), &proto.QueryRequest{Query: q})
	if err != nil {
		return nil, err
	}

	var hosts []Host
	for _, h := range resp.Hosts {
		var fps []Fingerprint
		for _, f := range h.Fingerprints {
			fps = append(fps, Fingerprint{
				ID:       uint(f.Id),
				HostID:   uint(f.HostID),
				Data:     f.Data,
				Service:  f.Service,
				Protocol: f.Protocol,
				Port:     f.Port,
			})
		}

		hosts = append(hosts, Host{
			ID:           uint(h.Id),
			Address:      h.Address,
			Fingerprints: fps,
		})
	}
	return hosts, nil
}

func (m *GRPCAdapterClient) GetHost(id uint) (Host, error) {
	resp, err := m.client.GetHost(context.Background(), &proto.IDRequest{
		Id: uint32(id),
	})

	if err != nil {
		return Host{}, err
	}

	return Host{
		ID:      uint(resp.Id),
		Address: resp.Address,
	}, nil
}

func (m *GRPCAdapterClient) GetSource(id uint) (Source, error) {
	resp, err := m.client.GetSource(context.Background(), &proto.IDRequest{
		Id: uint32(id),
	})

	if err != nil {
		return Source{}, err
	}

	return Source{
		ID:     uint(resp.Id),
		Path:   resp.Path,
		Format: resp.Format,
	}, nil
}

func (m *GRPCAdapterClient) GetScan(id uint) (Scan, error) {
	resp, err := m.client.GetScan(context.Background(), &proto.IDRequest{
		Id: uint32(id),
	})

	if err != nil {
		return Scan{}, err
	}

	return Scan{
		ID:      uint(resp.Id),
		Targets: resp.Targets,
		Module:  resp.Module,
		Args:    resp.Flags, // TODO: this needs unmarshaling into json
	}, nil
}

func (m *GRPCAdapterClient) Label(lab Label) error {
	_, err := m.client.AddLabel(context.Background(), &proto.Label{
		HostID: uint32(lab.HostID),
		Label:  lab.Label,
	})
	return err
}

func (m *GRPCAdapterClient) Fingerprint(fp Fingerprint) error {
	data, err := json.Marshal(fp.Data)
	if err != nil {
		return err
	}

	_, err = m.client.AddFingerprint(context.Background(), &proto.Fingerprint{
		HostID: uint32(fp.HostID),
		Data:   data,
	})
	return err
}

func (m *GRPCAdapterClient) Source(src Source) error {
	_, err := m.client.AddSource(context.Background(), &proto.Source{
		Format:  src.Format,
		Scanner: src.Scanner,
		Path:    src.Path,
	})
	return err
}

func (m *GRPCAdapterClient) Scan(scn Scan) error {
	_, err := m.client.AddScan(context.Background(), &proto.Scan{
		Targets: scn.Targets,
		Flags:   scn.Args,
		Module:  scn.Module,
	})
	return err
}

// Host-side (this happens in DICE), receives plugin calls
type GRPCAdapterServer struct {
	proto.UnimplementedAdapterServer
	Impl Adapter
}

func (s *GRPCAdapterServer) Query(ctx context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
	resp, err := s.Impl.Query(req.Query)
	if err != nil {
		return nil, err
	}

	var hosts []*proto.Host
	for _, h := range resp {
		var fps []*proto.Fingerprint
		for _, fp := range h.Fingerprints {
			fps = append(fps, &proto.Fingerprint{
				Id:       uint32(fp.ID),
				HostID:   uint32(fp.HostID),
				Data:     fp.Data,
				Service:  fp.Service,
				Protocol: fp.Protocol,
				Port:     fp.Port,
			})
		}

		hosts = append(hosts, &proto.Host{
			Id:           uint32(h.ID),
			Address:      h.Address,
			Fingerprints: fps,
		})
	}
	return &proto.QueryResponse{Hosts: hosts}, nil
}

func (s *GRPCAdapterServer) GetHost(ctx context.Context, req *proto.IDRequest) (*proto.Host, error) {
	host, err := s.Impl.GetHost(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return &proto.Host{
		Id:      uint32(host.ID),
		Address: host.Address,
	}, nil
}

func (s *GRPCAdapterServer) GetSource(ctx context.Context, req *proto.IDRequest) (*proto.Source, error) {
	src, err := s.Impl.GetSource(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return &proto.Source{
		Id:     uint32(src.ID),
		Path:   src.Path,
		Format: src.Format,
	}, nil
}

func (s *GRPCAdapterServer) GetScan(ctx context.Context, req *proto.IDRequest) (*proto.Scan, error) {
	scan, err := s.Impl.GetScan(uint(req.Id))
	if err != nil {
		return nil, err
	}

	return &proto.Scan{
		Id:      uint32(scan.ID),
		Targets: scan.Targets,
		Module:  scan.Module,
		Flags:   scan.Args,
	}, nil
}

func (m *GRPCAdapterServer) AddLabel(ctx context.Context, req *proto.Label) (*proto.Empty, error) {
	lab := Label{uint(req.HostID), req.Label}
	return &proto.Empty{}, m.Impl.Label(lab)
}

func (m *GRPCAdapterServer) AddFingerprint(ctx context.Context, req *proto.Fingerprint) (*proto.Empty, error) {
	fp := Fingerprint{HostID: uint(req.HostID), Data: req.Data}
	return &proto.Empty{}, m.Impl.Fingerprint(fp)
}

func (m *GRPCAdapterServer) AddScan(ctx context.Context, req *proto.Scan) (*proto.Empty, error) {
	sc := Scan{
		Targets: req.Targets,
		Module:  req.Module,
		Args:    req.Flags,
	}
	return &proto.Empty{}, m.Impl.Scan(sc)
}

func (m *GRPCAdapterServer) AddSource(ctx context.Context, req *proto.Source) (*proto.Empty, error) {
	src := Source{
		Path:    req.Path,
		Format:  req.Format,
		Scanner: req.Scanner,
	}
	return &proto.Empty{}, m.Impl.Source(src)
}
