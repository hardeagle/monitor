package monitorlog

import (
	"bufio"
	"common/monitor/monitorcommon"
	"strconv"

	"github.com/ivanabc/log4go"
)

//metric对应的Label统一接口， 目前有四种类型的metric
type LabelInfo interface {
	updateData(value float64)
	writeFile(writer *bufio.Writer, baseLabel string)
	writeFileForAbnoraml(writer *bufio.Writer, baseLabel string)
	init(info *MetricBaseInfo, data *UpdateData)
}

//metric的基本信息，一个metric的所有label这些信息都相同
type MetricBaseInfo struct {
	name       string
	help       string
	metricType int
	labels     []string
	alert      bool
	limit      float64
	over       bool
	most       int
	//
	buckets   []float64
	objective map[float64]float64
}

//metric信息，一个metric中有多种不同的label组合
type MetricInfo struct {
	baseInfo *MetricBaseInfo

	//一个metric中有多个labels的组合    labelNames -> labelInfo
	LabelName map[string]LabelInfo
}

//getTypeNameByType - 根据metric类型返回对应的类型名
func (dm *MetricBaseInfo) getTypeNameByType() string {
	if dm == nil {
		return ""
	}
	switch dm.metricType {
	case monitorcommon.MetricTypeGauge:
		return "gauge"
	case monitorcommon.MetricTypeCounter:
		return "counter"
	case monitorcommon.MetricTypeSummary:
		return "summary"
	case monitorcommon.MetricTypeHistogram:
		return "histogram"
	}
	return ""
}

//updateAlertCount 根据更新的只判断是否出现异常
//@param value - 当前需要进行对比的值
//@return bool - 当前值是否异常
func (dm *MetricBaseInfo) updateAlertCount(value float64) bool {
	if dm.alert {
		if dm.over {
			if value > dm.limit {
				return true
			}
		} else {
			if value < dm.limit {
				return true
			}
		}
	}
	return false
}

//writeFile - 将一个metric的数据写入文件, 对应多个label
//@param writer - 写入文件流
func (dm *MetricInfo) writeFile(writer *bufio.Writer) {
	if gMonitorManage == nil {
		log4go.Debug("gMonitorManage writeFile not open")
		return
	}
	for _, labelInfo := range dm.LabelName {
		//正常
		baseLabel := ConstProductIdFormat + ",server_name=" + "\"" +
			gMonitorManage.serverName + "\",server_id=" + "\"" +
			strconv.FormatUint(gMonitorManage.serverId, 10) + "\""
		//
		labelInfo.writeFile(writer, baseLabel)
	}
}

//writeFileForAbnormal - 将一个metric的异常数据写入文件, 对应多个label
//@param writer - 写入文件流
func (dm *MetricInfo) writeFileForAbnormal(writer *bufio.Writer) {
	if gMonitorManage == nil {
		log4go.Debug("gMonitorManage writeFileForAbnormal not open")
		return
	}
	for _, labelInfo := range dm.LabelName {
		//异常
		baseLabel := ConstProductIdFormat + ",server_name=" + "\"" +
			gMonitorManage.serverName + "\",server_id=" + "\"" +
			strconv.FormatUint(gMonitorManage.serverId, 10) + "\""
		//
		labelInfo.writeFileForAbnoraml(writer, baseLabel)
	}
}
