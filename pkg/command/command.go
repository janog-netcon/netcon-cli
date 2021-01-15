package command

import "github.com/spf13/cobra"

const (
	cliName        = "netcon"
	cliDescription = "in progress"
)

func NewNetconCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   cliName,
		Short: cliDescription,
	}

	rootCmd.AddCommand(
		NewSchedulerCommand(),
		NewScoreserverCommand(),
		NewVmmsCommand(),
	)

	return rootCmd
}
