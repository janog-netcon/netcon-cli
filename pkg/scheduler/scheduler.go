package scheduler

import (
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/janog-netcon/netcon-cli/pkg/vmms"
)

func SchedulerReady(cfg *types.SchedulerConfig, ssClient *scoreserver.Client, vmmsClient *vmms.Client, lg *zap.Logger) error {
	lg.Info("Scheduler: Start Scheduling...")
	//SchedulerConfigをSchedulerInfoにまとめ,ZonePriotyを作る
	pis, zps := InitSchedulerInfo(cfg, lg)
	//ScoreServerのデータを取得し集計する
	pis, zps, err := AggregateInstance(pis, zps, ssClient, lg)
	if err != nil {
		lg.Error(err)
		return err
	}
	//ConfigよりInstanceの作成リストを削除リストを作る
	ciList, diList := SchedulingList(pis, lg)

	//Instance削除リストから対象Instanceを削除する
	err = DeleteScheduler(diList, vmmsClient, lg)
	if err != nil {
		lg.Error(err)
		return err
	}
	//Instance作成リストから対象ProblemのInstanceを作成する
	err = CreateScheduler(ciList, zps, vmmsClient, lg)
	if err != nil {
		lg.Error(err)
		return err
	}
	return nil
}

func InitSchedulerInfo(cfg *types.SchedulerConfig, lg *zap.Logger) (map[string]*types.ProblemInstance, []types.ZonePriority) {
	lg.Info("Scheduler: Init SchedulerInfo")
	var pis map[string]*types.ProblemInstance
	//Init pis
	for _, p := range cfg.Setting.Problems {
		pi := types.ProblemInstance{"", "", 0, 0, 0, 0, 0, p.KeepPool, nil, 0}
		pis[p.Name] = pi
	}
	var zps []types.ZonePriority
	//Init zps
	for _, p := range cfg.Setting.Projects {
		for _, z := range p.Zones {
			zp := types.ZonePriority{p.Name, z.Name, z.Priority, z.MaxInstance, 0}
			zps = append(zps, zp)
		}
	}
	return pis, zps
}

func AggregateInstance(pis map[string]*types.ProblemInstance, zps []types.ZonePriority, scoreserverClient *scoreserver.scoreserverClien, lg *zap.Logger) (map[string]*types.ProblemInstance, []types.ZonePriority, error) {
	lg.Info("Scheduler: Aggregate ScoreServer Info")
	//ScoreServerからデータを取得
	pes, err := scoreserverClient.ListProblemEnvironment()
	if err != nil {
		return nil, nil, err
	}

	for i, p := range pes {
		min := strings.Split(p.MachineImageName, "-")
		pn := min[0]
		pis[pn].MachineImageName = p.MachineImageName
		pis[pn].ProblemID = p.ProblemID
		switch p.InnerStatus {
		case "NOT_READY":
			pis[pn].NotReady = pis[pn].NotReady + 1
		case "READY":
			pis[pn].Ready = pis[pn].Ready + 1
			ki := types.KeepInstance{p.InstanceName, p.ProjectName, p.ZoneName, p.CreatedAt}
			pis[pn].KIS = append(pis[pn].KIS, ki)
		case "UNDER_CHALLENGE":
			pis[pn].UnderChallenge = pis[pn].UnderChallenge + 1
		case "UNDER_SCORING":
			pis[pn].UnderScoring = pis[pn].UnderScoring + 1
		case "ABANDONED":
			pis[pn].Abandoned = pis[pn].Abandoned + 1
		case "":
			pis[pn].Ready = pis[pn].Ready + 1
		}
		pis[pn].CurrentInstance = pis[pn].CurrentInstance + 1
		//ZoneごとのInstance数を集計する
		for _, zp := range zps {
			if zp.ProjectName == p.ProjectName && zp.ZoneName == p.ProjectName {
				zp.CurrentInstance = zp.CurrentInstance + 1
			}
		}
	}
	return pis, zps, nil
}

