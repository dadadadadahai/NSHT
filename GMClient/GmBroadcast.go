package main

import (
	"encoding/json"
	"strings"

	"github.com/golang/protobuf/proto"

	"git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

func HandleBroadcastShutdown(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	all := uint32(unibase.Atoi(task.R.FormValue("allzone"), 0))
	now := uint32(unitime.Time.Sec())
	if all > 0 {
		zoneid = 0
	}
	content := strings.TrimSpace(task.R.FormValue("content"))
	send := &Pmd.ServerShutDownLoginUserPmd_S{}
	send.Servertime = proto.Uint32(etime)
	if now >= etime {
		send.Retcode = proto.Uint32(1)
		send.Retdesc = proto.String("停机时间错误")
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	send.Lefttime = proto.Uint32(etime - now)
	send.Desc = proto.String(content)
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleBroadcastAdd(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	btype := uint32(unibase.Atoi(task.R.FormValue("btype"), 0))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	itime := uint32(unibase.Atoi(task.R.FormValue("intervaltime"), 0))
	all := uint32(unibase.Atoi(task.R.FormValue("allzone"), 0))
	title := strings.TrimSpace(task.R.FormValue("title"))
	if all > 0 {
		zoneid = 0
	}
	content := strings.TrimSpace(task.R.FormValue("content"))
	send := &Pmd.BroadcastNewGmUserPmd_C{}
	send.Data = &Pmd.BroadcastInfo{}
	send.Data.Gameid = proto.Uint32(gameid)
	send.Data.Zoneid = proto.Uint32(zoneid)
	send.Data.Starttime = proto.Uint32(stime)
	send.Data.Endtime = proto.Uint32(etime)
	send.Data.Intervaltime = proto.Uint32(itime * 60)
	send.Data.Btype = proto.Uint32(btype)
	send.Data.Content = proto.String(content)
	send.Data.Title = proto.String(title)
	send.Data.Sceneid = proto.Uint32(0)
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleBroadcastList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	endtime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	send := &Pmd.RequestBroadcastListGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Endtime = proto.Uint32(endtime)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleBroadcastDelete(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	taskid := uint32(unibase.Atoi(task.R.FormValue("taskid"), 0))
	send := &Pmd.BroadcastDeleteGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Taskid = proto.Uint32(taskid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}
