package monitorfile

import (
	"bufio"
	"fmt"
)

//以50ms为一个单位纪录
const ConstLoadDataStageTime = 50
const ConstLoadDataModuleName = "loaduserdata"

type ModuleMonitorLoadUserData struct {
	//纪录的加载时间列表
	loadTimes map[int64]int64
}

//示例接口
func NewMonitorLoadUserDataInstance() *ModuleMonitorLoadUserData {
	return &ModuleMonitorLoadUserData{
		loadTimes: make(map[int64]int64),
	}
}

func (dm *ModuleMonitorLoadUserData) FrameFinish(nowTime int64, tickTime int64) {
}

func (dm *ModuleMonitorLoadUserData) AddLoadUserDataTime(executeTime int64) {
	//取整
	loadStageTime := ((executeTime + ConstFrameStageTime - 1) / ConstFrameStageTime) * ConstFrameStageTime
	//计数+1
	dm.loadTimes[loadStageTime] += 1
}

func (dm *ModuleMonitorLoadUserData) PostMonitorData(nowTime int64) *MonitorData {
	monitorContent := dm.loadTimes
	monitorData := &MonitorData{
		monitorId:  ConstMonitorModuleLoadUserData,
		time:       nowTime,
		moduleName: ConstLoadDataModuleName,
		content:    monitorContent,
	}
	dm.loadTimes = make(map[int64]int64)
	return monitorData
}

func (dm *ModuleMonitorLoadUserData) WriteMonitorData(writer *bufio.Writer, data *MonitorData) {
	if data == nil || data.content == nil {
		return
	}
	//接口转换OK
	if loadTimes, ok := data.content.(map[int64]int64); ok {
		for stageTime, count := range loadTimes {
			fmt.Fprintf(writer, "%s{%s,process=\"%s\",server=\"%d\",name=\"loaduser\",time=\"%dms\"} %d\n",
				ConstLoadDataModuleName, ConstProductIdFormat,
				gMonitorManage.serverName, gMonitorManage.serverId, stageTime, count)
		}
	}
}
