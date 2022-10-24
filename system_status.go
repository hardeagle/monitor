package monitor

import (
	"common/core"
	"common/monitor/monitorcommon"
	"runtime"
)

/*
Copyright©,2020-2021,email:wanglil@yoozoo.com
Author: wanglil
Version: 1.0.0
Date: 2021/1/23 15:07
Description:
*/

//系统状态
var SysStatus struct {
	Uptime       string //服务运行时间
	NumGoroutine int    //当前goroutine数量

	//General statistics.
	MemAllocated uint64 //当前内存使用量 bytes
	MemTotal     uint64 //所有被分配的内存 bytes
	MemSys       uint64 //内存占用量 bytes
	Lookups      uint64 //指针查找次数
	MemMallocs   uint64 //内存分配次数
	MemFrees     uint64 //内存释放次数

	//Main allocation heap statistics.
	//HeapAlloc    string //当前Heap内存使用量
	//HeapIdle     string //Heap内存空闲量
	//HeapReleased string //被释放的Heap内存

	HeapObjects uint64 //Heap对象数量
	HeapInuse   uint64 //正在使用的Heap内存 bytes
	HeapSys     uint64 //Heap内存占用量 bytes

	//Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse uint64 //正在使用的启动Stack使用量 bytes
	StackSys   uint64 //被分配的Stack内存 bytes
	OtherSys   uint64 //其他被分配的系统内存 bytes

	//MSpanInuse  string //MSpan结构内存使用量
	//MSpanSys    string //被分配的MSpan结构内存
	//MCacheInuse string //MCache结构内存使用量
	//MCacheSys   string //被分配的MCache结构内存
	//BuckHashSys string //被分配的剖析哈希表内存
	//GCSys       string //被分配的GC元数据存储

	//Garbage collector statistics.
	//NextGcStr       string //下次GC内存回收量
	NextGC uint64 //下次GC内存回收量 bytes
	//LastGcStr       string //距离上次GC时间
	LastGC uint64 //距离上次GC时间 纳秒
	//PauseTotalNsStr string //GC暂停时间总量
	PauseTotalNs uint64 //GC暂停时间总量 纳秒
	//PauseNsStr      string //上次GC暂停时间
	PauseNs uint64 //上次GC暂停时间 纳秒
	NumGC   uint32 //GC执行次数
}

//UpdateSystemStatus 定时更新
func UpdateSystemStatus() {
	SysStatus.Uptime = core.TimeSinceHuman(core.InitTime)
	//
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	SysStatus.NumGoroutine = runtime.NumGoroutine()
	SysStatus.MemAllocated = m.Alloc
	SysStatus.MemTotal = m.TotalAlloc
	SysStatus.MemSys = m.Sys
	SysStatus.Lookups = m.Lookups
	SysStatus.MemMallocs = m.Mallocs
	SysStatus.MemFrees = m.Frees

	//SysStatus.HeapAlloc = core.FileSize(int64(m.HeapAlloc))
	//SysStatus.HeapIdle = core.FileSize(int64(m.HeapIdle))
	//SysStatus.HeapReleased = core.FileSize(int64(m.HeapReleased))
	SysStatus.HeapSys = m.HeapSys
	SysStatus.HeapInuse = m.HeapInuse
	SysStatus.HeapObjects = m.HeapObjects

	SysStatus.StackInuse = m.StackInuse
	SysStatus.StackSys = m.StackSys
	//SysStatus.MSpanInuse = core.FileSize(int64(m.MSpanInuse))
	//SysStatus.MSpanSys = core.FileSize(int64(m.MSpanSys))
	//SysStatus.MCacheInuse = core.FileSize(int64(m.MCacheInuse))
	//SysStatus.MCacheSys = core.FileSize(int64(m.MCacheSys))
	//SysStatus.BuckHashSys = core.FileSize(int64(m.BuckHashSys))
	//SysStatus.GCSys = core.FileSize(int64(m.GCSys))
	//SysStatus.OtherSys = m.OtherSys

	//SysStatus.NextGcStr = core.FileSize(int64(m.NextGC))
	SysStatus.NextGC = m.NextGC
	//SysStatus.LastGcStr = fmt.Sprintf("%.1fs", float64(core.Now().UnixNano()-int64(m.LastGC))/1000/1000/1000)
	SysStatus.LastGC = uint64(core.Now().UnixNano()) - m.LastGC
	//SysStatus.PauseTotalNsStr = fmt.Sprintf("%.1fs", float64(m.PauseTotalNs)/1000/1000/1000)
	SysStatus.PauseTotalNs = m.PauseTotalNs
	//SysStatus.PauseNsStr = fmt.Sprintf("%.3fs", float64(m.PauseNs[(m.NumGC+255)%256])/1000/1000/1000)
	SysStatus.PauseNs = m.PauseNs[(m.NumGC+255)%256]
	SysStatus.NumGC = m.NumGC

	//更新基础模块
	UpdateCommon()
}

