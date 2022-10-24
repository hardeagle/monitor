package monitorlog

// 打印的监控日志路径
const ConstMonitorLogPath = "/usr/local/umonitor/magic/"

//const ConstMonitorLogPath = "../../umonitor/"
const ConstProductIdFormat = "_product_id=\"352\""

const startRecordTime = 20

const updateChanLen = 102400

// update数据
type UpdateData struct {
	//metric 名字
	name string
	//labelsValue 一组label值，与name组成成唯一的监控对象
	labelsValue []string
	//当前的监控值
	value float64
}
