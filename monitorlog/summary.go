package monitorlog

import (
	"bufio"
	"fmt"
	"sort"
)

//histogram类型的数据
type SummaryLabelInfo struct {
	//基本信息
	baseInfo *MetricBaseInfo
	//一组label值
	labelValue []string
	//对应的一组百分比
	objective []float64
	//上一次写入到这一次写入的所有值
	value []float64
	//出现异常的次数
	alertCount int
}

//updateData - 统计summary类型的数据
//@param value - 统计值
func (dm *SummaryLabelInfo) updateData(v float64) {
	dm.value = append(dm.value, v)
	if dm.baseInfo.updateAlertCount(v) {
		dm.alertCount++
	}
}

//init 初始化summary类型的label
//@param info - 基本数据
//@param data - 用于初始化label对应的值
func (dm *SummaryLabelInfo) init(info *MetricBaseInfo, data *UpdateData) {
	dm.labelValue = data.labelsValue
	dm.baseInfo = info
	dm.objective = make([]float64, 0)
	for obj := range info.objective {
		dm.objective = append(dm.objective, obj)
	}
	sort.Float64s(dm.objective)
	dm.value = make([]float64, 0)
}

//writeFile 将当前的label信息写入文件
//@param writer - 写入文件流
//@param baseLabel - 一个服务器通用的基础label，一个服务器的所有label都相同
func (dm *SummaryLabelInfo) writeFile(writer *bufio.Writer, baseLabel string) {
	if nil == dm.baseInfo || len(dm.labelValue) != len(dm.baseInfo.labels) {
		return
	}
	labelsName := baseLabel
	for i := 0; i < len(dm.baseInfo.labels); i++ {
		labelsName += "," + dm.baseInfo.labels[i] + "=" + "\"" + dm.labelValue[i] + "\""
	}
	sort.Float64s(dm.value)
	valueCount := len(dm.value)
	for _, buck := range dm.objective {
		index := int(buck * float64(valueCount))
		if index >= 0 && index < valueCount {
			fmt.Fprintf(writer, "%s{%s,quantile=\"%f\"} %f\n", dm.baseInfo.name, labelsName, buck, dm.value[index])
		}
	}
}

//writeFile 将当前的label异常信息写入文件
//@param writer - 写入文件流
//@param baseLabel - 一个服务器通用的基础label，一个服务器的所有label都相同
func (dm *SummaryLabelInfo) writeFileForAbnoraml(writer *bufio.Writer, baseLabel string) {
	if dm.baseInfo != nil && dm.baseInfo.alert {
		labelsName := baseLabel
		for i := 0; i < len(dm.baseInfo.labels); i++ {
			labelsName += "," + dm.baseInfo.labels[i] + "=" + "\"" + dm.labelValue[i] + "\""
		}
		fmt.Fprintf(writer, "%s{%s} %d\n", "abnormal__"+dm.baseInfo.name, labelsName, dm.alertCount)
	}
}
