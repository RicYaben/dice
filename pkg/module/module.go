package module

import (
	"github.com/dice/pkg/database"
)

type Module interface {
	// Returns the record updated
	Do(h *database.Host, r *database.Record) error
}
