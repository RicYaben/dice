package sdk

import (
	"io"
	"iter"
)

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
