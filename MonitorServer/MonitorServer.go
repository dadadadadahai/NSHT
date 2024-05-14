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

	"code.google.com/p/go.net/websocket"
	_ "github.com/go-sql-driver/mysql"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/goredis"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/signal"
	"git.code4.in/mobilegameserver/unibase/uniutil"
)

var (
	loginTaskManager    = NewLoginTaskManager()
	gameTaskManager     = NewGameTaskManager()
	monitorTaskManager  = NewMonitorTaskManager()
	GlobaleUrlClientLog string
	db_monitor          *sql.DB
	redis_handle        *goredis.Redis
	VERSION             = "v1.00.01"
)

func HandleUSerCmdMessage(ws *websocket.Conn) {
	task := NewMonitorTask(ws, nettask.ParseReflectCommand)
	if unibase.VTM.AddVerifyTask(task) == true {
		task.SetRemoveMeFunc(func() {
			unibase.VTM.RemoveVerifyTask(task)
		})
		nettask.LoopRecive(task)
	}
}

func HandleUSerCmdMessageJson(ws *websocket.Conn) {
	task := NewMonitorTaskJson(ws, nettask.ParseReflectCommand)
	if unibase.VTM.AddVerifyTask(task) == true {
		task.SetRemoveMeFunc(func() {
			unibase.VTM.RemoveVerifyTask(task)
		})
		nettask.LoopRecive(task)
	}
}

func HandleZoneSmdMessage(ws *websocket.Conn) {
	task := NewZoneTask(ws, nettask.ParseReflectCommand)
	if unibase.VTM.AddVerifyTask(task) == true {
		task.SetRemoveMeFunc(func() {
			unibase.VTM.RemoveVerifyTask(task)
		})
		nettask.LoopRecive(task)
	}
}

func HandleLoginSmdMessage(ws *websocket.Conn) {
	task := NewLoginTask(ws, nettask.ParseReflectCommand)
	if unibase.VTM.AddVerifyTask(task) == true {
		task.SetRemoveMeFunc(func() {
			unibase.VTM.RemoveVerifyTask(task)
		})
		nettask.LoopRecive(task)
	}
}

func HandleResetZoneList(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	logging.Debug("HandleResetZoneList:%s", req.RemoteAddr)
	unibase.ResetZoneList()
}

