package test

import (
	"testing"

	"github.com/dice"
)

type projTester struct {
	projects []string
	studies  []string
}

func (t *projTester) runTest(test *testing.T) {
	conf, err := testConfig()
	if err != nil {
		test.Fatalf("failed to load configuration: %v", err)
	}

	ad := dice.MakeAdapters(nil, conf).Projects()

	for _, p := range t.projects {
		proj, err := dice.MakeProject(p, conf)
		if err != nil {
			test.Fatalf("failed to make project %s: %v", p, err)
		}

		if err := ad.InitProject(proj); err != nil {
			test.Fatalf("failed to initialize project %v: %v", proj, err)
		}
	}
}

var projTests = [...]*projTester{
	{
		projects: []string{
			"test-project", // a project in some location
			"-",            // no project, use default paths
			".",            // use the current folder as project
		},
		studies: []string{"study-1"},
	},
}

func TestProjects(t *testing.T) {
	for _, cfg := range projTests {
		cfg.runTest(t)
	}
}
