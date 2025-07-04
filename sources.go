// Project sources. Sources of data
//
// Usage:
// newSource("targets", input, handler)
//
// source.Records() // get an Iterator over the input records
package dice

import (
	"encoding/json"
	"io"
	"iter"

	"github.com/pkg/errors"
)

// Makes a new source from arguments
func makeTargetArgsSource(args []string) (*SourceModel, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return &SourceModel{
		Name: "targets",
		Type: SourceArgs,
		Args: b,
	}, nil
}

type RecordsIterator iter.Seq2[[]byte, error]
type RecordsReader func(io.Reader) RecordsIterator

// A single data source that provides interfaces to iterate over records with different
// handlers.This is mainly used when loading
type Source struct {
	// name of the source
	Name string
	// some input
	input io.Reader
	// a function that takes a reader and returns an iterator
	handler RecordsReader
}

func newSource(name string, input io.Reader, handler RecordsReader) *Source {
	return &Source{name, input, handler}
}

func (s *Source) Records() RecordsIterator {
	return s.handler(s.input)
}
