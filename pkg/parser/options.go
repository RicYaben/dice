package parser

import (
	"fmt"
	"strings"
	"unicode"
)

type Option interface {
	string | []string
}

// 1. Get the option value
// 2. Separate by comma
// 3. For each value, check if there

func IsGroup(v string) bool {
	return (strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]"))
}

func GroupToSlice(v string) []string {
	v = strings.TrimPrefix(v, "[")
	v = strings.TrimSuffix(v, "]")
	return strings.Split(v, ",")
}

type Symbols struct {
	input string
	pos   int
}

func (s *Symbols) NextSymbol() rune {
	if s.pos >= len(s.input) {
		return rune(0) // End of input
	}
	ch := rune(s.input[s.pos])
	s.pos++
	return ch
}

// PeekSymbol peeks at the next symbol without advancing the position.
func (s *Symbols) PeekSymbol() rune {
	if s.pos >= len(s.input) {
		return rune(0) // End of input
	}
	return rune(s.input[s.pos])
}

func (s *Symbols) SkipWhitespace() {
	for unicode.IsSpace(s.PeekSymbol()) {
		s.NextSymbol()
	}
}

func (s *Symbols) ParseKey() string {
	var key []rune
	s.SkipWhitespace()
	for ch := s.PeekSymbol(); unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-'; ch = s.PeekSymbol() {
		key = append(key, ch)
		s.NextSymbol()
	}
	return string(key)
}

// ParseString parses a quoted string.
func (s *Symbols) ParseString() string {
	s.SkipWhitespace()
	if s.PeekSymbol() != '"' {
		return ""
	}
	s.NextSymbol() // Skip opening quote

	var str []rune
	for ch := s.PeekSymbol(); ch != rune(0) && ch != '"'; ch = s.PeekSymbol() {
		str = append(str, ch)
		s.NextSymbol()
	}

	s.NextSymbol() // Skip closing quote
	return string(str)
}

func (s *Symbols) ParseArray() string {
	s.SkipWhitespace()
	if s.PeekSymbol() != '[' {
		return ""
	}
	s.NextSymbol() // Skip opening bracket

	var arr []rune
	for ch := s.PeekSymbol(); ch != rune(0) && ch != ']'; ch = s.PeekSymbol() {
		arr = append(arr, ch)
		s.NextSymbol()
	}

	s.NextSymbol() // Skip closing bracket
	return string(arr)
}

func (s *Symbols) ParseValue() string {
	s.SkipWhitespace()
	// Check if it's a string
	if s.PeekSymbol() == '"' {
		return s.ParseString()
	}
	// Check if it's an array
	if s.PeekSymbol() == '[' {
		return s.ParseArray()
	}
	// Otherwise, it's a simple word (boolean, number, etc.)
	var value []rune
	for ch := s.PeekSymbol(); unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-'; ch = s.PeekSymbol() {
		value = append(value, ch)
		s.NextSymbol()
	}
	return string(value)
}

func (s *Symbols) Parse(line string) (map[string]string, error) {
	flags := make(map[string]string)
	s.pos = 0
	s.input = line

	for {
		s.SkipWhitespace()
		key := s.ParseKey() // Parse the key
		if key == "" {
			break
		}
		s.SkipWhitespace()

		// Parse the value (can be empty)
		value := s.ParseValue()

		// Store the key-value pair
		flags[key] = value

		// Skip whitespace and check for the next pair
		s.SkipWhitespace()

		// end of input
		if s.PeekSymbol() == rune(0) {
			break
		}

		if s.PeekSymbol() != ',' {
			return nil, fmt.Errorf("expected ',' between key-value pairs")
		}
		s.NextSymbol() // Skip the comma
	}

	return flags, nil
}
