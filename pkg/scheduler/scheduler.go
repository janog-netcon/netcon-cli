package scheduler

import (
	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/janog-netcon/netcon-cli/pkg/scoreserver"
	"github.com/janog-netcon/netcon-cli/pkg/types"
)



	//情報をsiにまとめる(siを問題ごとに初期化する)
	//var si types.ScheduleInfo
	//var pis []types.ProblemInstance
	//score serverのデータを取得し集計する
	//si, err := aggregateInstance(endpoint, si)
	//configよりInstanceの作成リストを削除リストを作る
	//Instance削除リストから対象Instanceを削除する
	//Instance作成リストから対象ProblemのInstanceを作成する

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

//---VMを削除するリストの作成
//問題のReady+NotReady+Abandoned数がKeepInstanceを超えてはいけない。超えてたら削除対象。
//---VMを作成するリストの作成
//問題のReady+NotReady+Abandoned数がKeepPoolより少ない場合は作成対象にする
// 作成リストと削除リストを返す
//→ Instanceをどこに作成するかのアルゴリズムはまた別
func schedulingList(si types.ScheduleInfo) []string, []string {
	var createInstanceList []types.CreateInstance
    var deleteInstanceList []types.DeleteInstance
	for pn, pi := range si.ProblemInstances {
		//問題のReady+NotReady+Abandoned数がKeepInstanceを超えてはいけない。超えてたら削除対象。
		if( pi.Ready + pi.NotReady + pi.Abandoned > pi.KeepPool) {
			deleteInstanceList = append(deleteInstanceList, types.DeleteInstance{pi.InstanceName, pi.ProjectName, pi.ZoneName})
		}
		//問題のReady+NotReady+Abandoned数がKeepPoolより少ない場合は作成対象にする
		if( pi.Ready + pi.NotReady + pi.Abandoned < pi.KeepPool) {
			createInstanceList = append(createInstanceList, types.CreateInstance{pi.ProblemID, pi.MachineImageName, pi.ProjectName, pi.ZoneName})
		}
	}
	return createInstanceList, deleteInstanceList
}

//削除対象Instanceを全て削除する
//空いてるZoneの中で優先Zoneに作成対象Instanceを作成する
func scheduler(createInstanceList, deleteInstanceList, si types.ScheduleInfo) {
	
}