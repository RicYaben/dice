package dice

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type DatabaseLocation string

const (
	NO_DATABASE       DatabaseLocation = ""
	INMEMORY_DATABASE DatabaseLocation = "file::memory:?cache=shared"
)

type Repository interface {
	WithTransaction(fn func(*gorm.DB) error) error
	connect() (*gorm.DB, error)
	changeLocation(l string)
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

func (r *repository) changeLocation(l string) {
	r.location = l
	r.db = nil
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
	wk := r.conf.WorkspaceFs()

	var srcs []*Source
	// globs are just fine. It takes some time to iterate through all the
	// patterns, but it adds flexibility
	withGlob := func(glob string) ([]*Source, error) {

		var gSrcs []*Source
		err := afero.Walk(wk, ".", func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			match, err := filepath.Match(glob, path)
			if err != nil {
				return err
			}

			if !match {
				return nil
			}

			format := filepath.Ext(path)
			if !slices.Contains(ext, format) {
				return nil
			}

			gSrcs = append(gSrcs, &Source{
				Name:     info.Name(),
				Location: path,
				Type:     SourceFile,
				Format:   format,
			})
			return nil
		})

		if err != nil {
			return nil, err
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
		q := d.Find(&h, Hook{ObjectID: id})
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
		}).Create(&h)

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
		q := d.Create(&f)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create fingerprint(s)")
		}
		return nil
	})
}

func (r *cosmosRepo) addLabel(l ...*Label) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Create(&l)
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
		q := d.Create(&s)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create scan(s)")
		}
		return nil
	})
}

func (r *cosmosRepo) addSource(s ...*Source) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		q := d.Create(&s)
		if err := q.Error; err != nil {
			return errors.Wrap(err, "failed to create source(s)")
		}
		return nil
	})
}

func (r *cosmosRepo) find(m, q any, args ...any) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		res := d.Where(q, args...).Find(m)
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
		q := d.
			Omit(clause.Associations).
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&s)

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
		}).Create(&m)

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

func (r *signatureRepo) find(m, q any, args ...any) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		res := d.Where(q, args...).Find(m)
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

	return sig, nil
}

func (r *signatureRepo) findFiles(t string, globs []string) ([]string, error) {
	g := make([]string, len(globs))
	copy(g, globs)

	var fs afero.Fs
	switch t {
	case "signature":
		fs = r.conf.SignaturesFs()
		for i, gl := range g {
			if !strings.HasSuffix(gl, ".dice") {
				g[i] = gl + ".dice"
			}
		}
	case "module":
		fs = r.conf.ModulesFs()
	default:
		return nil, errors.Errorf("unable to find DICE-related files of type %s", t)
	}

	withGlob := func(glob string) ([]string, error) {
		var gPaths []string
		err := afero.Walk(fs, ".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			match, err := filepath.Match(glob, path)
			if err != nil || !match || info.IsDir() {
				return err
			}

			gPaths = append(gPaths, path)
			return nil
		})

		if err != nil {
			return nil, err
		}
		return gPaths, nil
	}

	var sPaths []string
	for _, glob := range g {
		globF, err := withGlob(glob)
		if err != nil {
			return nil, err
		}
		sPaths = append(sPaths, globF...)
	}
	return sPaths, nil
}

type projectRepo struct {
	Repository
}

// Add a new project to the database and initialize it
func (r *projectRepo) addProject(proj ...*Project) (err error) {
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

func (r *projectRepo) find(m, q any, args ...any) error {
	return r.WithTransaction(func(d *gorm.DB) error {
		res := d.Where(q, args...).Find(m)
		if err := res.Error; err != nil {
			return err
		}
		return nil
	})
}

type repositoryBuilder struct {
	dbDir        string
	workspaceDir string
	location     string
	config       *gorm.Config
	models       []any
}

// Creates a repository builder.
// The home string indicates where databases are located
// The workspace indicates where files will be created
func newRepositoryBuilder(dbDir, wkDir string) *repositoryBuilder {
	return &repositoryBuilder{
		dbDir:        dbDir,
		workspaceDir: wkDir,
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
	switch b.dbDir {
	case "-":
		return b.setLocation(string(INMEMORY_DATABASE))
	default:
		return b.setLocation(path.Join(b.dbDir, n))
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
		builder: newRepositoryBuilder(conf.Databases(), conf.Workspace()),
	}
}

func (r *repositoryRegistry) Signatures() *signatureRepo {
	if r.signatures != nil {
		return r.signatures
	}

	models := []any{&Signature{}, &Module{}, &Node{}}
	repo := r.builder.setModels(models).setName("signatures.db").build()
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
	repo := r.builder.setModels([]any{&Project{}}).setName("projects.db").build()
	r.projects = &projectRepo{repo}
	return r.projects
}

func (r *repositoryRegistry) Cosmos() *cosmosRepo {
	if r.cosmos != nil {
		return r.cosmos
	}

	// cosmos goes into the current directory
	// TODO: not sure about this. Cosmos should go to the workspace
	b := newRepositoryBuilder(r.conf.Workspace(), r.conf.Workspace())
	repo := b.
		setModels([]any{&Host{}, &Fingerprint{}, &Label{}, &Hook{}}).
		setName("cosmos.db").
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
