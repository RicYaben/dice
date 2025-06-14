package module

import (
	"fmt"
	"time"

	"github.com/dice/pkg/api"
	"github.com/dice/pkg/database"
	"gorm.io/gorm"
)

type scanner struct {
	// scan description
	description *database.Scan
	// queues of targets
	queues chan *database.Record
	// ticker time duration to force scans to happen
	tick time.Duration
	// capacity of the batch and queues
	capacity int
	// module
	do func(batch []*database.Record) []*api.ScanResult
}

func newScanner(scan *database.Scan, tick time.Duration, capacity int) *scanner {
	return &scanner{
		description: scan,
		tick:        tick,
		capacity:    capacity,
		queues:      make(chan *database.Record, capacity),
	}
}

func (s *scanner) init() error {
	// TODO FIXME: implement this
	// we should load the mapper and create
	// the pipe between scanners

	// 1. check if we need to run an L4 scanner
	// 2. prepare the pipe
	// 3. prepare the L6 scanner
	var do = func(batch []*database.Record) []*api.ScanResult {
		return nil
	}
	s.do = do
	panic("not implemented yet")
}

func (s *scanner) scan(batch []*database.Record) {
	resultsMap := make(map[string][]*api.ScanResult)
	for _, res := range s.do(batch) {
		resultsMap[res.IP] = append(resultsMap[res.IP], res)
	}

	// TODO FIXME: same as always, how can I access the database from here?
	var db *gorm.DB
	var records []database.Record
	for ip, results := range resultsMap {
		var host database.Host
		if err := db.Model(&database.Host{}).First(&host, "ip = ?", ip).Error; err != nil {
			panic(err)
		}

		for _, result := range results {
			record := database.Record{
				ScanID: s.description.ID,
				HostID: host.ID,
				Data:   result.Data,
			}
			records = append(records, record)
		}
	}

	// A Hook? AfterCreate -> trigger notify subscribers on Host.
	if err := db.CreateInBatches(&records, 100).Error; err != nil {
		panic(err)
	}
}

func (s *scanner) Do(h *database.Host, r *database.Record) error {
	// TODO: query the host as well
	s.queues <- r
	return nil
}

func (s *scanner) Run() {
	// TODO: add a cancel?
	defer close(s.queues)
	t := time.NewTicker(s.tick)

	for {
		// Fill a batch until the ticker fires or the batch is full
		// then, send the current batch to scan
		var batch []*database.Record
		for {
			select {
			case <-t.C:
				goto SCAN
			case q := <-s.queues:
				batch = append(batch, q)
				if len(batch) == s.capacity {
					goto SCAN
				}
			}
		}

	SCAN:
		if len(batch) > 0 {
			s.scan(batch)
		}
	}
}

func Scanner(scan *database.Scan) (*scanner, error) {
	s := newScanner(scan, 30*time.Second, 100)
	if err := s.init(); err != nil {
		return nil, fmt.Errorf("failed to load scanner")
	}
	return s, nil
}
