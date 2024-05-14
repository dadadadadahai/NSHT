package main

import (
	"encoding/json"
	"reflect"
	"sort"

	sjson "github.com/bitly/go-simplejson"

	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/config"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	Smd "git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/nettask"
)

type ZoneInfo struct {
	ZoneId   uint32
	ZoneName string
}

type GameInfo struct {
	GameId   uint32
	GameName string
	ZoneList []*ZoneInfo
}

type GMMessage interface {
	proto.Message
	GetClientid() uint64
}

type GMClientTask struct {
	*nettask.ClientNetTask
	Data           *Pmd.GmUserInfo
	GameList       []*GameInfo
	LoginIp        uint32 //如果创建账号的时候设置了绑定IP，则只能通过该IP登陆
	VerifyOk       bool   //验证通过的链接，后继请求才会被处理，否则不作处理
	Password       string //保存密码用来断线重连
	LastChanTaskId uint64 //保存当前请求的http任务ID，以便能够返回数据
	LastSecretKey  string //保存secret key，下次请求的时候验证
}

func ParseReflectCommand(task nettask.NetTaskInterFace, nmd *Pmd.ForwardNullUserPmd_CS) bool {
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
	if _, ok := CallSdkLuaFunc("GmClient."+callname, "GmClient."+defaultname, task.GetId(), rev); !ok {
		call := reflect.ValueOf(task).MethodByName(callname)
		if !call.IsValid() {
			//task.Debug("call GoFunc %s not found", callname)
			call = reflect.ValueOf(task).MethodByName(defaultname)
		}
		if call.IsValid() {
			//task.Info("call GoFunc %s", callname)
			call.Call([]reflect.Value{reflect.ValueOf(rev)})
		} else {
			task.Error("%s not found", callname)
			return false
		}
	}
	return false
}

// 使用首次登陆的http临时ID作为task的ID
func NewGMClientTask(clientId uint64, username, password string, ip uint32) nettask.ClientTaskInterface {
	task := &GMClientTask{ClientNetTask: nettask.NewClientNetTask(ParseReflectCommand, GMCM)}
	task.SetId(clientId)
	task.SetName(username)
	task.VerifyOk = false
	task.Password = password
	task.LoginIp = ip
	task.LastChanTaskId = clientId
	task.GetEntryName = func() string {
		return "GMClientTask(" + task.Name + ")"
	}
	return task
}

func InitGMClientTask(clientId uint64, username, password string, ip uint32) {
	url := config.GetConfigStr("gm_server_url")
	origin := config.GetConfigStr("gm_server_origin")
	task := NewGMClientTask(clientId, username, password, ip)
	task.Dial(url, origin, task, "")
	task.Debug("Dail:%s,%s", url, origin)
}

func (self *GMClientTask) OnDialFail(err error) bool {
	self.Error("Dial to gm server failed")
	return false
}

func (self *GMClientTask) Verify(zoneKey string) {
	compress := config.GetConfigStr("compress_login")
	enc := config.GetConfigStr("encrypt_login")
	enc_key := config.GetConfigStr("encrypt_login_key")

	send := &Pmd.RequestLoginGmUserPmd_C{
		Key:        proto.String(zoneKey),
		Version:    proto.Uint32(uint32(Smd.Config_Version_Gm)),
		Compress:   proto.String(compress),
		Encrypt:    proto.String(enc),
		Encryptkey: proto.String(enc_key),
		Username:   proto.String(self.GetName()),
		Password:   proto.String(self.Password),
		Logintype:  proto.Uint32(1),
		Loginip:    proto.Uint32(self.LoginIp),
	}
	//WHJ 这里只能立即发送并且之后才能设计
	self.SetBlockSending(true)
	go func() {
		self.SendCmdImBlock(send)
		self.SetCompress(compress, 0)
		self.SetEncrypt(enc, enc_key)
	}()
}

func (self *GMClientTask) GetChanHttpTask(clientid uint64) *unibase.ChanHttpTask {
	if clientid == 0 {
		clientid = self.LastChanTaskId
	}
	entry := unibase.VTM.GetEntryById(clientid)
	task, ok := entry.(*unibase.ChanHttpTask)
	if ok == false {
		return nil
	}
	self.LastSecretKey = unibase.Rand.RandString(32)
	SetCookie(task.W, "key", self.LastSecretKey)
	return task
}

func (self *GMClientTask) SendChanHttpTaskBinary(clientid uint64, send []byte) {
	if clientid == 0 {
		if sj, _ := sjson.NewJson(send); sj != nil {
			clientid = sj.Get("clientid").MustUint64()
		}
	}
	task := self.GetChanHttpTask(clientid)
	if task != nil {
		task.SendBinary(send)
	} else {
		self.Warning("can not find ChanHttpTask:%d, data:%s", clientid, string(send))
	}
}

func (self *GMClientTask) SendChanHttpTaskCmd(send GMMessage) {
	data, err := json.Marshal(send)
	if err != nil {
		self.Error("json.Marshal %v error", send)
		return
	}
	self.Debug("SendChanHttpTaskCmd data:%s", unibase.GetProtoString(send.String()))
	self.SendChanHttpTaskBinary(send.GetClientid(), data)
}

type SortedGameInfo []*GameInfo

func (self SortedGameInfo) Len() int {
	return len(self)
}

func (self SortedGameInfo) Less(i, j int) bool {
	return self[i].GameId < self[j].GameId
}

func (self SortedGameInfo) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type SortedZoneInfo []*ZoneInfo

func (self SortedZoneInfo) Len() int {
	return len(self)
}

func (self SortedZoneInfo) Less(i, j int) bool {
	return self[i].ZoneId < self[j].ZoneId
}

