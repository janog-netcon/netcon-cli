package scheduler

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/janog-netcon/netcon-cli/pkg/vmms"
)

type Problem struct {
	MachineImageName string
	ProblemID        string
	NotReady         int
	Ready            int
	UnderChallenge   int
	UnderScoring     int
	Abandoned        int
	KeepPool         int
	KeptInstances    []Instance
	CurrentInstance  int
}

type Instance struct {
	InstanceName string
	ProjectName  string
	ZoneName     string
	InnerStatus  *string
	CreatedAt    time.Time
}

type ZonePriority struct {
	ProjectName     string
	ZoneName        string
	Priority        int
	MaxInstance     int
	CurrentInstance int
}

type CreationTargetInstance struct {
	ProblemName      string
	ProblemID        string
	MachineImageName string
}

type DeletionTargetInstance struct {
	ProblemName  string
	InstanceName string
	ProjectName  string
	ZoneName     string
}

func SchedulerReady(cfg *types.SchedulerConfig, ssClient *scoreserver.Client, vmmsClient *vmms.Client, lg *zap.Logger) error {
	lg.Info("Scheduler: SchedulerReady")

	// configファイルから設定を読み込む
	problems, zonePriorities := InitScheduler(cfg, lg)

	// ScoreServer からデータを取得し、現在のインスタンス状況を集計する
	problems, zonePriorities, abandonedInstances, err := AggregateInstance(problems, zonePriorities, ssClient, lg)

	if err != nil {
		lg.Error("Scheduler Aggregate: " + err.Error())
		return err
	}

	// ロギング
	PISLogging(problems, lg)
	ZPSLogging(zonePriorities, lg)

	// 作成対象のインスタンスと削除対象のインスタンスを列挙する
	creationTargetInstances, deletionTargetInstances := SchedulingList(problems, lg)

	// abandoned なインスタンスを削除する
	err = DeleteInstances(abandonedInstances, vmmsClient, cfg.Setting.Scheduler.InstanceDeletionInterval, lg)
	if err != nil {
		lg.Error("Scheduler DeleteScheduler: AbandonedInstance. " + err.Error())
		return err
	}

	// 削除対象のインスタンスを削除する
	err = DeleteInstances(deletionTargetInstances, vmmsClient, cfg.Setting.Scheduler.InstanceDeletionInterval, lg)
	if err != nil {
		lg.Error("Scheduler DeleteScheduler: " + err.Error())
		return err
	}

	// 作成対象のインスタンスを作成する
	err = CreateInstances(creationTargetInstances, zonePriorities, vmmsClient, cfg.Setting.Scheduler.InstanceCreationInterval, lg)
	if err != nil {
		lg.Error("Scheduler CreateScheduler: " + err.Error())
		return err
	}
	return nil
}

// InitScheduler VMとGCP Projectに関する情報を設定ファイルから取得し、Schedulerで扱える形式に変換する
// config.Problems -> Problems
// config.Projects -> ZonePriorities
func InitScheduler(cfg *types.SchedulerConfig, lg *zap.Logger) (map[string]*Problem, []*ZonePriority) {

	lg.Info("Scheduler: InitSchedulerInfo")

	// config から問題情報を取得する
	problems := map[string]*Problem{}

	for _, p := range cfg.Setting.Problems {
		problems[p.MachineImageName] = &Problem{
			MachineImageName: p.MachineImageName,
			ProblemID:        p.ProblemID,
			NotReady:         0,
			Ready:            0,
			UnderChallenge:   0,
			UnderScoring:     0,
			Abandoned:        0,
			KeepPool:         p.KeepPool,
			KeptInstances:    []Instance{},
			CurrentInstance:  0,
		}
	}

	// config からzone情報を取得する
	zonePriorities := []*ZonePriority{}

	for _, p := range cfg.Setting.Projects {
		for _, z := range p.Zones {
			zonePriorities = append(zonePriorities, &ZonePriority{
				ProjectName:     p.Name,
				ZoneName:        z.Name,
				Priority:        z.Priority,
				MaxInstance:     z.MaxInstance,
				CurrentInstance: 0,
			})
		}
	}
	return problems, zonePriorities
}

