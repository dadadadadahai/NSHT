package main

import (
	"net"
	"time"

	"code.google.com/p/go.net/websocket"
	"code.google.com/p/goprotobuf/proto"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	Smd "git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unisocket"
)

var (
	loggerTaskTempId = 10000
)

type LoginTask struct {
	nettask.NetTaskInterFace
	Version     uint32
	serverState *Pmd.GameZoneServerState
	VerifyOk    bool
}

func NewLoginTask(ws *websocket.Conn, parsefunc nettask.HandleForwardFunc) *LoginTask {
	task := &LoginTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTask(unisocket.NewUniSocketWs(ws, websocket.BinaryFrame), parsefunc, 50*time.Millisecond, gameTaskManager)
	task.SetEntryName(func() string { return "LoginTask" })
	task.serverState = &Pmd.GameZoneServerState{
		Gamezone: &Pmd.GameZoneInfo{},
	}
	return task
}

func NewLoginTaskBw(conn *net.TCPConn) *LoginTask {
	task := &LoginTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTaskBw(unisocket.NewUniSocketBw(conn), nettask.ParseBwReflectCommand, 50*time.Millisecond, gameTaskManager)
	task.SetEntryName(func() string { return "LoginTaskBw" })
	task.SetTaskVersion(nettask.TaskVersion)
	task.serverState = &Pmd.GameZoneServerState{
		Gamezone: &Pmd.GameZoneInfo{},
	}
	return task
}

func (self *LoginTask) GetGameId() uint32 {
	return unibase.GetGameId(self.GetId())
}

func (self *LoginTask) GetZoneId() uint32 {
	return unibase.GetZoneId(self.GetId())
}

func (self *LoginTask) ParseStartUpGameRequestMonitorSmd_C(rev *Smd.StartUpGameRequestMonitorSmd_C) bool {
	if rev.GetCompress() != "" {
		self.SetCompress(rev.GetCompress(), 0)
	}
	if rev.GetEncrypt() != "" {
		self.SetEncrypt(rev.GetEncrypt(), rev.GetEncryptkey())
	}
	loggerTaskTempId++
	self.SetId(uint64(loggerTaskTempId))
	self.SetName(self.GetRemoteAddr())
	send := &Smd.StartUpGameReturnMonitorSmd_S{Zoneinfo: &Pmd.GameZoneInfo{}, Ret: proto.Bool(true)}
	self.SendCmd(send)
	self.Info("ParseStartUpGameRequestMonitorSmd_C: %s", unibase.GetProtoString(rev.String()))
	return true
}

// 由于loginserver只在账号首次登陆才发这个协议，所以统计账号登陆相关的需要在角色登陆的地方更新
func (self *LoginTask) ParseUserLoginOkFromLoingServerMonitorSmd_C(rev *Smd.UserLoginOkFromLoingServerMonitorSmd_C) bool {
	// tblname := get_user_login_table(rev.GetGameid(), uint32(unitime.Time.YearMonthDay()))
	// if !check_table_exists(tblname) {
	// 	create_user_login(rev.GetGameid(), uint32(unitime.Time.YearMonthDay()))
	// }
	// str := fmt.Sprintf("insert into %s (zoneid,accountid,accountname) values(?,?,?)", tblname)
	// _, err := db_monitor.Exec(str, rev.GetZoneid(), rev.GetAccountid(), rev.GetAccountname())
	// if err != nil && strings.Split(err.Error(), ":")[0] != "Error 1062" {
	// 	self.Error("ParseUserLoginOkFromLoingServerMonitorSmd_C insert err:%s", err.Error())
	// }
	// tblname = get_user_imei_table(rev.GetGameid())
	// if !check_table_exists(tblname) {
	// 	create_user_imei(rev.GetGameid())
	// }
	// str = fmt.Sprintf("insert ignore into %s (zoneid,accountid,accountname,imei,osname) values(?,?,?,?,?)", tblname)
	// _, err = db_monitor.Exec(str, rev.GetZoneid(), rev.GetAccountid(), rev.GetAccountname(), rev.GetImei(), rev.GetOsname())
	// if err != nil && strings.Split(err.Error(), ":")[0] != "Error 1062" {
	// 	self.Error("ParseUserLoginOkFromLoingServerMonitorSmd_C insert err:%s", err.Error())
	// }
	// var mobile = ""
	// LoginAccount(rev.GetGameid(), rev.GetAccountid(), rev.GetAccountname(), int(rev.GetPlatid()), rev.GetAdcode(),, rev.GetIp(), rev.GetImei(), 0, 0, mobile)
	// self.Info("ParseUserLoginOkFromLoingServerMonitorSmd_C :%s", unibase.GetProtoString(rev.String()))
	return true
}
