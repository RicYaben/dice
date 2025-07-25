package dice

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hashicorp/golang-lru/v2/expirable"
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
	conf *Configuration
}

func (r *sourceRepo) addSource(s ...*Source) error {
	return r.WithTransaction(func(conn *gorm.DB) error {
		sourceQ := conn.Create(s)
		if err := sourceQ.Error; err != nil {
			return errors.Wrap(err, "failed to create source")
		}
		return nil
	})
}

// Locates source files inside a scan by the name of the source
func (r *sourceRepo) findSourceFiles(globs, ext []string) ([]*Source, error) {
	// current workspace
	fpath := r.conf.Workspace()

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
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue // we cannot stat this, or is a dir
			}

			format := filepath.Ext(match)
			if !slices.Contains(ext, format) {
				continue
			}

			gSrcs = append(gSrcs, &Source{
				Name:     info.Name(),
				Location: match,
				Type:     SourceFile,
				Format:   format,
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
	cache *expirable.LRU[uint, *Host]
}

// returns a host by id
func (r *cosmosRepo) getHost(id uint) (*Host, error) {
	if host, ok := r.cache.Get(id); ok {
		return host, nil
	}

	var h *Host
	return h, r.WithTransaction(func(d *gorm.DB) error {
		q := d.First(h, id)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find host")
		}
		r.cache.Add(h.ID, h)
		return nil
	})
}

func (r *cosmosRepo) getFingerprint(id uint) (*Fingerprint, error) {
	var fp *Fingerprint
	return fp, r.WithTransaction(func(d *gorm.DB) error {
		q := d.First(fp, id)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find fingerprint")
		}
		return nil
	})
}

func (r *cosmosRepo) getLabel(id uint) (*Label, error) {
	var lab *Label
	return lab, r.WithTransaction(func(d *gorm.DB) error {
		q := d.First(lab, id)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find label")
		}
		return nil
	})
}

func (r *cosmosRepo) getScan(id uint) (*Scan, error) {
	var sc *Scan
	return sc, r.WithTransaction(func(d *gorm.DB) error {
		q := d.First(sc, id)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find scan")
		}
		return nil
	})
}

func (r *cosmosRepo) getSource(id uint) (*Source, error) {
	var sc *Source
	return sc, r.WithTransaction(func(d *gorm.DB) error {
		q := d.First(sc, id)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find source")
		}
		return nil
	})
}

func (r *cosmosRepo) getHooks(id uint) ([]*Hook, error) {
	var h []*Hook
	return h, r.WithTransaction(func(d *gorm.DB) error {
		q := d.Find(h, Hook{ObjectID: id})
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find labels")
		}
		return nil
	})
}

func (r *cosmosRepo) addHost(h ...*Host) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(h)

		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create host(s)")
		}

		for _, host := range h {
			r.cache.Add(host.ID, host)
		}
		return nil
	})
}

func (r *cosmosRepo) addFingerprint(f ...*Fingerprint) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Create(f)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create fingerprint(s)")
		}
		return nil
	})
}

func (r *cosmosRepo) addLabel(l ...*Label) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Create(l)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create label(s)")
		}

		// Expire the hosts linked to these labels
		for _, lab := range l {
			r.cache.Remove(lab.HostID)
		}
		return nil
	})
}

func (r *cosmosRepo) addScan(s ...*Scan) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Create(s)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create scan(s)")
		}
		return nil
	})
}

func (r *cosmosRepo) addSource(s ...*Source) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Create(s)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create source(s)")
		}
		return nil
	})
}

func (r *cosmosRepo) find(m any, q any) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		res := d.Where(m).Find(q)
		if err := res.Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *cosmosRepo) query(m any) ([]*Host, error) {
	var hosts []*Host
	return hosts, r.WithTransaction(func(d *gorm.DB) error {
		res := d.Where(m).Find(hosts)
		if err := res.Error; err != nil {
			return err
		}
		return nil
	})
}

type signatureRepo struct {
	Repository
	parser Parser
	conf   *Configuration
}

func (r *signatureRepo) addSignature(s ...*Signature) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(s)

		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create signature(s)")
		}
		return nil
	})
}

func (r *signatureRepo) addModule(m ...*Module) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(m)

		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create module(s)")
		}
		return nil
	})
}

func (r *signatureRepo) getSignature(u uint) (*Signature, error) {
	var res *Signature
	return res, r.WithTransaction(func(d *gorm.DB) error {
		q := d.First(res, u)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find fingerprint")
		}
		return nil
	})
}

func (r *signatureRepo) getModule(u uint) (*Module, error) {
	var res *Module
	return res, r.WithTransaction(func(d *gorm.DB) error {
		q := d.First(res, u)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find module")
		}
		return nil
	})
}

func (r *signatureRepo) remove(m any, q any) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Delete(m, q)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to remove fingerprints")
		}
		return nil
	})
}

func (r *signatureRepo) deleteAll() error {
	return r.WithTransaction(func(conn *gorm.DB) error {
		q := conn.Session(&gorm.Session{AllowGlobalUpdate: true})
		if err := q.Unscoped().Select(clause.Associations).Delete(&Signature{}, &Module{}).Error; err != nil {
			return fmt.Errorf("failed to delete signatures with associations: %w", err)
		}
		return nil
	})
}

