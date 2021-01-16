package command

import (
	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/spf13/cobra"
)

func NewSchedulerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "scheduler",
	}

	cmd.AddCommand(
		NewSchedulerStartCommand(),
	)

	flags := cmd.PersistentFlags()
	flags.StringP("endpoint", "", "http://127.0.0.1:8905", "Score Server API Endpoint")
	flags.StringP("config", "", "./netcon.conf", "Scheduler Configuration")

	return cmd
}

func NewSchedulerStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "start",
		RunE: schedulerStartCommandFunc,
	}

	return cmd
}

func schedulerStartCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	endpoint, err := flags.GetString("endpoint")
	if err != nil {
		return err
	}
	//configをパースする

	//情報をsiにまとめる(siを問題ごとに初期化する)
	var si types.ScheduleInfo
	var pis []types.ProblemInstance
	//score serverのデータを取得し集計する
	si, err := aggregateInstance(endpoint, si)
	//configよりInstanceの作成リストを削除リストを作る
	//Instance削除リストから対象Instanceを削除する
	//Instance作成リストから対象ProblemのInstanceを作成する
	return nil
}

func aggregateInstance(endpoint string, si ScheduleInfo) {
	//Score Serverからデータを取得
	cli := scoreserver.NewClient(endpoint)
	pes, err := cli.ListProblemEnvironment()
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(pes)
	if err != nil {
		return nil, err
	}

	for i, p := range b {
		min := strings.Split(p.MachineImageName, "-")
		pn := min[-2:-1]
		si.ProblemInstances[pn].MachineImageName = p.MachineImageName
		switch p.Status {
		case "NOT_READY":
			si.ProblemInstances[pn].NotReady = si.ProblemInstances[pn].NotReady + 1
		case "READY":
			si.ProblemInstances[pn].Ready = si.ProblemInstances[pn].Ready + 1
		case "UNDER_CHALLENGE":
			si.ProblemInstances[pn].UnderChallenge = si.ProblemInstances[pn].UnderChallenge + 1
		case "UNDER_SCORING":
			si.ProblemInstances[pn].UnderScoring = si.ProblemInstances[pn].UnderScoring + 1
		case "ABANDONED":
			si.ProblemInstances[pn].Abandoned = si.ProblemInstances[pn].Abandoned + 1
		case "":
			si.ProblemInstances[pn].Ready = si.ProblemInstances[pn].Ready + 1
		}
	}

	return si, nil
}
