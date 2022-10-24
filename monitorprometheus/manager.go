package monitorprometheus

import (
	"common/dmlog"
	"common/dmsignal"
	"net/http"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ivanabc/log4go"
)

/*
Copyright©,2020-2021,email:wanglil@yoozoo.com
Author: wanglil
Version: 1.0.0
Date: 2020/12/29 21:56
Description:
*/

//prometheus管理器
type PrometheusManager struct {
	//服务器Id
	serverId uint64
	//服务器名字
	serverName string
}

//Run - 运行prometheus实例
//@param zone - 服务名
//@param port - 对应的端口号
func (dm *PrometheusManager) Run(zone, port string) {
	if dm == nil {
		log4go.Info("PrometheusManager Run not open")
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log4go.Critical("[PrometheusManager] Run 服务异常,请排查,err:%v stack:%s", err, debug.Stack())
			dmsignal.PN.NotifyPanic(dmsignal.PanicMonitorWritePrometheus, "[PrometheusManager] Run 出现异常请排查")
		}
	}()
	//
	worker := NewCollectorManager(zone)
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(worker)
	gatherers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		reg,
	}

	h := promhttp.HandlerFor(gatherers,
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		})
	//监控函数
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})

	// 运行http服务器
	/*
		http.Handle("/metrics", promhttp.HandlerFor(
			//prometheus.DefaultGatherer,
			gatherers,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
	*/
	if err := http.ListenAndServe(port, nil); err != nil {
		dmlog.ExitGame("普罗米修斯监听地址异常,无法启动,port:%s err:%v", port, err)
	}
}
