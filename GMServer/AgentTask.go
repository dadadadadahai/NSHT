package main

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"

	"code.google.com/p/goprotobuf/proto"
	sjson "github.com/bitly/go-simplejson"

	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/nettask"
)

type GmHandlerFunc func(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, ip uint32)

var jsmsgMainMap map[string]GmHandlerFunc = map[string]GmHandlerFunc{
	"RequestLoginGmUserPmd_C":          ParseRequestLoginGmUserPmd_C,
	"AgencyConditionGmUserPmd_CS":      ParseAgencyConditionGmUserPmd_CS,
	"SearchAgentlistGmUserPmd_CS":      ParseSearchAgentlistGmUserPmd_CS,
	"AgentManagerGmUserPmd_CS":         ParseAgentManagerGmUserPmd_CS,
	"SearchProxyRechargeGmUserPmd_CS":  ParseSearchProxyRechargeGmUserPmd_CS,
	"SearchRechargeConfigGmUserPmd_CS": ParseSearchRechargeConfigGmUserPmd_CS,
	"RechargeManagerGmUserPmd_CS":      ParseRechargeManagerGmUserPmd_CS,
	"SearchRechargeGmUserPmd_CS":       ParseSearchRechargeGmUserPmd_CS,
	"AgentAddGmUserPmd_CS":             ParseAgentAddGmUserPmd_CS,
	"DrawcashListGmUserPmd_CS":         ParseDrawcashListGmUserPmd_CS,
	"DrawcashManagerGmUserPmd_CS":      ParseDrawcashManagerGmUserPmd_CS,
	"AgencyConfigSetGmUserPmd_CS":      ParseAgencyConfigSetGmUserPmd_CS,
	"AgencyConfigGmUserPmd_CS":         ParseAgencyConfigGmUserPmd_CS,
	"AgentStockGmUserPmd_CS":           ParseAgentStockGmUserPmd_CS,
	"AgentStockModGmUserPmd_CS":        ParseAgentStockModGmUserPmd_CS,
	"AgentRechargeGmUserPmd_CS":        ParseAgentRechargeGmUserPmd_CS,
	"AgentSearchGmUserPmd_CS":          ParseAgentSearchGmUserPmd_CS,
	//"AgentAddGmUserPmd_CS":				ParseAgentAddGmUserPmd_CS,
}

func ParseRequestLoginGmUserPmd_C(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	account, passwd := sj.Get("username").MustString(), sj.Get("password").MustString()
	result, data := checkAccount(account, passwd, loginip)
	if result != 0 {
		task.SendBinary([]byte(`{"retcode":1, "retdesc":"登陆失败,请输入正确的账号和密码"}`))
		return
	}
	gmid := strconv.Itoa(int(data.GetGmid()))
	SetSecureCookie(task.W, "gmid", gmid)
	SetCookie(task.W, "name", data.GetName())
	SessionFlush(gmid)
	SessionSet(gmid, "login", "1")
	SessionSet(gmid, "name", data.GetName())
	task.SendBinary([]byte(`{"retcode":0,"retdesc":"登陆成功","redirecturl":"/games"}`))
}

func ParseAgencyConditionGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseSearchAgentlistGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseAgentManagerGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseSearchProxyRechargeGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseSearchRechargeConfigGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseRechargeManagerGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseSearchRechargeGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseAgentAddGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseDrawcashListGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseDrawcashManagerGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseAgencyConfigSetGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseAgencyConfigGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseAgentStockGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseAgentStockModGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseAgentSearchGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ParseAgentRechargeGmUserPmd_CS(protoname string, sj *sjson.Json, task *unibase.ChanHttpTask, loginip uint32) {
	ForwardGmCommand(protoname, task, sj, true)
}

func ForwardGmCommand(protoname string, task *unibase.ChanHttpTask, sj *sjson.Json, wrap bool) {
	data, err := sj.MarshalJSON()
	if err != nil {
		task.Debug("ForwardGmCommand MarshalJSON error:%s, proto:%s", err.Error(), protoname)
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return
	}
	send := nettask.GetCmdByFullName("*Pmd."+protoname, data)
	if send == nil {
		task.Debug("ForwardGmCommand GetCmdByFullName proto:%s", protoname)
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return
	}
	if protoname == "AgencyConditionGmUserPmd_CS" {
		// tmp := send.(*Pmd.AgencyConditionGmUserPmd_CS)
		// data := sj.Get("data").MustArray()
		// for i, t := range data {
		// 	td := t.(map[string]interface{})
		// 	tmp.Data[i].Desc = proto.String(url.QueryEscape(td["desc"].(string)))
		// }
		// send = tmp
	} else if protoname == "AgencyConfigSetGmUserPmd_CS" {
		tmp := send.(*Pmd.AgencyConfigSetGmUserPmd_CS)
		data := sj.Get("data").MustMap()
		tmp.Data.Code = proto.String(url.QueryEscape(data["code"].(string)))
		tmp.Data.Name = proto.String(url.QueryEscape(data["name"].(string)))
		tmp.Data.Desc = proto.String(url.QueryEscape(data["desc"].(string)))
		tmp.Data.Ext = proto.String(url.QueryEscape(data["ext"].(string)))
		send = tmp
	}
	var sendwrap proto.Message
	gameid, zoneid, gmid := uint32(sj.Get("gmdata").Get("gameid").MustInt()), uint32(sj.Get("gmdata").Get("zoneid").MustInt()), uint32(sj.Get("gmdata").Get("gmid").MustInt())
	if wrap {
		data, _ := json.Marshal(send)
		sj := sjson.New()
		sj.Set("data", string(data))
		sj.Set("do", reflect.TypeOf(send).String()[1:])
		data, _ = sj.MarshalJSON()
		send2 := &Pmd.RequestExecGmCommandGmPmd_SC{}
		send2.Gameid = proto.Uint32(gameid)
		send2.Zoneid = proto.Uint32(zoneid)
		send2.Gmid = proto.Uint32(uint32(gmid))
		send2.Msg = proto.String(string(data))
		sendwrap = send2
	} else {
		sendwrap = send
	}
	if zoneid == 0 {
		zoneTaskManager.BroadcastToGame(gameid, sendwrap)
		logging.Debug("ForwardGmCommand:%s, %s", protoname, unibase.GetProtoString(sendwrap.String()))
	} else {
		zonetask := zoneTaskManager.GetZoneTaskById(unibase.GetGameZone(gameid, zoneid))
		if zonetask == nil {
			logging.Error("ForwardGmCommand:%s, %s, not found zonetask", protoname, unibase.GetProtoString(sendwrap.String()))
			return
		}
		logging.Debug("ForwardGmCommand:%s, %s", protoname, unibase.GetProtoString(sendwrap.String()))
		zonetask.SendCmd(sendwrap)
	}
}
