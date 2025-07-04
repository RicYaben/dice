package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dice"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Project -> Experiment -> Source -> data
// A group of experiments is a `Deck`, or a `Measurement`
// A group of non-overlapping `Decks` `Measurements` is a Study
func Command(conf dice.Configuration) *cobra.Command {
	var pm *projectManager

	// base flags
	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"p"},
		Short:   "Manage projects, experiments, and sources",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			store, err := Store(conf.Data())
			if err != nil {
				return errors.Wrap(err, "failed while loading project store")
			}
			pm.store = store

			fpath := filepath.Join(conf.Data(), "project.db")
			repo, err := ProjectRepo(fpath)
			if err != nil {
				return errors.Wrap(err, "failed while loading project repository")
			}
			pm.repo = repo
			return nil
		},
	}

	cmd.AddCommand(
		newCommand(pm),
		deleteCommand(pm),
		listCommand(pm),
	)

	return cmd
}

func newCommand(pm *projectManager) *cobra.Command {
	return &cobra.Command{
		Use:     "new [project] [experiment] [source]",
		Short:   "Create new projects, experiments and sources",
		Args:    rangeArgsValidateFuncE(1, 3),
		Example: "new s123* e123",
		RunE: withProjectArgs(func(p, e, s string) error {
			return pm.New(p, e, s)
		}),
	}
}

func deleteCommand(pm *projectManager) *cobra.Command {
	return &cobra.Command{
		Use:     "delete [project] [experiment] [source]",
		Short:   "Delete matching sources",
		Args:    rangeArgsValidateFuncE(1, 3),
		Example: example("list"),
		RunE: withProjectArgs(func(p, e, s string) error {
			return pm.Delete(p, e, s)
		}),
	}
}

func listCommand(pm *projectManager) *cobra.Command {
	return &cobra.Command{
		Use:     "list [project] [experiment] [source]",
		Short:   "List matching projects, experiments and sources",
		Args:    rangeArgsValidateFuncE(1, 3),
		Example: example("list"),
		RunE: withProjectArgs(func(p, e, s string) error {
			return pm.List(p, e, s)
		}),
	}
}

func example(com string) string {
	var ret []string
	args := []string{"s123*", "e1*", "scan-*"}
	for i := 0; i < len(args); i++ {
		s := fmt.Sprintf("%s %s", com, strings.Join(args[:i], " "))
		ret = append(ret, s)
	}
	return strings.Join(ret, "\n")
}

func rangeArgsValidateFuncE(min, max int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.RangeArgs(min, max)(cmd, args); err != nil {
			return err
		}

		for i, arg := range args {
			if strings.TrimSpace(arg) == "" {
				return errors.Errorf("argument %d must not be empty", i+1)
			}
		}
		return nil
	}
}

func toProjectArgs(args []string) ([]string, error) {
	// defaults
	ret := []string{"-", "-", "-"}
	// replace defaults
	for i, v := range args {
		ret[i] = v
	}

	peek := func(l []string, curr int) (string, bool) {
		if !(len(l) > curr+1) {
			return "", false
		}
		return l[curr+1], true
	}

	for i, v := range ret {
		if v == "-" {
			next, ok := peek(ret, i)
			if next != "-" && ok {
				return nil, errors.New("preceeding arguments must be set")
			}
		}
	}

	return ret, nil
}

func withProjectArgs(
	handler func(p, e, s string) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		pArgs, err := toProjectArgs(args)
		if err != nil {
			return errors.Wrap(err, "failed to parse positional arguments")
		}
		return handler(pArgs[0], pArgs[1], pArgs[2])
	}
}
