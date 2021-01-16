package scheduler

import (
	"github.com/janog-netcon/netcon-cli/pkg/types"
)

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