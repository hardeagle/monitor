package monitorlog

import (
	"bufio"
	"fmt"
)

//histogram类型的数据
type HistogramLabelInfo struct {
	//基本信息
	baseInfo *MetricBaseInfo
	//label值
	labelValue []string
	//对应的桶
	buckets []float64
	//桶对应的映射
	value map[float64]float64
	//桶内的所有数据
	maxCount float64
	//出现异常的次数
	alertCount int
}

//updateData - 统计histogram类型的数据
//@param value - 统计值
func (dm *HistogramLabelInfo) updateData(value float64) {
	for _, buck := range dm.buckets {
		if value <= buck {
			dm.value[buck] += 1
		}
	}
	dm.maxCount += 1
	//
	if dm.baseInfo.updateAlertCount(value) {
		dm.alertCount++
	}
}

//init 初始化histogram类型的label
//@param info - 基本数据
//@param data - 用于初始化label对应的值
func (dm *HistogramLabelInfo) init(info *MetricBaseInfo, data *UpdateData) {
	dm.baseInfo = info
	dm.labelValue = data.labelsValue
	dm.buckets = info.buckets
	dm.value = make(map[float64]float64)
	for _, value := range dm.buckets {
		dm.value[value] = 0
	}
}

//writeFile 将当前的label信息写入文件
//@param writer - 写入文件流
//@param baseLabel - 一个服务器通用的基础label，一个服务器的所有label都相同
func (dm *HistogramLabelInfo) writeFile(writer *bufio.Writer, baseLabel string) {
	if nil == dm.baseInfo {
		return
	}
	if len(dm.labelValue) != len(dm.baseInfo.labels) {
		return
	}
	labelsName := baseLabel
	for i := 0; i < len(dm.baseInfo.labels); i++ {
		labelsName += dm.baseInfo.labels[i] + "=" + "\"" + dm.labelValue[i] + "\""
		if i < len(dm.baseInfo.labels)-1 {
			labelsName += ","
		}
	}
	for _, buck := range dm.buckets {
		value, ok := dm.value[buck]
		if ok {
			fmt.Fprintf(writer, "%s{%s,le=\"%f\"} %f\n", dm.baseInfo.name, labelsName, buck, value)
		}
	}
	fmt.Fprintf(writer, "%s{%s,le=\"Inf\"} %f\n", dm.baseInfo.name, labelsName, dm.maxCount)
}

//writeFile 将当前的label异常信息写入文件
//@param writer - 写入文件流
//@param baseLabel - 一个服务器通用的基础label，一个服务器的所有label都相同
func (dm *HistogramLabelInfo) writeFileForAbnoraml(writer *bufio.Writer, baseLabel string) {
	if dm.baseInfo != nil && dm.baseInfo.alert {
		labelsName := baseLabel
		for i := 0; i < len(dm.baseInfo.labels); i++ {
			labelsName += "," + dm.baseInfo.labels[i] + "=" + "\"" + dm.labelValue[i] + "\""
		}
		fmt.Fprintf(writer, "%s{%s} %d\n", "abnormal__"+dm.baseInfo.name, labelsName, dm.alertCount)
	}
}
