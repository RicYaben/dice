package main

import (
	"fmt"

	"github.com/dice/pkg/database"
	"github.com/dice/pkg/signature"
)

type Flags struct {
	Signatures []string `json:"signatures" description:""`

	Signature signature.Flags
	ScanFlags signature.ScanFlags
}

func parseFlags() error {
	flags := new(Flags)

	// Create a connection to the database
	conf := database.GetDatabase(database.SIG_DB)
	sigDB, err := database.Open(conf)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Scan folders for signatures
	// TODO FIXME: this should return the IDs of the required signatures
	ids, err := signature.Scan(sigDB, flags.ScanFlags)
	if err != nil {
		return fmt.Errorf("failed while scanning signatures: %w", err)
	}

	// Load signatures
	var hyper []signature.Node
	reg := signature.Registry(sigDB, flags.Signature)
	for _, id := range ids {
		// TODO: Load must take an id, not the name anymore
		// we get the ids from the scan.
		g, err := reg.FindOrLoad(id)
		if err != nil {
			return fmt.Errorf("failed to load signature: %w", err)
		}
		hyper = append(hyper, g)
	}

	// Send to the right command and run it
	return cmd.Do(hyper)
}
