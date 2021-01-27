package command

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/janog-netcon/netcon-cli/pkg/scheduler"
	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/janog-netcon/netcon-cli/pkg/vmms"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
	flags.StringP("config", "", "./netcon.conf", "Scheduler Configuration")

	return cmd
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

	configPath, err := flags.GetString("config")
	if err != nil {
		return err
	}

	// logger
	lg, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	// read mapping file
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	cfg := types.SchedulerConfig{}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return err
	}

	lg.Info(fmt.Sprintf("[INFO] config: %#v\n", cfg))

	// schedulerの起動
	scoreserverClient := scoreserver.NewClient(cfg.Setting.Scoreserver.Endpoint)
	vmmsClient := vmms.NewClient(cfg.Setting.Vmms.Endpoint, cfg.Setting.Vmms.Credential)

	c := cron.New()
	c.AddFunc(cfg.Setting.Cron, func() {
		lg.Info("cron start!!")
		if err := scheduler.SchedulerReady(&cfg, scoreserverClient, vmmsClient, lg); err != nil {
			fmt.Println(err)
		}
		lg.Info("cron finish!!")
	})
	c.Start()

	for {
		time.Sleep(time.Second * 10)
	}

	return nil
}
