package cmd

import (
	"reflect"

	"github.com/dice"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Workspace
type WorkspaceFlags struct {
	Project string
	Study   string
}

// What to do
type EngineFlags struct {
	Scan     bool `label:"scanner"`
	Identify bool `label:"identifier"`
	Classify bool `label:"classifier"`
}

func (e EngineFlags) toList() []string {
	val := reflect.ValueOf(e)
	typ := val.Type()

	tags := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).Bool() {
			tags = append(tags, typ.Field(i).Tag.Get("label"))
		}
	}
	return tags
}

// What to load
type ComposerFlags struct {
	Signatures []string
	Modules    []string
}

// Inputs
type InputFlags struct {
	Sources []string
}

type lazyLoader struct {
	conf     dice.Configuration
	adapters dice.Adapters
	composer dice.Composer
	emitter  dice.Emitter
}

func newLazyEngineLoader(conf dice.Configuration) *lazyLoader {
	return &lazyLoader{conf: conf}
}

func (l *lazyLoader) preRunE(w WorkspaceFlags, e EngineFlags, c ComposerFlags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// TODO: not sure how to implement this
		wk := dice.MakeWorkspace(w.Project, w.Study)

		em := dice.NewEmitter()
		ad := dice.MakeAdapters(em.Emit, l.conf.Paths.STATE_HOME, wk)
		comp := dice.NewComposer(ad.Composer())

		if err := comp.Stage(dice.STAGE_MODULE, c.Modules...); err != nil {
			return errors.Wrap(err, "failed to stage modules")
		}

		if err := comp.Stage(dice.STAGE_SIGNATURE, c.Signatures...); err != nil {
			return errors.Wrap(err, "failed to stage signatures")
		}

		components, err := comp.Compose(e.toList())
		if err != nil {
			return errors.Wrap(err, "failed to create components")
		}
		em.Subscribe(components...)

		l.adapters = ad
		l.composer = comp
		l.emitter = em
		return nil
	}
}

func (l *lazyLoader) runE(i InputFlags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ad := l.adapters.Engine()
		engine := dice.NewEngine(ad, l.emitter)

		s := l.adapters.Cosmos()
		// find file sources
		srcs, err := s.FindSources(i.Sources, []string{"json", "csv", "txt"})
		if err != nil {
			return errors.Wrap(err, "failed to find sources")
		}

		// a list of targets.
		if len(args) > 0 {
			src, err := dice.MakeTargetArgsSource(args)
			if err != nil {
				return err
			}
			srcs = append(srcs, src)
		}

		// For now, no stdin
		// srcs.append(dice.NewSource("stdin", io.Stdin))
		return engine.Run(srcs)
	}
}

func (l *lazyLoader) makeCommand(e EngineFlags) *cobra.Command {
	var (
		w WorkspaceFlags
		c ComposerFlags
		i InputFlags
	)

	cmd := &cobra.Command{
		GroupID: "run",
		PreRunE: l.preRunE(w, e, c),
		RunE:    l.runE(i),
	}

	flags := cmd.Flags()

	// workspace flags
	flags.StringVar(&w.Project, "project", "-", "Project to use. Defaults to current project")
	flags.StringVar(&w.Project, "study", "-", "Study to use. Defaults to current study")

	// composer flags
	flags.StringArrayVarP(&c.Signatures, "signatures", "S", []string{"*"}, "Signatures to load. By default loads all signatures")
	flags.StringArrayVarP(&c.Modules, "modules", "M", []string{}, "Modules to load")

	// input flags
	flags.StringArrayVarP(&i.Sources, "sources", "s", []string{}, "Sources to use as input.")

	return cmd
}

func engineCommands(conf dice.Configuration) []*cobra.Command {
	lazy := newLazyEngineLoader(conf)
	return []*cobra.Command{
		scanCommand(lazy),
		classifyCommand(lazy),
	}
}

func scanCommand(l *lazyLoader) *cobra.Command {
	cmd := l.makeCommand(EngineFlags{Scan: true, Classify: true, Identify: true})
	cmd.Use = "scan [targets] [-S signatures] [-M modules] [-s sources] [--project] [--study]"
	cmd.Short = "Orchestrate a new scan"
	return cmd
}

func classifyCommand(l *lazyLoader) *cobra.Command {
	cmd := l.makeCommand(EngineFlags{Scan: false, Classify: true, Identify: true})
	cmd.Use = "classify [targets] [-S signatures] [-M modules] [-s sources] [--project] [--study]"
	cmd.Short = "Classify sources"
	return cmd
}
