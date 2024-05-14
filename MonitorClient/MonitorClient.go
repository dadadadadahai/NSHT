package main

import (
	"database/sql"
	"flag"
	"net"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/goredis"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/signal"
	"git.code4.in/mobilegameserver/unibase/uniutil"
)

var (
	db_monitor   *sql.DB
	redis_handle *goredis.Redis
	MTCM         = NewMonitorClientTaskManager()
	VERSION      = "v1.00.01"
)

func DBPing() {
	if db_monitor == nil {
		return
	}
	if err := db_monitor.Ping(); err != nil {
		logging.Error("Lost db_monitor connect:%s", err.Error())
		InitDBConnect()
	}
}

func InitDBConnect() error {
	mysqlurl := config.GetConfigStr("mysql")
	if ok, err := regexp.MatchString("^mysql://.*:.*@.*/.*$", mysqlurl); ok == false || err != nil {
		logging.Error("mysql config syntax err:mysql,%s,shutdown", mysqlurl)
		return err
	}
	mysqlurl = strings.Replace(mysqlurl, "mysql://", "", 1)
	db, err := sql.Open("mysql", mysqlurl)
	if err != nil {
		logging.Error(err.Error())
		return err
	}
	dbname := strings.Split(mysqlurl, "/")
	config.SetConfig("mysql_dbname", dbname[len(dbname)-1])
	logging.Info("mysql_dbname:%s", config.GetConfigStr("mysql_dbname"))
	db_monitor = db
	db_monitor.SetMaxOpenConns(5)
	db_monitor.SetMaxIdleConns(2)
	return nil
}

func main() {
	flag_port := flag.String("port", "7002", "http port ")
	flag_logfilename := flag.String("logfilename", "tmp/mtclient.log", "log file name")
	flag_redis := flag.String("redis", "tcp://127.0.0.1:6379/0?timeout=60s&maxidle=10", "redis url ")
	flag_mtserver_origin := flag.String("monitor_server_origin", "http://127.0.0.1", "monitor server websocket origin")
	flag_mtserver_url := flag.String("monitor_server_url", "http://127.0.0.1:7002/monitor/user", "monitor server websocket url")
	if unibase.InitConfig("MonitorClient", true, VERSION) == false {
		return
	}
	if config.GetConfigStr("port") == "" {
		config.SetConfig("port", *flag_port)
	}
	if config.GetConfigStr("logfilename") == "" {
		config.SetConfig("logfilename", *flag_logfilename)
	}
	if config.GetConfigStr("monitor_server_url") == "" {
		config.SetConfig("monitor_server_url", *flag_mtserver_url)
	}
	if config.GetConfigStr("monitor_server_origin") == "" {
		config.SetConfig("monitor_server_origin", *flag_mtserver_origin)
	}
	if config.GetConfigStr("redis") == "" {
		config.SetConfig("redis", *flag_redis)
	}
	if unibase.InitServerLogger("MTC") != nil {
		return
	}

	if err := InitDBConnect(); err != nil {
		return
	}

	redisurl := config.GetConfigStr("redis")
	redis, err := goredis.DialURL(redisurl)
	if err != nil {
		logging.Error("redis init fail:%s, %s", err.Error(), redisurl)
		return
	} else {
		logging.Info("redis init ok:%s", redisurl)
	}
	redis_handle = redis

	logging.Info("server start...")

	if unibase.RegisterCommonCommandType() == false {
		logging.Error("unibase.RegisterCommonCommandType err")
		return
	}
	InitTimeTick()
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
	if InitMonitorClient() == false {
		logging.Error("InitMonitorClient error")
		return
	}
	InitTable()
	InitWebApp()
	InitHttpMsgMapMain()
	signal.InitSignal()
	http.HandleFunc("/monitor/http", unibase.HttpHandleFuncWrapper(unibase.HandleHttpChan))
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
	url_http := config.GetConfigStr("ip") + ":" + config.GetConfigStr("port") + "/monitor/http"
	logging.Info("Listening:%s", url_http)
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
	MTCM.CloseAll()
	logging.Info("server stop...:%d", runtime.NumGoroutine())
	time.Sleep(time.Second)
	logging.Info("server stop...")
	if err := uniutil.ReviewDumpPanic(dumpPanicFile); err != nil {
		logging.Error("review dump panic error: %s", err.Error())
	}
}
