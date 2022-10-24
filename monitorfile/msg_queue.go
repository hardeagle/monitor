package monitorfile

import (
	"bufio"
	"fmt"
)

//消息数量纪录
const ConstMsgQueueModuleName = "msgqueue"

type ModuleMonitorMsgQueue struct {
	msgQueues map[string]uint32 //纪录消息队列的数量
}

//实例接口
func NewMonitorMsgQueueInstance() *ModuleMonitorMsgQueue {
	return &ModuleMonitorMsgQueue{
		msgQueues: make(map[string]uint32),
	}
}

func (dm *ModuleMonitorMsgQueue) FrameFinish(nowTime int64, tickTime int64) {
}

func (dm *ModuleMonitorMsgQueue) PostMonitorData(nowTime int64) *MonitorData {
	msgQueues := dm.msgQueues
	monitorData := &MonitorData{
		monitorId:  ConstMonitorModuleMsgQueue,
		time:       nowTime,
		moduleName: ConstMsgQueueModuleName,
		content:    msgQueues,
	}

	dm.msgQueues = make(map[string]uint32)
	return monitorData
}

func (dm *ModuleMonitorMsgQueue) WriteMonitorData(writer *bufio.Writer, data *MonitorData) {
	if data == nil || data.content == nil {
		return
	}
	//接口转换OK
	if msgQueues, ok := data.content.(map[string]uint32); ok {
		for queueName, queueSize := range msgQueues {
			fmt.Fprintf(writer, "%s{%s,process=\"%s\",server=\"%d\",name=\"%s\"} %d\n",
				ConstMsgQueueModuleName, ConstProductIdFormat,
				gMonitorManage.serverName, gMonitorManage.serverId, queueName, queueSize)
		}
	}
}

func (dm *ModuleMonitorMsgQueue) AddMsgQueue(queueName string, queueSize uint32) {
	dm.msgQueues[queueName] = queueSize
}
