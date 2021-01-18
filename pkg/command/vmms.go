package command

import (
	"encoding/json"
	"fmt"

	"github.com/janog-netcon/netcon-cli/pkg/vmms"
	"github.com/spf13/cobra"
)

func NewVmmsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "vmms",
	}

	cmd.AddCommand(
		NewVmmsInstanceCommand(),
	)

	flags := cmd.PersistentFlags()
	flags.StringP("endpoint", "", "http://127.0.0.1:8950", "vm-management-server Endpoint")
	flags.StringP("credential", "", "", "Token")

	return cmd
}

func NewVmmsInstanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "instance",
	}

	cmd.AddCommand(
		NewVmmsInstanceCreateCommand(),
		NewVmmsInstanceDeleteCommand(),
	)

	return cmd
}

func NewVmmsInstanceCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create",
		RunE: vmmsInstanceCreateCommandFunc,
	}

	flags := cmd.Flags()
	flags.StringP("problem-id", "", "", "Problem ID")
	flags.StringP("machine-image-name", "", "", "Machine Image Name")
	flags.StringP("project", "", "", "Project")
	flags.StringP("zone", "", "", "Zone")

	cmd.MarkFlagRequired("problem-id")
	cmd.MarkFlagRequired("machine-image-name")
	cmd.MarkFlagRequired("project")
	cmd.MarkFlagRequired("zone")

	return cmd
}

func vmmsInstanceCreateCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	endpoint, err := flags.GetString("endpoint")
	if err != nil {
		return err
	}
	credential, err := flags.GetString("credential")
	if err != nil {
		return err
	}
	problemID, err := flags.GetString("problem-id")
	if err != nil {
		return err
	}
	machineImageName, err := flags.GetString("machine-image-name")
	if err != nil {
		return err
	}
	project, err := flags.GetString("project")
	if err != nil {
		return err
	}
	zone, err := flags.GetString("zone")
	if err != nil {
		return err
	}

	cli := vmms.NewClient(endpoint, credential)
	pes, err := cli.CreateInstance(problemID, machineImageName, project, zone)
	if err != nil {
		return err
	}

	b, err := json.Marshal(pes)
	fmt.Println(string(b))

	return nil
}

func NewVmmsInstanceDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "delete",
		RunE: vmmsInstanceDeleteCommandFunc,
	}

	flags := cmd.Flags()
	flags.StringP("instance-name", "", "", "instance name")
	flags.StringP("project", "", "", "Project")
	flags.StringP("zone", "", "", "Zone")

	cmd.MarkFlagRequired("instance-name")
	cmd.MarkFlagRequired("project")
	cmd.MarkFlagRequired("zone")

	return cmd
}

func vmmsInstanceDeleteCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	endpoint, err := flags.GetString("endpoint")
	if err != nil {
		return err
	}
	credential, err := flags.GetString("credential")
	if err != nil {
		return err
	}
	instanceName, err := flags.GetString("instance-name")
	if err != nil {
		return err
	}
	project, err := flags.GetString("project")
	if err != nil {
		return err
	}
	zone, err := flags.GetString("zone")
	if err != nil {
		return err
	}

	cli := vmms.NewClient(endpoint, credential)
	if err := cli.DeleteInstance(instanceName, project, zone); err != nil {
		return err
	}

	fmt.Printf("[INFO] Deleted successfully: %s\n", instanceName)

	return nil
}
