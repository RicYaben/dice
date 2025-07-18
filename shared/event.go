package shared

type Event interface {
	Type() string
	ID() uint

	Host() *Host
	Source() *Source
	Scan() *Scan

	WithHost(*Host) Event
	WithSource(*Source) Event
	WithScan(*Scan) Event
}

type baseEvent struct {
	typ  string
	id   uint
	host *Host
	src  *Source
	scan *Scan
}

func NewEvent(typ string, id uint) Event {
	return &baseEvent{
		typ: typ,
		id:  id,
	}
}

func (e *baseEvent) Type() string    { return e.typ }
func (e *baseEvent) ID() uint        { return e.id }
func (e *baseEvent) Host() *Host     { return e.host }
func (e *baseEvent) Source() *Source { return e.src }
func (e *baseEvent) Scan() *Scan     { return e.scan }

func (e *baseEvent) WithHost(h *Host) Event {
	return &baseEvent{typ: e.typ, id: e.id, host: h, src: e.src, scan: e.scan}
}

func (e *baseEvent) WithSource(s *Source) Event {
	return &baseEvent{typ: e.typ, id: e.id, host: e.host, src: s, scan: e.scan}
}

func (e *baseEvent) WithScan(sc *Scan) Event {
	return &baseEvent{typ: e.typ, id: e.id, host: e.host, src: e.src, scan: sc}
}
