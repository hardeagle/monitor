package monitorlog

import (
	"bufio"
	"common/dmsignal"
	"common/monitor/monitorcommon"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/ivanabc/log4go"
)

// 监控管理器
type Manager struct {
	serverId   uint64           // 进程id
	serverName string           // 进程名
	updateChan chan *UpdateData //从游戏业务发送到数据更新的管道,目的是不阻塞游戏业务

}

type AllMetric struct {
	lock       sync.RWMutex
	mapMetrics map[string]*MetricInfo // 监控模块列表
}

// writeMonitor 协程里打印数据
func (dm *Manager) writeMonitor() {
	if dm == nil {
		log4go.Debug("gMonitorManage writeMonitor not open")
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log4go.Critical("[Monitor] writeMonitor 服务异常,请排查,err:%v stack:%s", err, debug.Stack())
			dmsignal.PN.NotifyPanic(dmsignal.PanicMonitorWriteLog, "[Monitor] writeMonitor 出现异常请排查")
		}
	}()

	ticker := time.NewTicker(time.Duration(1) * time.Second)

	for {
		select {
		case <-ticker.C:
			second := time.Now().Second()
			if second == startRecordTime {
				dm.WriteMonitorFile()
			}
		}
	}
}

//monitorUpdate  在协程里面更新数据
func (dm *Manager) monitorUpdate() {
	if dm == nil {
		log4go.Error("[Manager] monitorUpdate not open")
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log4go.Critical("[Manager] monitorUpdate 服务异常,请排查,err:%v stack:%s", err, debug.Stack())
			dmsignal.PN.NotifyPanic(dmsignal.PanicMonitorMonitorUpdate, "[Manager] monitorUpdate 出现异常请排查")
		}
	}()

	for {
		select {
		case updateDate := <-dm.updateChan:
			dm.updateDate(updateDate)
		}
	}
}

//updateDate  更新数据
//@param data - 需要更新的数据
func (dm *Manager) updateDate(data *UpdateData) {
	if dm == nil {
		log4go.Debug("gMonitorManage updateDate not open")
		return
	}
	gAllMetric.lock.Lock()
	defer gAllMetric.lock.Unlock()
	//
	metricInfo, ok := gAllMetric.mapMetrics[data.name]
	if (!ok) || nil == metricInfo {
		log4go.Info("未注册指标: %s", data.name)
		return
	}
	labelsName := getLabelsName(data.labelsValue)
	labelInfo, ok := metricInfo.LabelName[labelsName]
	if (!ok) || nil == labelInfo {
		labelInfo = dm.getLabelInfoByType(metricInfo.baseInfo.metricType)
		//统一初始化
		labelInfo.init(metricInfo.baseInfo, data)
		metricInfo.LabelName[labelsName] = labelInfo
	}
	labelInfo.updateData(data.value)
}

//getLabelInfoByType  一个简单的工厂，根据类型创建metric对应的label对象
//@param metricType - metric类型
//@return LabelInfo - label对象
func (dm *Manager) getLabelInfoByType(metricType int) LabelInfo {
	var info LabelInfo
	switch metricType {
	case monitorcommon.MetricTypeGauge:
		info = &GuageLabelInfo{}
	case monitorcommon.MetricTypeCounter:
		info = &CounterLabelInfo{}
	case monitorcommon.MetricTypeSummary:
		info = &SummaryLabelInfo{}
	case monitorcommon.MetricTypeHistogram:
		info = &HistogramLabelInfo{}
	}
	return info
}

//WriteMonitorFile 监控数据写入文件
func (dm *Manager) WriteMonitorFile() {
	if dm == nil {
		log4go.Debug("gMonitorManage WriteMonitorFile not open")
		return
	}
	filename := fmt.Sprintf("%s%s_%d.txt", ConstMonitorLogPath, dm.serverName, dm.serverId)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log4go.Error("Open Monitor File %s err=%v", filename, err)
		return
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	gAllMetric.lock.Lock()
	defer gAllMetric.lock.Unlock()

	for _, metric := range gAllMetric.mapMetrics {
		if metric.baseInfo == nil {
			continue
		}
		fmt.Fprintf(w, "# HELP %s\n", metric.baseInfo.help+" "+metric.baseInfo.getTypeNameByType())
		fmt.Fprintf(w, "# TYPE %s\n", metric.baseInfo.name+" "+metric.baseInfo.getTypeNameByType())
		metric.writeFile(w)

		//写异常
		if metric.baseInfo != nil && metric.baseInfo.alert {
			fmt.Fprintf(w, "# HELP %s\n", metric.baseInfo.help+" abnormal"+" "+metric.baseInfo.getTypeNameByType())
			fmt.Fprintf(w, "# TYPE %s\n", "abnormal__"+metric.baseInfo.name+" "+metric.baseInfo.getTypeNameByType())
			metric.writeFileForAbnormal(w)
		}

		metric.LabelName = make(map[string]LabelInfo)
	}

	err = w.Flush()
	if err != nil {
		log4go.Error("Write Monitor File %s err=%v", filename, err)
	}
}

// IsExist - 判断文件/文件夹是否存在
//@param filename - 对应的文件/文件夹目录
//@return bool - 文件/文件夹是否创建成功
func (dm *Manager) IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

//CreateDir - 创建文件夹
//@param path - 文件夹路径
func (dm *Manager) CreateDir(path string) {
	err := os.MkdirAll(path, os.ModePerm)
	if err == nil {
		err = os.Chmod(path, os.ModePerm)
		if err != nil {
			log4go.Error("Chmod path %s err=%v", path, err)
		}
	} else {
		log4go.Error("Create path %s err=%v", path, err)
	}
}
