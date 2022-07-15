package scheduler

import (
	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/janog-netcon/netcon-cli/pkg/vmms"
	"go.uber.org/zap"
)

func Dump(cfg *types.SchedulerConfig, ssClient *scoreserver.Client, vmmsClient *vmms.Client, lg *zap.Logger) (map[string]*Problem, []*ZonePriority, error) {
	// configファイルから設定を読み込む
	problems, zonePriorities := InitScheduler(cfg, lg)

	// ScoreServer からデータを取得し、現在のインスタンス状況を集計する
	problems, zonePriorities, _, err := AggregateInstance(problems, zonePriorities, ssClient, lg)
	if err != nil {
		lg.Error("Scheduler Aggregate: " + err.Error())
		return nil, nil, err
	}

	return problems, zonePriorities, nil
}
