package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"

	sjson "github.com/bitly/go-simplejson"

	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/luaginx"
)

func HandleHttpGm(task *unibase.ChanHttpTask) bool {
	defer func() {
		if err := recover(); err != nil {
			task.Error("HandleHttpGm error:%v, url:%s, body:%s", err, task.R.RequestURI, task.Rawdata)
		}
	}()
	defer task.R.Body.Close()
	rawdata := task.Rawdata
	if rawdata == nil || len(rawdata) == 0 {
		if HandleGmCommand(task) == false {
			task.Error("HandleHttpGm HandleGmCommand err:%p,%s,%s", rawdata, task.R.RemoteAddr, task.R.URL.String())
			return false
		}
		return true
	}
	task.Debug("Handle %s, %s", task.R.RequestURI, string(rawdata))
	task.JSW = unibase.NewJSResponseWriter(task.W, nil, nil)
	sj, err := sjson.NewJson(rawdata)
	if err != nil {
		logging.Error("HandleHttpGm ParseJson error:%s, data:%s", err.Error(), string(rawdata))
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return false
	}
	cmd := task.R.FormValue("do")
	tmpid := GetSecureCookie(task.R, "gmid")
	gmid, _ := strconv.Atoi(tmpid)
	if cmd != "RequestLoginGmUserPmd_C" && (gmid == 0 || SessionGet(tmpid, "login") != "1") {
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"未登录","redirecturl":"/login"}`))
		return false
	}
	if cmd != "RequestLoginGmUserPmd_C" {
		gmdata, ok := sj.CheckGet("gmdata")
		if !ok {
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
			return false
		}
		zoneid := gmdata.Get("zoneid").MustInt()
		if zoneid != 0 {
			SessionSet(tmpid, "zoneid", strconv.Itoa(zoneid))
		}
		tmpgameid := int(unibase.Atoi(SessionGet(tmpid, "gameid"), 0))
		if tmpgameid != 0 {
			gmdata.Set("gameid", tmpgameid)
		}
		gmdata.Set("gmid", gmid)
		gmdata.Set("clientid", task.GetId())
	}
	ip := Ip2Int(GetRemoteIp(task.R))
	luafun := "GmHttp.Parse" + cmd
	if luaginx.LuaIsFunction(luafun) {
		retjson, err := luaginx.LuaDoFunc(luafun, cmd, sj, task, ip)
		if err != nil {
			logging.Error("%s error: %s", luafun, err.Error())
			return false
		}
		retstr, ok := retjson.(string)
		if ok {
			task.SendBinary([]byte(retstr))
		} else {
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"系统错误"}`))
		}
	} else if gofun, ok := jsmsgMainMap[cmd]; ok {
		gofun(cmd, sj, task, ip)
	} else {
		ForwardGmCommand(cmd, task, sj, true)
	}
	return true
}

//h5.bwgame.com.cn/SuperSlot/?platid=67&uid=10722&account=official&email=whj@whj.whj&gender=male&nickname=navy1125&timestamp=12345&sign=54ad6e9d12baac23ed73e5613913a6db
func gmGetLoginSign(account, email, gender, gameid, nickname, platid, uid string) string {
	key := config.GetConfigStr("key_" + platid + "_" + gameid)
	if key == "" {
		key = config.GetConfigStr("key_" + platid)
	}
	hashstr := account + email + gender + gameid + nickname + platid + /*strconv.Itoa(int(time.Now().Unix()))*/ "12345" + uid
	hash := md5.New()
	io.WriteString(hash, hashstr+key)
	sum := fmt.Sprintf("%x", hash.Sum(nil))
	return hashstr + ":" + sum
}

func HandleGmCommand(task *unibase.ChanHttpTask) bool {
	cmd := task.R.FormValue("cmd")
	task.Debug("HandleGmCommand:%s,gameid:%s,zoneid:%s,cmd:%s", task.R.URL.Path, task.R.FormValue("gameid"), task.R.FormValue("zoneid"), cmd)
	task.JSW = unibase.NewJSResponseWriter(task.W, nil, nil)
	if cmd == "getsign" {
		task.SendBinary([]byte(gmGetLoginSign(task.R.FormValue("account"), task.R.FormValue("email"), task.R.FormValue("gender"), task.R.FormValue("gameid"), task.R.FormValue("nickname"), task.R.FormValue("platid"), task.R.FormValue("uid"))))
		return true
	}
	gamezoneid := unibase.GetGameZone(uint32(unibase.Atoi(task.R.FormValue("gameid"), 0)), uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0)))
	zoneTask := zoneTaskManager.GetZoneTaskById(gamezoneid)
	if zoneTask == nil {
		task.SendBinary([]byte("server shutdown now"))
		return false
	}
	//WHJ这个临时兼容
	send := &Smd.HttpGmCommandLoginSmd_SC{}
	send.Logintempid = proto.Uint64(task.Id)
	send.Gamezoneid = proto.Uint64(zoneTask.GetId())
	send.Params = proto.String(task.R.FormValue("cmd"))
	zoneTask.SendCmd(send)
	return true
}
