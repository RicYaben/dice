package shared

type Host struct {
	ID           uint
	Address      string
	Fingerprints []Fingerprint
}

type Fingerprint struct {
	ID     uint
	HostID uint
	Data   []byte

	Service  string
	Protocol string
	Port     uint32
}

type Scan struct {
	ID      uint
	Targets []string
	Module  string
	Args    string
}

type Source struct {
	ID      uint
	Path    string
	Format  string
	Scanner string
}

type Label struct {
	HostID uint
	Label  string
}
