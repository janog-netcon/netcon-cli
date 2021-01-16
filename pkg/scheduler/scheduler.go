package scheduler

import (
	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/janog-netcon/netcon-cli/pkg/vmms"
	"github.com/janog-netcon/netcon-cli/pkg/types"
)


func schedulerReady(ssClient *scoreserver.scoreserverClient, vmmsClient *vmms.vmmsClient) error {
	//情報をsiにまとめる(siを問題ごとに初期化する)
	var si types.ScheduleInfo
	var pis []types.ProblemInstance
	//score serverのデータを取得し集計する
	si, err := aggregateInstance(si, ssClient)
	if(err != nil) {
		return err
	}

	//configよりInstanceの作成リストを削除リストを作る
	createInstanceList, deleteInstanceList := schedulingList(si)
	//Instance削除リストから対象Instanceを削除する
	err:= deleteScheduler(deleteInstanceList, vmmsClient)
	if(err != nil) {
		return err
	}
	//Instance作成リストから対象ProblemのInstanceを作成する
	cis:= createScheduler(createInstanceList, si, vmmsClient)
	if(cis != nil) {
		//Zoneが空いておらず作れていないVMがある
		return cis
	}
	return nil
}

func aggregateInstance(si ScheduleInfo, scoreserverClient *scoreserver.scoreserverClient) ScheduleInfo, error{
	//ScoreServerからデータを取得
	pes, err := scoreserverClient.ListProblemEnvironment()
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

//---VMを削除するリストの作成
//問題のReady+NotReady+Abandoned数がKeepInstanceを超えてはいけない。超えてたら削除対象。
//---VMを作成するリストの作成
//問題のReady+NotReady+Abandoned数がKeepPoolより少ない場合は作成対象にする
// 作成リストと削除リストを返す
//→ Instanceをどこに作成するかのアルゴリズムはまた別
func schedulingList(si types.ScheduleInfo) []string, []string {
	var ciList []types.CreateInstance
    var diList []types.DeleteInstance
	for pn, pi := range si.ProblemInstances {
		//問題のReady+NotReady+Abandoned数がKeepInstanceを超えてはいけない。超えてたら削除対象。
		if( pi.Ready + pi.NotReady + pi.Abandoned > pi.KeepPool) {
			diList = append(diList, types.DeleteInstance{pi.InstanceName, pi.ProjectName, pi.ZoneName})
		}
		//問題のReady+NotReady+Abandoned数がKeepPoolより少ない場合は作成対象にする
		if( pi.Ready + pi.NotReady + pi.Abandoned < pi.KeepPool) {
			ciList = append(ciList, types.CreateInstance{pi.ProblemID, pi.MachineImageName, pi.ProjectName, pi.ZoneName})
		}
	}
	return ciList, diList
}

//削除対象Instanceを全て削除する
func deleteScheduler(dis []types.DeleteInstance, vmmsClient *vmms.Client) {
	var err error
	for i, d := range dis {
		err = vmmsClient.DeleteInstance(d.InstanceName, d.ProjectName, d.ZoneName)
		if(err != nil) {
			return err
		}
	}
}


type ZPS []ZonePriority

func (a ZPS) Len() int           { return len(a) }
func (a ZPS) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ZPS) Less(i, j int) bool { return a[i].Priority < a[j].Priority }


//空いてるZoneの中で優先Zoneに作成対象Instanceを作成する
func createScheduler(cis []CreateInstance, zps []ZonePriority, vmmsClient *vmms.Client) cis []CreateInstance{	
	//優先Zone順に作っていく
	sort.Sort(ZPS(zps))

	//Instanceを順番につくっていく
	i := 0
	//優先度の高いZoneから作る
	for _, zp := range(zps) {
		//Zoneに空きがある限りはそこで作る
		for _; zp.MaxInstance - zp.CurrentInstance > 0; zp.CurrentInstance++ {
			err = vmmsClient.CreateInstance(cis[i].ProblemID, cis[i].MachineImageName, zp.ProjectName, zp.ZoneName)
			if(err != nil) {
				return err
			}
			//作れたら次のInstanceの処理に移る
			i++
			//Instanceがなくなったら終了
			if(len(cis) =< i){
				return nil
			}
		}
	}
	//作れていないInstance
	return cis[i-1:]
}