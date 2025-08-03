package dice

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// TODO: the configuration should not give paths to store files, but
// an FS based on the paths in the config!
type Settings interface {
	// Base path storing DICE documents. (read-only)
	Home() afero.Fs
	// Directory storing signatures (read-only)
	SignaturesFs() afero.Fs
	// Directory storing Modules (read-only)
	ModulesFs() afero.Fs
	// Project directory, i.e., where to locate runtime files (rw)
	ProjectFs() afero.Fs
	// Workspace directory, i.e., where to output (rw)
	WorkspaceFs() afero.Fs

	Workspace() string
	Databases() string
}

// Standard paths to use to store DICE related data
// https://specifications.freedesktop.org/basedir-spec/latest/
type StandardPaths struct {
	// Can be used to change the profile
	// Default: "dice"
	DICE_APPNAME string `json:"app_name,omitempty"`
	// Path to state directory
	// Default: "$XDG_STATE_HOME/$DICE_APPNAME" or "$HOME/.local/state/$DICE_APPNAME"
	STATE_HOME string `json:"state,omitempty"`
	// Path to data directory
	// Default: "$XDG_DATA_HOME/$DICE_APPNAME" or "$HOME/.local/share/$DICE_APPNAME"
	DATA_HOME string `json:"data,omitempty"`
	// Path to configuration directory.
	// Default: "$XDG_CONFIG_HOME/$DICE_APPNAME" or "$HOME/.config/$DICE_APPNAME"
	CONFIG_HOME string `json:"config,omitempty"`
}

func (s *StandardPaths) UnmarshalJSON(data []byte) error {
	type Aux StandardPaths
	aux := &Aux{
		DICE_APPNAME: "dice",
		STATE_HOME:   "-",
		DATA_HOME:    "-",
		CONFIG_HOME:  "-",
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	*s = StandardPaths(*aux)
	return nil
}

// This stdpaths overrides the values of the taken stdpaths if the values are unset.
// Unset values are "-"
func (s *StandardPaths) Overrides(stdpaths *StandardPaths) *StandardPaths {
	if stdpaths.STATE_HOME == "-" {
		stdpaths.STATE_HOME = s.STATE_HOME
	}
	if stdpaths.DATA_HOME == "-" {
		stdpaths.DATA_HOME = s.DATA_HOME
	}
	if stdpaths.CONFIG_HOME == "-" {
		stdpaths.CONFIG_HOME = s.CONFIG_HOME
	}
	return stdpaths
}

func DefaultPaths() *StandardPaths {
	return BindStandardPaths(&StandardPaths{})
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
	home string

	app    string
	config string
	state  string
	data   string
}

func newStdpathsBuilder() *stdpathsBuilder {
	return &stdpathsBuilder{home: os.Getenv("HOME")}
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
	return &StandardPaths{
		DICE_APPNAME: b.app,
		CONFIG_HOME:  b.config,
		STATE_HOME:   b.state,
		DATA_HOME:    b.data,
	}
}

// Overrides empty standard paths. More of a purgue or clean job.
func BindStandardPaths(stdpaths *StandardPaths) *StandardPaths {
	b := newStdpathsBuilder()
	return b.setApp(stdpaths.DICE_APPNAME).
		setConfig(stdpaths.CONFIG_HOME).
		setData(stdpaths.DATA_HOME).
		setState(stdpaths.STATE_HOME).
		build()
}

type Configuration struct {
	Paths   *StandardPaths `json:"paths"`
	Project *Project       `json:"project"`

	fs    afero.Fs
	study *Study
}

func newConfiguration() *Configuration {
	return &Configuration{
		fs: afero.NewOsFs(),
	}
}

// Returns the location where we store databases, modules, and signatures
func (c *Configuration) Home() afero.Fs {
	return afero.NewBasePathFs(c.fs, c.Paths.DATA_HOME)
}

func (c *Configuration) SignaturesFs() afero.Fs {
	return afero.NewBasePathFs(c.Home(), "signatures")
}

func (c *Configuration) ModulesFs() afero.Fs {
	return afero.NewBasePathFs(c.Home(), "modules")
}

func (c *Configuration) ProjectFs() afero.Fs {
	if c.Project != nil {
		return afero.NewBasePathFs(c.fs, c.Project.Path)
	}
	return c.Home()
}

// Returns the location where the current run will output data to
// If no project is set, return the data location
func (c *Configuration) WorkspaceFs() afero.Fs {
	if c.study != nil {
		return afero.NewBasePathFs(c.fs, c.study.Path)
	}
	return c.ProjectFs()
}

func (c *Configuration) Workspace() string {
	if c.study != nil {
		return c.study.Path
	}
	return c.Paths.DATA_HOME
}

func (c *Configuration) Databases() string {
	return filepath.Join(c.Paths.DATA_HOME, "databases")
}

func (c *Configuration) SetFS(fs afero.Fs) *Configuration {
	c.fs = fs
	return c
}

var (
	ErrNotInDICEProject = errors.New("not inside a DICE project")
)

func findDICERoot(curr string) (string, error) {
	dicePath := filepath.Join(curr, ".dice")
	info, err := os.Stat(dicePath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// Is a DICE project and this is the location
	// of the configuration folder
	if info.IsDir() {
		return dicePath, nil
	}

	parent := filepath.Dir(curr)
	if parent == curr {
		return "", ErrNotInDICEProject
	}
	return findDICERoot(parent)
}

// Attempts to find the root folder of '.dice' to find out whether
// we are in a DICE project and we can load the project configuration.
func projConf(stdpaths *StandardPaths) (*Configuration, error) {
	var err error
	curr, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current directory")
	}

	root, err := findDICERoot(curr)
	if err != nil {
		return nil, err
	}

	return fpathConf(filepath.Join(root, "conf.json"), stdpaths)
}

// Reads a configuration from a file located in some path
// DICE configuration files should not reach more than 1MB right now, so
func fpathConf(p string, stdpaths *StandardPaths) (*Configuration, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open configuration file")
	}
	defer f.Close()

	buf := make([]byte, 0, 1024)
	r := bufio.NewReader(f)
	if _, err := io.ReadFull(r, buf[:cap(buf)]); err != nil {
		return nil, errors.Wrap(err, "failed to read configuration file")
	}

	var conf Configuration
	if err := json.Unmarshal(buf, &conf); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal configuration")
	}
	stdpaths.Overrides(conf.Paths)

	// override the configuration if it points to another file
	// NOTE: I am not sure about this right now
	// if conf.Paths.CONFIG_HOME != "" {
	// 	oc, err := fpathConf(c.Paths.CONFIG_HOME)
	// 	if err != nil {
	// 		return nil, errors.Wrapf(err, "failed to load overriding configuration %s", c.Paths.CONFIG_HOME)
	// 	}
	// 	c.Override(oc)
	// 	return c, nil
	// }
	return &conf, nil
}