//注册基础监控模块
func RegisterCommon() {
	Register(monitorcommon.MetricTypeGauge,
		"goroutine_num",
		"the number of goroutines that currently exist",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		nil,
	)
	Register(monitorcommon.MetricTypeGauge,
		"gc_count",
		"the number of completed GC cycles",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		nil,
	)
	Register(monitorcommon.MetricTypeHistogram,
		"gc_pause_time",
		"a circular buffer of recent GC stop-the-world(pause times in nanoseconds)",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		[]float64{1000, 1000000, 5000000, 1000000000, 10000000000},
	)
	Register(monitorcommon.MetricTypeHistogram,
		"gc_last_time",
		"the time the last garbage collection finished",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		[]float64{1000, 1000000, 5000000, 1000000000, 10000000000},
	)
	Register(monitorcommon.MetricTypeHistogram,
		"gc_heap_size",
		"the target heap size of the next GC cycle",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		[]float64{1024, 262144, 524288, 1048576, 10485760, 104857600},
	)
	Register(monitorcommon.MetricTypeHistogram,
		"mem_allocated",
		"bytes of allocated heap objects",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		[]float64{1024, 262144, 524288, 1048576, 10485760, 104857600},
	)
	Register(monitorcommon.MetricTypeHistogram,
		"mem_allocated_total",
		"cumulative bytes allocated for heap objects",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		[]float64{1024, 262144, 524288, 1048576, 10485760, 104857600},
	)
	Register(monitorcommon.MetricTypeGauge,
		"mem_sys",
		"the total bytes of memory obtained from the OS",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		nil,
	)
	Register(monitorcommon.MetricTypeGauge,
		"mem_heap_count",
		"the number of allocated heap objects",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		nil,
	)
	Register(monitorcommon.MetricTypeGauge,
		"mem_heap_sys",
		"bytes of heap memory obtained from the OS",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		nil,
	)
	Register(monitorcommon.MetricTypeGauge,
		"mem_heap_inuse",
		"bytes in in-use spans",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		nil,
	)
	Register(monitorcommon.MetricTypeGauge,
		"mem_stack_sys",
		"bytes of stack memory obtained from the OS",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		nil,
	)
	Register(monitorcommon.MetricTypeGauge,
		"mem_stack_inuse",
		"bytes in stack spans",
		[]string{},
		true,
		200,
		true,
		monitorcommon.MostTypeNone,
		nil,
		nil,
	)
}

func UpdateCommon() {
	Update("goroutine_num", nil, float64(SysStatus.NumGoroutine))
	//GC
	Update("gc_count", nil, float64(SysStatus.NumGC))
	Update("gc_pause_time", nil, float64(SysStatus.PauseNs))
	Update("gc_last_time", nil, float64(SysStatus.LastGC))
	Update("gc_heap_size", nil, float64(SysStatus.NextGC))
	//Mem
	Update("mem_allocated", nil, float64(SysStatus.MemAllocated))
	Update("mem_allocated_total", nil, float64(SysStatus.MemTotal))
	Update("mem_sys", nil, float64(SysStatus.MemSys))

	Update("mem_heap_count", nil, float64(SysStatus.HeapObjects))
	Update("mem_heap_sys", nil, float64(SysStatus.HeapSys))
	Update("mem_heap_inuse", nil, float64(SysStatus.HeapInuse))

	Update("mem_stack_sys", nil, float64(SysStatus.StackSys))
	Update("mem_stack_inuse", nil, float64(SysStatus.StackInuse))
}
