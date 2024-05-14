package main

import (
	"encoding/json"
	"fmt"

	sjson "github.com/bitly/go-simplejson"

	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/nettask"
)

var jsmsgAPIMap map[string]GmHandlerFunc = map[string]GmHandlerFunc{
	"PunishUserGmUserPmd_C":             ParsePunishUserGmUserPmd_C,
	"RequestSendMailGmUserPmd_CS":       ParseRequestSendMailGmUserPmd_CS,
	"RequestUsePackageCodeGmUserPmd_CS": ParseRequestUsePackageCodeGmUserPmd_CS,
}

func HandleGmApi(task *unibase.ChanHttpTask) bool {
	defer func() {
		if err := recover(); err != nil {
			task.Error("HandleGmApi error:%v, url:%s, body:%s", err, task.R.RequestURI, task.Rawdata)
		}
	}()
	defer task.R.Body.Close()
	rawdata := task.Rawdata
	if rawdata == nil || len(rawdata) == 0 {
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
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
	sj.Set("clientid", task.GetId())
	cmd := task.R.FormValue("do")
	ip := Ip2Int(GetRemoteIp(task.R))
	if gofun, ok := jsmsgAPIMap[cmd]; ok {
		gofun(cmd, sj, task, ip)
	} else {
		ForwardGmApiCommand(cmd, task, sj, true)
	}
	return true
}

func ForwardGmApiCommand(protoname string, task *unibase.ChanHttpTask, sj *sjson.Json, wrap bool) {
	gameid, zoneid := uint32(sj.Get("gameid").MustInt()), uint32(sj.Get("zoneid").MustInt())
	if gameid > 0 {
		if send := ParseProtoMsg(protoname, task, sj); send != nil {
			SendProtoMsg(protoname, gameid, zoneid, send)
		}
	}
}

func ParseProtoMsg(protoname string, task *unibase.ChanHttpTask, sj *sjson.Json) proto.Message {
	data, err := sj.MarshalJSON()
	if err != nil {
		task.Debug("ParseProtoMsg MarshalJSON error:%s, proto:%s", err.Error(), protoname)
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return nil
	}
	send := nettask.GetCmdByFullName("*Pmd."+protoname, data)
	if send == nil {
		task.Debug("ParseProtoMsg GetCmdByFullName proto:%s", protoname)
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return nil
	}
	return send
}

func SendProtoMsg(protoname string, gameid, zoneid uint32, send proto.Message) {
	if zoneid == 0 {
		zoneTaskManager.BroadcastToGame(gameid, send)
		logging.Debug("SendProtoMsg:%s, %s", protoname, unibase.GetProtoString(send.String()))
	} else {
		zonetask := zoneTaskManager.GetZoneTaskById(unibase.GetGameZone(gameid, zoneid))
		if zonetask == nil {
			logging.Error("SendProtoMsg:%s, %s, not found zonetask", protoname, unibase.GetProtoString(send.String()))
			return
		}
		logging.Debug("SendProtoMsg:%s, %s", protoname, unibase.GetProtoString(send.String()))
		zonetask.SendCmd(send)
	}
}

func ParsePunishUserGmUserPmd_C(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	fmt.Println("api")
	gameid, zoneid, gmid := uint32(sj.Get("gameid").MustInt()), uint32(sj.Get("zoneid").MustInt()), uint32(sj.Get("gmid").MustInt())
	if send, ok := ParseProtoMsg(protoname, task, sj).(*Pmd.PunishUserGmUserPmd_C); ok {
		savePunish(send.GetData(), gmid)
		//send.Clientid = proto.Uint64(task.GetId())
		SendProtoMsg(protoname, gameid, zoneid, send)
	}
}

func ParseRequestSendMailGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	gameid, zoneid, gmid := uint32(sj.Get("gameid").MustInt()), uint32(sj.Get("zoneid").MustInt()), uint32(sj.Get("gmid").MustInt())
	if send, ok := ParseProtoMsg(protoname, task, sj).(*Pmd.RequestSendMailGmUserPmd_CS); ok {
		send.Data.Id = proto.Uint64(saveGmMail(send.GetData(), gmid))
		//send.Clientid = proto.Uint64(task.GetId())
		SendProtoMsg(protoname, gameid, zoneid, send)
	}
}
func ParseRequestUsePackageCodeGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {

	gameid, zoneid := uint32(sj.Get("usedgameid").MustInt()), uint32(sj.Get("usedzoneid").MustInt())
	useduid, accid, codeid, platid, packageid := uint64(sj.Get("useduid").MustInt()), uint32(sj.Get("accid").MustInt()), sj.Get("codeid").MustString(), uint32(sj.Get("platid").MustInt()), uint32(sj.Get("packageid").MustInt())
	if send, ok := ParseProtoMsg(protoname, task, sj).(*Pmd.RequestUsePackageCodeGmUserPmd_CS); ok {
		if useduid == 0 {
			useduid = uint64(accid)
		}
		ret, codetype := UsePackageCode(gameid, zoneid, useduid, codeid, platid, packageid)
		logging.Error("ret:%d ,codetype:%d", ret, codetype)
		send.Ret = proto.Uint32(ret)
		send.Codetype = proto.Uint32(codetype)
		send.Items = CheckoutItemsByCodetype(gameid, codetype)

		//		self.Info("UsePackageCode data:%s", unibase.GetProtoString(rev.String()))
		//send.Clientid = proto.Uint64(task.GetId())
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		//task.SendCmd(0,send)
		//SendProtoMsg(protoname, gameid, zoneid, send)
	}
}
