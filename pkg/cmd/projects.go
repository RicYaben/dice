package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/dice"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func projectCommands(conf dice.Configuration) []*cobra.Command {
	init := &cobra.Command{
		Use:     "init [project_name]",
		Short:   "Initialize a DICE project",
		GroupID: "init",
		Example: `
		$ pwd
		/home/d2/dice/projects/coin
		$ dice init . 
		DICE project "coin" initialized in "/home/d2/dice/projects/coin" for d2...
		$ dice projects list d*
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
			p, err := os.Getwd()
			if err != nil {
				return err
			}
			proj := &dice.Project{Path: p}

			var aliases = []string{".", "-"}
			switch {
			case len(args) > 0 && !slices.Contains(aliases, args[0]):
				proj.Name = args[0]
			default:
				proj.Name = filepath.Base(p)
			}

			ad := dice.MakeAdapters(nil, &conf)
			if err := ad.Projects().AddProject(proj); err != nil {
				return err
			}

			log.Info().Msgf(`DICE project "%s" initialized in %s`, proj.Name, proj.Path)
			return nil
		},
	}

	var pAd dice.ProjectAdapter
	projs := &cobra.Command{
		Use: "projects",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ad := dice.MakeAdapters(nil, &conf)
			pAd = ad.Projects()
		},
	}

	list := &cobra.Command{
		Use: "list [name]...",
		RunE: func(cmd *cobra.Command, args []string) error {
			var query []any
			for _, name := range args {
				name = globToSQLLike(name)
				query = append(query, name)
			}
			query = append([]any{"name LIKE ?"}, query...)

			var ms []dice.Project
			if err := pAd.Find(&ms, query); err != nil {
				return err
			}

			fmt.Printf("%-20s | %-50s\n", "Name", "Location")
			for _, m := range ms {
				fmt.Printf("%-20s | %-50s\n", m.Name, m.Path)
			}
			return nil
		},
	}

	projs.AddCommand(list)
	return []*cobra.Command{init, projs}
}
