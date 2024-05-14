package main

import (
	"flag"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/signal"
	"git.code4.in/mobilegameserver/unibase/uniutil"
)

var (
	GMCM    = NewGMClientTaskManager()
	VERSION = "v1.00.01"
)

func main() {

	flag_logfilename := flag.String("logfilename", "tmp/gmclient.log", "log file name")
	flag_port := flag.String("port", "7008", "http port ")
	secret_key := flag.String("secret_key", "123456", "cookie secret key")
	flag_gmserver_url := flag.String("gm_server_url", "http://127.0.0.1:7005/gm/user", "gm server websocket url")
	flag_gmserver_origin := flag.String("gm_server_origin", "http://127.0.0.1", "gm server websocket origin")
	if unibase.InitConfig("GMClient", true, VERSION) == false {
		return
	}
	if config.GetConfigStr("logfilename") == "" {
		config.SetConfig("logfilename", *flag_logfilename)
	}
	if config.GetConfigStr("port") == "" {
		config.SetConfig("port", *flag_port)
	}
	if config.GetConfigStr("secret_key") == "" {
		config.SetConfig("port", *secret_key)
	}
	if config.GetConfigStr("gm_server_url") == "" {
		config.SetConfig("gm_server_url", *flag_gmserver_url)
	}
	if config.GetConfigStr("gm_server_origin") == "" {
		config.SetConfig("gm_server_origin", *flag_gmserver_origin)
	}
	if InitConfig(false) != nil || unibase.InitServerLogger("GMC") != nil {
		logging.Error("init lua or logger error")
		return
	}
	logging.Info("server start...")

	InitHttpMsgMapMain()
	if unibase.RegisterCommonCommandType() == false {
		logging.Error("unibase.RegisterCommonCommandType err")
		return
	}
	InitTimeTick()
	signal.InitSignal()
	InitWebApp()

	http.HandleFunc("/gm/http", unibase.HttpHandleFuncWrapper(unibase.HandleHttpChan))
	///TODO WHJ 以下部分都存在多线程问题，有空统一解决
	http.TimeoutHandler(http.DefaultServeMux, 30*time.Second, "timeout")

	if config.GetConfigStr("ip") == "" {
		ifname, err := net.InterfaceByName(config.GetConfigStr("ifname"))
		if err != nil {
			logging.Error("net.InterfaceByName err:%s,%s", config.GetConfigStr("ifname"), err.Error())
		} else {
			addrs, _ := ifname.Addrs()
			if len(addrs) == 0 {
				logging.Error("no addr in ifname:%s", config.GetConfigStr("ifname"))
			} else {
				config.SetConfig("ip", strings.Split(addrs[0].String(), "/")[0])
			}
		}
	}
	dumpPanicFile, err := uniutil.DumpPanic()
	if err != nil {
		logging.Error("init dump panic error: %s", err.Error())
	}
	go func() {
		err = http.ListenAndServe(config.GetConfigStr("ip")+":"+config.GetConfigStr("port"), nil)
		if err != nil {
			logging.Error("ListenAndServe err: %s", err.Error())
			signal.ChanRunning <- false
		}
	}()
	<-signal.ChanRunning
	GMCM.CloseAll()
	logging.Info("server stop...:%d", runtime.NumGoroutine())
	time.Sleep(time.Second)
	logging.Info("server stop...")
	if err := uniutil.ReviewDumpPanic(dumpPanicFile); err != nil {
		logging.Error("review dump panic error: %s", err.Error())
	}
}