// AggregateInstance スコアサーバから問題環境情報を取得し、現在のインスタンス情報について集計を行う
func AggregateInstance(problems map[string]*Problem, zonePriorities []*ZonePriority, ssClient *scoreserver.Client, lg *zap.Logger) (map[string]*Problem, []*ZonePriority, []DeletionTargetInstance, error) {
	lg.Info("Scheduler: AggregateInstance")

	// ScoreServer から問題環境データを取得する
	problemEnvironments, err := ssClient.ListProblemEnvironment()
	if err != nil {
		return nil, nil, nil, err
	}

	lg.Info("Scheduler: Aggregate. Got ProblemEnvironments")

	// 削除するインスタンスリスト
	abandonedInstances := []DeletionTargetInstance{}

	for _, p := range *problemEnvironments {

		if _, ok := problems[*p.MachineImageName]; !ok {
			lg.Error("Scheduler: Aggregate. This problem name not exists. The value is " + *p.MachineImageName)
			continue
		}

		// エラーメッセージは出力するが処理は継続する
		// configファイルに書かれているMachineImageNameを正とする
		if problems[*p.MachineImageName].MachineImageName != *p.MachineImageName {
			lg.Error(fmt.Sprintf(
				"Scheduler: Aggregate. Inconsistent settings: Scheduler Value: %s, ScoreServer value: %s",
				problems[*p.MachineImageName].MachineImageName,
				*p.MachineImageName,
			))
		}

		// エラーメッセージは出力するが処理は継続する
		// configファイルに書かれているProblemIDを正とする
		if problems[*p.MachineImageName].ProblemID != p.ProblemID {
			lg.Error(fmt.Sprintf(
				"Scheduler: Aggregate. Inconsistent settings: Scheduler Value: %s, ScoreServer value: %s",
				problems[*p.MachineImageName].ProblemID,
				p.ProblemID,
			))
		}

		if p.InnerStatus == nil {
			problems[*p.MachineImageName].Ready++
			problems[*p.MachineImageName].KeptInstances = append(
				problems[*p.MachineImageName].KeptInstances,
				Instance{
					InstanceName: p.Name,
					ProjectName:  p.ProjectName,
					ZoneName:     p.ZoneName,
					InnerStatus:  p.InnerStatus,
					CreatedAt:    p.CreatedAt,
				},
			)
		} else {
			switch *p.InnerStatus {
			case types.ProblemEnvironmentInnerStatusNotReady:
				problems[*p.MachineImageName].NotReady++
			case types.ProblemEnvironmentInnerStatusReady:
				problems[*p.MachineImageName].Ready++
				problems[*p.MachineImageName].KeptInstances = append(
					problems[*p.MachineImageName].KeptInstances,
					Instance{
						InstanceName: p.Name,
						ProjectName:  p.ProjectName,
						ZoneName:     p.ZoneName,
						InnerStatus:  p.InnerStatus,
						CreatedAt:    p.CreatedAt,
					},
				)
			case types.ProblemEnvironmentInnerStatusUnderChallenge:
				problems[*p.MachineImageName].UnderChallenge++
			case types.ProblemEnvironmentInnerStatusUnderScoring:
				problems[*p.MachineImageName].UnderScoring++
			case types.ProblemEnvironmentInnerStatusAbandoned:
				problems[*p.MachineImageName].Abandoned++
				// 削除するインスタンス
				abandonedInstances = append(abandonedInstances, DeletionTargetInstance{
					ProblemName:  *p.MachineImageName,
					InstanceName: p.Name,
					ProjectName:  p.ProjectName,
					ZoneName:     p.ZoneName,
				})
			case "":
				// スコアサーバがまだ触れていないインスタンスのInnerStatusにはnil(デフォルト)が設定されている
				// そのため、scheduler的には InnerStatus に nil が設定されているインスタンスはReady扱いになる
				problems[*p.MachineImageName].Ready++
				problems[*p.MachineImageName].KeptInstances = append(problems[*p.MachineImageName].KeptInstances, Instance{
					InstanceName: p.Name,
					ProjectName:  p.ProjectName,
					ZoneName:     p.ZoneName,
					InnerStatus:  p.InnerStatus,
					CreatedAt:    p.CreatedAt,
				})
			}
		}

		problems[*p.MachineImageName].CurrentInstance++

		// ZoneごとのInstance数を集計する
		for _, zp := range zonePriorities {
			if zp.ProjectName == p.ProjectName && zp.ZoneName == p.ZoneName {
				zp.CurrentInstance++
			}
		}
	}

	return problems, zonePriorities, abandonedInstances, nil
}

