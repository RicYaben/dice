package dice

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Makes a new source from arguments
func MakeTargetArgsSource(args []string) (*Source, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return &Source{
		Name: "targets",
		Type: SourceArgs,
		Args: b,
	}, nil
}

// Filter values from a slice
func Filter[T any](s []T, fn func(T) bool) []T {
	var r []T
	for _, t := range s {
		if fn(t) {
			r = append(r, t)
		}
	}
	return r
}