func main() {
	flag_logfilename := flag.String("logfilename", "tmp/monitorserver.log", "log file name")
	// flag_redis := flag.String("redis", "tcp://127.0.0.1:6379/0?timeout=60s&maxidle=10", "redis url ")
	flag_port := flag.String("port", "7002", "http port ")
	flag_port_monitor := flag.String("port_bw_monitor", "7010", "user bw port ")
	flag_port_zone := flag.String("port_bw_zone", "7011", "zone bw port ")
	if unibase.InitConfig("MonitorServer", true, VERSION) == false {
		return
	}
	if config.GetConfigStr("logfilename") == "" {
		config.SetConfig("logfilename", *flag_logfilename)
	}
	// if config.GetConfigStr("redis") == "" {
	// 	config.SetConfig("redis", *flag_redis)
	// }
	if config.GetConfigStr("port") == "" {
		config.SetConfig("port", *flag_port)
	}
	if config.GetConfigStr("port_bw_monitor") == "" {
		config.SetConfig("port_bw_monitor", *flag_port_monitor)
	}
	if config.GetConfigStr("port_bw_zone") == "" {
		config.SetConfig("port_bw_zone", *flag_port_zone)
	}
	if unibase.InitServerLogger("MS") != nil {
		return
	}
	logging.Info("server start...")

	mysqlurl := config.GetConfigStr("mysql")
	if ok, err := regexp.MatchString("^mysql://.*:.*@.*/.*$", mysqlurl); ok == false || err != nil {
		logging.Error("mysql config syntax err:mysql,%s,shutdown", mysqlurl)
		return
	}
	mysqlurl = strings.Replace(mysqlurl, "mysql://", "", 1)
	db, err := sql.Open("mysql", mysqlurl)
	if err != nil {
		logging.Error(err.Error())
		return
	}
	dbname := strings.Split(mysqlurl, "/")
	config.SetConfig("mysql_dbname", dbname[len(dbname)-1])
	logging.Info("mysql_dbname:%s", config.GetConfigStr("mysql_dbname"))
	db_monitor = db

	redisurl := config.GetConfigStr("redis")
	redis, err := goredis.DialURL(redisurl)
	if err != nil {
		logging.Error("redis init fail:%s, %s", err.Error(), redisurl)
		return
	} else {
		logging.Info("redis init ok:%s", redisurl)
	}
	redis_handle = redis
	if unibase.ResetZoneList() == true {
		logging.Info("init zoneInfo Ok...")
	} else {
		logging.Info("init zoneInfo Error...,shutdown")
		//return
	}
	if unibase.RegisterCommonCommandType() == false {
		logging.Error("unibase.RegisterCommonCommandType err")
		return
	}

	if InitClientLogFile() != nil {
		return
	}

	InitTimeTick()
	initTable()
	InitHttpMsgMapMain()
	go func() {
		hourgetadjustreport()
	}()

	signal.InitSignal()
	handlerZoneSmd := websocket.Server{}
	handlerZoneSmd.Handler = HandleZoneSmdMessage
	handlerLoginSmd := websocket.Server{}
	handlerLoginSmd.Handler = HandleLoginSmdMessage
	handlerUserCmd := websocket.Server{}
	handlerUserCmd.Handler = HandleUSerCmdMessage
	handlerJsonCmd := websocket.Server{}
	handlerJsonCmd.Handler = HandleUSerCmdMessageJson
	http.HandleFunc("/monitor/zone", unibase.HttpHandlerWrapper(handlerZoneSmd))
	http.HandleFunc("/monitor/login", unibase.HttpHandlerWrapper(handlerLoginSmd))
	http.HandleFunc("/monitor/user", unibase.HttpHandlerWrapper(handlerUserCmd))
	http.HandleFunc("/monitor/json", unibase.HttpHandlerWrapper(handlerJsonCmd))
	http.HandleFunc("/monitor/http", unibase.HttpHandleFuncWrapper(unibase.HandleHttpChan))
	///TODO WHJ 以下部分都存在多线程问题，有空统一解决
	http.HandleFunc("/monitor/clientlog", unibase.HttpHandleFuncWrapper(WriteClientLog))
	http.HandleFunc("/monitor/clientlogview", unibase.HttpHandleFuncWrapper(ViewClientLog))
	http.HandleFunc("/monitor/resetzoneList", unibase.HttpHandleFuncWrapper(HandleResetZoneList))
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
	url_zone := config.GetConfigStr("ip") + ":" + config.GetConfigStr("port") + "/monitor/zone"
	url_user := config.GetConfigStr("ip") + ":" + config.GetConfigStr("port") + "/monitor/user"
	GlobaleUrlClientLog = "http://" + config.GetConfigStr("ip") + ":" + config.GetConfigStr("port") + "/shen/clientlog"
	logging.Info("Listening:%s", url_zone)
	logging.Info("Listening:%s", url_user)
	logging.Info("Listening:%s", GlobaleUrlClientLog)
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
	go func() {
		raddr, err := net.ResolveTCPAddr("tcp", config.GetConfigStr("ip")+":"+config.GetConfigStr("port_bw_monitor"))
		if err != nil {
			logging.Error("net.ResolveTCPAddr err: %s", err.Error())
			signal.ChanRunning <- false
		}
		logging.Info("Listening TCP:%s", raddr)
		listen, err := net.ListenTCP("tcp", raddr)
		if err != nil {
			logging.Error("listen err:%s", err.Error())
			signal.ChanRunning <- false
		} else {
			for {
				conn, err := listen.AcceptTCP()
				if err != nil {
					logging.Error("listen err:%s", err.Error())
					continue
					//<-signal.ChanRunning
				}
				logging.Info("new connection:%s", conn.RemoteAddr())
				task := NewMonitorTaskBw(conn)
				if unibase.VTM.AddVerifyTask(task) == true {
					task.SetRemoveMeFunc(func() {
						unibase.VTM.RemoveVerifyTask(task)
					})
					nettask.LoopRecive(task)
				}
			}
		}
	}()
	go func() {
		raddr, err := net.ResolveTCPAddr("tcp", config.GetConfigStr("ip")+":"+config.GetConfigStr("port_bw_zone"))
		if err != nil {
			logging.Error("net.ResolveTCPAddr err: %s", err.Error())
			signal.ChanRunning <- false
		}
		logging.Info("Listening TCP:%s", raddr)
		listen, err := net.ListenTCP("tcp", raddr)
		if err != nil {
			logging.Error("listen err:%s", err.Error())
			signal.ChanRunning <- false
		} else {
			for {
				conn, err := listen.AcceptTCP()
				if err != nil {
					logging.Error("listen err:%s", err.Error())
					continue
					//<-signal.ChanRunning
				}
				logging.Info("new connection:%s", conn.RemoteAddr())
				task := NewZoneTaskBw(conn)
				if unibase.VTM.AddVerifyTask(task) == true {
					task.SetRemoveMeFunc(func() {
						unibase.VTM.RemoveVerifyTask(task)
					})
					nettask.LoopRecive(task)
				}
			}
		}
	}()
	signal.ServerRunning = <-signal.ChanRunning
	logging.Info("server stop...:%d,%v", runtime.NumGoroutine(), signal.ServerRunning)
	gameTaskManager.CloseAll()
	loginTaskManager.CloseAll()
	monitorTaskManager.CloseAll()
	for TimeTickStoped == false {
		time.Sleep(time.Millisecond)
	}
	logging.Info("server stop...")
	if err := uniutil.ReviewDumpPanic(dumpPanicFile); err != nil {
		logging.Error("review dump panic error: %s", err.Error())
	}
}
