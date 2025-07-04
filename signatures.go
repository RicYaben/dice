package dice

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const S_EXT = ".dice"

type SignatureFlags struct {
	// Paths where to find signatures
	Allowed []string `json:"allowed"`
	// Required Required
	// defaults to []string{"*"}
	Required []string `json:"required"`
	// Scanning Mode
	Mode string `json:"mode"`
}

type SignatureDBMode string

const (
	// Skip scanning
	S_DB_SKIP SignatureDBMode = "skip"
	// Scan only for missing signatures
	S_DB_MISSING_ONLY SignatureDBMode = "missing"
	// Scan and parse signatures. Updates entries in db
	S_DB_UPDATE SignatureDBMode = "update"
	// Delete signatures from the
	// database and scan signatures
	S_RESET SignatureDBMode = "reset"
)

type SignatureLoader interface {
	Scan(required []string) error
}

type signatureLoader struct {
	allowed []string
	seen    []string
	check   bool

	parser *signatureParser
	repo   *signatureRepo
}

// get the name of the referenced file and remove the file extension
func (s *signatureLoader) slugify(fpath string) string {
	name := filepath.Base(fpath)
	name = strings.TrimSuffix(name, S_EXT)

	// Slugging the name even further may create issues
	// strings.ReplaceAll(name, " ", "-")
	return name
}

// searches for a glob of .dice signatures, e.g.: **/*
func (s *signatureLoader) searchGlob(patterns []string) ([]string, error) {
	var fpaths []string
	search := func(fpath string, info os.FileInfo, err error) error {
		for _, pattern := range patterns {
			fpath := filepath.Join(fpath, pattern+S_EXT)
			matches, err := filepath.Glob(fpath)
			if err != nil {
				return fmt.Errorf("failed to search with pattern %s: %w", pattern, err)
			}
			fpaths = append(fpaths, matches...)
		}
		return nil
	}

	for _, root := range s.allowed {
		if err := filepath.Walk(root, search); err != nil {
			return nil, err
		}
	}
	return fpaths, nil
}

// searches dice signatures by walking the allowed paths
func (s *signatureLoader) searchFiles(names []string) ([]string, error) {
	var fpaths []string
	search := func(fpath string, info os.FileInfo, err error) error {
		name := s.slugify(fpath)
		if slices.Contains(names, name) {
			fpaths = append(fpaths, fpath)
		}
		return nil
	}

	for _, root := range s.allowed {
		if err := filepath.Walk(root, search); err != nil {
			return nil, err
		}
	}
	return fpaths, nil
}

// searches for globs or paths
func (s *signatureLoader) search(names []string, glob bool) ([]string, error) {
	if glob {
		return s.searchGlob(names)
	}
	return s.searchFiles(names)
}

// parse a signature given its filepath
func (s *signatureLoader) parse(fpath string) (*Signature, error) {
	sig := Signature{
		Name: s.slugify(fpath),
	}

	f, err := os.Open(fpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open signature file: %w", err)
	}
	defer f.Close()

	if err := s.parser.Parse(&sig, f); err != nil {
		return nil, fmt.Errorf("failed to parse signature file: %w", err)
	}

	s.seen = append(s.seen, sig.Name)
	return &sig, nil
}

// check for signatures not stored from a list of names
func (s *signatureLoader) findMissing(names []string) ([]string, error) {

	// check signatures already stored
	found, err := s.repo.findSignatures(names)
	if err != nil {
		return nil, fmt.Errorf("failed to find signatures in database: %w", err)
	}

	var missing []string
	for _, name := range names {
		if slices.Contains(found, name) {
			continue
		}
		missing = append(missing, name)
	}
	return missing, nil
}

// find embedded signatures
func (s *signatureLoader) embedded(sigs []*Signature) []string {
	var names []string
	for _, sig := range sigs {
		for _, rule := range sig.Rules {
			if RuleMode(rule.Mode) == R_SOURCE {
				names = append(names, rule.Source)
			}
		}
	}
	return names
}

// find signatures, store them, and recursively iterate the embedded ones until all signatures are stored
func (s *signatureLoader) scan(names []string, glob bool) error {
	if len(names) == 0 {
		return nil
	}

	if s.check {
		missing, err := s.findMissing(names)
		if err != nil {
			return fmt.Errorf("failed to find missing signatures: %w", err)
		}
		names = missing
	}

	fpaths, err := s.search(names, glob)
	if err != nil {
		return fmt.Errorf("failed to find signature files: %w", err)
	}

	var signatures []*Signature
	for _, fpath := range fpaths {
		sig, err := s.parse(fpath)
		if err != nil {
			return fmt.Errorf("failed to parse signatures: %w", err)
		}
		signatures = append(signatures, sig)
	}

	if err := s.repo.saveSignatures(signatures); err != nil {
		return fmt.Errorf("failed to store signature: %w", err)
	}

	embedded := s.embedded(signatures)
	return s.scan(embedded, false)
}

type missingSignatureLoader struct {
	signatureLoader
}

func MissingSignaturesLoader(scn signatureLoader) SignatureLoader {
	scn.check = true
	return &missingSignatureLoader{
		signatureLoader: scn,
	}
}

func (s *missingSignatureLoader) Scan(patterns []string) error {
	return s.scan(patterns, true)
}

type updateSignatureLoader struct {
	signatureLoader
}

func UpdateSignaturesLoader(scn signatureLoader) SignatureLoader {
	scn.check = false
	return &updateSignatureLoader{
		signatureLoader: scn,
	}
}

func (s *updateSignatureLoader) Scan(patterns []string) error {
	return s.scan(patterns, true)
}

type resetSignatureLoader struct {
	signatureLoader
}

func ResetSignaturesLoader(scn signatureLoader) SignatureLoader {
	return &resetSignatureLoader{
		signatureLoader: scn,
	}
}

func (s *resetSignatureLoader) Scan(patterns []string) error {
	if err := s.repo.deleteSignatures(); err != nil {
		return err
	}
	scn := UpdateSignaturesLoader(s.signatureLoader)
	return scn.Scan(patterns)
}

// Scan for signatures in the given folder.
// The scan mode stays how and which signatures to scan for.
func LoadSignatures(repo *signatureRepo, flags SignatureFlags) error {
	var wrap func(signatureLoader) SignatureLoader
	switch mode := SignatureDBMode(flags.Mode); mode {
	case S_DB_SKIP:
		return nil
	case S_DB_MISSING_ONLY:
		wrap = MissingSignaturesLoader
	case S_DB_UPDATE:
		wrap = UpdateSignaturesLoader
	case S_RESET:
		wrap = ResetSignaturesLoader
	default:
		return fmt.Errorf("mode not found: %v", mode)
	}

	base := signatureLoader{
		repo:    repo,
		allowed: flags.Allowed,
		parser:  makeSignatureParser(),
	}

	scn := wrap(base)
	return scn.Scan(flags.Required)
}
