package command

import (
	"encoding/json"
	"fmt"

	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/spf13/cobra"
)

func NewScoreserverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "scoreserver",
	}

	cmd.AddCommand(
		NewScoreserverInstanceCommand(),
	)

	flags := cmd.PersistentFlags()
	flags.StringP("endpoint", "", "http://127.0.0.1:8905", "Score Server API Endpoint")

	return cmd
}

func NewScoreserverInstanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "instance",
	}

	cmd.AddCommand(
		NewScoreserverInstanceListCommand(),
		NewScoreserverInstanceGetCommand(),
	)

	return cmd
}

func NewScoreserverInstanceListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "list",
		RunE: scoreserverInstanceListCommandFunc,
	}

	return cmd
}

func scoreserverInstanceListCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	endpoint, err := flags.GetString("endpoint")
	if err != nil {
		return err
	}

	cli := scoreserver.NewClient(endpoint)
	pes, err := cli.ListProblemEnvironment()
	if err != nil {
		return err
	}

	b, err := json.Marshal(pes)
	fmt.Println(string(b))

	return nil
}

func NewScoreserverInstanceGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "get",
		RunE: scoreserverInstanceGetCommandFunc,
	}

	flags := cmd.Flags()
	flags.StringP("name", "", "", "vm name")

	cmd.MarkFlagRequired("name")

	return cmd
}

func scoreserverInstanceGetCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	endpoint, err := flags.GetString("endpoint")
	if err != nil {
		return err
	}
	name, err := flags.GetString("name")
	if err != nil {
		return err
	}

	cli := scoreserver.NewClient(endpoint)
	pes, err := cli.GetProblemEnvironment(name)
	if err != nil {
		return err
	}

	b, err := json.Marshal(pes)
	fmt.Println(string(b))

	return nil
}
