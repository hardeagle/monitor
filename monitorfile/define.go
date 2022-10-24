package monitorfile

import "bufio"

//打印的监控日志路径
const ConstMonitorFilePath = "/usr/local/umonitor/magic/"
const ConstProductIdFormat = "_product_id=\"352\""

//帧率监控模块id
const ConstMonitorModuleFrame = 1
const ConstMonitorModuleMsgExecute = 2
const ConstMonitorModuleMsgQueue = 3
const ConstMonitorModuleLoadUserData = 4

//监控数据
type MonitorData struct {
	monitorId  uint32      //监控id
	time       int64       //时间
	moduleName string      //监控模块名
	content    interface{} //监控数据
}

//监控数据转化的接口  TODO  监控模块务必实现本接口
type MonitorInterface interface {
	//每帧结束纪录
	//nowTime 当前时间( ms )
	//tickTime 当前循环帧消耗的时间( ms )
	FrameFinish(nowTime int64, tickTime int64)
	//投递监控数据
	//nowTime 当前时间( ms )
	PostMonitorData(nowTime int64) *MonitorData
	//监控数据写入文件
	WriteMonitorData(writer *bufio.Writer, data *MonitorData)
}
