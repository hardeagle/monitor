package monitorprometheus

import (
	"common/monitor/monitorcommon"
	"strconv"
	"strings"
	"sync"

	"github.com/ivanabc/log4go"
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeDes struct {
	general  *prometheus.GaugeVec
	abnormal *prometheus.CounterVec
}

type CounterDes struct {
	general  *prometheus.CounterVec
	abnormal *prometheus.CounterVec
}

type SummaryDes struct {
	general  *prometheus.SummaryVec
	abnormal *prometheus.CounterVec
}

type HistogramDes struct {
	general  *prometheus.HistogramVec
	abnormal *prometheus.CounterVec
}

type metricIndex struct {
	lock       sync.RWMutex
	mapMetrics map[string]*MetricInfo
	//四种类型的metric生成器
	gaugeMetric     map[string]*GaugeDes
	counterMetric   map[string]*CounterDes
	summaryMetric   map[string]*SummaryDes
	histogramMetric map[string]*HistogramDes
}

//所有的监控信息
var allMetric *metricIndex

//实例接口
var gPrometheusManage *PrometheusManager

//InitPrometheus 创建监控管理器
//@param name - 服务器名， 如logic
//@param id - 服务器id，
//@param port - 被监听的端口号
//@param registerMetric - 注册统一接口，应用层在里面注册所有的指标
//@param open - 该功能是否开启
func InitPrometheus(name string, id uint64, port string, registerCommon, registerMetric func(), open bool) {
	//未开启
	if !open {
		log4go.Info("[InitPrometheus] not open,%t", open)
		return
	}
	log4go.Info("[InitPrometheus] start,name:%s id:%d port:%s", name, id, port)
	gPrometheusManage = &PrometheusManager{
		serverId:   id,
		serverName: name,
	}
	allMetric = &metricIndex{}
	allMetric.mapMetrics = make(map[string]*MetricInfo)
	allMetric.gaugeMetric = make(map[string]*GaugeDes)
	allMetric.counterMetric = make(map[string]*CounterDes)
	allMetric.summaryMetric = make(map[string]*SummaryDes)
	allMetric.histogramMetric = make(map[string]*HistogramDes)

	//注册基础监控
	registerCommon()
	//注册逻辑监控
	registerMetric()
	//开启一个协程处理
	go gPrometheusManage.Run(name, port)
}

//getLabelsName  将一组label值映射成为一个唯一名
//@param labels - 一组label值
//@return string - 唯一值
func getLabelsName(labels []string) string {
	sortLabels := make([]string, len(labels))
	copy(sortLabels, labels)
	return strings.Join(sortLabels, "")
}

//addServerToLabels - 将serverType和serverId加入到基本的label中
//@param label - 需要添加的label
//@return [] string - 返回组合的一个新label
func addServerToLabels(labels []string) []string {
	allLabels := []string{"serverType", "serverId"}
	return append(allLabels, labels...)
}

//addServerToLabelsValue - 将serverType和serverId对应的值加入到基本的label值中
//@param label - 需要添加的label
//@return [] string - 返回组合的一个新label
func addServerToLabelsValue(labels []string) []string {
	if gPrometheusManage == nil {
		log4go.Info("[addServerToLabelsValue] not open")
		return []string{""}
	}
	allLabels := []string{gPrometheusManage.serverName, strconv.FormatInt(int64(gPrometheusManage.serverId), 10)}
	return append(allLabels, labels...)
}

//generateGaugeMetric - 生成一个gauge类型metric对象
//@param name - metric对应的名字
//@param help - metric对应的帮助信息
//@param allLabels metric对应的label值
//@return GaugeVec 生成的gauge类型的metric对象
func generateGaugeMetric(name string, help string, allLabels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		allLabels,
	)
}

//generateCounterMetric - 生成一个counter类型metric对象
//@param name - metric对应的名字
//@param help - metric对应的帮助信息
//@param allLabels metric对应的label值
//@return CounterVec 生成的counter类型的metric对象
func generateCounterMetric(name string, help string, allLabels []string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		allLabels,
	)
}

//generateSummaryMetric - 生成一个summary类型metric对象
//@param name - metric对应的名字
//@param help - metric对应的帮助信息
//@param allLabels metric对应的label值
//@param objectives 对应的百分比分段
//@return SummaryVec 生成的Summary类型的metric对象
func generateSummaryMetric(name string, help string, allLabels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		allLabels,
	)
}

//generateHistogramMetric - 生成一个histogram类型metric对象
//@param name - metric对应的名字
//@param help - metric对应的帮助信息
//@param allLabels metric对应的label值
//@param buckets 对应分段区间
//@return HistogramVec 生成的histogram类型的metric对象
func generateHistogramMetric(name string, help string, allLabels []string, buckets []float64) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		allLabels,
	)
}

