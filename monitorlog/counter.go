package monitorlog

import (
	"bufio"
	"fmt"
)

//一个counter类型数据的数据结构，目前counter类型的数据没有异常
type CounterLabelInfo struct {
	//基本信息
	baseInfo *MetricBaseInfo
	//对应的label值
	labelValue []string
	//当前counter的累计数
	value float64
}

//updateData - 累加counter类型的数据
//@param value - 累加值
func (dm *CounterLabelInfo) updateData(value float64) {
	dm.value += value
}

//init 初始化counter类型的label
//@param info - 基本数据
//@param data - 用于初始化label对应的值
func (dm *CounterLabelInfo) init(info *MetricBaseInfo, data *UpdateData) {
	dm.baseInfo = info
	dm.labelValue = data.labelsValue
}

//writeFile 将当前的label信息写入文件
//@param writer - 写入文件流
//@param baseLabel - 一个服务器通用的基础label，一个服务器的所有label都相同
func (dm *CounterLabelInfo) writeFile(writer *bufio.Writer, baseLabel string) {
	if nil == dm.baseInfo || len(dm.labelValue) != len(dm.baseInfo.labels) {
		return
	}
	labelsName := baseLabel
	for i := 0; i < len(dm.baseInfo.labels); i++ {
		labelsName += "," + dm.baseInfo.labels[i] + "=" + "\"" + dm.labelValue[i] + "\""
	}
	fmt.Fprintf(writer, "%s{%s} %f\n", dm.baseInfo.name, labelsName, dm.value)
}

//writeFile 将当前的label异常信息写入文件
//@param writer - 写入文件流
//@param baseLabel - 一个服务器通用的基础label，一个服务器的所有label都相同
func (dm *CounterLabelInfo) writeFileForAbnoraml(writer *bufio.Writer, baseLabel string) {

}
