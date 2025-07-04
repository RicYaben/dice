package cmd

import (
	"github.com/dice"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Flags struct {
	Paths dice.StandardPaths
	//Config dice.ConfigFlags
	Output dice.LogsFlags
}

// TODO: interactive cli?
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
			return dice.LoadConfiguration(f.Paths, f.Output, &conf)
		},
	}

	// This set of flags propagates
	fl := com.PersistentFlags()

	stdpaths := &f.Paths
	pathFlags := pflag.NewFlagSet("Standard Paths", pflag.ExitOnError)
	pathFlags.StringVar(&stdpaths.DICE_APPNAME, "stdpath.app", dice.UnsetFlag, "App name")
	pathFlags.StringVar(&stdpaths.CONFIG_HOME, "stdpath.config", dice.UnsetFlag, "Configuration directory")
	pathFlags.StringVar(&stdpaths.STATE_HOME, "stdpath.state", dice.UnsetFlag, "State directory")
	pathFlags.StringVar(&stdpaths.DATA_HOME, "stdpath.data", dice.UnsetFlag, "Data directory")
	fl.AddFlagSet(pathFlags)

	out := &f.Output
	outFlags := pflag.NewFlagSet("Output", pflag.ExitOnError)
	outFlags.BoolVar(&out.Debug, "debug", false, "Debug mode. Very verbose (not recommended)")
	outFlags.BoolVar(&out.Quiet, "quiet", false, "Run DICE in quiet mode (recommended). Check the logs to see output messages.")

	// Config flags -- not used atm
	// cfg := &f.Config
	// cfgFlags := pflag.NewFlagSet("Configuration", pflag.ExitOnError)
	// cfgFlags.StringVarP(&cfg.Profile, "profile", "u", string(dice.NO_PROFILE), "Initialize a DICE profile. ")
	// fl.AddFlagSet(cfgFlags)

	com.AddCommand(dice.Commands(conf)...)

	return com.Execute()
}
