package monitorcommon

//模块监控方式
const (
	MonitorTypeNone       = iota
	MonitorTypeFile       //自定义文件方式,写死无法自定义
	MonitorTypePrometheus //提供端口方式
	MonitorTypeLog        //写metric日志方式,上层逻辑自定义
)

const (
	MostTypeNone = iota
	MostTypeMax
	MostTypeMin
)

const (
	MetricTypeNone      = iota
	MetricTypeGauge     //可增可减的仪表盘
	MetricTypeCounter   //只增不减的计数器
	MetricTypeSummary   //统计和分析样本的分布情况
	MetricTypeHistogram //统计和分析样本的分布情况
)
