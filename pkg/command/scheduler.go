package command

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewSchedulerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "scheduler",
	}

	cmd.AddCommand(
		NewSchedulerStartCommand(),
	)

	flags := cmd.PersistentFlags()
	flags.StringP("scoreserver-endpoint", "", "http://127.0.0.1:8905", "Score Server API Endpoint")
	flags.StringP("vmms-endpoint", "", "http://127.0.0.1:8950", "vm-management-server Endpoint")
	flags.StringP("vmms-credential", "", "", "Token")

	return cmd
}

type schedulerConfig struct {
}

func NewSchedulerStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "start",
		RunE: schedulerStartCommandFunc,
	}

	flags := cmd.Flags()
	flags.StringP("config", "c", "", "config file path")

	cmd.MarkFlagRequired("config")

	return cmd
}

func schedulerStartCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	scoreserverEndpoint, err := flags.GetString("scoreserver-endpoint")
	if err != nil {
		return err
	}
	vmmsEndpoint, err := flags.GetString("vmms-endpoint")
	if err != nil {
		return err
	}
	vmmsCredential, err := flags.GetString("vmms-credential")
	if err != nil {
		return err
	}
	configPath, err := flags.GetString("config")
	if err != nil {
		return err
	}

	// read mapping file
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	cfg := []schedulerConfig{}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return err
	}

	fmt.Printf("[INFO] config: %#v\n", cfg)

	// schedulerの起動

	return nil
}
