package test

import (
	"testing"

	"github.com/dice"
)

type sigTester struct {
	signatures []string
	modules    []string
}

func (t *sigTester) runTest(test *testing.T) {
	conf, err := testConfig()
	if err != nil {
		test.Fatalf("failed to load configuration: %v", err)
	}

	ad := dice.MakeAdapters(nil, conf).Signatures()

	// Add all missing
	if _, err := ad.AddMissingModules(t.modules...); err != nil {
		test.Fatalf("failed to add modules: %v", err)
	}

	if _, err := ad.AddMissingSignatures(t.signatures...); err != nil {
		test.Fatalf("failed to add signatures: %v", err)
	}

	// Query and list
	mods := []*dice.Module{}
	if err := ad.Find(&mods, "name IN ?", t.modules); err != nil {
		test.Fatalf("failed to find modules: %v", err)
	}

	if len(mods) != len(t.modules) {
		test.Fatalf("expected %d, got %d modules", len(t.modules), len(mods))
	}

	sigs := []*dice.Signature{}
	if err := ad.Find(&sigs, "name IN ?", t.signatures); err != nil {
		test.Fatalf("failed to find signatures: %v", err)
	}

	if len(sigs) != len(t.signatures) {
		test.Fatalf("expected %d, got %d signatures", len(t.signatures), len(sigs))
	}
}

var sigTests = [...]*sigTester{
	{
		signatures: []string{"test", "embedded"},
		modules:    []string{"test"},
	},
}

func TestSignatures(t *testing.T) {
	for _, cfg := range sigTests {
		cfg.runTest(t)
	}
}
