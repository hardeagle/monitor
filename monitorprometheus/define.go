package monitorprometheus

type FrameMetricInfo struct {
	//帧分步
	frames map[int64]int64
	//累积异常次数
	accumulateCount int64
	//运维两次来获取的过程中 处理一帧的最长时长
	maxTimeWaste int64
}

//一个监控指标的所有信息
type MetricInfo struct {
	//metric名字
	name string
	//metric 帮助信息
	info string
	//metric 对应的label名
	labelsName []string
	//metric 类型
	metricType int
	//最值类型
	mostType int
	//阈值
	limit float64
	//是否报警
	alert bool
	//是否是超过报警
	over bool
	//metric对应的所有label
	labelsInfo map[string]*MetricLabelInfo
}

//metric 一个label对应的所有信息
type MetricLabelInfo struct {
	//label对应的一组值
	labelsValue []string
	//监控的值
	value float64
}