func baseConf(paths *StandardPaths) *Configuration {
	return &Configuration{
		Paths: paths,
	}
}

func pwdConf() (*Configuration, error) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return &Configuration{
		Paths: &StandardPaths{"dice", wd, wd, wd},
	}, nil
}

// Load settings from a given path.
// If no path is given, look for a DICE config file locally;
// if this fails, look for a config file in the default config path;
// otherwise, return a default config.
// If the path is unset (i.e., "-"), then we use a "no-configuration"
// setting, where DICE will use the current directory as the main location
// for related files (e.g., modules, signatures, databases, etc.).
// If the path is another location, DICE will strictly look for a config file
// in that location and load it, returning an error if fails.
func LoadSettings(p string, stdpaths *StandardPaths) (*Configuration, error) {
	switch p {
	// Default mode. No config given, check for the default config file path
	// If no config is there, return default settings
	case "":
		// This config is the one of the project
		// TODO: maybe improve this so we can understand when we are in a project
		// before this happens, that way we can also load the project
		conf, lErr := projConf(stdpaths)
		if lErr != nil {
			if errors.Is(lErr, ErrNotInDICEProject) {
				goto CONFILE
			}
			return nil, errors.Wrap(lErr, "failed to load project settings")
		}
		return conf, nil

	CONFILE:
		conf, fErr := fpathConf(filepath.Join(stdpaths.CONFIG_HOME, "conf.json"), stdpaths)
		if fErr != nil {
			if errors.Is(fErr, os.ErrNotExist) {
				goto CONFBASE
			}
			return nil, errors.Wrap(fErr, "failed to load settings from config directory")
		}
		return conf, nil

	CONFBASE:
		return baseConf(stdpaths), nil
	// Without settings, everything in the current path
	case "-":
		conf, err := pwdConf()
		if err != nil {
			return nil, errors.Wrap(err, "failed load without settings")
		}
		return conf, nil
	// Another path, check if there is a config file there
	// If there isn't, return an error
	default:
		conf, err := fpathConf(p, stdpaths)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load settings from location")
		}
		return conf, nil
	}
}
