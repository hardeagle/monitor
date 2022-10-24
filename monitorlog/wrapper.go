package monitorlog

import (
	"common/monitor/monitorcommon"
	"strings"

	"github.com/ivanabc/log4go"
)

// 实例接口
var gMonitorManage *Manager
var gAllMetric *AllMetric

//InitMetric 创建监控实例
//@param name - 服务器名， 如logic
//@param id - 服务器id，
//@param registerMetric - 注册统一接口，应用层在里面注册所有的指标
//@param open - 该功能是否开启
//@return *Manager - 监控实例
func InitMetric(name string, id uint64, registerCommon, registerMetric func(), open bool) {
	if !open {
		log4go.Info("[InitMetric] not open,%t", open)
		return
	}
	log4go.Info("[InitMetric] start,name:%s id:%d", name, id)
	gMonitorManage = &Manager{
		serverId:   id,
		serverName: name,
		updateChan: make(chan *UpdateData, updateChanLen),
	}

	gAllMetric = &AllMetric{}
	gAllMetric.mapMetrics = make(map[string]*MetricInfo)

	// 创建目录
	if !gMonitorManage.IsExist(ConstMonitorLogPath) {
		gMonitorManage.CreateDir(ConstMonitorLogPath)
	}
	//注册基础监控模块
	registerCommon()
	//注册逻辑监控模块
	registerMetric()
	// 开启模块更新
	go gMonitorManage.monitorUpdate()
	// 启动写文件协程
	go gMonitorManage.writeMonitor()
}

//getLabelsName  将一组label值映射成为一个唯一名
//@param labels - 一组label值
//@return string - 唯一值
func getLabelsName(labels []string) string {
	sortLabels := make([]string, len(labels))
	copy(sortLabels, labels)
	return strings.Join(sortLabels, "")
}

//RegisterMetric  将一个metric注册到监控系统中
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
func RegisterMetric(metricType int, name string, help string, labels []string, alert bool, limit float64, over bool, most int, objective map[float64]float64, buckets []float64) {
	if gMonitorManage == nil {
		log4go.Debug("gMonitorManage RegisterMetric not open")
		return
	}
	if metricType == monitorcommon.MetricTypeSummary && (objective == nil || len(objective) < 1) {
		log4go.Info("summary类型未添加objective")
		return
	}
	if metricType == monitorcommon.MetricTypeHistogram && (buckets == nil || len(buckets) < 1) {
		log4go.Info("Histogram类型未添加buckets")
		return
	}
	metricInfo, ok := gAllMetric.mapMetrics[name]
	if (!ok) || nil == metricInfo {
		metricInfo = &MetricInfo{
			baseInfo: &MetricBaseInfo{
				name:       name,
				help:       help,
				metricType: metricType,
				alert:      alert,
				labels:     labels,
				limit:      limit,
				over:       over,
				most:       most,
				objective:  objective,
				buckets:    buckets,
			},
			LabelName: make(map[string]LabelInfo),
		}
		gAllMetric.mapMetrics[name] = metricInfo
	} else {
		log4go.Info("这个指标已存在")
	}
}

//UpdateAMetricValue  更新metric的监控值
//@param name - 对应着一个metric
//@param labelsValue - 对应着metric里面的唯一一组监控指标
//@param value - 当前的监控值
func UpdateAMetricValue(name string, labelsValue []string, value float64) {
	if gMonitorManage == nil || gAllMetric == nil || gAllMetric.mapMetrics == nil {
		log4go.Debug("gMonitorManage UpdateAMetricValue not open")
		return
	}
	gMonitorManage.updateChan <- &UpdateData{name: name, labelsValue: labelsValue, value: value}
}
