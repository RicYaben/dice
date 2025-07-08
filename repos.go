package dice

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DatabaseLocation string

const (
	NO_DATABASE       DatabaseLocation = ""
	INMEMORY_DATABASE DatabaseLocation = ":memory:"
)

type Repository interface {
	WithTransaction(fn func(*gorm.DB) error) error
	connect() (*gorm.DB, error)
}

type repository struct {
	Repository

	db *gorm.DB

	location string
	config   *gorm.Config
	models   []any
}

// do whatever within a separate withTransaction
func (r *repository) WithTransaction(fn func(conn *gorm.DB) error) error {
	if _, err := r.connect(); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx) // pass new repo to handler
	})
}

func (r *repository) connect() (*gorm.DB, error) {
	if r.db != nil {
		return r.db, nil
	}

	db, err := gorm.Open(sqlite.Open(r.location), r.config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database connection")
	}

	db = db.Exec("PRAGMA foreign_keys = ON")
	if err := db.AutoMigrate(r.models...); err != nil {
		return nil, err
	}
	r.db = db

	return db, nil
}

type sourceRepo struct {
	Repository

	currProject Project
	currScan    string
}

func (r *sourceRepo) addSource(s *Source) error {
	return r.WithTransaction(func(conn *gorm.DB) error {
		sourceQ := conn.Create(s)
		if err := sourceQ.Error; err != nil {
			return errors.Wrap(err, "failed to create source")
		}
		return nil
	})
}

func (r *sourceRepo) getSources(u ...uint) ([]*Source, error) {
	panic("not implemented yet")
}

// func (e *engine) findSources(scan string, names, globs []string) ([]*SourceModel, error) {
// 	var srcs []*SourceModel

// 	// Location of the scan
// 	for _, name := range names {
// 		// Scan source location
// 		sr, err := e.conn.sources.find(scan, name, globs)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "failed to locate source")
// 		}
// 		srcs = append(srcs, sr...)
// 	}
// 	return srcs, nil
// }

// Locates source files inside a scan by the name of the source
func (r *sourceRepo) findSourceFiles(spath, sname string, globs []string) ([]*Source, error) {
	fpath := filepath.Join(spath, sname)
	info, err := os.Stat(spath)
	if os.IsNotExist(err) {
		return nil, errors.Wrap(err, "source not found")
	}

	if !info.IsDir() {
		return nil, errors.Wrap(err, "source path not a directory")
	}

	var srcs []*Source
	// globs are just fine. It takes some time to iterate through all the
	// patterns, but it adds flexibility
	withGlob := func(glob string) ([]*Source, error) {
		matches, err := filepath.Glob(filepath.Join(fpath, glob))
		if err != nil {
			return nil, errors.Wrap(err, "invalid glob pattern")
		}

		var gSrcs []*Source
		for _, match := range matches {
			_, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue // we cannot stat this, or is a dir
			}

			gSrcs = append(gSrcs, &Source{
				Name:     sname,
				Location: match,
				Type:     SourceFile,
				Format:   filepath.Ext(match),
			})
		}
		return gSrcs, nil
	}

	for _, glob := range globs {
		globSrcs, err := withGlob(glob)
		if err != nil {
			return nil, err
		}
		srcs = append(srcs, globSrcs...)
	}
	return srcs, nil
}

type cosmosRepo struct {
	Repository
}

// returns a host by id
func (r *cosmosRepo) getHost(id uint) (*Host, error) {
	panic("not implemented yet")
}

func (r *cosmosRepo) getHosts(id ...uint) ([]*Host, error) {
	panic("not implemented yet")
}

func (r *cosmosRepo) getLabels(id ...uint) ([]*Label, error) {
	panic("not implemented yet")
}

func (r *cosmosRepo) getFingerprints(id ...uint) ([]*Fingerprint, error) {
	panic("not implemented yet")
}

func (r *cosmosRepo) addHost(h *Host) error {
	panic("not implemented yet")
}

func (r *cosmosRepo) addFingerprint(f *Fingerprint) error {
	panic("not implemented yet")
}

func (r *cosmosRepo) addLabel(l *Label) error {
	panic("not implemented yet")
}

type eventRepo struct {
	Repository
}

func (r *eventRepo) addEvent(event Event) error {
	panic("not implemented yet")
}

func (r *eventRepo) getEvents(u ...uint) ([]*Event, error) {
	panic("not impplemented yet")
}

type signatureRepo struct {
	Repository
}

func (r *signatureRepo) addSignature(sig *Signature) error {
	panic("not implemented yet")
}

func (r *signatureRepo) getSignatures(u ...uint) (*Signature, error) {
	panic("not implemented yet")
}

func (r *signatureRepo) removeSignatures(u ...uint) error {
	panic("not implemented yet")
}

// TODO: remove - just remove the db
func (r *signatureRepo) deleteSignatures() error {
	return r.WithTransaction(func(conn *gorm.DB) error {
		q := conn.Session(&gorm.Session{AllowGlobalUpdate: true})
		if err := q.Unscoped().Select(clause.Associations).Delete(&Signature{}).Error; err != nil {
			return fmt.Errorf("failed to delete signatures with associations: %w", err)
		}
		return nil
	})
}

