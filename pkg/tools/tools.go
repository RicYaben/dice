package tools

import (
	"github.com/dice"
	"github.com/dice/pkg/tools/project"
	"github.com/spf13/cobra"
)

func Command(conf dice.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use: "tool",
	}
	cmd.AddCommand(project.Command(conf))
	return cmd
}
