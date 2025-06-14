package signature

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dice/pkg/database"
	"github.com/dice/pkg/signature/parser"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const S_EXT = ".dice"

type ScanFlags struct {
	// Paths where to find signatures
	Allowed []string `json:"allowed"`
	// Required Required
	// defaults to []string{"*"}
	Required []string `json:"required"`
	// Scanning Mode
	Mode string `json:"mode"`
}

type ScanMode string

const (
	// Skip scanning
	S_SKIP ScanMode = "skip"
	// Scan only for missing signatures
	S_MISSING_ONLY ScanMode = "missing"
	// Scan and parse signatures. Updates entries in db
	S_UPDATE ScanMode = "update"
	// Delete signatures from the
	// database and scan signatures
	S_RESET ScanMode = "reset"
)

type Scanner interface {
	Scan(required []string) error
}

type scanner struct {
	allowed []string
	seen    []string
	check   bool

	parser *parser.SignatureParser
	tx     *gorm.DB
}

func (s *scanner) slugify(fpath string) string {
	name := filepath.Base(fpath)
	name = strings.TrimSuffix(name, S_EXT)

	// Slugging the name even further may create issues
	// strings.ReplaceAll(name, " ", "-")
	return name
}

func (s *scanner) searchGlob(patterns []string) ([]string, error) {
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

func (s *scanner) searchFiles(names []string) ([]string, error) {
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

func (s *scanner) search(names []string, glob bool) ([]string, error) {
	if glob {
		return s.searchGlob(names)
	}
	return s.searchFiles(names)
}

func (s *scanner) parse(fpath string) (*database.Signature, error) {
	sig := database.Signature{
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

func (s *scanner) searchNames(names []string) ([]string, error) {
	var found []string
	q := s.tx.Model(&database.Signature{}).Where("name IN ?", names).Pluck("name", &found)
	if err := q.Error; err != nil {
		return nil, fmt.Errorf("failed to find names: %w", err)
	}
	return found, nil
}

func (s *scanner) findMissing(names []string) ([]string, error) {

	// check signatures already stored
	found, err := s.searchNames(names)
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

func (s *scanner) store(signatures []*database.Signature) error {
	if len(signatures) == 0 {
		// no signatures to store
		return nil
	}

	q := s.tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).
		Omit("Scans.Probes", "Scans.Rules", "Rules.Track").
		Create(signatures)

	if err := q.Error; err != nil {
		s.tx.Rollback()
		return fmt.Errorf("failed to create signatures: %w", err)
	}

	if err := s.tx.Save(signatures).Error; err != nil {
		s.tx.Rollback()
		return fmt.Errorf("failed to save signatures: %w", err)
	}

	return nil
}

func (s *scanner) embedded(sigs []*database.Signature) []string {
	var names []string
	for _, sig := range sigs {
		for _, rule := range sig.Rules {
			if parser.RuleMode(rule.Mode) == parser.R_SOURCE {
				names = append(names, rule.Source)
			}
		}
	}
	return names
}

func (s *scanner) scan(names []string, glob bool) error {
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

	var signatures []*database.Signature
	for _, fpath := range fpaths {
		sig, err := s.parse(fpath)
		if err != nil {
			return fmt.Errorf("failed to parse signatures: %w", err)
		}
		signatures = append(signatures, sig)
	}

	if err := s.store(signatures); err != nil {
		return fmt.Errorf("failed to store signature: %w", err)
	}

	embedded := s.embedded(signatures)
	return s.scan(embedded, false)
}

type missingScanner struct {
	scanner
}

func Missing(scn scanner) Scanner {
	scn.check = true
	return &missingScanner{
		scanner: scn,
	}
}

func (s *missingScanner) Scan(patterns []string) error {
	return s.scan(patterns, true)
}

type updateScanner struct {
	scanner
}

func Update(scn scanner) Scanner {
	scn.check = false
	return &updateScanner{
		scanner: scn,
	}
}

func (s *updateScanner) Scan(patterns []string) error {
	return s.scan(patterns, true)
}

type resetScanner struct {
	scanner
}

func Reset(scn scanner) Scanner {
	return &resetScanner{
		scanner: scn,
	}
}

func (s *resetScanner) deleteSignatures() error {
	q := s.tx.Session(&gorm.Session{AllowGlobalUpdate: true})
	if err := q.Unscoped().Select(clause.Associations).Delete(&database.Signature{}).Error; err != nil {
		q.Rollback()
		return fmt.Errorf("failed to delete signatures with associations: %w", err)
	}
	return nil
}

func (s *resetScanner) Scan(patterns []string) error {
	if err := s.deleteSignatures(); err != nil {
		return err
	}
	scn := Update(s.scanner)
	return scn.Scan(patterns)
}

// Scan for signatures in the given folder.
// The scan mode stays how and which signatures to scan for.
func Scan(db *gorm.DB, flags ScanFlags) error {
	var wrap func(scanner) Scanner
	switch mode := ScanMode(flags.Mode); mode {
	case S_SKIP:
		return nil
	case S_MISSING_ONLY:
		wrap = Missing
	case S_UPDATE:
		wrap = Update
	case S_RESET:
		wrap = Reset
	default:
		return fmt.Errorf("mode not found: %v", mode)
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	base := scanner{
		tx:      tx,
		allowed: flags.Allowed,
		parser:  parser.Signature(),
	}

	scn := wrap(base)
	if err := scn.Scan(flags.Required); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to scan signatures: %w", err)
	}

	return tx.Commit().Error
}