func PISLogging(pis map[string]*Problem, lg *zap.Logger) {
	for pn, pi := range pis {
		lg.Info("--------Problem Environments--------")
		lg.Info("Problem Name: " + pn)
		lg.Info("Problem ID: " + pi.ProblemID)
		lg.Info("Ready: " + strconv.Itoa(pi.Ready))
		lg.Info("NotReady: " + strconv.Itoa(pi.NotReady))
		lg.Info("UnderChallenge: " + strconv.Itoa(pi.UnderChallenge))
		lg.Info("UnderScoring: " + strconv.Itoa(pi.UnderScoring))
		lg.Info("Abandoned: " + strconv.Itoa(pi.Abandoned))
		lg.Info("CurrentInstance: " + strconv.Itoa(pi.CurrentInstance))
	}
}

func ZPSLogging(zps []*ZonePriority, lg *zap.Logger) {
	for _, zp := range zps {
		lg.Info("--------Zone Priority Info--------")
		lg.Info("Project Name: " + zp.ProjectName)
		lg.Info("Zone Name:" + zp.ZoneName)
		lg.Info("Priority: " + strconv.Itoa(zp.Priority))
		lg.Info("MaxInstance: " + strconv.Itoa(zp.MaxInstance))
		lg.Info("CurrentInstance: " + strconv.Itoa(zp.CurrentInstance))
	}
}

type KeptInstances []Instance

func (a KeptInstances) Len() int           { return len(a) }
func (a KeptInstances) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a KeptInstances) Less(i, j int) bool { return a[j].CreatedAt.Before(a[i].CreatedAt) }

// statusが Ready, NotReady, nil なインスタンスでfilterして返す
func filterInstances(instances []Instance) []Instance {
	filteredInstances := []Instance{}

	for _, instance := range instances {
		if instance.InnerStatus == nil ||
			*instance.InnerStatus == types.ProblemEnvironmentInnerStatusReady ||
			*instance.InnerStatus == types.ProblemEnvironmentInnerStatusNotReady {
			filteredInstances = append(filteredInstances, instance)
		}
	}

	return filteredInstances
}

// SchedulingList 作成・削除するインスタンスを列挙する
// 削除するインスタンスは、作成日時が新しいインスタンスから削除する。
// (インスタンスを作成してからインスタンス内部でプロビジョニングを行っているため、削除する順番は新しいインスタンスからにしている)
//
// 処理について
// 「Ready, NotReady」なインスタンスが KeepInstance を超えていたらインスタンスを削除する
// ただし、このReadyは InnerStatus が nil なインスタンス(まだスコアサーバから使用されていないインスタンス) も含んでいる
func SchedulingList(problems map[string]*Problem, lg *zap.Logger) ([]CreationTargetInstance, []DeletionTargetInstance) {
	lg.Info("Scheduler: SchedulingList")

	creationTargetInstances := []CreationTargetInstance{}
	deletionTargetInstances := []DeletionTargetInstance{}

	for key, problem := range problems {
		// 新しく作成されたインスタンスから削除対象にするために作成日時でソートする
		sort.Sort(KeptInstances(problem.KeptInstances))
		// 問題に挑戦中のVMが削除されないようにReadyとNotReadyでfilterする
		filteredKeepInstances := filterInstances(problem.KeptInstances)

		// Ready と NotReady なインスタンスを保持したいインスタンスとしてカウントする
		validInstanceCount := problem.Ready + problem.NotReady

		// Ready + NotReady なインスタンスが KeepPool を超えていたらインスタンスの削除を行う
		for i := 0; validInstanceCount > problem.KeepPool && len(filteredKeepInstances) > i; i++ {
			deletionTargetInstances = append(deletionTargetInstances, DeletionTargetInstance{
				ProblemName:  key,
				InstanceName: filteredKeepInstances[i].InstanceName,
				ProjectName:  filteredKeepInstances[i].ProjectName,
				ZoneName:     filteredKeepInstances[i].ZoneName,
			})
			validInstanceCount--
		}

		// Ready + NotReady なインスタンスが KeepPool より少ない場合は作成対象にする
		for validInstanceCount < problem.KeepPool {
			creationTargetInstances = append(creationTargetInstances, CreationTargetInstance{
				ProblemName:      key,
				ProblemID:        problem.ProblemID,
				MachineImageName: problem.MachineImageName,
			})
			validInstanceCount++
		}
	}

	return creationTargetInstances, deletionTargetInstances
}

