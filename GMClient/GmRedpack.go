package main

//日期：2017-05-04
//描述：麻将红包兑换码接口

import (
	"strings"

	"github.com/golang/protobuf/proto"

	"git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
)

func HandleRedpackCodeSearch(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	code := strings.TrimSpace(task.R.FormValue("code"))
	if code == "" || gameid == 0 {
		task.Error("Search redpackcode gameid:%d, code:%s", gameid, code)
		task.SendBinary([]byte(`{"ret":1,retcode":1,"retdesc":"参数错误"}`))
		return
	}

	send := &Pmd.RedpackCodeSearchGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Code = proto.String(code)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleRedpackCodeOperate(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	code := strings.TrimSpace(task.R.FormValue("code"))
	state := int32(unibase.Atoi(task.R.FormValue("state"), 0))
	if code == "" || gameid == 0 || state != 0 {
		task.Error("Operate redpackcode gameid:%d, code:%s, state:%d", gameid, code, state)
		task.SendBinary([]byte(`{"ret":1,retcode":1,"retdesc":"参数错误"}`))
		return
	}

	send := &Pmd.RedPackCodeOperateGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Code = proto.String(code)
	send.State = proto.Int32(1)
	ForwardGmCommand(task, send, 0, 0, false)
}
