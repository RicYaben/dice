package project

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// A project is a collection of measurements. The project dictates what is accepted
// to do during the measurements.
type Project struct {
	gorm.Model

	Serial      string
	Title       string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	// License, approval from the ethics comittee?

	Tags         datatypes.JSON
	Measurements []*Experiment
}

func (s *Project) Summary() map[string]string {
	return map[string]string{
		"serial":      s.Serial,
		"title":       s.Title,
		"description": s.Description,
		"start-date":  s.StartDate.String(),
		"end-date":    s.EndDate.String(),
	}
}

func (s *Project) Save(store FileStore) error {
	fpath := filepath.Join(s.Serial, "metadata.json")

	summary, _ := json.Marshal(s.Summary())
	if err := store.Save(fpath, summary); err != nil {
		return fmt.Errorf("failed to store study %s: %w", s.Serial, err)
	}
	return nil
}

type Experiment struct {
	gorm.Model

	ProjectID uint

	Serial string
	Path   string
	// summary of sources, protocols, stuff like that?
	Metadata datatypes.JSON
	Sources  []*Source
}

func (m *Experiment) Summary() string {
	return m.Metadata.String()
}

func (m *Experiment) Save(store FileStore) error {
	fpath := filepath.Join(m.Path, "metadata.json")

	summary, _ := json.Marshal(m.Summary())
	if err := store.Save(fpath, summary); err != nil {
		return fmt.Errorf("failed to store study %s: %w", m.Serial, err)
	}
	return nil
}

type Source struct {
	gorm.Model

	ExperimentID uint

	Serial string
	Path   string

	Metadata datatypes.JSON
	Hash     string
}

func (s *Source) Save(store FileStore) error {
	fpath := filepath.Join(s.Path, fmt.Sprintf("%s.tar", s.Serial))

	if err := store.Save(fpath, []byte{}); err != nil {
		return fmt.Errorf("failed to create source tarball: %w", err)
	}
	return nil
}