func (self SortedZoneInfo) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self *GMClientTask) ParseReturnLoginGmUserPmd_S(rev *Pmd.ReturnLoginGmUserPmd_S) bool {
	if rev.GetRetcode() == 0 {
		self.VerifyOk = true
		self.Data = rev.GetData()
		self.AddToTaskManager()
		self.Info("GM: %s login succeed!", self.Name)
	} else {
		self.RequestClose("verify failed")
		self.Error("GM: %s login failed, error: %d", self.Name, rev.GetRetcode())
	}
	task := self.GetChanHttpTask(rev.GetClientid())
	if task != nil {
		gamelist := make([]*GameInfo, 0)
		for _, info := range rev.GetZoneinfo() {
			gameid := info.GetGameid()
			var gameinfo *GameInfo
			for _, info := range gamelist {
				if info.GameId == gameid {
					gameinfo = info
				}
			}
			if gameinfo == nil {
				gameinfo = &GameInfo{}
				gamelist = append(gamelist, gameinfo)
				gameinfo.GameId = gameid
				gameinfo.GameName = info.GetGamename()
				gameinfo.ZoneList = make([]*ZoneInfo, 0)
			}
			gameinfo.ZoneList = append(gameinfo.ZoneList, &ZoneInfo{info.GetZoneid(), info.GetZonename()})
		}
		self.GameList = gamelist
		sort.Sort(SortedGameInfo(self.GameList))
		for _, gameinfo := range self.GameList {
			sort.Sort(SortedZoneInfo(gameinfo.ZoneList))
		}
		HandleGMLoginSuccess(task, rev)
	} else {
		self.Error("ParseStartUpGameReturnGmPmd_S error, lost the HttpTask: %d", self.GetId())
		self.RequestClose("Lost HttpTask")
	}
	return true
}

func (self *GMClientTask) ParseSelectGamezoneGmUserPmd_SC(rev *Pmd.SelectGamezoneGmUserPmd_SC) bool {
	entry := unibase.VTM.GetEntryById(self.GetId())
	if entry != nil {
		task, ok := entry.(*unibase.ChanHttpTask)
		if ok == false {
			self.Error("ParseSelectGamezoneGmUserPmd_SC type error")
			return false
		}
		b, _ := json.Marshal(rev)
		task.SendBinary(b)
	}
	return true
}

func (self *GMClientTask) AddToTaskManager() {
	oldTask := GMCM.GetGMClientTaskById(self.GetId())
	if oldTask != nil {
		if config.GetConfigStr("debug") != "false" {
			oldTask.Error("被新用户踢下线")
			oldTask.RemoveMe()
			oldTask.RequestClose("被新用户踢下线")
		}
	}
	if GMCM.AddGMClientTask(self) == true {
		self.SetRemoveMeFunc(func() {
			GMCM.RemoveGMClientTask(self)
		})
	}
}

func (self *GMClientTask) ParseReconnectKickoutGmSmd_S(rev *Pmd.ReconnectKickoutGmSmd_S) bool {
	self.RemoveMe()
	return true
}

func (self *GMClientTask) ParseHttpGmCommandLoginSmd_SC(rev *Smd.HttpGmCommandLoginSmd_SC) bool {
	entry := unibase.VTM.GetEntryById(rev.GetLogintempid())
	if entry != nil {
		task, ok := entry.(*unibase.ChanHttpTask)
		if ok == false {
			self.Error("ParseHttpGmCommandLoginSmd_SC type err")
			return false
		}
		task.SendBinary([]byte(rev.GetParams()))
	} else {
		self.Error("login back err:%d,%s", rev.GetLogintempid(), rev.String())
	}
	return true
}

func (self *GMClientTask) ParseRequestExecGmCommandGmPmd_SC(rev *Pmd.RequestExecGmCommandGmPmd_SC) bool {
	sj, _ := sjson.NewJson([]byte(rev.GetMsg()))
	data, _ := sj.Get("data").MarshalJSON()
	self.SendChanHttpTaskBinary(rev.GetClientid(), []byte(data))
	return true
}

func (self *GMClientTask) ParseZoneInfoListLoginUserPmd_S(rev *Pmd.ZoneInfoListLoginUserPmd_S) bool {
	gameid := self.Data.GetGameid()
	if gameid == 0 || gameid == rev.GetGameid() {
		data := &GameInfo{}
		data.GameId = rev.GetGameid()
		data.GameName = rev.GetGamename()
		data.ZoneList = make([]*ZoneInfo, 0)
		zoneid := self.Data.GetZoneid()
		for _, zonedata := range rev.GetZonelist() {
			if zonedata.GetState() == *(Pmd.ZoneState_Normal.Enum()) && (zoneid == 0 || zoneid == zonedata.GetZoneid()) {
				zone := &ZoneInfo{}
				zone.ZoneId = zonedata.GetZoneid()
				zone.ZoneName = zonedata.GetZonename()
				data.ZoneList = append(data.ZoneList, zone)
			}
		}
		exists := false
		for _, gamedata := range self.GameList {
			if data.GameId == gamedata.GameId {
				gamedata.ZoneList = data.ZoneList
				exists = true
				break
			}
		}
		if !exists {
			self.GameList = append(self.GameList, data)
		}
	}
	return true
}

// todo 导出文件
// func (self *GMClientTask) ParseRequestChatMessageGmUserPmd_CS(rev *Pmd.RequestChatMessageGmUserPmd_CS) bool {
// 	self.SendChanHttpTaskCmd(rev)
// 	return true
// }

func (self *GMClientTask) ParseCmdDefault(rev GMMessage) bool {
	self.Debug("ParseCmdDefault data:%s", unibase.GetProtoString(rev.String()))
	self.SendChanHttpTaskCmd(rev)
	return true
}
