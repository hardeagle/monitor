package monitorprometheus

import (
	"common/monitor/monitorcommon"
	"strconv"

	"github.com/ivanabc/log4go"
	"github.com/prometheus/client_golang/prometheus"
)

//这是一个收集器
//不用默认的收集器是因为，默认的收集器无法进行逻辑运算
type CollectorManager struct {
	//收集器名
	zone string
	//对应的所有收集器
	mapDesc map[string]*prometheus.Desc
}

//NewCollectorManager 创建一个新的收集管理器
//@param zone - 服务器名
func NewCollectorManager(zone string) *CollectorManager {
	collectorManager := &CollectorManager{zone: zone, mapDesc: make(map[string]*prometheus.Desc)}
	for name, metricInfo := range allMetric.mapMetrics {
		if metricInfo.metricType == monitorcommon.MetricTypeGauge && metricInfo.mostType != monitorcommon.MostTypeNone {
			labels := []string{"serverType", "serverId"}
			labels = append(labels, metricInfo.labelsName...)
			collectorManager.mapDesc[name] = prometheus.NewDesc(name, metricInfo.info, labels, prometheus.Labels{"_product_name": zone})
		}
	}
	return collectorManager
}

//Describe simply sends the two Descs in the struct to the channel.
//Describe - 统一接口，接收收集器的监控数据
//@param ch - 接收的管道
func (c *CollectorManager) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.mapDesc {
		ch <- desc
	}
}

//Collect - 收集器，prometheus后端每发送一次请求，则该函数执行一次
//@param ch 统一收集管道
func (c *CollectorManager) Collect(ch chan<- prometheus.Metric) {
	if gPrometheusManage == nil {
		log4go.Debug("[CollectorManager] Collect not open")
		return
	}
	allMetric.lock.Lock()
	defer allMetric.lock.Unlock()
	for name, metricInfo := range allMetric.mapMetrics {
		for _, labelInfo := range metricInfo.labelsInfo {
			labels := []string{gPrometheusManage.serverName, strconv.FormatUint(gPrometheusManage.serverId, 10)}
			labels = append(labels, labelInfo.labelsValue...)
			desc := c.mapDesc[name]
			if nil != desc {
				ch <- prometheus.MustNewConstMetric(
					desc,
					prometheus.GaugeValue,
					labelInfo.value,
					labels...,
				)
			}
		}
	}
	//完成之后清空指标
	for _, metricInfo := range allMetric.mapMetrics {
		metricInfo.labelsInfo = make(map[string]*MetricLabelInfo)
	}
}
