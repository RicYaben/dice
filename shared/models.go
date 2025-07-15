package shared

type Host struct {
	ID uint
}

type Fingerprint struct {
	ID uint
}

type Scan struct {
	ID     uint
	Module string
	Flags  map[string]any
}

type Source struct {
	ID     uint
	Path   string
	Format string
}
