package cmd

import (
	"github.com/dice"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const unset = "-"

type Flags struct {
	Paths dice.StandardPaths
	//Config dice.ConfigFlags
}

func Run() error {
	var conf dice.Configuration
	var f Flags

	com := &cobra.Command{
		Use:   "dice",
		Short: "The Engine",
		Args:  cobra.ArbitraryArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initial checks. Checks environment variables and standard paths

			// 1. bind the paths. Overrides defaults.
			dice.BindStandardPaths(&f.Paths)
			// 2. load and validate the configuration
			return dice.LoadConfiguration(f.Paths, &conf)
		},
	}

	// This set of flags propagates
	fl := com.PersistentFlags()

	stdpaths := &f.Paths
	pathFlags := pflag.NewFlagSet("Standard Paths", pflag.ExitOnError)
	pathFlags.StringVar(&stdpaths.DICE_APPNAME, "stdpath.app", unset, "App name")
	pathFlags.StringVar(&stdpaths.CONFIG_HOME, "stdpath.config", unset, "Configuration directory")
	pathFlags.StringVar(&stdpaths.STATE_HOME, "stdpath.state", unset, "State directory")
	pathFlags.StringVar(&stdpaths.DATA_HOME, "stdpath.data", unset, "Data directory")
	fl.AddFlagSet(pathFlags)

	// Config flags -- not used atm
	// cfg := &f.Config
	// cfgFlags := pflag.NewFlagSet("Configuration", pflag.ExitOnError)
	// cfgFlags.StringVarP(&cfg.Profile, "profile", "u", string(dice.NO_PROFILE), "Initialize a DICE profile. ")
	// fl.AddFlagSet(cfgFlags)

	com.AddCommand(engineCommands(conf)...)
	com.AddCommand(projectCommands(conf)...)
	com.AddCommand(signatureCommands(conf))

	return com.Execute()
}
