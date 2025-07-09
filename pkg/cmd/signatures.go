package cmd

import (
	"fmt"
	"strings"

	"github.com/dice"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type operatorCmds struct {
	// DICE configuration (paths)
	conf dice.Configuration
	// Adapter factory
	adapters dice.Adapters
	// name of the command
	name string
	// Operations
	add    func(*operatorCmds, []string) error
	remove func(*operatorCmds, []string) error
	list   func(*operatorCmds, []string) error
	update func(*operatorCmds) error
}

func (o *operatorCmds) makeAddCommand() *cobra.Command {
	return &cobra.Command{
		GroupID: o.name,
		Use:     "add [name]...",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.add(o, args)
		},
	}
}

func (o *operatorCmds) makeRemoveCommand() *cobra.Command {
	return &cobra.Command{
		GroupID: o.name,
		Use:     "remove [name]...",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.remove(o, args)
		},
	}
}

func (o *operatorCmds) makeListCommand() *cobra.Command {
	return &cobra.Command{
		GroupID: o.name,
		Use:     "list [name]...",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.list(o, args)
		},
	}
}

func (o *operatorCmds) makeCommand() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: o.name,
		Use:     o.name,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			o.adapters = dice.MakeAdapters(nil, o.conf.Paths.STATE_HOME, ".")
		},
	}

	cmd.AddCommand(
		o.makeAddCommand(),
		o.makeListCommand(),
		o.makeRemoveCommand(),
	)
	return cmd
}

func signatureCommands(conf dice.Configuration) *cobra.Command {
	o := &operatorCmds{name: "signature", conf: conf}

	o.add = func(oc *operatorCmds, args []string) error {
		var query []any
		for _, name := range args {
			name = globToSQLLike(name)
			query = append(query, name)
		}
		query = append([]any{"name LIKE ?"}, query...)

		sAdapter := oc.adapters.Signatures()
		locs, err := sAdapter.Locate(&dice.Module{}, query)
		if err != nil {
			return err
		}

		var psigs []dice.Signature
		parser := dice.NewParser()
		for _, loc := range locs {
			if loc.ObjectID > -1 {
				continue
			}

			r, err := loc.Open()
			if err != nil {
				return errors.Wrapf(err, "failed to open signature %s", loc.Name)
			}

			psig, err := parser.Parse(loc.Name, r)
			if err != nil {
				return errors.Wrapf(err, "failed to parse signature %s", loc.Name)
			}
			psigs = append(psigs, psig)
		}

		// This is weird, because some signatures may be referencing others
		// so we need to add all of them at once in case there are
		// dependencies between them
		if err := sAdapter.AddSignatures(psigs...); err != nil {
			return err
		}
		return nil
	}

	o.remove = deleteHandler(&dice.Signature{})
	o.list = listHandler(&dice.Signature{})
	o.update = func(oc *operatorCmds) error {
		return o.adapters.Signatures().Update()
	}
	return o.makeCommand()
}

func moduleCommands(conf dice.Configuration) *cobra.Command {
	o := &operatorCmds{name: "module", conf: conf}

	o.add = func(oc *operatorCmds, args []string) error {
		var query []any
		for _, name := range args {
			name = globToSQLLike(name)
			query = append(query, name)
		}
		query = append([]any{"name LIKE ?"}, query...)

		sAdpater := oc.adapters.Signatures()
		locs, err := sAdpater.Locate(&dice.Module{}, query)
		if err != nil {
			return err
		}

		// A plugin pAdapter. It can start a plugin, extract
		// metadata, etc.
		pAdapter := o.adapters.Plugins()
		for _, loc := range locs {
			if loc.ObjectID > -1 {
				continue
			}

			// get the module metadata. i.e., name, type, requirements,
			// query, maintainer, version, help, description, etc.
			p, err := pAdapter.Load(loc.Location)
			if err != nil {
				return err
			}

			if err := sAdpater.AddModule(p.Metadata()); err != nil {
				return err
			}
		}
		return nil
	}

	o.remove = deleteHandler(&dice.Module{})
	o.list = listHandler(&dice.Module{})
	o.update = func(oc *operatorCmds) error {
		return o.adapters.Signatures().Update()
	}
	return o.makeCommand()
}

func listHandler[M dice.Signature | dice.Module](model *M) func(oc *operatorCmds, args []string) error {
	return func(oc *operatorCmds, args []string) error {
		var query []any
		for _, name := range args {
			name = globToSQLLike(name)
			query = append(query, name)
		}
		query = append([]any{"name LIKE ?"}, query...)

		var ms []M
		adapter := oc.adapters.Signatures()
		if err := adapter.Find(&ms, query); err != nil {
			return err
		}

		fmt.Printf("%-5s | %-20s\n", "ID", "Name")
		for _, m := range ms {
			fmt.Printf("%-5d | %-20s\n", m.ID, m.Name)
		}
		return nil
	}
}

func deleteHandler[M dice.Signature | dice.Module](model *M) func(oc *operatorCmds, args []string) error {
	return func(oc *operatorCmds, args []string) error {
		var query []any
		for _, name := range args {
			name = globToSQLLike(name)
			query = append(query, name)
		}
		query = append([]any{"name LIKE ?"}, query...)

		adapter := oc.adapters.Signatures()
		if err := adapter.Remove(query); err != nil {
			return err
		}
		return nil
	}
}

func globToSQLLike(glob string) string {
	// Escape SQL LIKE wildcards
	glob = strings.ReplaceAll(glob, "%", "\\%")
	glob = strings.ReplaceAll(glob, "_", "\\_")
	// Convert glob wildcards to SQL LIKE
	glob = strings.ReplaceAll(glob, "*", "%")
	glob = strings.ReplaceAll(glob, "?", "_")
	return glob
}
