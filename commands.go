package dice

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func Commands(conf Configuration) []*cobra.Command {
	return []*cobra.Command{
		// DICE group for profile management
		initCommand(conf), // init a project with .dice folder and register in cache db
		// DICE cache management
		addCommand(conf),   // add projects or signatures to their db
		listCommand(conf),  // list signatures or projects
		clearCommand(conf), // clear databases
		// DICE group to scan and classify
		scanCommand(conf),     // scan with signatures
		classifyCommand(conf), // classify dataset with signatures (no scan)
	}
}

func initCommand(conf Configuration) *cobra.Command {
	var f InitFlags

	cmd := &cobra.Command{
		Use:     "init [project_name] [--profile profile]",
		Short:   "Initialize a DICE project",
		GroupID: "init",
		// TODO: swap the hardcoded messages for mockup function calls?
		Example: `
		$ pwd
		/home/d2/dice/projects/coin
		$ dice init . --profile d2
		DICE project "coin" initialized in "/home/d2/dice/projects/coin" for d2...
		$ dice list -P coin --profile d2
		coin /home/d2/dice/projects/coin
		`,
		Long: `
		DICE creates a ".dice" folder in the given location. The command takes a single argument,
		a 'project_name', which is the path to the root directory where the project will live.
		If no 'project_name' is given, the project will be initialized in the current directory,
		and DICE will use it's name to register it.
		`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "."
			if len(args) > 0 {
				name = args[0]
			}
			return initProject(name, f, conf)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&f.NoCache, "no-cache", false, "Do not track this project")
	return cmd
}

type ModifierFlags struct {
	Include []string
	Exclude []string
}

func addCommand(conf Configuration) *cobra.Command {
	var (
		mods        ModifierFlags
		projs, sigs []string
	)

	cmd := &cobra.Command{
		Use:     "add (-S signatures | -P projects)... [-i include]... [-e exclude]...",
		Short:   "Add signatures or projects to the registry",
		GroupID: "cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("signatures") {
				return addSignatures(sigs, mods, conf)
			}
			return addProjects(projs, mods, conf)
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&sigs, "signatures", "S", []string{}, "Signatures to register")
	flags.StringArrayVarP(&projs, "projects", "P", []string{}, "Projects to register")
	flags.StringArrayVarP(&mods.Include, "include", "i", []string{}, "Only in the given paths")
	flags.StringArrayVarP(&mods.Exclude, "exclude", "e", []string{}, "Exclude paths")

	cmd.MarkFlagsMutuallyExclusive("signatures", "projects")
	cmd.MarkFlagsOneRequired("signatures", "projects")

	return cmd
}

func clearCommand(conf Configuration) *cobra.Command {
	var (
		mods        ModifierFlags
		projs, sigs []string
	)

	cmd := &cobra.Command{
		Use:     "clear (-S signatures | -P projects)... [-i include]... [-e exclude]...",
		Short:   "Clear all registered signatures or projects",
		GroupID: "cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("signatures") {
				return clearSignatures(sigs, mods, conf)
			}
			return clearProjects(projs, mods, conf)
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&sigs, "signatures", "S", []string{"*"}, "Signatures to clear. Default clears all")
	flags.StringArrayVarP(&projs, "projects", "P", []string{"*"}, "Projects to clear. Default clears all")
	flags.StringArrayVarP(&mods.Include, "include", "i", []string{}, "Only in the given paths")
	flags.StringArrayVarP(&mods.Exclude, "exclude", "e", []string{}, "Exclude paths")

	cmd.MarkFlagsMutuallyExclusive("signatures", "projects")
	cmd.MarkFlagsOneRequired("signatures", "projects")

	return cmd
}

func listCommand(conf Configuration) *cobra.Command {
	var (
		mods        ModifierFlags
		projs, sigs []string
	)

	cmd := &cobra.Command{
		Use:     "list (-S signatures | -P projects)... [-i include]... [-e exclude]...",
		Short:   "List all registered signatures or projects",
		GroupID: "cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("signatures") {
				return listSignatures(sigs, mods, conf)
			}
			return listProjects(projs, mods, conf)
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&sigs, "signatures", "S", []string{"*"}, "Signatures to list. Default clears all")
	flags.StringArrayVarP(&projs, "projects", "P", []string{"*"}, "Projects to list. Default clears all")
	flags.StringArrayVarP(&mods.Include, "include", "i", []string{}, "Only in the given paths")
	flags.StringArrayVarP(&mods.Exclude, "exclude", "e", []string{}, "Exclude paths")

	cmd.MarkFlagsMutuallyExclusive("signature", "project")
	cmd.MarkFlagsOneRequired("signature", "project")

	return cmd
}

