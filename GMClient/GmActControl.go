package main

//日期：2017-05-04
//描述：活动控制模块

import (
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
)

func HandleActionControl(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	if gameid == 0 {
		task.SendBinary([]byte(`{"ret":1,retcode":1,"retdesc":"参数错误"}`))
		return
	}
	actid := uint32(unibase.Atoi(task.R.FormValue("actid"), 0))
	actname := strings.TrimSpace(task.R.FormValue("actname"))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	recordid, _ := strconv.ParseUint(task.R.FormValue("recordid"), 10, 64)
	stime, _ := strconv.ParseUint(task.R.FormValue("stime"), 10, 64)
	etime, _ := strconv.ParseUint(task.R.FormValue("etime"), 10, 64)
	packid := uint32(unibase.Atoi(task.R.FormValue("packid"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))

	send := &Pmd.ActionControlGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Optype = proto.Uint32(optype)
	send.Rdata = &Pmd.ActionControlData{}
	send.Rdata.Recordid = proto.Uint64(recordid)
	send.Rdata.Actid = proto.Uint32(actid)
	send.Rdata.Actname = proto.String(actname)
	send.Rdata.Packid = proto.Uint32(packid)
	send.Rdata.Platid = proto.Uint32(platid)
	send.Rdata.State = proto.Uint32(state)
	send.Rdata.Stime = proto.Uint64(stime)
	send.Rdata.Etime = proto.Uint64(etime)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleActionControlSearch(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	if gameid == 0 {
		task.SendBinary([]byte(`{"ret":1,retcode":1,"retdesc":"参数错误"}`))
		return
	}

	recordid, _ := strconv.ParseUint(task.R.FormValue("recordid"), 10, 64)
	stime, _ := strconv.ParseUint(task.R.FormValue("stime"), 10, 64)
	etime, _ := strconv.ParseUint(task.R.FormValue("etime"), 10, 64)
	packid := uint32(unibase.Atoi(task.R.FormValue("packid"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	if etime == 0 {
		etime = uint64(time.Now().Unix())
	}

	send := &Pmd.ActionControlSearchGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Recordid = proto.Uint64(recordid)
	send.Stime = proto.Uint64(stime)
	send.Etime = proto.Uint64(etime)
	send.Packid = proto.Uint32(packid)
	send.Platid = proto.Uint32(platid)
	send.Perpage = proto.Uint32(perpage)
	send.Curpage = proto.Uint32(curpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSystemClassifyControl(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	room := uint32(unibase.Atoi(task.R.FormValue("room"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	systype := uint32(unibase.Atoi(task.R.FormValue("systype"), 0))
	classify := strings.TrimSpace(task.R.FormValue("classify"))

	send := &Pmd.SystemClassifyControlGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Subgameid = proto.Uint32(subgameid)
	send.Room = proto.Uint32(room)
	send.Optype = proto.Uint32(optype)
	send.Systype = proto.Uint32(systype)

	for _, tmpatt := range strings.Split(classify, ",") {
		att := strings.Split(tmpatt, ":")
		if len(att) != 2 {
			continue
		}
		level, _ := strconv.ParseUint(att[0], 10, 64)
		//value, _ := strconv.ParseUint(att[1], 10, 64)
		data := &Pmd.ClassifyData{}
		data.Level = proto.Uint64(level)
		//data.Value = proto.Uint64(value)
		send.Data = append(send.Data, data)
	}

	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchCurrentBetinfo(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	room := uint32(unibase.Atoi(task.R.FormValue("room"), 0))

	send := &Pmd.CurrentBetInfoGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Subgameid = proto.Uint32(subgameid)
	send.Room = proto.Uint32(room)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleExportClassifyProfit(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	room := uint32(unibase.Atoi(task.R.FormValue("room"), 0))

	send := &Pmd.ClassifyProfitExportGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Subgameid = proto.Uint32(subgameid)
	send.Room = proto.Uint32(room)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchRobotList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	room := uint32(unibase.Atoi(task.R.FormValue("room"), 0))

	send := &Pmd.RequestRobotListGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Subgameid = proto.Uint32(subgameid)
	send.Room = proto.Uint32(room)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleOperateRobot(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	room := uint32(unibase.Atoi(task.R.FormValue("room"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))

	recordid, _ := strconv.ParseUint(task.R.FormValue("id"), 10, 64)
	num := uint32(unibase.Atoi(task.R.FormValue("num"), 0))
	stime, _ := strconv.ParseUint(task.R.FormValue("stime"), 10, 64)
	etime, _ := strconv.ParseUint(task.R.FormValue("etime"), 10, 64)
	mincoin, _ := strconv.ParseUint(task.R.FormValue("mincoin"), 10, 64)
	maxcoin, _ := strconv.ParseUint(task.R.FormValue("maxcoin"), 10, 64)

	send := &Pmd.OperateRobotGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Subgameid = proto.Uint32(subgameid)
	send.Room = proto.Uint32(room)
	send.Optype = proto.Uint32(optype)

	send.Data = &Pmd.RobotData{}
	send.Data.Id = proto.Uint64(recordid)
	send.Data.Num = proto.Uint32(num)
	send.Data.Stime = proto.Uint64(stime)
	send.Data.Etime = proto.Uint64(etime)
	send.Data.Mincoin = proto.Uint64(mincoin)
	send.Data.Maxcoin = proto.Uint64(maxcoin)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleRobotControl(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	room := uint32(unibase.Atoi(task.R.FormValue("room"), 0))
	minnum := uint32(unibase.Atoi(task.R.FormValue("minnum"), 0))
	maxnum := uint32(unibase.Atoi(task.R.FormValue("maxnum"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	//betprobalility := uint32(unibase.Atoi(task.R.FormValue("betprobalility"), 0))

	send := &Pmd.RobotBetControlGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Subgameid = proto.Uint32(subgameid)
	send.Optype = proto.Uint32(optype)
	send.Room = proto.Uint32(room)
	send.Minnum = proto.Uint32(minnum)
	send.Maxnum = proto.Uint32(maxnum)
	//send.Betprobalility = proto.Uint32(betprobalility)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleServerListSearch(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	if gameid == 0 {
		task.SendBinary([]byte(`{"ret":1,retcode":1,"retdesc":"参数错误"}`))
		return
	}

	recordid := uint32(unibase.Atoi(task.R.FormValue("recordid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	externalip := strings.TrimSpace(task.R.FormValue("externalip"))
	externalport := uint32(unibase.Atoi(task.R.FormValue("externalport"), 0))
	internalip := strings.TrimSpace(task.R.FormValue("internalip"))
	internalname := strings.TrimSpace(task.R.FormValue("internalname"))
	logicid := uint32(unibase.Atoi(task.R.FormValue("logicid"), 0))
	logicname := strings.TrimSpace(task.R.FormValue("logicname"))
	logicopenstatus := uint32(unibase.Atoi(task.R.FormValue("logicopenstatus"), 0))
	logicspecialflags := uint32(unibase.Atoi(task.R.FormValue("logicspecialflags"), 0))
	order := uint32(unibase.Atoi(task.R.FormValue("order"), 0))
	notice := strings.TrimSpace(task.R.FormValue("notice"))

	send := &Pmd.ServerListGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Optype = proto.Uint32(optype)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Recordid = proto.Uint32(recordid)
	send.Perpage = proto.Uint32(perpage)
	send.Curpage = proto.Uint32(curpage)
	send.ExternalIp = proto.String(externalip)
	send.ExternalPort = proto.Uint32(externalport)
	send.InternalIp = proto.String(internalip)
	send.InternalName = proto.String(internalname)
	send.LogicId = proto.Uint32(logicid)
	send.LogicName = proto.String(logicname)
	send.LogicOpenStatus = proto.Uint32(logicopenstatus)
	send.LogicSpecialFlags = proto.Uint32(logicspecialflags)
	send.Order = proto.Uint32(order)
	send.Notice = proto.String(notice)

	ForwardGmCommand(task, send, 0, 0, false)
}