//updateAbnormalInfo 异常处理
//@param abnormal - 异常指标
//@param metricInfo - 基本的metric信息
//@param allLabel - 基本的label值
//@param value - 需要更新的值
func updateAbnormalInfo(abnormal *prometheus.CounterVec, metricInfo *MetricInfo, allLabel []string, value float64) {
	if abnormal == nil || metricInfo == nil {
		return
	}
	if metricInfo.over {
		if value < metricInfo.limit {
			abnormal.WithLabelValues(allLabel...).Inc()
		}
	} else {
		if value > metricInfo.limit {
			abnormal.WithLabelValues(allLabel...).Inc()
		}
	}
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
func RegisterMetric(metricType int, name string, info string, labelNames []string,
	alert bool, limit float64, over bool, mostType int,
	objectives map[float64]float64, buckets []float64) {

	if gPrometheusManage == nil {
		log4go.Debug("gPrometheusManage RegisterMetric not open")
		return
	}

	if (metricType != monitorcommon.MetricTypeGauge) &&
		(metricType != monitorcommon.MetricTypeCounter) &&
		(metricType != monitorcommon.MetricTypeSummary) &&
		(metricType != monitorcommon.MetricTypeHistogram) {
		log4go.Info("指标的metricType不存在")
		return
	}
	metricInfo, ok := allMetric.mapMetrics[name]
	if (!ok) || nil == metricInfo {
		metricInfo = &MetricInfo{
			name:       name,
			info:       info,
			labelsName: labelNames,
			metricType: metricType,
			mostType:   mostType,
			limit:      limit,
			alert:      alert,
			over:       over,
			labelsInfo: make(map[string]*MetricLabelInfo),
		}
		allMetric.mapMetrics[name] = metricInfo
	} else {
		log4go.Info("这个指标已存在")
	}
	//如果是这种类型，不用在这里注册
	if metricInfo.metricType == monitorcommon.MetricTypeGauge && metricInfo.mostType != monitorcommon.MostTypeNone {
		return
	}
	//用于注册
	abnormalMetric := generateCounterMetric("abnormal__"+name, info, addServerToLabels(labelNames))
	switch metricType {
	case monitorcommon.MetricTypeGauge:
		{
			gaugeMetric := generateGaugeMetric(name, info, addServerToLabels(labelNames))
			allMetric.gaugeMetric[name] = &GaugeDes{general: gaugeMetric, abnormal: abnormalMetric}
			prometheus.MustRegister(gaugeMetric)
		}
	case monitorcommon.MetricTypeCounter:
		{
			counterMetric := generateCounterMetric(name, info, addServerToLabels(labelNames))
			allMetric.counterMetric[name] = &CounterDes{general: counterMetric, abnormal: abnormalMetric}
			prometheus.MustRegister(counterMetric)
		}
	case monitorcommon.MetricTypeSummary:
		{
			summaryMetric := generateSummaryMetric(name, info, addServerToLabels(labelNames), objectives)
			allMetric.summaryMetric[name] = &SummaryDes{general: summaryMetric, abnormal: abnormalMetric}
			prometheus.MustRegister(summaryMetric)
		}
	case monitorcommon.MetricTypeHistogram:
		{
			histogramMetric := generateHistogramMetric(name, info, addServerToLabels(labelNames), buckets)
			allMetric.histogramMetric[name] = &HistogramDes{general: histogramMetric, abnormal: abnormalMetric}
			prometheus.MustRegister(histogramMetric)
		}
	}
	prometheus.MustRegister(abnormalMetric)
}

//UpdateAMetricValue  更新metric的监控值
//@param name - 对应着一个metric
//@param labelsValue - 对应着metric里面的唯一一组监控指标
//@param value - 当前的监控值
func UpdateAMetricValue(name string, labelsValue []string, value float64) {
	if gPrometheusManage == nil {
		log4go.Debug("[UpdateAMetricValue] not open")
		return
	}
	allMetric.lock.Lock()
	defer allMetric.lock.Unlock()

	metricInfo, ok := allMetric.mapMetrics[name]
	if (!ok) || nil == metricInfo {
		log4go.Info("未注册该指标")
		return
	}
	//
	if metricInfo.metricType == monitorcommon.MetricTypeGauge && metricInfo.mostType != monitorcommon.MostTypeNone {
		labelsName := getLabelsName(labelsValue)
		labelInfo, ok := metricInfo.labelsInfo[labelsName]
		if (!ok) || nil == labelInfo {
			labelInfo = &MetricLabelInfo{}
			metricInfo.labelsInfo[labelsName] = labelInfo
		}
		labelInfo.labelsValue = labelsValue
		if labelInfo.value <= 0 {
			labelInfo.value = value
		}
		switch metricInfo.mostType {
		case monitorcommon.MostTypeMax:
			if value > labelInfo.value {
				labelInfo.value = value
			}
		case monitorcommon.MostTypeMin:
			if value < labelInfo.value {
				labelInfo.value = value
			}
		}
		return
	}
	//
	switch metricInfo.metricType {
	case monitorcommon.MetricTypeGauge:
		gaugeMetric := allMetric.gaugeMetric[name]
		if gaugeMetric != nil {
			if gaugeMetric.general != nil {
				gaugeMetric.general.WithLabelValues(addServerToLabelsValue(labelsValue)...).Set(value)
			}
			updateAbnormalInfo(gaugeMetric.abnormal, metricInfo, addServerToLabelsValue(labelsValue), value)
		}
	case monitorcommon.MetricTypeCounter:
		//累计好像没有什么异常，只是计数而已
		counterMetric := allMetric.counterMetric[name]
		if counterMetric != nil {
			if counterMetric.general != nil {
				counterMetric.general.WithLabelValues(addServerToLabelsValue(labelsValue)...).Add(value)
			}
		}
	case monitorcommon.MetricTypeSummary:
		summaryMetric := allMetric.summaryMetric[name]
		if summaryMetric != nil {
			if summaryMetric.general != nil {
				summaryMetric.general.WithLabelValues(addServerToLabelsValue(labelsValue)...).Observe(value)
			}
			updateAbnormalInfo(summaryMetric.abnormal, metricInfo, addServerToLabelsValue(labelsValue), value)
		}
	case monitorcommon.MetricTypeHistogram:
		histogramMetric := allMetric.histogramMetric[name]
		if histogramMetric != nil {
			if histogramMetric.general != nil {
				histogramMetric.general.WithLabelValues(addServerToLabelsValue(labelsValue)...).Observe(value)
			}
			updateAbnormalInfo(histogramMetric.abnormal, metricInfo, addServerToLabelsValue(labelsValue), value)
		}
	}
}
