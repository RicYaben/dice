package dice

import (
	"github.com/spf13/cobra"
)

type InitFlags struct {
	NoCache bool
}

func Commands(conf Configuration) []*cobra.Command {
	return []*cobra.Command{
		// DICE group for profile management
		initCommand(conf), // init a project with .dice folder and register in cache db
		// DICE cache management
		addCommand(conf),   // add projects or signatures to their db
		listCommand(conf),  // list signatures or projects
		clearCommand(conf), // clear databases

		// TODO: this is better?
		// engineCommands(conf) // scan, classify
		// signatureCommand(conf) // add, remove, list
		// moduleCommand(conf)
		// projectCommand(conf) // switch
		// studyCommand(conf) // switch
		// cosmosCommand(conf) // query
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