func (r *signatureRepo) getRoots(id uint) ([]*Node, error) {
	var roots []*Node
	return roots, r.WithTransaction(func(conn *gorm.DB) error {
		q := conn.Raw(`
			SELECT *
			FROM nodes AS n
			WHERE n.signature_id = ?
			AND NOT EXISTS (
				SELECT 1
				FROM node_children AS nc
				JOIN nodes AS parent ON nc.node_id = parent.id
				WHERE nc.child_id = n.id
				AND parent.signature_id = n.signature_id
			)
		`, id).Scan(&roots)

		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to find signature roots")
		}
		return nil
	})
}

func (r *signatureRepo) find(m any, q any) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		res := d.Where(m).Find(q)
		if err := res.Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *signatureRepo) parseSignatureFile(fpath string) (*Signature, error) {
	info, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}

	f, oErr := os.Open(fpath)
	if oErr != nil {
		return nil, err
	}
	defer f.Close()

	sig, err := r.parser.Parse(info.Name(), f)
	if err != nil {
		return nil, err
	}

	return &sig, nil
}

func (r *signatureRepo) findFiles(t string, globs []string) ([]string, error) {
	var fpath string
	switch t {
	case "signature":
		fpath = r.conf.Signatures()
		for i, g := range globs {
			if !strings.HasPrefix(g, ".dice") {
				globs[i] = g + ".dice"
			}
		}
	case "module":
		fpath = r.conf.Modules()
	default:
		return nil, errors.Errorf("unable to find DICE-related files of type %s", t)
	}

	withGlob := func(glob string) ([]string, error) {
		matches, err := filepath.Glob(filepath.Join(fpath, glob))
		if err != nil {
			return nil, errors.Wrap(err, "invalid glob pattern")
		}

		var gPaths []string
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue // we cannot stat this, or is a dir
			}

			gPaths = append(gPaths, match)
		}
		return gPaths, nil
	}

	var sPaths []string
	for _, glob := range globs {
		globSigs, err := withGlob(glob)
		if err != nil {
			return nil, err
		}
		sPaths = append(sPaths, globSigs...)
	}
	return sPaths, nil
}

type projectRepo struct {
	Repository
}

// Add projects to the database
// If successfully added, it creates DICE project files
// in the project location
// NOTE: projects have unique name-location pairs!
func (r *projectRepo) remove(proj ...*Project) (err error) {
	for _, p := range proj {
		if rerr := os.RemoveAll(path.Join(p.Path, ".dice")); err != nil {
			err = errors.Wrap(rerr, "failed to clear project")
		}
	}
	return
}

func (r *projectRepo) addProject(proj ...*Project) (err error) {
	defer func() {
		if err != nil {
			r.remove(proj...)
		}
	}()

	for _, p := range proj {
		f, err := os.Create(path.Join(p.Path, ".dice"))
		if err != nil {
			return err
		}

		// Store the settings. Noramlly empty on creation
		if _, err := f.Write(p.Settings); err != nil {
			return err
		}
	}

	return r.WithTransaction(func(conn *gorm.DB) error {
		q := conn.Create(proj)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create project(s)")
		}
		return nil
	})
}

func (r *projectRepo) addStudy(s ...*Study) (err error) {
	panic("not implemented yet")
}

func (r *projectRepo) find(m any, q any) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		res := d.Where(m).Find(q)
		if err := res.Error; err != nil {
			return err
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
	conf    *Configuration
	builder *repositoryBuilder

	signatures *signatureRepo
	projects   *projectRepo
	cosmos     *cosmosRepo
	sources    *sourceRepo
}

func newRepositoryFactory(conf *Configuration) *repositoryRegistry {
	return &repositoryRegistry{
		conf:    conf,
		builder: newRepositoryBuilder(conf.Home(), conf.Workspace()),
	}
}

func (r *repositoryRegistry) Signatures() *signatureRepo {
	if r.signatures != nil {
		return r.signatures
	}

	models := []any{&Signature{}, &Module{}, &Node{}}
	repo := r.builder.setModels(models).setName("signatures").build()
	r.signatures = &signatureRepo{
		repo,
		NewParser(),
		r.conf,
	}
	return r.signatures
}

func (r *repositoryRegistry) Projects() *projectRepo {
	if r.projects != nil {
		return r.projects
	}
	repo := r.builder.setModels([]any{&Project{}}).setName("projects").build()
	r.projects = &projectRepo{repo}
	return r.projects
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
		setName("cosmos").
		build()

	cache := expirable.NewLRU[uint, *Host](1e3, nil, 5*time.Minute)
	r.cosmos = &cosmosRepo{repo, cache}
	return r.cosmos
}

func (r *repositoryRegistry) Sources() *sourceRepo {
	// Sources goes into memory
	b := newRepositoryBuilder("-", "-")
	repo := b.setModels([]any{&Source{}}).build()
	r.sources = &sourceRepo{Repository: repo}
	return r.sources
}
