package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/nettask"
)

type MonitorClientTask struct {
	*nettask.ClientNetTask
	VerifyOk bool
}

var notice_time_map map[uint64]int64 = make(map[uint64]int64)

func NewMonitorClientTask() nettask.ClientTaskInterface {
	task := &MonitorClientTask{ClientNetTask: nettask.NewClientNetTask(nettask.ParseReflectCommand, MTCM)}
	task.SetEntryName(func() string { return "MonitorClientTask" })
	return task
}

func InitMonitorClient() bool {
	url := config.GetConfigStr("monitor_server_url")
	origin := config.GetConfigStr("monitor_server_origin")
	if !(url != "" && origin != "") {
		return false
	}
	task := NewMonitorClientTask()
	task.Dial(url, origin, task, "")
	task.Debug("Dail:%s,%s", url, origin)
	return true
}

func (self *MonitorClientTask) OnDialFail(err error) bool {
	self.Error("Dial to monitor server failed")
	return false
}

func (self *MonitorClientTask) Verify(zoneKey string) {
	enc := config.GetConfigStr("encrypt_login")
	compress := config.GetConfigStr("compress_login")
	enc_key := config.GetConfigStr("encrypt_login_key")
	send := &Pmd.StartUpGameRequestMonitorPmd_C{
		Key:        proto.String(zoneKey),
		Version:    proto.Uint32(uint32(Smd.Config_Version_Monitor)),
		Compress:   proto.String(compress),
		Encrypt:    proto.String(enc),
		Encryptkey: proto.String(enc_key),
	}
	self.SetBlockSending(true)
	go func() {
		self.SendCmdImBlock(send)
		self.SetCompress(compress, 1)
		self.SetEncrypt(enc, enc_key)
	}()
}

func (self *MonitorClientTask) ParseStartUpGameReturnMonitorPmd_S(rev *Pmd.StartUpGameReturnMonitorPmd_S) bool {
	self.Debug("ParseStartUpGameReturnMonitorPmd_S retcode:%t, retdesc:%s", rev.GetRet(), rev.GetRetdesc())
	self.VerifyOk = true
	return true
}

func (self *MonitorClientTask) ParseZoneInfoListLoginUserPmd_S(rev *Pmd.ZoneInfoListLoginUserPmd_S) bool {
	unibase.GlobalZoneListMap[rev.GetGameid()] = rev
	return true
}

func (self *MonitorClientTask) ParseRefreshServerStateMonitorPmd_CSC(rev *Pmd.RefreshServerStateMonitorPmd_CSC) bool {
	return true
}

func (self *MonitorClientTask) ParseStErrorLogMonitorUserCmd_S(rev *Pmd.StErrorLogMonitorUserCmd_S) bool {
	now := time.Now().Unix()
	gameid := rev.GetGamezone().GetGameid()
	zoneid := rev.GetGamezone().GetZoneid()
	ts, ok := notice_time_map[unibase.GetGameZone(gameid, zoneid)]
	if ok && ts+120 > now {
		self.Warning("Gameid:%d-Zoneid:%d have already send message in two minute", gameid, zoneid)
		return true
	}
	notice_time_map[unibase.GetGameZone(gameid, zoneid)] = now
	msg := fmt.Sprintf("[Game:%s(%d)-Zone:%s(%d)-IP:%s]-%s", rev.GetGamezone().GetGamename(), gameid, rev.GetGamezone().GetZonename(), zoneid, rev.GetRemoteaddr(), rev.GetLogger())
	/*
		if sendCustomMessage(msg) != nil {
			sendMassMessage(msg)
		}
	*/
	if config.GetConfigStr("sendsmmsg") == "true" {
		sendSMMessage(msg, config.GetConfigStr(fmt.Sprintf("game%d", rev.GetGamezone().GetGameid())))
	}
	return true
}

func (self *MonitorClientTask) ParseForwardOnlineNumMonitorSmd_S(rev *Smd.ForwardOnlineNumMonitorSmd_S) bool {
	//self.Debug("ParseForwardOnlineNumMonitorSmd_S data:%s", unibase.GetProtoString(rev.String()))
	if config.GetConfigStr("sendwarning") != "true" {
		return true
	}
	gamezone := rev.GetGamezone()
	gameid, zoneid := gamezone.GetGameid(), gamezone.GetZoneid()
	gameidlist := strings.Split(config.GetConfigStr("gameidlist"), ",")
	if len(gameidlist) == 0 {
		return true
	}
	//检查需要预警的游戏ID
	warning := false
	for _, tmpgameid := range gameidlist {
		if tmpgameid == strconv.Itoa(int(gameid)) {
			warning = true
			break
		}
	}
	if !warning {
		return true
	}
	now := time.Now().Unix()
	ts, ok := notice_time_map[unibase.GetGameZone(gameid, zoneid)]
	if ok && ts+600 > now {
		return true
	}
	tblname := get_user_data_table(gameid)
	str := fmt.Sprintf("select count(distinct accid) from %s where zoneid=%d", tblname, zoneid)
	row := db_monitor.QueryRow(str)
	var accnum int64
	row.Scan(&accnum)

	if rev.GetOnlinenum() >= uint32(config.GetConfigInt("onlinelimit")) || accnum >= config.GetConfigInt64("accountlimit") {
		msg := fmt.Sprintf(config.GetConfigStr("msgcontent"), gamezone.GetGamename(), gamezone.GetZonename(), gamezone.GetZoneid(), accnum, rev.GetOnlinenum())
		sendSMMessage(msg, config.GetConfigStr("firephones"))
		notice_time_map[unibase.GetGameZone(gameid, zoneid)] = now
	}

	return true
}