// DeleteInstances 削除対象のinstanceを全て削除する
func DeleteInstances(instances []DeletionTargetInstance, vmmsClient *vmms.Client, interval int, lg *zap.Logger) error {
	lg.Info("Scheduler: DeleteScheduler")

	for i, instance := range instances {

		// 1秒待たないとEOFエラーになる `Post "http://vm-management-service:81/instance": EOF`
		time.Sleep(time.Duration(interval) * time.Second)

		if err := vmmsClient.DeleteInstance(instance.InstanceName, instance.ProjectName, instance.ZoneName); err != nil {
			msg := ""
			for _, v := range instances[i:] {
				msg = msg + v.InstanceName + ", "
			}
			// FIXME: VM不整合が起きた時に404エラーになって処理が止まってしまうのでログを出力するだけにしている
			// return fmt.Errorf("scheduler: delete scheduler. %w remains on the delete_instance_list. %s", err, msg)
			lg.Error(fmt.Sprintf("scheduler: delete scheduler. %s remains on the delete_instance_list. %s", err.Error(), msg))
		}
		lg.Info("DeletedInstance: " + instance.ProblemName + " " + instance.InstanceName)
	}

	return nil
}

type ZonePriorities []*ZonePriority

func (a ZonePriorities) Len() int           { return len(a) }
func (a ZonePriorities) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ZonePriorities) Less(i, j int) bool { return a[i].Priority < a[j].Priority }

// CreateInstance 作成対象のinstanceを作成する
// 作成時はZonePriorityを参照し、Zoneの優先順に作成していく
func CreateInstances(instances []CreationTargetInstance, zonePriorities []*ZonePriority, vmmsClient *vmms.Client, interval int, lg *zap.Logger) error {
	lg.Info("Scheduler: CreateScheduler")

	// Zoneを優先順に並び替える
	sort.Sort(ZonePriorities(zonePriorities))

	// インスタンスを作成した回数 (作成したインスタンス数)
	i := 0

	// 優先度の高いZoneからinstanceを作成する
	for _, zonePriority := range zonePriorities {

		// Zoneに空きがある限りは対象のZoneにインスタンスを作成する
		creatableInstanceCount := zonePriority.MaxInstance - zonePriority.CurrentInstance

		for creatableInstanceCount > 0 && len(instances) > i {

			// 1秒待たないとEOFエラーになる `Post "http://vm-management-service:81/instance": EOF`
			time.Sleep(time.Duration(interval) * time.Second)

			newInstance, err := vmmsClient.CreateInstance(
				instances[i].ProblemID,
				instances[i].MachineImageName,
				zonePriority.ProjectName,
				zonePriority.ZoneName,
			)

			if err != nil {
				lg.Error("CreatedInstance: Failed to CreateInstance. " + err.Error())

				msg := ""
				for _, v := range instances[i:] {
					msg = msg + v.ProblemName + ", "
				}

				return fmt.Errorf("scheduler: create scheduler. remains on the create_instance_list. %s", msg)
			}

			lg.Info("CreatedInstance: " + newInstance.InstanceName)

			i++
			creatableInstanceCount--
		}
	}

	return nil
}
