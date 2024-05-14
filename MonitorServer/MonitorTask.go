package main

import (
	"net"
	"time"

	"code.google.com/p/go.net/websocket"
	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unisocket"
)

type GameMap map[uint32]string

type MonitorTask struct {
	nettask.NetTaskInterFace
	VerifyOk bool
	gameMap  GameMap //保存想监控的区服信息
}

func NewMonitorTask(ws *websocket.Conn, parsefunc nettask.HandleForwardFunc) *MonitorTask {
	task := &MonitorTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTask(unisocket.NewUniSocketWs(ws, websocket.BinaryFrame), parsefunc, 50*time.Millisecond, monitorTaskManager)
	task.SetEntryName(func() string { return "MonitorTask" })
	return task
}
func NewMonitorTaskJson(ws *websocket.Conn, parsefunc nettask.HandleForwardFunc) *MonitorTask {
	task := &MonitorTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTaskJson(unisocket.NewUniSocketWs(ws, websocket.TextFrame), parsefunc, 50*time.Millisecond, monitorTaskManager)
	task.SetEntryName(func() string { return "MonitorTaskJson" })
	return task
}
func NewMonitorTaskBw(conn *net.TCPConn) *MonitorTask {
	task := &MonitorTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTaskBw(unisocket.NewUniSocketBw(conn), nettask.ParseBwReflectCommand, 50*time.Millisecond, monitorTaskManager)
	task.SetEntryName(func() string { return "MonitorTaskBw" })
	task.SetTaskVersion(nettask.TaskVersion)
	return task
}
func (self *MonitorTask) RefreshServerStateList() bool {
	send := &Pmd.RefreshServerStateListMonitorPmd_S{}

	callback := func(v entry.EntryInterface) bool {
		task := v.(*ZoneTask)
		//WHJ 如果有指定游戏区，则必须匹配
		if len(self.gameMap) != 0 {
			if _, ok := self.gameMap[task.GetGameId()]; ok == false {
				return true
			}
		}
		send.Statelist = append(send.Statelist, task.serverState)
		return true
	}
	gameTaskManager.ExecEvery(callback)
	self.SendCmd(send)
	return true
}

func (self *MonitorTask) ParseStartUpGameRequestMonitorPmd_C(rev *Pmd.StartUpGameRequestMonitorPmd_C) bool {
	if rev.GetCompress() != "" {
		self.SetCompress(rev.GetCompress(), 0)
	}
	if rev.GetEncrypt() != "" {
		self.SetEncrypt(rev.GetEncrypt(), rev.GetEncryptkey())
	}
	send := &Pmd.StartUpGameReturnMonitorPmd_S{
		Ret: proto.Bool(true),
	}
	self.SetId(monitor_task_tempid)
	monitor_task_tempid++
	if monitorTaskManager.AddMonitorTask(self) == true {
		self.SetRemoveMeFunc(func() {
			monitorTaskManager.RemoveMonitorTask(self)
		})
		send.Ret = proto.Bool(true)
	} else {
		self.Error("duplicate login")
	}
	self.Info("login ok:%s,%s", self.GetRemoteAddr(), unibase.GetProtoString(rev.String()))
	self.SendCmd(send)
	return true
}

func (self *MonitorTask) ParseSupportGameZoneListSdkPmd_C(rev *Pmd.SupportGameZoneListSdkPmd_C) bool {
	for _, gamezone := range rev.GetGamezonelist() {
		self.gameMap[gamezone.GetGameid()] = gamezone.GetGamename()
	}
	self.RefreshServerStateList()
	return true
}
