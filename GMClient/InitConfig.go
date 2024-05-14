package main

import (
	"bytes"
	"path/filepath"
	//	"reflect"
	"io/ioutil"
	"runtime"
	"sync"
	"time"

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
	luastate.LuaRegister("go", "forwardCmd2HttpTask", ForwardCmd2HttpTask)
	luastate.LuaRegister("go", "forwardCmd2GmClient", ForwardCmd2GmClient)
	luastate.LuaRegister("go", "parseMultipartForm", ParseMultipartForm)
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

//调用sdk的lua函数，存在则返回true，否则返回false；调用处依据b来判断是否调用go相应的函数
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
	logging.Info("call LuaFunc %s", callname)
	var err error
	if ret, err = luaginx.LuaDoFunc(callname, args...); err != nil {
		if _, err = luaginx.LuaDoFunc("unilight.scripterror", err); err != nil {
			logging.Error("%s unilight.scripterror error: %s", callname, err.Error())
		}
	}
	return ret, true
}

func ForwardCmd2HttpTask(taskid uint64, send GMMessage) {
	if task := GMCM.GetGMClientTaskById(taskid); task != nil {
		task.SendChanHttpTaskCmd(send)
	} else {
		logging.Warning("can not find GmClientTask:%d", taskid)
	}
}

func ForwardCmd2GmClient(taskid uint64, gameid, zoneid uint32, send GMMessage) {
	entry := unibase.VTM.GetEntryById(taskid)
	if entry != nil {
		if httptask, ok := entry.(*unibase.ChanHttpTask); ok {
			ForwardGmCommand(httptask, send, gameid, zoneid, false)
			return
		}
	}
	logging.Warning("can not find ChanHttpTask:%d", taskid)
}

func ParseMultipartForm(taskid uint64, body string) (params map[string]string) {
	entry := unibase.VTM.GetEntryById(taskid)
	if entry != nil {
		if task, ok := entry.(*unibase.ChanHttpTask); ok {
			oldbody := task.R.Body
			bf := bytes.NewBuffer([]byte(body))
			task.R.Body = ioutil.NopCloser(bf)
			task.R.ParseMultipartForm(32 << 20)
			task.R.Body.Close()
			task.R.Body = oldbody
			params = make(map[string]string)
			for k, vl := range task.R.Form {
				if len(vl) > 0 {
					params[k] = vl[0]
				} else {
					params[k] = ""
				}
			}
			return params
		}
	}
	return nil
}
