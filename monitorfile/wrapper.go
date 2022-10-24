package monitorfile

import (
	"bufio"
	"common/core"
	"common/dmsignal"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/ivanabc/log4go"
)

//打印的间隔时间60秒
const ConstFrameLogIntervalTime = 60000

//监控纪录管理
//纪录一段时间内的服务器情况, 并打印出来(metric日志), 供平台收录, 并呈现在web端, 可以供查询
//例子: module{process="logic",server="1108",name="mainloop",time="50ms"} 10 timestamp

//监控管理器
type ModuleMonitorManage struct {
	serverId        uint64                      //进程id
	serverName      string                      //进程名
	nextLogTime     int64                       //下次打印纪录时间( 毫秒 )
	frameStartTime  int64                       //纪录循环帧起始时间
	monitorDataChan chan *MonitorData           //需要纪录的监控数据
	moduleList      map[uint32]MonitorInterface //监控模块列表
}

//实例接口
var gMonitorManage *ModuleMonitorManage

//InitMonitor - 创建监控管理器
//@param name - 服务器名， 如logic
//@param id - 服务器id，
//@param open - 该功能是否开启
func InitMonitor(name string, id uint64, open bool) {
	if !open {
		log4go.Info("[InitMonitor] not open,%t", open)
		return
	}
	log4go.Info("[InitMonitor] start,name:%s id:%d", name, id)
	gMonitorManage = &ModuleMonitorManage{
		nextLogTime:     0,
		frameStartTime:  0,
		serverId:        id,
		serverName:      name,
		monitorDataChan: make(chan *MonitorData, 102400),
		moduleList:      make(map[uint32]MonitorInterface),
	}
	//注册模块
	gMonitorManage.RegMonitorModule(ConstMonitorModuleFrame, NewMonitorFrameInstance())
	gMonitorManage.RegMonitorModule(ConstMonitorModuleMsgExecute, NewMonitorMsgExecuteInstance())
	gMonitorManage.RegMonitorModule(ConstMonitorModuleMsgQueue, NewMonitorMsgQueueInstance())

	//创建目录
	if !gMonitorManage.IsExist(ConstMonitorFilePath) {
		gMonitorManage.CreateDir(ConstMonitorFilePath)
	}

	//启动写文件协程
	go gMonitorManage.WriteMonitor()
}

//RegMonitorModule 添加监控模块
func (dm *ModuleMonitorManage) RegMonitorModule(moduleId uint32, module MonitorInterface) {
	if dm == nil {
		log4go.Debug("[ModuleMonitorManage] RegMonitorModule not open")
		return
	}
	if _, ok := dm.moduleList[moduleId]; ok {
		log4go.Error("monitor module:%d already reg", moduleId)
		return
	}

	dm.moduleList[moduleId] = module
}

//Start 纪录帧开始的时间
func (dm *ModuleMonitorManage) Start() {
	if dm == nil {
		log4go.Debug("[ModuleMonitorManage] Start not open")
		return
	}
	dm.frameStartTime = core.Now().UnixNano() / 1000000
}

//Finish 纪录循环帧结束的时间, 返回是否需要写日志文件
func (dm *ModuleMonitorManage) Finish() {
	if dm == nil {
		log4go.Debug("[ModuleMonitorManage] Finish not open")
		return
	}
	nowTime := core.Now().UnixNano() / 1000000
	tickTime := nowTime - dm.frameStartTime

	//每帧结束回调
	for _, module := range dm.moduleList {
		module.FrameFinish(nowTime, tickTime)
	}

	//判断是否需要打印
	if nowTime < dm.nextLogTime {
		return
	}

	//每分钟的0秒打印
	dm.nextLogTime = (nowTime/ConstFrameLogIntervalTime + 1) * ConstFrameLogIntervalTime

	//开始打印日志
	for _, module := range dm.moduleList {
		monitorData := module.PostMonitorData(nowTime)
		dm.monitorDataChan <- monitorData
	}
}

//WriteMonitor 协程里打印数据
func (dm *ModuleMonitorManage) WriteMonitor() {
	if dm == nil {
		log4go.Error("[ModuleMonitorManage] WriteMonitor not open")
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log4go.Critical("[Monitor] WriteMonitor 服务异常,请排查,err:%v stack:%s", err, debug.Stack())
			dmsignal.PN.NotifyPanic(dmsignal.PanicMonitorWriteFile, "[Monitor] WriteMonitor 出现异常请排查")
		}
	}()

	for {
		select {
		case monitorData := <-dm.monitorDataChan:
			dm.WriteMonitorFile(monitorData)
		}
	}
}

//WriteMonitorFile 监控数据写入文件
func (dm *ModuleMonitorManage) WriteMonitorFile(monitorData *MonitorData) {
	if dm == nil {
		log4go.Debug("[ModuleMonitorManage] WriteMonitorFile not open")
		return
	}
	filename := fmt.Sprintf("%s%s_%s_%d.txt", ConstMonitorFilePath, monitorData.moduleName, dm.serverName, dm.serverId)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log4go.Error("Open Monitor File %s err=%v", filename, err)
		return
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	if module, ok := gMonitorManage.moduleList[monitorData.monitorId]; ok {
		module.WriteMonitorData(w, monitorData)
	}

	err = w.Flush()
	if err != nil {
		log4go.Error("Write Monitor File %s err=%v", filename, err)
	}
}

//CreateDir 创建文件夹
func (dm *ModuleMonitorManage) CreateDir(path string) {
	if dm == nil {
		log4go.Debug("[ModuleMonitorManage] CreateDir not open")
		return
	}
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

//IsExist 判断文件/文件夹是否存在
func (dm *ModuleMonitorManage) IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//循环帧开始
func LoopStart() {
	if gMonitorManage != nil {
		gMonitorManage.Start()
	}
}

//循环帧结束
func LoopFinish() {
	if gMonitorManage != nil {
		gMonitorManage.Finish()
	}
}

//纪录消息数量
func AddMsgExecute(msgId uint32, executeTime int64) {
	if gMonitorManage == nil {
		return
	}
	//
	if module, ok := gMonitorManage.moduleList[ConstMonitorModuleMsgExecute]; ok {
		if v, ok1 := module.(*ModuleMonitorMsgExecute); ok1 {
			v.AddMsgExecute(msgId, executeTime)
		}
	}
}

//纪录消息队列长度
func AddMsgQueue(queueName string, queueSize uint32) {
	if gMonitorManage == nil {
		return
	}
	//
	if module, ok := gMonitorManage.moduleList[ConstMonitorModuleMsgQueue]; ok {
		if v, ok1 := module.(*ModuleMonitorMsgQueue); ok1 {
			v.AddMsgQueue(queueName, queueSize)
		}
	}
}

//纪录加载玩家数据时间
func AddLoadUserDataTime(executeTime int64) {
	if gMonitorManage == nil {
		return
	}

	if module, ok := gMonitorManage.moduleList[ConstMonitorModuleLoadUserData]; ok {
		if v, ok1 := module.(*ModuleMonitorLoadUserData); ok1 {
			v.AddLoadUserDataTime(executeTime)
		}
	}
}