//---VMを削除するリストの作成
//問題のReady+NotReady+Abandoned数がKeepInstanceを超えてはいけない。超えてたら削除対象。
//削除するInstanceは新しく出来たものから。
//---VMを作成するリストの作成
//問題のReady+NotReady+Abandoned数がKeepPoolより少ない場合は作成対象にする
// 作成リストと削除リストを返す

type KIS []types.KeepInstance

func (a KIS) Len() int           { return len(a) }
func (a KIS) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a KIS) Less(i, j int) bool { return a[j].CreatedAt.After(a[i].CreatedAt) }

func SchedulingList(pis map[string]*types.ProblemInstance, lg *zap.Logger) ([]types.CreateInstance, []types.DeleteInstance) {
	lg.Info("Scheduler: Create Operation List")
	var ciList []types.CreateInstance
	var diList []types.DeleteInstance
	for pn, pi := range pis {
		sort.Sort(KIS(pi.KIS))
		//問題のReady+NotReady+Abandoned数がKeepInstanceを超えてはいけない。超えてたら削除対象。
		for i := 0; pi.Ready+pi.NotReady+pi.Abandoned > pi.KeepPool; i++ {
			diList = append(diList, types.DeleteInstance{pn, pi.KIS[i].InstanceName, pi.KIS[i].ProjectName, pi.KIS[i].ZoneName})
			pi.KeepPool++
		}
		//問題のReady+NotReady+Abandoned数がKeepPoolより少ない場合は作成対象にする
		for i := 0; pi.Ready+pi.NotReady+pi.Abandoned < pi.KeepPool; i++ {
			ciList = append(ciList, types.CreateInstance{pi.ProblemID, pi.MachineImageName, pi.KIS[i].ProjectName, pi.KIS[i].ZoneName})
			pi.NotReady++
		}
	}
	return ciList, diList
}

//削除対象Instanceを全て削除する
//各問題のdefaultのInstance数以下の場合は削除しない
func DeleteScheduler(dis []types.DeleteInstance, vmmsClient *vmms.Client, lg *zap.Logger) error {
	lg.Info("Scheduler: Delete Instances")
	var err error
	for i, d := range dis {
		err = vmmsClient.DeleteInstance(d.InstanceName, d.ProjectName, d.ZoneName)
		if err != nil {
			msg := ""
			for _, v := range dis[i-1:] {
				msg = msg + v.ProblemName + ","
			}
			return fmt.Errorf("Remains on the CreateInstanceList. %s", msg)
		}
	}
	return nil
}

type ZPS []types.ZonePriority

func (a ZPS) Len() int           { return len(a) }
func (a ZPS) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ZPS) Less(i, j int) bool { return a[i].Priority < a[j].Priority }

//空いてるZoneの中で優先Zoneに作成対象Instanceを作成する
func CreateScheduler(cis []types.CreateInstance, zps []types.ZonePriority, vmmsClient *vmms.Client, lg *zap.Logger) error {
	lg.Info("Scheduler: Create Instances")
	//優先Zone順に作っていく
	sort.Sort(ZPS(zps))
	//Instanceを順番につくっていく
	i := 0
	var err error
	err = nil
	//優先度の高いZoneから作る
	for _, zp := range zps {
		//Zoneに空きがある限りはそこで作る
		for zp.MaxInstance-zp.CurrentInstance > 0 {
			ci, err := vmmsClient.CreateInstance(cis[i].ProblemID, cis[i].MachineImageName, zp.ProjectName, zp.ZoneName)
			if err != nil {
				break
			}
			lg.Info("CreateInstance: %s %s", cis[i].ProblemName, ci.InstanceName)
			//作れたら次のInstanceの処理に移る
			i++
			zp.CurrentInstance++
			//Instanceがなくなったら終了
			if len(cis) <= i {
				return nil
			}
		}
		//errが入ってる場合は処理を終わらせerr処理をする
		if err != nil {
			break
		}
	}
	//err処理。
	msg := ""
	for _, v := range cis[i-1:] {
		msg = msg + v.ProblemName
	}
	return fmt.Errorf("Error. %s Remains on the CreateInstanceList. %s", err, msg)
}
