package monitorlog

import (
	"bufio"
	"common/monitor/monitorcommon"
	"fmt"
)

//gauge类型的数据
type GuageLabelInfo struct {
	//基本信息
	baseInfo *MetricBaseInfo
	//label值
	labelValue []string
	//当前的监控值
	value float64
	//当前的最值。 最大值或最小值由baseInfo中的most字段决定
	mostValue float64
	//出现异常的次数
	alertCount int
}

//updateData - 更新gauge类型的数据
//@param value - 更新值
func (dm *GuageLabelInfo) updateData(value float64) {
	dm.value = value
	switch dm.baseInfo.most {
	case monitorcommon.MostTypeNone:
		dm.value = value
	case monitorcommon.MostTypeMax:
		if dm.mostValue == 0.0 || dm.mostValue < value {
			dm.mostValue = value
		}
	case monitorcommon.MostTypeMin:
		if dm.mostValue == 0.0 || dm.mostValue > value {
			dm.mostValue = value
		}
	}

	if dm.baseInfo.updateAlertCount(value) {
		dm.alertCount++
	}
}

//init 初始化gauge类型的label
//@param info - 基本数据
//@param data - 用于初始化label对应的值
func (dm *GuageLabelInfo) init(info *MetricBaseInfo, data *UpdateData) {
	dm.baseInfo = info
	dm.labelValue = data.labelsValue
	dm.mostValue = 0.0
}

//writeFile 将当前的label信息写入文件
//@param writer - 写入文件流
//@param baseLabel - 一个服务器通用的基础label，一个服务器的所有label都相同
func (dm *GuageLabelInfo) writeFile(writer *bufio.Writer, baseLabel string) {
	if nil == dm.baseInfo || len(dm.labelValue) != len(dm.baseInfo.labels) {
		return
	}
	labelsName := baseLabel
	for i := 0; i < len(dm.baseInfo.labels); i++ {
		labelsName += "," + dm.baseInfo.labels[i] + "=" + "\"" + dm.labelValue[i] + "\""
	}
	switch dm.baseInfo.most {
	case monitorcommon.MostTypeNone:
		fmt.Fprintf(writer, "%s{%s} %f\n", dm.baseInfo.name, labelsName, dm.value)
	case monitorcommon.MostTypeMin:
		fmt.Fprintf(writer, "%s{%s} %f\n", dm.baseInfo.name, labelsName, dm.mostValue)
	case monitorcommon.MostTypeMax:
		fmt.Fprintf(writer, "%s{%s} %f\n", dm.baseInfo.name, labelsName, dm.mostValue)
	}
}

//writeFile 将当前的label异常信息写入文件
//@param writer - 写入文件流
//@param baseLabel - 一个服务器通用的基础label，一个服务器的所有label都相同
func (dm *GuageLabelInfo) writeFileForAbnoraml(writer *bufio.Writer, baseLabel string) {
	if dm.baseInfo != nil && dm.baseInfo.alert {
		labelsName := baseLabel
		for i := 0; i < len(dm.baseInfo.labels); i++ {
			labelsName += "," + dm.baseInfo.labels[i] + "=" + "\"" + dm.labelValue[i] + "\""
		}
		fmt.Fprintf(writer, "%s{%s} %d\n", "abnormal__"+dm.baseInfo.name, labelsName, dm.alertCount)
	}
}
