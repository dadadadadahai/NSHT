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
	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/signal"
	"git.code4.in/mobilegameserver/unibase/uniutil"
	_ "github.com/go-sql-driver/mysql"
)

var (
	zoneTaskManager    = NewZoneTaskManager()
	gmTaskManager      = NewGmTaskManager()
	db_gm              *sql.DB
	VERSION            = "v1.00.01"
	ChanHttpRecive_api = make(chan *unibase.ChanHttpTask, 1024)
)

func DBPing() {
	if db_gm == nil {
		return
	}
	if err := db_gm.Ping(); err != nil {
		logging.Error("Lost db_gm connect:%s", err.Error())
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
	db_gm = db
	return nil
}

func HandleUSerCmdMessage(ws *websocket.Conn) {
	task := NewGmTask(ws, ParseReflectGmTaskCommand)
	if unibase.VTM.AddVerifyTask(task) == true {
		task.SetRemoveMeFunc(func() {
			unibase.VTM.RemoveVerifyTask(task)
		})
		nettask.LoopRecive(task)
	}
}

func HandleUSerCmdMessageJson(ws *websocket.Conn) {
	task := NewGmTaskJson(ws, ParseReflectGmTaskCommand)
	if unibase.VTM.AddVerifyTask(task) == true {
		task.SetRemoveMeFunc(func() {
			unibase.VTM.RemoveVerifyTask(task)
		})
		nettask.LoopRecive(task)
	}
}

func HandleZoneSmdMessage(ws *websocket.Conn) {
	task := NewZoneTask(ws, ParseReflectCommand)
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
	flag_logfilename := flag.String("logfilename", "tmp/gmserver.log", "log file name")
	flag_port := flag.String("port", "7005", "http port ")
	flag_port_user := flag.String("port_bw_user", "7007", "user bw port ")
	flag_port_zone := flag.String("port_bw_zone", "7006", "zone bw port ")
	if unibase.InitConfig("GMServer", true, VERSION) == false {
		return
	}
	if config.GetConfigStr("logfilename") == "" {
		config.SetConfig("logfilename", *flag_logfilename)
	}
	if config.GetConfigStr("port") == "" {
		config.SetConfig("port", *flag_port)
	}
	if config.GetConfigStr("port_bw_user") == "" {
		config.SetConfig("port_bw_user", *flag_port_user)
	}
	if config.GetConfigStr("port_bw_zone") == "" {
		config.SetConfig("port_bw_zone", *flag_port_zone)
	}
	if unibase.InitServerLogger("GMS") != nil {
		return
	}
	logging.Info("server start...")
	if err := InitConfig(false); err != nil {
		logging.Error("InitConfig error ")
		return
	}
	InitPushmsgConfig()
	if err := InitDBConnect(); err != nil {
		return
	}

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
	InitTimeTick()
	signal.InitSignal()
	check_db_init()
	check_db_update()
	InitWebService()

	handlerZoneSmd := websocket.Server{}
	handlerZoneSmd.Handler = HandleZoneSmdMessage
	handlerUserCmd := websocket.Server{}
	handlerUserCmd.Handler = HandleUSerCmdMessage
	handlerJsonCmd := websocket.Server{}
	handlerJsonCmd.Handler = HandleUSerCmdMessageJson
	http.HandleFunc("/gm/zone", unibase.HttpHandlerWrapper(handlerZoneSmd))
	http.HandleFunc("/gm/user", unibase.HttpHandlerWrapper(handlerUserCmd))
	http.HandleFunc("/gm/json", unibase.HttpHandlerWrapper(handlerJsonCmd))
	http.HandleFunc("/gm/http", unibase.HttpHandleFuncWrapper(unibase.HandleHttpChan))
	http.HandleFunc("/gm/api", unibase.HttpHandleFuncWrapperWithChan(unibase.HandleHttpChanRecive, ChanHttpRecive_api))
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
	go func() {
		raddr, err := net.ResolveTCPAddr("tcp", config.GetConfigStr("ip")+":"+config.GetConfigStr("port_bw_user"))
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
				task := NewGmTaskBw(conn)
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
	url_zone := config.GetConfigStr("ip") + ":" + config.GetConfigStr("port") + "/gm/zone"
	logging.Info("Listening:ws://%s", url_zone)
	logging.Info("Listening:tcp://%s", config.GetConfigStr("ip")+":"+config.GetConfigStr("port_bw_zone"))
	logging.Info("Listening:tcp://%s", config.GetConfigStr("ip")+":"+config.GetConfigStr("port_bw_user"))
	<-signal.ChanRunning
	logging.Info("server stop...:%d", runtime.NumGoroutine())
	zoneTaskManager.CloseAll()
	time.Sleep(time.Second)
	logging.Info("server stop...")
	if err := uniutil.ReviewDumpPanic(dumpPanicFile); err != nil {
		logging.Error("review dump panic error: %s", err.Error())
	}
}
