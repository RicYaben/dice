package shared

type Host struct {
	ID           uint           `json:"id"`
	Ip           string         `json:"ip"`
	Domain       string         `json:"domain"`
	Fingerprints []*Fingerprint `json:"fingerprints"`
	Labels       []*Label       `json:"labels"`
}

type Fingerprint struct {
	ID       uint   `json:"id"`
	HostID   uint   `json:"host_id"`
	Data     []byte `json:"data"`
	Service  string `json:"service"`
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port"`
}

type Scan struct {
	ID      uint     `json:"id"`
	Targets []string `json:"targets"`
	Args    []byte   `json:"args"`
}

type Label struct {
	ShortName   string `json:"short_name"`
	LongName    string `json:"long_name"`
	Description string `json:"description"`
	Mitigation  string `json:"mitigation"`
}

type Source struct {
	ID       uint   `json:"id"`
	Location string `json:"location"`
	Format   string `json:"format"`
	Scanner  string `json:"scanner"`
	Type     string `json:"type"`
	Args     []byte `json:"args"`
}
