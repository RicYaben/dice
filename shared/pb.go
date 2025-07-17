package shared

import "github.com/dice/pb"

func ToProtoHost(h *Host) *pb.Host {
	if h == nil {
		return nil
	}

	fps := make([]*pb.Fingerprint, len(h.Fingerprints))
	for i, f := range h.Fingerprints {
		fps[i] = ToProtoFingerprint(f)
	}

	labels := make([]*pb.Label, len(h.Labels))
	for i, l := range h.Labels {
		labels[i] = ToProtoLabel(l)
	}

	return &pb.Host{
		Id:           uint32(h.ID),
		Ip:           h.Ip,
		Domain:       h.Domain,
		Fingerprints: fps,
		Labels:       labels,
	}
}

func FromProtoHost(h *pb.Host) *Host {
	if h == nil {
		return nil
	}

	fps := make([]*Fingerprint, len(h.Fingerprints))
	for i, f := range h.Fingerprints {
		fps[i] = FromProtoFingerprint(f)
	}

	labels := make([]*Label, len(h.Labels))
	for i, l := range h.Labels {
		labels[i] = FromProtoLabel(l)
	}

	return &Host{
		ID:           uint(h.Id),
		Ip:           h.Ip,
		Domain:       h.Domain,
		Fingerprints: fps,
		Labels:       labels,
	}
}

func ToProtoFingerprint(f *Fingerprint) *pb.Fingerprint {
	if f == nil {
		return nil
	}
	return &pb.Fingerprint{
		Id:       uint32(f.ID),
		HostId:   uint32(f.HostID),
		Data:     f.Data,
		Service:  f.Service,
		Protocol: f.Protocol,
		Port:     uint32(f.Port),
	}
}

func FromProtoFingerprint(f *pb.Fingerprint) *Fingerprint {
	if f == nil {
		return nil
	}
	return &Fingerprint{
		ID:       uint(f.Id),
		HostID:   uint(f.HostId),
		Data:     f.Data,
		Service:  f.Service,
		Protocol: f.Protocol,
		Port:     uint16(f.Port),
	}
}

func ToProtoLabel(l *Label) *pb.Label {
	if l == nil {
		return nil
	}
	return &pb.Label{
		ShortName:   l.ShortName,
		LongName:    l.LongName,
		Description: l.Description,
		Mitigation:  l.Mitigation,
	}
}

func FromProtoLabel(l *pb.Label) *Label {
	if l == nil {
		return nil
	}
	return &Label{
		ShortName:   l.ShortName,
		LongName:    l.LongName,
		Description: l.Description,
		Mitigation:  l.Mitigation,
	}
}

func ToProtoScan(s *Scan) *pb.Scan {
	if s == nil {
		return nil
	}
	return &pb.Scan{
		Id:      uint32(s.ID),
		Targets: s.Targets,
		Args:    s.Args,
	}
}

func FromProtoScan(s *pb.Scan) *Scan {
	if s == nil {
		return nil
	}
	return &Scan{
		ID:      uint(s.Id),
		Targets: s.Targets,
		Args:    s.Args,
	}
}

func ToProtoSource(s *Source) *pb.Source {
	if s == nil {
		return nil
	}
	return &pb.Source{
		Id:       uint32(s.ID),
		Location: s.Location,
		Format:   s.Format,
		Scanner:  s.Scanner,
		Type:     s.Type,
		Args:     s.Args,
	}
}

func FromProtoSource(s *pb.Source) *Source {
	if s == nil {
		return nil
	}
	return &Source{
		ID:       uint(s.Id),
		Location: s.Location,
		Format:   s.Format,
		Scanner:  s.Scanner,
		Type:     s.Type,
		Args:     s.Args,
	}
}
