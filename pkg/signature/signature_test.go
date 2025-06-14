package signature

import (
	"testing"

	"github.com/dice/pkg/database"
)

type scanTester struct {
	flags ScanFlags
}

func (t *scanTester) runTest(test *testing.T, name string) {
	conf := database.GetDatabase(0)
	db, err := database.Open(conf)
	if err != nil {
		test.Errorf("[%s] failed to open db connection: %v", name, err)
		return
	}

	if err := Scan(db, t.flags); err != nil {
		test.Errorf("[%s] failed to scan: %v", name, err)
		return
	}
}

var tests = map[string]*scanTester{
	"update": {
		flags: ScanFlags{
			Allowed:  []string{"../../examples/rules"},
			Required: []string{"iot"},
			Mode:     "reset",
		},
	},
}

func TestScan(t *testing.T) {
	for name, conf := range tests {
		conf.runTest(t, name)
	}
}
