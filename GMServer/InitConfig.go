package main

import (
	"encoding/json"
	"path/filepath"
	//	"reflect"
	"runtime"
	"sync"
	"time"

	"code.google.com/p/goprotobuf/proto"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/luaginx"
	"git.code4.in/mobilegameserver/unibase/mysqldb"
	"git.code4.in/mobilegameserver/unibase/uniutil"
)

var (
	scriptFile = "script"
	tableFile  = "table"
)

func InitConfig(first bool) error {
	if err := initLuaEngine(luaginx.GetDefaultLuaState()); err != nil {
		logging.Error("initLuaEngine err:%s", err.Error())
		return err
	}
	logging.Info("lua engine initialize ok")

	cpuNum := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuNum)
	logging.Info("set runtime.GOMAXPROCS = %d", cpuNum)
	setupMySQLDB(luaginx.GetDefaultLuaState(), config.GetConfigStr("mysql_lua"))
	if first == false {
		//luaRegisterDBHandler(nil)
		startOver()
	}
	return nil
}

func initLuaEngine(luastate *luaginx.LuaState) error {
	luastate.ResetTimer()
	luastate.LuaRegisterBasic(scriptFile, tableFile)
	luastate.LuaRegister("go", "sessionset", SessionSet)
	luastate.LuaRegister("go", "sessionget", SessionGet)
	luastate.LuaRegister("go", "sessionflush", SessionFlush)
	luastate.LuaRegister("go", "deletesession", DeleteSession)
	luastate.LuaRegister("go", "forwardCmd2ZoneTask", ForwardCmd2ZoneTask)
	luastate.LuaRegister("go", "sendCmd2ZoneTask", SendCmd2ZoneTask)
	luastate.LuaRegister("go", "sendCmd2GmTask", SendCmd2GmTask)
	luastate.LuaRegister("go", "sendCmd2HttpTask", SendCmd2HttpTask)
	initLua := filepath.Join(scriptFile, "init.lua")
	var err error
	var ok bool
	if ok, err = uniutil.Exists(initLua); err != nil {
		logging.Warning("Check file exists error: %s, %s", initLua, err.Error())
	} else if ok {
		err = luastate.LuaExecInit(initLua)
	}
	luastate.ClearOldTimer()
	return err
}

func setupMySQLDB(luastate *luaginx.LuaState, v string) {
	if v == "" {
		return
	}
	mdb, err := mysqldb.NewMySQLDB(v)
	if err != nil {
		logging.Error("Connect MySQL Error: %s, %s", err.Error(), v)
		return
	}
	if _, err := luastate.LuaDoFunc("unilight.mysqldbready", mdb); err != nil {
		logging.Error(err.Error())
	} else {
		logging.Info("MySQLDB handler register to lua OK")
	}
}
func setupRuntimeStats(ds string) error {
	if ds == "" {
		return nil
	}
	d, err := time.ParseDuration(ds)
	if err != nil {
		return err
	}
	(&sync.Once{}).Do(func() {
		go uniutil.RuntimeStats(func(s string) {
			time.Sleep(d)
			r, err := luaginx.LuaDoFunc("gcinfo")
			if err == nil {
				logging.Debug("RuntimeStats Lua: %s, %s", uniutil.HumanSize(uint64(r.(float64))*1024), s)
			} else {
				logging.Debug("RuntimeStats %s", s)
			}
		})
	})
	return nil
}

func checkRuntimeStats() error {
	uniutil.RuntimeStats(func(s string) {
		r, err := luaginx.LuaDoFunc("gcinfo")
		if err == nil {
			logging.Debug("RuntimeStats Lua: %s, %s", uniutil.HumanSize(uint64(r.(float64))*1024), s)
		} else {
			logging.Debug("RuntimeStats %s", s)
		}
	})
	return nil
}

func startOver() {
	if luaginx.LuaIsFunction("StartOver") {
		_, err := luaginx.LuaDoFunc("StartOver")
		if err != nil {
			logging.Error("Call StartOver error:%s", err.Error())
		}
	}
}

//调用lua函数，存在则返回true，否则返回false；调用处依据b来判断是否调用go相应的函数
func CallSdkLuaFunc(method, defmethod string, args ...interface{}) (ret interface{}, b bool) {
	ret, b = nil, false
	callname := ""
	if len(method) > 0 && luaginx.LuaIsFunction(method) {
		callname = method
	} else if len(defmethod) > 0 && luaginx.LuaIsFunction(defmethod) {
		callname = defmethod
	} else {
		return
	}
	var err error
	if ret, err = luaginx.LuaDoFunc(callname, args...); err != nil {
		if _, err = luaginx.LuaDoFunc("unilight.scripterror", err); err != nil {
			logging.Error("%s unilight.scripterror error: %s", callname, err.Error())
		}
	}
	return ret, true
}

//gmtask转发到游戏服的数据
func ForwardCmd2ZoneTask(taskid uint64, gameid, zoneid uint32, cmd, retcmd proto.Message) bool {
	if gmtask := gmTaskManager.GetGmTaskById(taskid); gmtask != nil {
		gmtask.ForwardGmCommand(cmd, gameid, zoneid, true, retcmd)
		return true
	}
	logging.Error("ForwardCmd2ZoneTask error, GmTask:%d not found, data:%s", taskid, unibase.GetProtoString(cmd.String()))
	return false
}

//回复ZoneTask的消息
func SendCmd2ZoneTask(taskid uint64, cmd proto.Message) bool {
	if zonetask := zoneTaskManager.GetZoneTaskById(taskid); zonetask != nil {
		zonetask.SendCmd(cmd)
		return true
	}
	logging.Error("SendCmd2ZoneTask error, Zonetask:%d not found, data:%s", taskid, unibase.GetProtoString(cmd.String()))
	return false
}

func SendCmd2HttpTask(taskid uint64, cmd proto.Message) bool {
	entry := unibase.VTM.GetEntryById(taskid)
	if task, ok := entry.(*unibase.ChanHttpTask); ok {
		data, _ := json.Marshal(cmd)
		task.SendBinary(data)
		return true
	}
	logging.Error("SendCmd2HttpTask error, HttpTask:%d not found, data:%s", taskid, unibase.GetProtoString(cmd.String()))
	return false
}

func SendCmd2GmTask(taskid uint32, cmd proto.Message) bool {
	if gmtask := gmTaskManager.GetGmTaskById(uint64(taskid)); gmtask != nil {
		gmtask.SendCmd(cmd)
		return true
	}
	logging.Error("SendCmd2GmTask error, GmTask:%d not found, data:%s", taskid, unibase.GetProtoString(cmd.String()))
	return false
}
