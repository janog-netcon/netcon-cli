package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/janog-netcon/netcon-cli/pkg/scheduler"
	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/janog-netcon/netcon-cli/pkg/vmms"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

func NewSchedulerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "scheduler",
	}

	cmd.AddCommand(
		NewSchedulerStartCommand(),
		NewSchedulerDumpCommand(),
	)

	flags := cmd.PersistentFlags()
	flags.StringP("config", "", "./netcon.conf", "Scheduler Configuration")

	cmd.MarkFlagRequired("config")

	return cmd
}

func NewSchedulerStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "start",
		RunE: schedulerStartCommandFunc,
	}

	flags := cmd.Flags()
	flags.BoolP("oneshot", "", false, "cronでの繰り返し実行を行わずに1度のみ実行する")
	flags.StringP("log-file-path", "", "./scheduler.log", "Scheduler logfile")

	return cmd
}

func schedulerStartCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	configPath, err := flags.GetString("config")
	if err != nil {
		return err
	}
	oneshot, err := flags.GetBool("oneshot")
	if err != nil {
		return err
	}
	logFilePath, err := flags.GetString("log-file-path")
	if err != nil {
		return err
	}

	// logger
	/*
		lg, err := zap.NewDevelopment()
		if err != nil {
			return err
		}
	*/
	lg := newLogger(logFilePath)

	// read mapping file
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	cfg := types.SchedulerConfig{}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return err
	}

	// lg.Info(fmt.Sprintf("[INFO] config: %#v\n", cfg))

	// schedulerの起動
	scoreserverClient := scoreserver.NewClient(cfg.Setting.Scoreserver.Endpoint)
	vmmsClient := vmms.NewClient(cfg.Setting.Vmms.Endpoint, cfg.Setting.Vmms.Credential)

	// oneshotオプション
	if oneshot {
		scheduler.SchedulerReady(&cfg, scoreserverClient, vmmsClient, lg)
		if err != nil {
			return err
		}
		return nil
	}

	c := cron.New()
	// lock
	mutex := &sync.Mutex{}
	c.AddFunc(cfg.Setting.Cron, func() {
		// lg.Info("cron start!!")
		mutex.Lock()
		defer mutex.Unlock()
		if err := scheduler.SchedulerReady(&cfg, scoreserverClient, vmmsClient, lg); err != nil {
			fmt.Println(err)
		}
		// lg.Info("cron finish!!")
	})
	c.Start()

	for {
		time.Sleep(time.Second * 10)
	}

	return nil
}

func NewSchedulerDumpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "dump",
		RunE: schedulerDumpCommandFunc,
	}

	flags := cmd.Flags()
	flags.StringP("log-file-path", "", "./scheduler.log", "Scheduler logfile")

	return cmd
}

func schedulerDumpCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	configPath, err := flags.GetString("config")
	if err != nil {
		return err
	}
	logFilePath, err := flags.GetString("log-file-path")
	if err != nil {
		return err
	}

	lg := newLogger(logFilePath)

	// read mapping file
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	cfg := types.SchedulerConfig{}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return err
	}

	// schedulerの起動
	scoreserverClient := scoreserver.NewClient(cfg.Setting.Scoreserver.Endpoint)
	vmmsClient := vmms.NewClient(cfg.Setting.Vmms.Endpoint, cfg.Setting.Vmms.Credential)

	problems, zonePriorities, err := scheduler.Dump(&cfg, scoreserverClient, vmmsClient, lg)
	if err != nil {
		return err
	}

	j := struct {
		Problems       map[string]*scheduler.Problem `json:"problems"`
		ZonePriorities []*scheduler.ZonePriority     `json:"zone_priorities"`
	}{
		Problems:       problems,
		ZonePriorities: zonePriorities,
	}

	b, err := json.MarshalIndent(&j, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)

	return nil
}

// https://k1low.hatenablog.com/entry/2018/08/15/100000
func newLogger(logFilePath string) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	file, _ := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	logCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(file),
		zapcore.InfoLevel,
	)

	logger := zap.New(zapcore.NewTee(
		consoleCore,
		logCore,
	))

	return logger
}
