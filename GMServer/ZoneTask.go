package main

import (
	"encoding/json"
	"fmt"
	"net"

	//	"net/url"
	"reflect"
	"time"

	sjson "github.com/bitly/go-simplejson"

	"code.google.com/p/go.net/websocket"
	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/config"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	Smd "git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/goroutine"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unisocket"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

var Rank *Pmd.RequestInviteRankGmUserPmd_CS
var RankTime int64

func MonitorLog(format string, v ...interface{}) {
	zoneTaskManager.MonitorLog(format, v...)
}

type ZoneTask struct {
	nettask.NetTaskInterFace
	Version         uint32
	serverState     *Pmd.GameZoneServerState
	_timer_one_min  *unitime.Timer
	_timer_five_min *unitime.Timer
	_clock_zero     *unitime.Clocker
	VerifyOk        bool
}

type GMClientMessage interface {
	proto.Message
	GetGmid() uint32
}

type GmHttpMessage interface {
	proto.Message
	GetGmdata() *Pmd.GmRequestBaseData
}

func ParseReflectCommand(task nettask.NetTaskInterFace, nmd *Pmd.ForwardNullUserPmd_CS) bool {
	defer func() {
		if err := recover(); err != nil {
			task.Error("ParseReflectCommand error:%v, data:%s", err, unibase.GetProtoString(nmd.String()))
		}
	}()
	byCmd, byParam, msgdata := nmd.GetByCmd(), nmd.GetByParam(), nmd.GetData()
	if task.GetTaskVersion() < nettask.TaskVersion {
		alllen := 0
		alllen, msgdata = nettask.GetMsgData(msgdata)
		if msgdata == nil || alllen != len(nmd.GetData()) {
			task.Error("协议解析错误,尝试兼容式自动升级,请对方尽快完善指定协议号:%d,%d,%d,%d", byCmd, byParam, task.GetTaskVersion(), nettask.TaskVersion)
			task.SetTaskVersion(nettask.TaskVersion)
			msgdata = nmd.GetData()
		}
	}
	if byCmd == 0 {
		return nettask.ParseNullUserCmd(task, byParam, msgdata, task.GetForwardfunc())
	}
	revt := reflect.TypeOf(nettask.GetMessageInstance(byCmd, byParam))
	if revt == nil {
		task.Error("ParseReflectCommand type err:%d,%d", byCmd, byParam)
		return false
	}
	msgsname := nettask.GetMessageShotName(byCmd, byParam)
	rev := reflect.New(revt).Interface().(proto.Message)
	err := task.GetCmd(rev, msgdata)
	if err != nil {
		task.Error("ParseReflectCommand err:Parse%s,%s", msgsname, err.Error())
		return false
	}

	callname, defaultname := "Parse"+msgsname, "ParseCmdDefault"
	if _, ok := CallSdkLuaFunc("ZoneTask."+callname, "ZoneTask."+defaultname, task.GetId(), rev); !ok {
		call := reflect.ValueOf(task).MethodByName(callname)
		if !call.IsValid() {
			call = reflect.ValueOf(task).MethodByName(defaultname)
		}
		if call.IsValid() {
			call.Call([]reflect.Value{reflect.ValueOf(rev)})
		} else {
			task.Error("%s not found, data:%s", callname, unibase.GetProtoString(rev.String()))
			return false
		}
	}
	return false
}

func NewZoneTask(ws *websocket.Conn, parsefunc nettask.HandleForwardFunc) *ZoneTask {
	task := &ZoneTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTask(unisocket.NewUniSocketWs(ws, websocket.BinaryFrame), parsefunc, 50*time.Millisecond, zoneTaskManager)
	task.SetEntryName(func() string { return "ZoneTask" })
	task.serverState = &Pmd.GameZoneServerState{
		Gamezone: &Pmd.GameZoneInfo{},
	}
	task._timer_one_min = unitime.NewTimer(unitime.Time.Now(), 60*1000, false)
	task._timer_five_min = unitime.NewTimer(unitime.Time.Now(), 5*60*1000, false)
	task._clock_zero = unitime.NewClocker(unitime.Time.Now(), 0, 24*3600)
	return task
}

func NewZoneTaskBw(conn *net.TCPConn) *ZoneTask {
	task := &ZoneTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTaskBw(unisocket.NewUniSocketBw(conn), ParseReflectCommand /*nettask.ParseBwReflectCommand,*/, 50*time.Millisecond, zoneTaskManager)
	task.SetEntryName(func() string { return "ZoneTaskBw" })
	task.SetTaskVersion(nettask.TaskVersion)
	task.serverState = &Pmd.GameZoneServerState{
		Gamezone: &Pmd.GameZoneInfo{},
	}
	task._timer_one_min = unitime.NewTimer(unitime.Time.Now(), 60*1000, false)
	task._timer_five_min = unitime.NewTimer(unitime.Time.Now(), 5*60*1000, false)
	task._clock_zero = unitime.NewClocker(unitime.Time.Now(), 0, 24*3600)
	return task
}

func (self *ZoneTask) GetGameId() uint32 {
	return unibase.GetGameId(self.GetId())
}

func (self *ZoneTask) GetZoneId() uint32 {
	return unibase.GetZoneId(self.GetId())
}

func (self *ZoneTask) Loop() {
	//self.Debug("%d,%d", unitime.Time.Now().Unix(), unitime.Time.Now().Nanosecond())
	if self._timer_one_min.Check(unitime.Time.Now()) == true {
		self.BroadcastLoop()
	}
	//五分钟导出聊天记录
	if self._timer_five_min.Check(unitime.Time.Now()) == true {
		//self.Info("导出聊天记录, gameid:%d, zoneid:%d, time:%d", self.GetGameId(), self.GetZoneId(), unitime.Time.Sec())
		//ExportChatMsg(self.GetGameId(), self.GetZoneId(), uint32(unitime.Time.YearMonthDay()))
	}
	if self._clock_zero.Check(unitime.Time.Now()) == true {
		self.Info("零点重置")
		//create_user_chat(self.GetGameId(), self.GetZoneId(), uint32(unitime.Time.YearMonthDay()))
		drop_user_chat(self.GetGameId(), self.GetZoneId(), uint32(unitime.Time.YearMonthDay(-3)))
	}
}

func (self *ZoneTask) BroadcastLoop() {
	if _, ok := self.NetTaskInterFace.(*nettask.NetTaskBw); ok {
		retl := checkNextBroadcast(self.GetGameId(), self.GetZoneId())
		for _, data := range retl {
			cmd := &Pmd.BroadcastNewGmUserPmd_C{}
			cmd.Gameid = proto.Uint32(self.GetGameId())
			cmd.Zoneid = proto.Uint32(self.GetZoneId())
			/*
				if self.GetGameId() != 3002 {
					data.Content = proto.String(url.QueryEscape(data.GetContent()))
				}
			*/
			cmd.Data = data
			if data.GetZoneid() == 0 {
				zoneTaskManager.BroadcastToGame(self.GetGameId(), cmd)
			} else {
				self.SendCmd(cmd)
			}

		}
		self.Debug("BroadcastLoop:%d", len(retl))
	}
}

func (self *ZoneTask) ParseStartUpGameRequestGmSmd_C(rev *Smd.StartUpGameRequestGmSmd_C) bool {
	if rev.GetCompress() != "" {
		self.SetCompress(rev.GetCompress(), 0)
	}
	if rev.GetEncrypt() != "" {
		self.SetEncrypt(rev.GetEncrypt(), rev.GetEncryptkey())
	}
	send := &Smd.StartUpGameReturnGmSmd_S{Ret: proto.Bool(false)}
	zone := unibase.CheckZoneInfo(rev.GetKey())
	if rev.GetCompress() != "" {
		self.SetCompress(rev.GetCompress(), 0)
	}
	if rev.GetEncrypt() != "" {
		self.SetEncrypt(rev.GetEncrypt(), rev.GetEncryptkey())
	}
	if rev.GetVersion() != uint32(Smd.Config_Version_Gm) {
		send := &Pmd.CheckVersionUserPmd_CS{
			Versionserver: proto.Uint32(uint32(Smd.Config_Version_Gm)),
			Versionclient: rev.Version,
		}
		self.SendCmd(send)
		self.Error("版本号错误:%d,%d", rev.GetVersion(), uint32(Smd.Config_Version_Gm))
	}
	var oldTask *ZoneTask
	self.Info("login request:%s", unibase.GetProtoString(rev.String()))
	if zone != nil {
		send.Zoneinfo = zone
		self.SetId(unibase.GetGameZone(zone.GetGameid(), zone.GetZoneid()))
		self.SetName(zone.GetGamename() + "-" + zone.GetZonename())
		self.Version = rev.GetVersion()
		self.serverState.Gamezone.Gameid = zone.Gameid
		self.serverState.Gamezone.Zoneid = zone.Zoneid
		self.serverState.Gamezone.Gamename = zone.Gamename
		self.serverState.Gamezone.Zonename = zone.Zonename
		oldTask = zoneTaskManager.GetZoneTaskById(self.GetId())
		if oldTask != nil {
			oldTask.Error("duplicate login:%s,%s,%s", oldTask.GetRemoteAddr(), self.GetRemoteAddr(), unibase.GetProtoString(rev.String()))
			if config.GetConfigStr("debug") != "false" && oldTask.GetRemoteIp() != self.GetRemoteIp() {
				oldTask.Error("被新用户踢下线:%s,%s", oldTask.GetRemoteAddr(), self.GetRemoteAddr())
				if rev.GetLastseq() != 0 {
					self.SetReconnectData(oldTask)
				}
				zoneTaskManager.RemoveZoneTask(oldTask)
				send := &Smd.ReconnectKickoutGmSmd_S{}
				send.Desc = proto.String(fmt.Sprintf("被新用户踢下线:%s,%s", oldTask.GetRemoteAddr(), self.GetRemoteAddr()))
				oldTask.SetBlockSending(true)
				go func() {
					oldTask.SendCmdImBlock(send)
				}()
			}
		}
		if zoneTaskManager.AddZoneTask(self) == true {
			self.SetRemoveMeFunc(func() {
				zoneTaskManager.RemoveZoneTask(self)
				zoneTaskManager.ResetServerState(self.GetGameId())
				serverlist, ok := unibase.GlobalZoneListMap[self.GetGameId()]
				if ok == true {
					gmTaskManager.Broadcast(serverlist)
				}
			})
			send.Ret = proto.Bool(true)
		} else {
			self.Error("duplicate login")
		}
		self.Info("login ok:%d, %s,%s", self.GetId(), self.GetRemoteAddr(), unibase.GetProtoString(rev.String()))
	} else {
		send.Zoneinfo = &Pmd.GameZoneInfo{}
		send.Retdesc = proto.String("未找到区信息:" + rev.GetKey())
		self.Error("login err:%s,%s", self.GetRemoteAddr(), unibase.GetProtoString(rev.String()))
	}
	if rev.GetLastseq() == 0 || self.SendReconnectData(self, rev.GetLastseq()) == false { //这里说明不进行断线重连或者断线重连尝试失败,就当重新登陆
		self.SendCmd(send)
	}

	if goroutine.MonitorLog == nil {
		goroutine.MonitorLog = MonitorLog
	}
	return true
}

func (self *ZoneTask) ParseRequestExecGmCommandGmSmd_SC(rev *Smd.RequestExecGmCommandGmSmd_SC) bool {
	entry := unibase.VTM.GetEntryById(rev.GetGmuserid())
	if entry != nil {
		switch loginTask := entry.(type) {
		case *unibase.ChanHttpTask:
			loginTask.SendBinary([]byte(rev.GetParams()))
		case *GmTask:
			loginTask.SendCmd(rev)
		default:
			self.Error("ParseHttpGmCommandLoginSmd_SC type err")
			return false
		}
	} else {
		self.Error("login back err:%d,%s", rev.GetGmuserid(), rev.String())
	}
	return true
}

func (self *ZoneTask) ParseHttpGmCommandLoginSmd_SC(rev *Smd.HttpGmCommandLoginSmd_SC) bool {
	entry := unibase.VTM.GetEntryById(rev.GetLogintempid())
	if entry != nil {
		switch task := entry.(type) {
		case *unibase.ChanHttpTask:
			task.SendBinary([]byte(rev.GetParams()))
		case *GmTask:
			task.SendCmd(rev)
		default:
			self.Error("ParseHttpGmCommandLoginSmd_SC type err")
			return false
		}
	} else {
		self.Error("login back err:%d,%s", rev.GetLogintempid(), rev.String())
	}
	return true
}

func (self *ZoneTask) ParseRequestExecGmCommandGmPmd_SC(rev *Pmd.RequestExecGmCommandGmPmd_SC) bool {
	defer func() {
		if err := recover(); err != nil {
			self.Error("ParseRequestExecGmCommandGmPmd_SC error:%v", err)
		}
	}()

	sj, _ := sjson.NewJson([]byte(rev.GetMsg()))
	/*
		if sj.Get("do").MustString() == "Pmd.AgencyConditionGmUserPmd_CS" {
			tmpdata := sj.Get("data")
			desc, _ := url.QueryUnescape(tmpdata.Get("desc").MustString())
			tmpdata.Set("desc", desc)
		}
	*/
	data, _ := sj.Get("data").MarshalJSON()
	gmdata, ok := sj.Get("data").CheckGet("gmdata")
	if ok {
		self.Debug("ParseRequestExecGmCommandGmPmd_SC: %s", unibase.GetProtoString(rev.String()))
		clientid := gmdata.Get("clientid").MustUint64()
		entry := unibase.VTM.GetEntryById(clientid)
		if task, ok := entry.(*unibase.ChanHttpTask); ok {
			task.SendBinary([]byte(data))
			return true
		}
		gmid := gmdata.Get("gmid").MustUint64()
		gmtask := gmTaskManager.GetGmTaskById(gmid)
		if gmtask != nil {
			gmtask.SendCmd(rev)
			return true
		}
		self.Error("ParseRequestExecGmCommandGmPmd_SC error: not find task!")
		return false
	}
	self.Debug("ParseRequestExecGmCommandGmPmd_SC: %s", unibase.GetProtoString(rev.String()))
	gmtask := gmTaskManager.GetGmTaskById(uint64(rev.GetGmid()))
	if gmtask != nil {
		gmtask.SendCmd(rev)
		return true
	}
	self.Error("ParseRequestExecGmCommandGmPmd_SC error: not find task!")
	return false
}

func (self *ZoneTask) ParseQueryPackageCodeGmUserPmd_CS(rev *Pmd.QueryPackageCodeGmUserPmd_CS) bool {
	info := CheckoutPackcodeInfo(self.GetGameId(), rev.GetCodeid())
	if info != nil {
		rev.Ret = proto.Uint32(0)
		rev.Flag = info.Flag
		rev.Usedzoneid = info.Uzoneid
		rev.Useduid = info.Ucharid
		rev.Usedtime = proto.Uint64(uint64(info.GetUtime()))
		rev.Createtime = proto.Uint64(uint64(info.GetCreatetime()))
		rev.Codetype = info.Codetype
	} else {
		rev.Ret = proto.Uint32(1)
	}
	self.SendCmd(self.WrapProtoCmd(rev, true))
	return true
}

func (self *ZoneTask) ParseRequestUsePackageCodeGmUserPmd_CS(rev *Pmd.RequestUsePackageCodeGmUserPmd_CS) bool {
	if rev.GetUseduid() == 0 {
		rev.Useduid = proto.Uint64(rev.GetAccid())
	}
	ret, codetype := UsePackageCode(self.GetGameId(), self.GetZoneId(), rev.GetUseduid(), rev.GetCodeid(), rev.GetPlatid(), rev.GetPackageid())
	rev.Ret = proto.Uint32(ret)
	rev.Codetype = proto.Uint32(codetype)
	rev.Items = CheckoutItemsByCodetype(self.GetGameId(), codetype)
	self.Info("UsePackageCode data:%s", unibase.GetProtoString(rev.String()))
	self.SendCmd(self.WrapProtoCmd(rev, true))
	return true
}

func (self *ZoneTask) ParseRequestBroadcastListGmUserPmd_C(rev *Pmd.RequestBroadcastListGmUserPmd_C) bool {
	gameid, zoneid := self.GetGameId(), self.GetZoneId()
	maxpage, data := checkBroadcastList(gameid, zoneid, rev.GetCountryid(), rev.GetSceneid(), rev.GetBtype(), rev.GetEndtime(), rev.GetCurpage(), rev.GetPerpage())
	/*
		if gameid != 3002 {
			for _, tmp := range data {
				tmp.Content = proto.String(url.QueryEscape(tmp.GetContent()))
			}
		}
	*/

	send := &Pmd.ReturnBroadcastListGmUserPmd_S{}
	send.Data = data
	send.Curpage = proto.Uint32(rev.GetCurpage())
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(rev.GetPerpage())
	self.SendCmd(send)
	self.Debug("ParseRequestBroadcastListGmUserPmd_C:%s", unibase.GetProtoString(send.String()))
	return true
}

func (self *ZoneTask) ParseReturnPunishUserGmUserPmd_S(rev *Pmd.ReturnPunishUserGmUserPmd_S) bool {
	fmt.Println("惩罚返回")
	if rev.GetRetcode() != 0 {
		fmt.Println("惩罚失败，删除记录")
		deletePunish(rev.GetTaskid())
	}
	if self.ForwardCommand2Httptask(rev.GetClientid(), rev) {
		return true
	}
	if !self.ForwardGmCommand(rev.GetGmid(), rev) {
		self.Warning("ParseReturnPunishUserGmUserPmd_S cannot find GmTask")
	}
	return true
}

func (self *ZoneTask) ParseFeedbackGmUserPmd_CS(rev *Pmd.FeedbackGmUserPmd_CS) bool {
	saveFeedback(rev.GetFeedbacktype(), rev.GetData())
	self.Debug("ParseFeedbackGmUserPmd_CS %s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseRequestFeedbackListGmUserPmd_C(rev *Pmd.RequestFeedbackListGmUserPmd_C) bool {
	var maxpage uint32
	send := &Pmd.FeedbackGmUserPmd_CS{}
	maxpage, send.Data = checkFeedbackList(rev.GetGameid(), rev.GetZoneid(), rev.GetPlatid(), rev.GetFeedbacktype(), rev.GetState(),
		rev.GetCharid(), rev.GetCharname(), rev.GetStarttime(), rev.GetEndtime(), rev.GetCurpage(), rev.GetPerpage())
	send.Curpage = proto.Uint32(rev.GetCurpage())
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(rev.GetPerpage())
	self.SendCmd(send)
	self.Info("ParseRequestFeedbackListGmUserPmd_C data:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseRequestSendMailGmUserPmd_CS(rev *Pmd.RequestSendMailGmUserPmd_CS) bool {
	if self.ForwardCommand2Httptask(rev.GetClientid(), rev) {
		return true
	}
	if !self.ForwardGmCommand(rev.GetGmid(), rev) {
		self.Warning("ParseRequestSendMailGmUserPmd_CS cannot find GmTask")
	}
	return true
}

func (self *ZoneTask) ParseActionControlSearchGmUserPmd_CS(rev *Pmd.ActionControlSearchGmUserPmd_CS) bool {
	tblname := get_act_control_table()

	var where string
	if rev.GetRecordid() != 0 {
		where = fmt.Sprintf("id=%d", rev.GetRecordid())
	} else {
		var etime uint64 = rev.GetEtime()
		if etime == 0 {
			etime = uint64(time.Now().Unix())
		}
		where = fmt.Sprintf("gameid=%d and (stime<=%d and %d<=etime)", self.GetGameId(), etime, etime)
	}

	str := fmt.Sprintf("select id, packid, platid, actid, actname, state, stime, etime from %s where %s", tblname, where)
	self.Info("Action sql:%s", str)
	rows, err := db_gm.Query(str)
	if err != nil {
		rev.Retcode = proto.Uint32(1)
		rev.Retdesc = proto.String(err.Error())
		self.SendCmd(rev)
		return false
	}
	defer rows.Close()
	rev.Rdata = make([]*Pmd.ActionControlData, 0)
	for rows.Next() {
		data := &Pmd.ActionControlData{}
		if err = rows.Scan(&data.Recordid, &data.Packid, &data.Platid, &data.Actid, &data.Actname, &data.State, &data.Stime, &data.Etime); err != nil {
			continue
		}
		rev.Rdata = append(rev.Rdata, data)
	}
	self.SendCmd(rev)
	return true
}

// func (self *ZoneTask) ParseGmRequestPushMessageUserListGmUserPmd_CS(rev *Pmd.GmRequestPushMessageUserListGmUserPmd_CS) bool {
// 	//todo添加推送接口
// 	//PushMsgToIOSDevice(rev.GetTitle(), rev.GetDesc(), rev.GetData())
// 	apnsm.AddMsg(rev.GetTitle(), rev.GetDesc(), rev.GetData())
// 	if !self.ForwardGmCommand(rev.GetGmid(), rev) {
// 		self.Warning("ParsevGmRequestPushMessageUserListGmUserPmd_CS cannot find GmTask")
// 	}
// 	return true
// }

func (self *ZoneTask) ParseGameRequestPushMessageUserListGmUserPmd_C(rev *Pmd.GameRequestPushMessageUserListGmUserPmd_C) bool {
	//todo 添加推送接口
	self.Debug("GameRequestPushMessageUserListGmUserPmd_C: %s", unibase.GetProtoString(rev.String()))
	//PushMsgToIOSDevice(rev.GetTitle(), rev.GetDesc(), rev.GetData())
	apnsm.AddMsg(rev.GetTitle(), rev.GetDesc(), rev.GetData())
	return true
}

func (self *ZoneTask) ParseRequestSendMailExGmUserPmd_CS(rev *Pmd.RequestSendMailExGmUserPmd_CS) bool {
	updateMailExtState(rev.GetData().GetId(), rev.GetGameid(), rev.GetData().GetZoneid(), rev.GetData().GetCharids())
	return true
}

func (self *ZoneTask) ForwardGmCommand(gmid uint32, rev proto.Message) bool {
	gmtask := gmTaskManager.GetGmTaskById(uint64(gmid))
	if gmtask != nil {
		gmtask.SendCmd(rev)
		return true
	}
	return false
}

var ChatTypeDict = map[uint32]uint32{
	1:  1, //普通聊天
	7:  2, //世界聊天
	15: 3, //帮会聊天
	3:  4, //队伍聊天
	9:  5, //好友聊天，私聊
	27: 6, //喇叭
}

// func (self *ZoneTask) ParseChatMessageGmUserPmd_C(rev *Pmd.ChatMessageGmUserPmd_C) bool {
// 	tblname := get_user_chat_table(self.GetGameId(), self.GetZoneId(), uint32(unitime.Time.YearMonthDay()))
// 	ctime := rev.GetCreatedat()
// 	if ctime == 0 {
// 		ctime = uint64(unitime.Time.Sec())
// 	}
// 	ctype := ChatTypeDict[rev.GetType()]
// 	if ctype == 0 {
// 		ctype = 7
// 	}
// 	if !check_table_exists(tblname) {
// 		create_user_chat(self.GetGameId(), self.GetZoneId(), uint32(unitime.Time.YearMonthDay()))
// 	}
// 	str := fmt.Sprintf("insert into %s(cpid, platid, charid, charname, type, otherid, othername, content, created_at, accid) values(?,?,?,?,?,?,?,?,?,?)", tblname)
// 	_, err := db_gm.Exec(str, rev.GetCpid(), rev.GetPlatid(), rev.GetCharid(), rev.GetCharname(), ctype, rev.GetOtherid(), rev.GetOthername(), rev.GetContent(), ctime, rev.GetAccid())
// 	if err != nil {
// 		self.Error("ParseChatMessageGmUserPmd_C error:%s", err.Error())
// 	}
// 	self.Debug("ParseChatMessageGmUserPmd_C rev:%s", unibase.GetProtoString(rev.String()))
// 	return true
// }

func (self *ZoneTask) ParseCmdDefault(rev proto.Message) bool {
	self.Debug("ParseCmdDefault: %s", unibase.GetProtoString(rev.String()))
	switch cmd := rev.(type) {
	case GmHttpMessage:
		clientid := cmd.GetGmdata().GetClientid()
		self.ForwardCommand2Httptask(clientid, cmd)
		return true
	case GMClientMessage:
		if gmtask := gmTaskManager.GetGmTaskById(uint64(cmd.GetGmid())); gmtask != nil {
			gmtask.SendCmd(cmd)
			return true
		}
	default:
		self.Warning("cannot find GmTask, data:%s", unibase.GetProtoString(rev.String()))
		return false
	}
	return false
}

func (self *ZoneTask) ParseRequestInviteRankGmUserPmd_CS(rev *Pmd.RequestInviteRankGmUserPmd_CS) bool {
	self.Debug("ParseRequestInviteRankGmUserPmd_CS: %d", unitime.Time.Now())
	Rank = rev
	RankTime = unitime.Time.Sec()
	return true
}

/*
func (self *ZoneTask) ParseAgencyConfigSetGmUserPmd_CS(rev *Pmd.AgencyConfigSetGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseAgencyConfigSetGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseAgencyConfigGmUserPmd_CS(rev *Pmd.AgencyConfigGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseAgencyConfigGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseAgencyConditionGmUserPmd_CS(rev *Pmd.AgencyConditionGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseAgencyConditionGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseSearchAgentlistGmUserPmd_CS(rev *Pmd.SearchAgentlistGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseSearchAgentlistGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseAgentManagerGmUserPmd_CS(rev *Pmd.AgentManagerGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseAgentManagerGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseAgentAddGmUserPmd_CS(rev *Pmd.AgentAddGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseAgentAddGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseSearchProxyRechargeGmUserPmd_CS(rev *Pmd.SearchProxyRechargeGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseSearchProxyRechargeGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseSearchRechargeConfigGmUserPmd_CS(rev *Pmd.SearchRechargeConfigGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseSearchRechargeConfigGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseRechargeManagerGmUserPmd_CS(rev *Pmd.RechargeManagerGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseRechargeManagerGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseSearchRechargeGmUserPmd_CS(rev *Pmd.SearchRechargeGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseSearchRechargeGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseDrawcashListGmUserPmd_CS(rev *Pmd.DrawcashListGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseSearchRechargeGmUserPmd_CS", clientid, rev)
	return true
}

func (self *ZoneTask) ParseDrawcashManagerGmUserPmd_CS(rev *Pmd.DrawcashManagerGmUserPmd_CS) bool {
	clientid := rev.GetGmdata().GetClientid()
	self.ForwardCommand2Httptask("ParseSearchRechargeGmUserPmd_CS", clientid, rev)
	return true
}
*/

func (self *ZoneTask) ForwardCommand2Httptask(clientid uint64, rev proto.Message) bool {
	if clientid <= 0 {
		return false
	}
	entry := unibase.VTM.GetEntryById(clientid)
	if task, ok := entry.(*unibase.ChanHttpTask); ok {
		data, _ := json.Marshal(rev)
		task.SendBinary(data)
		return true
	}
	return false
}

func (self *ZoneTask) ParseCPaylistGmUserPmd_CS(rev *Pmd.CPaylistGmUserPmd_CS) bool {
	tblname := get_cpay_table()
	str := fmt.Sprintf("select packid,platid,payplatid,minlevel,maxlevel,minmoney,maxmoney,stime,etime,state from %s where gameid=? and state!=0", tblname)
	rows, err := db_gm.Query(str, self.GetGameId())
	self.Debug("CPaylist sql:%s", str)
	if err != nil {
		self.Error("ParseCPaylistGmUserPmd_CS query error:%s", err.Error())
		self.Debug("CPaylist data:%s", unibase.GetProtoString(rev.String()))
		self.SendCmd(rev)
		return false
	}
	rev.Data = make([]*Pmd.CPayData, 0)
	for rows.Next() {
		data := &Pmd.CPayData{}
		if err := rows.Scan(&data.Packageid, &data.Platid, &data.Payplatid, &data.Minlevel, &data.Maxlevel, &data.Minmoney, &data.Maxmoney, &data.Stime, &data.Etime, &data.State); err != nil {
			self.Error("ParseCPaylistGmUserPmd_CS scan error:%s", err.Error())
			continue
		}
		rev.Data = append(rev.Data, data)
	}
	rows.Close()
	self.Debug("CPaylist data:%s", unibase.GetProtoString(rev.String()))
	self.SendCmd(rev)
	return true
}

func (self *ZoneTask) ParseCPayGmUserPmd_CS(rev *Pmd.CPayGmUserPmd_CS) bool {
	if !self.ForwardGmCommand(rev.GetGmid(), rev) {
		self.Warning("ParseCPayGmUserPmd_CS cannot find GmTask")
	}
	return true
}

func (self *ZoneTask) WrapProtoCmd(send proto.Message, wrap bool) (sendwrap proto.Message) {
	if wrap && !zoneTaskManager.CheckZoneTaskBwPort(self.GetGameId()) {
		data, _ := json.Marshal(send)
		sj := sjson.New()
		sj.Set("data", string(data))
		sj.Set("do", reflect.TypeOf(send).String()[1:])
		data, _ = sj.MarshalJSON()
		send2 := &Pmd.RequestExecGmCommandGmPmd_SC{}
		send2.Gameid = proto.Uint32(self.GetGameId())
		send2.Zoneid = proto.Uint32(self.GetZoneId())
		send2.Gmid = proto.Uint32(uint32(self.GetId()))
		send2.Msg = proto.String(string(data))
		sendwrap = send2
	} else {
		sendwrap = send
	}
	return
}
