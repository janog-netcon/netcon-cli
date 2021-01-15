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
		NewScoreserverVMCommand(),
	)

	flags := cmd.PersistentFlags()
	flags.StringP("endpoint", "", "http://127.0.0.1:8905", "Score Server API Endpoint")

	return cmd
}

func NewScoreserverVMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "vm",
	}

	cmd.AddCommand(
		NewScoreserverVMListCommand(),
		NewScoreserverVMGetCommand(),
	)

	return cmd
}

func NewScoreserverVMListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "list",
		RunE: scoreserverVMListCommandFunc,
	}

	return cmd
}

func scoreserverVMListCommandFunc(cmd *cobra.Command, args []string) error {
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

func NewScoreserverVMGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "get",
		RunE: scoreserverVMGetCommandFunc,
	}

	flags := cmd.Flags()
	flags.StringP("name", "", "", "vm name")

	return cmd
}

func scoreserverVMGetCommandFunc(cmd *cobra.Command, args []string) error {
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
