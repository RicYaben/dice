package dice

import (
	"github.com/dice/shared"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func FromPluginHost(p *shared.Host) *Host {
	return &Host{
		Model:        gorm.Model{ID: p.ID},
		Ip:           p.Ip,
		Domain:       p.Domain,
		Fingerprints: FromPluginFingerprints(p.Fingerprints),
		Labels:       FromPluginLabels(p.Labels),
	}
}

func FromPluginFingerprint(f *shared.Fingerprint) *Fingerprint {
	return &Fingerprint{
		Model:    gorm.Model{ID: f.ID},
		HostID:   f.HostID,
		Data:     f.Data,
		Service:  f.Service,
		Protocol: f.Protocol,
		Port:     f.Port,
	}
}

func FromPluginLabel(l *shared.Label) *Label {
	return &Label{
		HostID:      l.HostID,
		ShortName:   l.ShortName,
		LongName:    l.LongName,
		Description: l.Description,
		Mitigation:  l.Mitigation,
	}
}

func FromPluginScan(p *shared.Scan) *Scan {
	return &Scan{
		Model:   gorm.Model{ID: p.ID},
		Targets: p.Targets,
		Args:    datatypes.JSON(p.Args),
	}
}

func FromPluginSource(p *shared.Source) *Source {
	return &Source{
		Model:    gorm.Model{ID: p.ID},
		Location: p.Location,
		Format:   p.Format,
		Type:     SourceType(p.Type),
		Args:     datatypes.JSON(p.Args),
	}
}

func FromPluginFingerprints(p []*shared.Fingerprint) []*Fingerprint {
	out := make([]*Fingerprint, len(p))
	for i, f := range p {
		out[i] = FromPluginFingerprint(f)
	}
	return out
}

func FromPluginLabels(p []*shared.Label) []*Label {
	out := make([]*Label, len(p))
	for i, l := range p {
		out[i] = FromPluginLabel(l)
	}
	return out
}

func ToPluginHost(m *Host) *shared.Host {
	return &shared.Host{
		ID:           m.ID,
		Ip:           m.Ip,
		Domain:       m.Domain,
		Fingerprints: ToPluginFingerprints(m.Fingerprints),
		Labels:       ToPluginLabels(m.Labels),
	}
}

func ToPluginFingerprint(f *Fingerprint) *shared.Fingerprint {
	if f == nil {
		return nil
	}
	return &shared.Fingerprint{
		ID:       f.ID,
		HostID:   f.HostID,
		Data:     f.Data,
		Service:  f.Service,
		Protocol: f.Protocol,
		Port:     f.Port,
	}
}

func ToPluginLabel(l *Label) *shared.Label {
	if l == nil {
		return nil
	}
	return &shared.Label{
		HostID:      l.HostID,
		ShortName:   l.ShortName,
		LongName:    l.LongName,
		Description: l.Description,
		Mitigation:  l.Mitigation,
	}
}

func ToPluginFingerprints(m []*Fingerprint) []*shared.Fingerprint {
	out := make([]*shared.Fingerprint, len(m))
	for i, f := range m {
		out[i] = ToPluginFingerprint(f)
	}
	return out
}

func ToPluginLabels(m []*Label) []*shared.Label {
	out := make([]*shared.Label, len(m))
	for i, l := range m {
		out[i] = ToPluginLabel(l)
	}
	return out
}

func ToPluginScan(m *Scan) *shared.Scan {
	return &shared.Scan{
		ID:      m.ID,
		Targets: m.Targets,
		Args:    []byte(m.Args),
	}
}

func ToPluginSource(m *Source) *shared.Source {
	return &shared.Source{
		ID:       m.ID,
		Location: m.Location,
		Format:   m.Format,
		Type:     string(m.Type),
		Args:     []byte(m.Args),
	}
}

type Connector struct {
	Impl CosmosAdapter
}

func (c *Connector) GetHost(id uint) (*shared.Host, error) {
	h, err := c.Impl.GetHost(id)
	if err != nil {
		return nil, err
	}
	return ToPluginHost(h), nil
}

func (c *Connector) GetSource(id uint) (*shared.Source, error) {
	s, err := c.Impl.GetSource(id)
	if err != nil {
		return nil, err
	}
	return ToPluginSource(s), nil
}

func (c *Connector) GetScan(id uint) (*shared.Scan, error) {
	s, err := c.Impl.GetScan(id)
	if err != nil {
		return nil, err
	}
	return ToPluginScan(s), nil
}

// Add

func (c *Connector) AddHost(h *shared.Host) error {
	return c.Impl.AddHost(FromPluginHost(h))
}

func (c *Connector) AddLabel(l *shared.Label) error {
	return c.Impl.AddLabel(FromPluginLabel(l))
}

func (c *Connector) AddFingerprint(f *shared.Fingerprint) error {
	return c.Impl.AddFingerprint(FromPluginFingerprint(f))
}

func (c *Connector) AddScan(s *shared.Scan) error {
	return c.Impl.AddScan(FromPluginScan(s))
}

func (c *Connector) AddSource(s *shared.Source) error {
	return c.Impl.AddSource(FromPluginSource(s))
}

// Query

func (c *Connector) Query(q string) ([]*shared.Host, error) {
	hosts, err := c.Impl.Query(q)
	if err != nil {
		return nil, err
	}

	var results []*shared.Host
	for _, host := range hosts {
		results = append(results, ToPluginHost(host))
	}

	return results, nil
}
