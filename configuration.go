package dice

import (
	"os"
	"path"
	"slices"

	"github.com/pkg/errors"
)

// Standard paths to use to store DICE related data
// https://specifications.freedesktop.org/basedir-spec/latest/
type StandardPaths struct {
	// Can be used to change the profile
	// Default: "dice"
	DICE_APPNAME string
	// Path to configuration directory.
	// Default: "$XDG_CONFIG_HOME/$DICE_APPNAME" or "$HOME/.config/$DICE_APPNAME" if unset
	CONFIG_HOME string
	// Path to state directory
	// Default: "$XDG_STATE_HOME/$DICE_APPNAME" or "$HOME/.local/state/$DICE_APPNAME" if unset
	STATE_HOME string
	// Path to data directory
	// Default: "$XDG_DATA_HOME/$DICE_APPNAME" or "$HOME/.local/share/$DICE_APPNAME"
	DATA_HOME string
}

func PWDStandardPaths() StandardPaths {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return StandardPaths{"dice", wd, wd, wd}
}

func (s StandardPaths) init() error {
	for _, p := range []string{s.CONFIG_HOME, s.STATE_HOME, s.DATA_HOME} {
		if err := os.MkdirAll(p, 0700); err != nil {
			return errors.Wrapf(err, "failed to create standard path: %s", p)
		}
	}
	return nil
}

type stdpathsBuilder struct {
	stdpaths *StandardPaths
	home     string

	app    string
	config string
	state  string
	data   string
}

func newStdpathsBuilder() *stdpathsBuilder {
	return &stdpathsBuilder{home: os.Getenv("HOME")}
}

func (b *stdpathsBuilder) withStdpaths(stdpaths *StandardPaths) *stdpathsBuilder {
	bcp := *b
	bcp.stdpaths = stdpaths
	return &bcp
}

func (b *stdpathsBuilder) isValid(val string) bool {
	return !slices.Contains([]string{"", "-"}, val)
}

func (b *stdpathsBuilder) bind(val, env, def string) string {
	if b.isValid(val) {
		return val
	}
	if v := os.Getenv(env); b.isValid(v) {
		return v
	}
	return def
}

func (b *stdpathsBuilder) bindToApp(val, env, def string) string {
	v := b.bind(val, env, def)
	if v == val {
		return val
	}
	return path.Join(v, b.app)
}

func (b *stdpathsBuilder) setApp(val string) *stdpathsBuilder {
	b.app = b.bindToApp(val, "DICE_APPNAME", "dice")
	return b
}

func (b *stdpathsBuilder) setConfig(val string) *stdpathsBuilder {
	b.config = b.bindToApp(val, "XDG_CONFIG_HOME", path.Join(b.home, ".config"))
	return b
}

func (b *stdpathsBuilder) setState(val string) *stdpathsBuilder {
	b.state = b.bindToApp(val, "XDG_STATE_HOME", path.Join(b.home, ".local", "state"))
	return b
}

func (b *stdpathsBuilder) setData(val string) *stdpathsBuilder {
	b.data = b.bindToApp(val, "XDG_DATA_HOME", path.Join(b.home, ".local", "share"))
	return b
}

func (b *stdpathsBuilder) build() *StandardPaths {
	stdpaths := b.stdpaths
	stdpaths.DICE_APPNAME = b.app
	stdpaths.CONFIG_HOME = b.config
	stdpaths.STATE_HOME = b.state
	stdpaths.DATA_HOME = b.data
	// Note: normally we reset here, but in this case is not needed
	return stdpaths
}

// Overrides empty standard paths. More of a purgue or clean job.
func BindStandardPaths(stdpaths *StandardPaths) *StandardPaths {
	b := newStdpathsBuilder().withStdpaths(stdpaths)
	return b.setApp(stdpaths.DICE_APPNAME).
		setConfig(stdpaths.CONFIG_HOME).
		setData(stdpaths.DATA_HOME).
		setState(stdpaths.STATE_HOME).
		build()
}

type Configuration struct {
	paths   StandardPaths
	project *Project
	study   *Study
}

// Returns the location where we store databases, modules, and signatures
func (c *Configuration) Home() string {
	return c.paths.DATA_HOME
}

func (c *Configuration) Project() string {
	if c.project != nil {
		return c.project.Path
	}
	return c.paths.STATE_HOME
}

// Returns the location where the current run will output data to
// If no project is set, return the data location
func (c *Configuration) Workspace() string {
	if c.study != nil {
		return c.study.Path
	}
	return c.Project()
}

func (c *Configuration) Signatures() string {
	return path.Join(c.Home(), "signatures")
}

func (c *Configuration) Modules() string {
	return path.Join(c.Home(), "modules")
}

func (c *Configuration) WithProject(p *Project) *Configuration {
	cp := *c
	cp.project = p
	return c
}

func LoadConfiguration(stdpaths StandardPaths, conf *Configuration) error {
	// initialize paths
	if err := stdpaths.init(); err != nil {
		return errors.Wrap(err, "failed to initialize standard paths")
	}

	conf.paths = stdpaths
	return nil
}