type EngineActions struct {
	Scan     bool
	Identify bool
	Classify bool
}

type EngineSetupFlags struct {
	Signatures []string
	Project    string
	Actions    EngineActions
}

type EngineRunFlags struct {
	Scan    string
	Sources []string
}

func lazyEngineSetup(engine *engine, sFlags EngineSetupFlags, conf Configuration) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Create the graph of scanners, identifiers and classifiers from
		// the loaded signatures
		engine.conf = conf
		if err := LoadEngine(engine, conf, sFlags); err != nil {
			return errors.Wrap(err, "failed to build engine")
		}
		return nil
	}
}

func scanCommand(conf Configuration) *cobra.Command {
	var (
		eFlags = EngineSetupFlags{
			Actions: EngineActions{
				true, true, true,
			},
		}
		rFlags EngineRunFlags
	)

	engine := new(engine)
	cmd := &cobra.Command{
		Use:     "scan [target]... [-S signatures]... [-P project] ([-s sources]... [--scan scan])",
		Short:   "Orchestrate a new scan",
		GroupID: "run",
		PreRunE: lazyEngineSetup(engine, eFlags, conf),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Scanning from sources means that we use previously collected results
			// as the seed for this scan. For example, if we use results from zmap,
			// we can scan now with a bunch of signatures to produce a similar result.
			// E.g 1: follow-up scans on known targets.
			// E.g 2: resume scans

			// load these sources. The engine should have access to all the needed
			// repos based on the config.
			srcs, err := engine.findSources(rFlags.Scan, rFlags.Sources, []string{"*.json", "*.csv", "*.txt"})
			if err != nil {
				return errors.Wrap(err, "failed to find sources")
			}

			// a list of targets.
			if len(args) > 0 {
				src, err := makeTargetArgsSource(args)
				if err != nil {
					return err
				}
				srcs = append(srcs, src)
			}

			// For now, no stdin
			// srcs.append(makeStdinSource())
			return engine.Run(srcs) // resume
		},
	}

	flags := cmd.Flags()
	// setup flags
	flags.StringArrayVarP(&eFlags.Signatures, "signature", "S", []string{"*"}, "Signatures to load. By default loads all signatures")
	flags.StringVarP(&eFlags.Project, "project", "P", "-", "Project to use. Defaults to current project")
	// run flags
	flags.StringArrayVarP(&rFlags.Sources, "sources", "s", []string{}, "Sources to use as input. If this flag is set, other inputs are ignored")
	flags.StringVar(&rFlags.Scan, "scan", "-", "Scan to use. Required if sources flag included")
	//flags.BoolVar(&rFlags.Resume, "resumev", false, "Resume a previous scan. This flag is ignore during new scans")
	cmd.MarkFlagsRequiredTogether("sources", "scan")

	return cmd
}

func classifyCommand(conf Configuration) *cobra.Command {
	var (
		eFlags = EngineSetupFlags{
			Actions: EngineActions{
				false, true, true,
			},
		}
		rFlags EngineRunFlags
	)

	engine := new(engine)
	cmd := &cobra.Command{
		Use:     "classify scan [-P project] [-S signatures]... [--sources source]...",
		Aliases: []string{"class", "cls"},
		Short:   "Classify a scan",
		GroupID: "run",
		Args:    cobra.ExactArgs(1),
		PreRunE: lazyEngineSetup(engine, eFlags, conf),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Classify a scan from a list of sources. Outputs a cosmos
			srcs, err := engine.findSources(args[0], rFlags.Sources, []string{"*.json", "*.csv", "*.txt"})
			if err != nil {
				return errors.Wrap(err, "failed to find sources")
			}
			return engine.Run(srcs)
		},
	}

	flags := cmd.Flags()
	// setup flags
	flags.StringArrayVarP(&eFlags.Signatures, "signature", "S", []string{"*"}, "Signatures to load. By default loads all signatures")
	flags.StringVarP(&eFlags.Project, "project", "P", "-", "Project to use. Defaults to current project")
	// run flags
	flags.StringArrayVarP(&rFlags.Sources, "sources", "s", []string{}, "Sources to use as input. If this flag is set, other inputs are ignored")

	return cmd
}