// TODO: remove
func (r *signatureRepo) getSignature(name string) (*Signature, error) {
	panic("not implemented yet")
}

// TODO: remove
func (r *signatureRepo) findSignatures(names []string) ([]string, error) {
	var found []string

	return found, r.WithTransaction(func(conn *gorm.DB) error {
		q := conn.Model(&Signature{}).Where("name IN ?", names).Pluck("name", &found)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find signatures")
		}
		return nil
	})
}

// TODO: remove
func (r *signatureRepo) saveSignatures(signatures []*Signature) error {
	if len(signatures) == 0 {
		// no signatures to store
		return nil
	}

	conn, err := r.connect()
	if err != nil {
		return err
	}

	q := conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).
		Omit("Scans.Probes", "Scans.Rules", "Rules.Track").
		Create(signatures)

	if err := q.Error; err != nil {
		return fmt.Errorf("failed to create signatures: %w", err)
	}

	if err := conn.Save(signatures).Error; err != nil {
		return fmt.Errorf("failed to save signatures: %w", err)
	}
	return nil
}

type projectRepo struct {
	Repository

	currProject Project
}

// Add projects to the database
// If successfully added, it creates DICE project files
// in the project location
// NOTE: projects have unique name-location pairs!
func (r *projectRepo) addProject(proj *Project) error {
	panic("not implemented yet")
}

// Retrieve projects from the database
func (r *projectRepo) getProjects(u ...uint) ([]*Project, error) {
	panic("not implemented yet")
}

// Remove projects form the database
func (r *projectRepo) removeProjects(projs ...*Project) error {
	panic("not implemented yet")
}

// Deletes all projects in the database
func (r *projectRepo) deleteProjects() error {
	return r.WithTransaction(func(conn *gorm.DB) error {
		q := conn.Session(&gorm.Session{AllowGlobalUpdate: true})
		if err := q.Unscoped().Select(clause.Associations).Delete(&Project{}).Error; err != nil {
			return errors.Wrap(err, "failed to delete projects with associations")
		}
		return nil
	})
}

type repositoryBuilder struct {
	home      string
	workspace string
	location  string
	config    *gorm.Config
	models    []any
}

func newRepositoryBuilder(home, workspace string) *repositoryBuilder {
	return &repositoryBuilder{
		home:      home,
		workspace: workspace,
		config: &gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
		},
	}
}

func (b *repositoryBuilder) setLocation(name string) *repositoryBuilder {
	b.location = name
	return b
}

func (b *repositoryBuilder) setName(n string) *repositoryBuilder {
	switch b.home {
	case "-":
		return b.setLocation(string(INMEMORY_DATABASE))
	default:
		return b.setLocation(path.Join(b.home, n))
	}
}

func (b *repositoryBuilder) setModels(m []any) *repositoryBuilder {
	b.models = m
	return b
}

func (b *repositoryBuilder) reset() {
	b.models = nil
	b.location = ""
}

func (b *repositoryBuilder) build() *repository {
	repo := &repository{
		config:   b.config,
		location: b.location,
		models:   b.models,
	}
	defer b.reset()
	return repo
}

type repositoryRegistry struct {
	builder *repositoryBuilder

	signatures *signatureRepo
	projects   *projectRepo
	events     *eventRepo
	cosmos     *cosmosRepo
	sources    *sourceRepo
}

func newRepositoryFactory(home, workspace string) *repositoryRegistry {
	return &repositoryRegistry{
		builder: newRepositoryBuilder(home, workspace),
	}
}

func (r *repositoryRegistry) Signatures() *signatureRepo {
	if r.signatures != nil {
		return r.signatures
	}

	models := []any{&Signature{}, &Module{}, &Node{}}
	repo := r.builder.setModels(models).setName("signatures.db").build()
	r.signatures = &signatureRepo{repo}
	return r.signatures
}

func (r *repositoryRegistry) Projects() *projectRepo {
	if r.projects != nil {
		return r.projects
	}
	repo := r.builder.setModels([]any{&Project{}}).setName("projects.db").build()
	r.projects = &projectRepo{Repository: repo}
	return r.projects
}

func (r *repositoryRegistry) Events() *eventRepo {
	if r.events != nil {
		return r.events
	}
	// Events db is always in memory
	b := newRepositoryBuilder("-", "-")
	repo := b.setModels([]any{&Event{}}).build()
	r.events = &eventRepo{repo}
	return r.events
}

func (r *repositoryRegistry) Cosmos() *cosmosRepo {
	if r.cosmos != nil {
		return r.cosmos
	}

	// cosmos goes into the current directory
	// TODO: not sure about this. Cosmos should go to the workspace
	b := newRepositoryBuilder(".", r.builder.workspace)
	repo := b.
		setModels([]any{&Host{}, &Fingerprint{}, &Label{}, &Hook{}}).
		setName("cosmos.db").
		build()
	r.cosmos = &cosmosRepo{repo}
	return r.cosmos
}

func (r *repositoryRegistry) Sources() *sourceRepo {
	// Sources goes into memory
	b := newRepositoryBuilder("-", "-")
	repo := b.setModels([]any{&Source{}}).build()
	r.sources = &sourceRepo{Repository: repo}
	return r.sources
}
