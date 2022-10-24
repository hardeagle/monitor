package monitorfile

import (
	"bufio"
	"fmt"
)

//以50ms为一个单位纪录
const ConstFrameStageTime = 50
const ConstFrameModuleName = "frame"

type ModuleMonitorFrame struct {
	//纪录的帧率列表
	frames map[int64]int64
}

//示例接口
func NewMonitorFrameInstance() *ModuleMonitorFrame {
	return &ModuleMonitorFrame{
		frames: make(map[int64]int64),
	}
}

func (dm *ModuleMonitorFrame) FrameFinish(nowTime int64, tickTime int64) {
	//取整
	frameStageTime := ((tickTime + ConstFrameStageTime - 1) / ConstFrameStageTime) * ConstFrameStageTime

	//计数+1
	dm.frames[frameStageTime] += 1
}

func (dm *ModuleMonitorFrame) PostMonitorData(nowTime int64) *MonitorData {
	monitorContent := dm.frames
	monitorData := &MonitorData{
		monitorId:  ConstMonitorModuleFrame,
		time:       nowTime,
		moduleName: ConstFrameModuleName,
		content:    monitorContent,
	}

	dm.frames = make(map[int64]int64)
	return monitorData
}

func (dm *ModuleMonitorFrame) WriteMonitorData(writer *bufio.Writer, data *MonitorData) {
	if data == nil || data.content == nil {
		return
	}
	//接口转换OK
	if frames, ok := data.content.(map[int64]int64); ok {
		for frameStageTime, count := range frames {
			fmt.Fprintf(writer, "%s{%s,process=\"%s\",server=\"%d\",name=\"mainloop\",time=\"%dms\"} %d\n",
				ConstFrameModuleName, ConstProductIdFormat,
				gMonitorManage.serverName, gMonitorManage.serverId, frameStageTime, count)
		}
	}
}
