package monitor

import (
	"common/monitor/monitorcommon"
	"common/monitor/monitorfile"
	"common/monitor/monitorlog"
	"common/monitor/monitorprometheus"
)

var gMonitorType int

//Start - 开启一个监控服务 供应用层使用
//@param monitorType - 开启哪种类型的监控服务
//@param name - 服务器名
//@param id - 服务器唯一id
//@param port - 作为服务被监听的端口号
//@param registerMetric - 注册函数
//@param open - 该功能是否开启
func Start(monitorType int, name string, id uint64, port string, registerMetric func(), open bool) {
	//
	gMonitorType = monitorType
	//
	switch monitorType {
	case monitorcommon.MonitorTypeFile:
		monitorfile.InitMonitor(name, id, open)
	case monitorcommon.MonitorTypePrometheus:
		monitorprometheus.InitPrometheus(name, id, port, RegisterCommon, registerMetric, open)
	case monitorcommon.MonitorTypeLog:
		monitorlog.InitMetric(name, id, RegisterCommon, registerMetric, open)
	}
}

//Register  将一个metric注册到监控系统中
//@param metricType - metric类型，有四种
//@param name - metric名字
//@param help - metric帮助信息
//@param label - metric的label名
//@param alert - 不在范围内是否报警
//@param limit - 警戒值
//@param over - 如果为true，则当前大于limit就记录异常，否者小于limit记录异常
//@param most - 记录的是当前值、最大值还是最小值
//@param objective - 与metric的summary类型关联
//@param buckets - 与metric的histogram类型关联
func Register(metricType int, name string, help string, labels []string, alert bool, limit float64, over bool, most int, objective map[float64]float64, buckets []float64) {
	switch gMonitorType {
	case monitorcommon.MonitorTypePrometheus:
		monitorprometheus.RegisterMetric(metricType, name, help, labels, alert, limit, over, most, objective, buckets)
	case monitorcommon.MonitorTypeLog:
		monitorlog.RegisterMetric(metricType, name, help, labels, alert, limit, over, most, objective, buckets)
	}
}

//Update  更新metric的监控值
//@param name - 对应着一个metric
//@param labelsValue - 对应着metric里面的唯一一组监控指标
//@param value - 当前的监控值
func Update(name string, labelsValue []string, value float64) {
	switch gMonitorType {
	case monitorcommon.MonitorTypePrometheus:
		monitorprometheus.UpdateAMetricValue(name, labelsValue, value)
	case monitorcommon.MonitorTypeLog:
		monitorlog.UpdateAMetricValue(name, labelsValue, value)
	}
}
