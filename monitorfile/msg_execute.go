package monitorfile

import (
	"bufio"
	"fmt"
	"sync"
)

//消息处理数量纪录
const ConstMsgExecuteModuleName = "msgexecute"
const ConstMsgExecuteCountModuleName = "msgcount"
const ConstMsgExecuteMinTimeModuleName = "msgmintime"
const ConstMsgExecuteMaxTimeModuleName = "msgmaxtime"

type MsgExecuteData struct {
	count   uint32 //执行次数
	minTime int64  //执行的最小时间
	maxTime int64  //执行的最大时间
}

type ModuleMonitorMsgExecute struct {
	//纪录的消息的数量( id, 数量 )
	msgExecuteDataList sync.Map
}

//NewMonitorMsgExecuteInstance 实例接口
func NewMonitorMsgExecuteInstance() *ModuleMonitorMsgExecute {
	return &ModuleMonitorMsgExecute{
		msgExecuteDataList: sync.Map{},
	}
}

//FrameFinish
func (dm *ModuleMonitorMsgExecute) FrameFinish(nowTime int64, tickTime int64) {
}

//PostMonitorData
func (dm *ModuleMonitorMsgExecute) PostMonitorData(nowTime int64) *MonitorData {
	monitorData := &MonitorData{
		monitorId:  ConstMonitorModuleMsgExecute,
		time:       nowTime,
		moduleName: ConstMsgExecuteModuleName,
		content:    "",
	}
	return monitorData
}

//WriteMonitorData
func (dm *ModuleMonitorMsgExecute) WriteMonitorData(writer *bufio.Writer, data *MonitorData) {
	dm.msgExecuteDataList.Range(func(k, v interface{}) bool {
		//接口转换OK
		if executeData, ok := v.(*MsgExecuteData); ok {
			if executeData.count > 0 {
				fmt.Fprintf(writer, "%s{%s,process=\"%s\",server=\"%d\",name=\"count\",msgId=\"%d\"} %d\n",
					ConstMsgExecuteCountModuleName, ConstProductIdFormat,
					gMonitorManage.serverName, gMonitorManage.serverId, k, executeData.count)

				minTime := executeData.minTime / 1000000
				if minTime > 0 {
					fmt.Fprintf(writer, "%s{%s,process=\"%s\",server=\"%d\",name=\"mintime\",msgId=\"%d\"} %d\n",
						ConstMsgExecuteMinTimeModuleName, ConstProductIdFormat,
						gMonitorManage.serverName, gMonitorManage.serverId, k, minTime)
				}

				maxTime := executeData.maxTime / 1000000
				if maxTime > 0 {
					fmt.Fprintf(writer, "%s{%s,process=\"%s\",server=\"%d\",name=\"maxtime\",msgId=\"%d\"} %d\n",
						ConstMsgExecuteMaxTimeModuleName, ConstProductIdFormat,
						gMonitorManage.serverName, gMonitorManage.serverId, k, maxTime)
				}

				executeData.count = 0
				executeData.minTime = 0
				executeData.maxTime = 0
			}
		}
		return true
	})
}

//AddMsgExecute
func (dm *ModuleMonitorMsgExecute) AddMsgExecute(msgId uint32, executeTime int64) {
	executeInterface, ok := dm.msgExecuteDataList.Load(msgId)
	if !ok {
		executeData := &MsgExecuteData{
			count:   1,
			minTime: executeTime,
			maxTime: executeTime,
		}
		dm.msgExecuteDataList.Store(msgId, executeData)
	} else {
		executeData := executeInterface.(*MsgExecuteData)
		executeData.count += 1
		if executeTime < executeData.minTime {
			executeData.minTime = executeTime
		}
		if executeTime > executeData.maxTime {
			executeData.maxTime = executeTime
		}
	}
}
