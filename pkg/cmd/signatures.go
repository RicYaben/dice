package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dice"
	"github.com/spf13/cobra"
)

type operatorCmds struct {
	// DICE configuration (paths)
	conf *dice.Configuration
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
			o.adapters = dice.MakeAdapters(nil, o.conf)
		},
	}

	cmd.AddCommand(
		o.makeAddCommand(),
		o.makeListCommand(),
		o.makeRemoveCommand(),
	)
	return cmd
}

func signatureCommand(conf *dice.Configuration) *cobra.Command {
	o := &operatorCmds{name: "signature", conf: conf}

	// Takes a number of glob-like names and adds them to the db
	o.add = func(oc *operatorCmds, args []string) error {
		sAdapter := oc.adapters.Signatures()

		// Query for all the signatures with a name like the globs
		var globs []string
		for _, name := range args {
			name = globToSQLLike(name)
			globs = append(globs, name)
		}

		_, err := sAdapter.AddMissingSignatures(globs...)
		return err
	}

	o.remove = deleteHandler[dice.Signature]()
	o.list = listHandler[dice.Signature]()
	o.update = func(oc *operatorCmds) error {
		return o.adapters.Signatures().Update()
	}
	return o.makeCommand()
}

func moduleCommand(conf *dice.Configuration) *cobra.Command {
	o := &operatorCmds{name: "module", conf: conf}

	// register one or more modules into the database (by globs)
	o.add = func(oc *operatorCmds, args []string) error {
		sAdapter := oc.adapters.Signatures()

		var globs []string
		for _, name := range args {
			name = globToSQLLike(name)
			globs = append(globs, name)
		}

		_, err := sAdapter.AddMissingModules(globs...)
		return err
	}

	o.remove = deleteHandler[dice.Module]()
	o.list = listHandler[dice.Module]()
	o.update = func(oc *operatorCmds) error {
		return o.adapters.Signatures().Update()
	}
	return o.makeCommand()
}

type NamedImpl interface {
	dice.Signature | dice.Module
}

func listHandler[M NamedImpl]() func(oc *operatorCmds, args []string) error {
	return func(oc *operatorCmds, args []string) error {
		var query []any
		for _, name := range args {
			name = globToSQLLike(name)
			query = append(query, name)
		}
		query = append([]any{"name LIKE ?"}, query...)

		var ms []*M
		adapter := oc.adapters.Signatures()
		if err := adapter.Find(&ms, query); err != nil {
			return err
		}

		fmt.Printf("%-5s | %-20s\n", "ID", "Name")
		for _, m := range ms {
			// I really do not want to add these methods to the structs
			v := reflect.ValueOf(m)
			id := v.FieldByName("ID").Uint()
			name := v.FieldByName("Name").String()
			fmt.Printf("%-5d | %-20s\n", id, name)
		}
		return nil
	}
}

func deleteHandler[M NamedImpl]() func(oc *operatorCmds, args []string) error {
	return func(oc *operatorCmds, args []string) error {
		var query []any
		for _, name := range args {
			name = globToSQLLike(name)
			query = append(query, name)
		}
		query = append([]any{"name LIKE ?"}, query...)

		adapter := oc.adapters.Signatures()

		var m M
		if err := adapter.Remove(&m, query); err != nil {
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
