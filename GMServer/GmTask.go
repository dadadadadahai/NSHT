package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	//"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	sjson "github.com/bitly/go-simplejson"

	"code.google.com/p/go.net/websocket"
	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	Smd "git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/entry"

	//"git.code4.in/mobilegameserver/unibase/luaginx"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unisocket"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type GmTask struct {
	nettask.NetTaskInterFace
	VerifyOk  bool
	Data      *Pmd.GmUserInfo //保存GM账号相关信息
	CurGameId uint32
	CurZoneId uint32
}

type GmTaskMessage interface {
	proto.Message
	GetGameid() uint32
	GetZoneid() uint32
}

func NewGmTask(ws *websocket.Conn, parsefunc nettask.HandleForwardFunc) *GmTask {
	task := &GmTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTask(unisocket.NewUniSocketWs(ws, websocket.BinaryFrame), parsefunc, 50*time.Millisecond, gmTaskManager)
	task.SetEntryName(func() string { return "GmTask" })
	return task
}

func NewGmTaskJson(ws *websocket.Conn, parsefunc nettask.HandleForwardFunc) *GmTask {
	task := &GmTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTaskJson(unisocket.NewUniSocketWs(ws, websocket.TextFrame), parsefunc, 50*time.Millisecond, gmTaskManager)
	task.SetEntryName(func() string { return "GmTaskJson" })
	return task
}

func NewGmTaskBw(conn *net.TCPConn) *GmTask {
	task := &GmTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTaskBw(unisocket.NewUniSocketBw(conn), nettask.ParseBwReflectCommand, 50*time.Millisecond, gmTaskManager)
	task.SetEntryName(func() string { return "GmTaskBw" })
	task.SetTaskVersion(nettask.TaskVersion)
	return task
}

func ParseReflectGmTaskCommand(task nettask.NetTaskInterFace, nmd *Pmd.ForwardNullUserPmd_CS) bool {
	defer func() {
		if err := recover(); err != nil {
			task.Error("ParseReflectGmTaskCommand error:%v, data:%s", err, unibase.GetProtoString(nmd.String()))
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
	if _, ok := CallSdkLuaFunc("GmTask."+callname, "GmTask."+defaultname, task.GetId(), rev); !ok {
		call := reflect.ValueOf(task).MethodByName(callname)
		if !call.IsValid() {
			call = reflect.ValueOf(task).MethodByName(defaultname)
		}
		if call.IsValid() {
			call.Call([]reflect.Value{reflect.ValueOf(rev)})
		} else {
			task.Error("%s not found", callname)
			return false
		}
	}
	return true
}

func (self *GmTask) ParseRequestLoginGmUserPmd_C(rev *Pmd.RequestLoginGmUserPmd_C) bool {
	if rev.GetCompress() != "" {
		self.SetCompress(rev.GetCompress(), 0)
	}
	if rev.GetEncrypt() != "" {
		self.SetEncrypt(rev.GetEncrypt(), rev.GetEncryptkey())
	}
	var loginIp uint32 = 0
	if rev.GetLogintype() == 1 {
		loginIp = rev.GetLoginip()
	} else {
		loginIp = self.GetRemoteIp()
	}
	self.VerifyLogin(rev.GetUsername(), rev.GetPassword(), loginIp)
	return true
}

func (self *GmTask) VerifyLogin(username, password string, loginIp uint32) {
	retcode, data := checkAccount(username, password, loginIp)
	send := &Pmd.ReturnLoginGmUserPmd_S{}
	send.Retcode = proto.Uint32(retcode)
	if retcode == 0 {
		send.Data = data
		self.Data = data
		self.VerifyOk = true
		self.SetId(uint64(data.GetGmid()))
		self.SetName(username)
		self.SetEntryName(func() string { return "GmTask(" + self.GetName() + ")" })
		oldtask := gmTaskManager.GetGmTaskById(self.GetId())
		if oldtask != nil {
			oldtask.Error("duplicate login:%s,%s,%s", oldtask.GetRemoteAddr(), self.GetRemoteAddr(), username)
			oldtask.Error("被新用户踢下线")
			gmTaskManager.RemoveGmTask(oldtask)
			send := &Pmd.ReconnectKickoutGmSmd_S{}
			send.Desc = proto.String(fmt.Sprintf("被新用户踢下线:%s,%s", oldtask.GetRemoteAddr(), self.GetRemoteAddr()))
			oldtask.SetBlockSending(true)
			go func() {
				oldtask.SendCmdImBlock(send)
			}()
		}
		if gmTaskManager.AddGmTask(self) == true {
			self.SetRemoveMeFunc(func() {
				gmTaskManager.RemoveGmTask(self)
			})
		}
	}
	send.Zoneinfo = self.GetGameZone(data.GetGameid(), data.GetZoneid())
	self.SendCmd(send)
	self.Info("ParseStartUpGameRequestGmPmd_C  gm:%s retcode:%d", username, send.GetRetcode())
}

func InSlice(e uint32, s []uint32) bool {
	for _, eid := range s {
		if e == eid {
			return true
		}
	}
	return false
}

func (self *GmTask) GetGameZone(gameid, zoneid uint32) (zonelist []*Pmd.GameZoneInfo) {
	zonelist = make([]*Pmd.GameZoneInfo, 0)
	tmpzone, gameids, _ := get_gamezone_by_accid(uint32(self.GetId()))
	gameids = append(gameids, gameid)

	callback := func(v entry.EntryInterface) bool {
		task, ok := v.(*ZoneTask)
		if ok {
			tgameid := task.serverState.GetGamezone().GetGameid()
			tzoneid := task.serverState.GetGamezone().GetZoneid()
			if ((gameid == 0 && tgameid >= 9000 && tgameid < 10000) || gameid == tgameid || InSlice(tgameid, gameids)) &&
				(zoneid == 0 || zoneid == tzoneid) {
				zonelist = append(zonelist, task.serverState.Gamezone)
			}
		}
		return true
	}
	zoneTaskManager.ExecEvery(callback)

	if zonelist == nil || len(zonelist) == 0 {
		zonelist = tmpzone
	} else {
		for _, zone := range tmpzone {
			exists := false
			for _, existszone := range zonelist {
				if zone.GetGameid() == existszone.GetGameid() && zone.GetZoneid() == existszone.GetZoneid() {
					exists = true
					break
				}
			}
			if !exists {
				zonelist = append(zonelist, zone)
			}
		}
	}
	return
}

func (self *GmTask) ParseSelectGamezoneGmUserPmd_SC(rev *Pmd.SelectGamezoneGmUserPmd_SC) bool {
	if !self.VerifyOk {
		return false
	}
	self.CurGameId = rev.GetGameid()
	self.CurZoneId = rev.GetZoneid()
	rev.Retcode = proto.Uint32(0)
	self.SendCmd(rev)
	return true
}

func (self *GmTask) ParseRequestModifyUserInfoGmUserPmd_C(rev *Pmd.RequestModifyUserInfoGmUserPmd_C) bool {
	charids := strings.Split(rev.GetCharids(), ",")

	for _, v := range charids {
		a, _ := strconv.ParseUint(v, 10, 64)

		rev.Charid = proto.Uint64(a)

		self.Info("gm:%s ParseRequestModifyUserInfoGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
		self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ReturnModifyUserInfoGmUserPmd_S{})
		tblname := get_user_mail_table()
		now := unitime.Time.Sec()
		var subject_str string
		if rev.GetOptype() == 3 {
			subject_str = "vip"
		} else if rev.GetOptype() == 7 {
			subject_str = "金币"
		} else if rev.GetOptype() == 22 {
			subject_str = "银币"
		} else if rev.GetOptype() == 15 {
			subject_str = "真实姓名"
		} else if rev.GetOptype() == 16 {
			subject_str = "CPF信息"
		} else if rev.GetOptype() == 17 {
			subject_str = "触发免费游戏"
		} else if rev.GetOptype() == 18 {
			subject_str = "触发爆池"
		} else if rev.GetOptype() == 19 {
			subject_str = "触发bonus"
		} else if rev.GetOptype() == 20 {
			subject_str = "邮箱"
		} else if rev.GetOptype() == 21 {
			subject_str = "电话"
		} else if rev.GetOptype() == 22 {
			subject_str = "银币"
		} else if rev.GetOptype() == 23 {
			subject_str = "累计充值金额"
		} else if rev.GetOptype() == 24 {
			subject_str = "提现通道"
		} else if rev.GetOptype() == 25 {
			subject_str = "清理玩家账号"
		}
		str := fmt.Sprintf("insert into %s(gameid,zoneid,type,gmid,charid,subject,content,attachment,recordtime,gold,goldbind,money,ext) values(?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
		_, err := db_gm.Exec(str, rev.GetGameid(), rev.GetZoneid(), 2, self.GetId(), rev.GetCharid(), "GM修改玩家属性", subject_str, rev.GetOpnum(), now, 0, 0, 0, 0)
		if err != nil {
			logging.Error("saveGmMailEx error:%s", err.Error())

		}
		var content = ""
		if rev.GetOptype() == 3 || rev.GetOptype() == 7 {
			content = fmt.Sprintf("修改玩家%d 的 %s 数值为 %d ", rev.GetCharid(), subject_str, rev.GetOpnum())
		} else if rev.GetOptype() == 17 || rev.GetOptype() == 18 || rev.GetOptype() == 19 {
			content = fmt.Sprintf("添加玩家%d 受控，方式为 %s ", rev.GetCharid(), subject_str)
		} else if rev.GetOptype() == 25 {
			content = fmt.Sprintf("%s , 玩家id为 %d ", subject_str, rev.GetCharid())
		} else {
			content = fmt.Sprintf("修改玩家%d 的 %s 数值为 %s ", rev.GetCharid(), subject_str, rev.GetExt())
		}

		add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)

	}

	return true
}
func (self *GmTask) ParseStRequestConvertVerifyPmd_CS(rev *Pmd.StRequestConvertVerifyPmd_CS) bool {

	if rev.GetOptype() == 1 {
		self.Info("gm:%s StRequestConvertVerifyPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
		self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.StRequestConvertVerifyPmd_CS{})

	} else {
		charids := strings.Split(rev.GetCharids(), ",")

		for _, v := range charids {

			a, _ := strconv.ParseUint(v, 10, 64)
			if a != 0 {

				rev.Orderid = proto.Uint64(uint64(a))

				self.Info("gm:%s StRequestConvertVerifyPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
				self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.StRequestConvertVerifyPmd_CS{})
				var subject_str = ""

				if rev.GetOpvalue() == 0 {
					subject_str = "同意"
				} else if rev.GetOpvalue() == 1 {
					subject_str = "拒绝"
				}
				content := fmt.Sprintf("对订单%d的兑现审核，方式为%s", rev.GetOrderid(), subject_str)

				add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
			}
		}

	}

	return true

}
func (self *GmTask) ParseStRequestSlotsGameParamPmd_CS(rev *Pmd.StRequestSlotsGameParamPmd_CS) bool {
	self.Info("gm:%s StRequestConvertVerifyPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.StRequestSlotsGameParamPmd_CS{})

	if rev.GetOptype() == 2 {

		content := fmt.Sprintf("修改游戏配置")
		add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	}
	return true
}
func (self *GmTask) ParseStRequestHundredGameParamPmd_CS(rev *Pmd.StRequestHundredGameParamPmd_CS) bool {
	self.Info("gm:%s StRequestConvertVerifyPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.StRequestHundredGameParamPmd_CS{})

	if rev.GetOptype() == 2 {

		content := fmt.Sprintf("修改游戏配置")
		add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	}
	return true
}

func (self *GmTask) ParseRequestUserRecordGmUserPmd_C(rev *Pmd.RequestUserRecordGmUserPmd_C) bool {
	self.Info("gm:%s ParseRequestUserRecordGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ReturnUserRecordGmUserPmd_S{})
	return true
}

// 广播、公告
func (self *GmTask) ParseBroadcastNewGmUserPmd_C(rev *Pmd.BroadcastNewGmUserPmd_C) bool {
	self.Info("gm:%s ParseBroadcastNewGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	saveBroadcast(rev.GetData(), uint32(self.GetId()))

	content := "发送公告" + *rev.GetData().Title
	add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)

	//if rev.GetData().GetGameid() == 3002 && rev.GetData().GetStarttime() > uint32(unitime.Time.Sec()) {
	if rev.GetData().GetStarttime() > uint32(unitime.Time.Sec()) {
		send := &Pmd.ReturnBroadcastNewGmUserPmd_S{}
		send.Retcode = proto.Uint32(0)
		send.Taskid = proto.Uint32(rev.GetData().GetTaskid())
		send.Gmid = proto.Uint32(uint32(self.GetId()))
		self.SendCmd(send)
		return true
	}
	/*
		if rev.GetData().GetGameid() != 3002 {
			rev.Data.Content = proto.String(url.QueryEscape(rev.GetData().GetContent()))
		}
	*/
	self.ForwardGmCommand(rev, rev.GetData().GetGameid(), rev.GetData().GetZoneid(), true, &Pmd.ReturnBroadcastNewGmUserPmd_S{})

	return true
}

// 广播、公告
func (self *GmTask) ParseServerShutDownLoginUserPmd_S(rev *Pmd.ServerShutDownLoginUserPmd_S) bool {
	self.Info("gm:%s ParseServerShutDownLoginUserPmd_S data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	saveShutdownBroadcast(rev.GetGameid(), rev.GetZoneid(), rev.GetGmid(), rev.GetServertime(), rev.GetLefttime(), rev.GetDesc())
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), false, &Pmd.ServerShutDownLoginUserPmd_S{})
	return true
}

// 删除公告,如果传来的zoneid为0，则表示全服都的执行该请求
func (self *GmTask) ParseBroadcastDeleteGmUserPmd_C(rev *Pmd.BroadcastDeleteGmUserPmd_C) bool {
	self.Info("gm:%s ParseBroadcastDeleteGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	updateBroadcastState(rev.GetTaskid(), 2)
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ReturnBroadcastDeleteGmUserPmd_S{})
	content := "删除公告"
	add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	return true
}

// 公告列表请求
func (self *GmTask) ParseRequestBroadcastListGmUserPmd_C(rev *Pmd.RequestBroadcastListGmUserPmd_C) bool {
	self.Info("gm:%s ParseRequestBroadcastListGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	maxpage, data := checkBroadcastList(rev.GetGameid(), rev.GetZoneid(), rev.GetCountryid(), rev.GetSceneid(), rev.GetBtype(), rev.GetEndtime(), rev.GetCurpage(), rev.GetPerpage())
	send := &Pmd.ReturnBroadcastListGmUserPmd_S{}
	send.Data = data
	send.Curpage = proto.Uint32(rev.GetCurpage())
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(rev.GetPerpage())
	self.SendCmd(send)
	return true
}

func (self *GmTask) ParsePunishUserGmUserPmd_C(rev *Pmd.PunishUserGmUserPmd_C) bool {

	var charids = strings.Split(rev.GetCharids(), ",")

	for _, v := range charids {
		a, _ := strconv.ParseUint(v, 10, 64)

		rev.Data.Charid = proto.Uint64(a)

		self.Info("gm:%s ParsePunishUserGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))

		b := getMaxid()
		rev.Data.Taskid = proto.Uint32(b)
		if self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ReturnPunishUserGmUserPmd_S{}) {

			self.Info("gm:%s ParsePunishUserGmUserPmd_C data:%s, success", self.GetName(), unibase.GetProtoString(rev.String()))
			savePunish(rev.GetData(), uint32(self.GetId()))

		}
		var subject_str = ""
		switch *rev.GetData().Ptype {

		case 1:
			subject_str = "禁言"
		case 2:
			subject_str = "踢下线"
		case 3:
			subject_str = "封号"
		case 4:
			subject_str = "点控"
		case 5:
			subject_str = "追踪"
		case 6:
			subject_str = "限制ip"
		case 7:
			subject_str = "限制机器码"

		}
		var content string
		if *rev.GetData().Ptype < 5 {
			content = fmt.Sprintf("添加对玩家%d的惩罚,方式为%s", rev.GetData().GetCharid(), subject_str)
		} else {
			content = fmt.Sprintf("%s:%s", subject_str, rev.GetData().GetPunishvalue())
		}

		add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)

	}

	return true
}

func (self *GmTask) ParseRechargeRewardGmUserPmd_CS(rev *Pmd.RechargeRewardGmUserPmd_CS) bool {

	self.Info("gm:%s ParseServerShutDownLoginUserPmd_S data:%s", self.GetName(), unibase.GetProtoString(rev.String()))

	if rev.GetOptype() == 2 {
		tblname := get_exclusive_rewards_table()
		str := fmt.Sprintf("insert into %s(gameid, zoneid, content, created_at) values (?,?,?,?)", tblname)
		result, err := db_gm.Exec(str, rev.GetGameid(), rev.GetZoneid(), rev.GetInterval(), unitime.Time.Sec())
		if err == nil {
			_, err = result.LastInsertId()
		}
	}

	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.RechargeRewardGmUserPmd_CS{})
	return true
}

func (self *GmTask) ParseRechargeRewardLogGmUserPmd_CS(rev *Pmd.RechargeRewardLogGmUserPmd_CS) bool {
	self.Info("gm:%s ParseRechargeRewardLogGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	if !self.VerifyOk {
		return false
	}
	maxpage, data := checkrechargerewardlog(rev.GetGameid(), rev.GetZoneid(), rev.GetCurpage(), rev.GetPerpage())
	send := &Pmd.RechargeRewardLogGmUserPmd_CS{}
	send.Data = data
	send.Perpage = proto.Uint32(rev.GetPerpage())
	send.Maxpage = proto.Uint32(maxpage)
	send.Curpage = proto.Uint32(rev.GetCurpage())
	self.SendCmd(send)
	return true
}

func (self *GmTask) ParseDeletePunishUserGmUserPmd_C(rev *Pmd.DeletePunishUserGmUserPmd_C) bool {
	self.Info("gm:%s ParseDeletePunishUserGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	if self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ReturnDeletePunishUserGmUserPmd_S{}) {
		self.Info("gm:%s ParseDeletePunishUserGmUserPmd_C data:%s sucess", self.GetName(), unibase.GetProtoString(rev.String()))
		updatePunishState(rev.GetTaskid(), 3)
	}
	var subject_str = ""
	switch rev.GetPtype() {

	case 1:
		subject_str = "禁言"
	case 2:
		subject_str = "踢下线"
	case 3:
		subject_str = "封号"
	case 4:
		subject_str = "点控"
	case 5:
		subject_str = "追踪"
	case 6:
		subject_str = "限制ip"
	case 7:
		subject_str = "限制机器码"

	}
	content := fmt.Sprintf("删除对玩家%d的%s惩罚", rev.GetCharid(), subject_str)
	add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	return true
}
func (self *GmTask) ParseRequestGetUserItemList_C(rev *Pmd.RequestGetUserItemList_C) bool {
	self.Info("gm:%s ParseRequestGetUserItemList_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ReturnGetUserItemList_S{})
	return true
}

func (self *GmTask) ParseModifyUserItem_C(rev *Pmd.ModifyUserItem_C) bool {
	self.Info("gm:%s ParseReModifyUserItem_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ModifyUserItem_C{})
	content := "修改角色背包道具信息"
	add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	return true
}
func (self *GmTask) ParseEmptyUserItem_C(rev *Pmd.EmptyUserItem_C) bool {
	self.Info("gm:%s ParseReEmptyUserItem_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.EmptyUserItem_C{})
	content := "清空角色背包道具信息"
	add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	return true
}
func (self *GmTask) ParseRequestUserInfoGmUserPmd_C(rev *Pmd.RequestUserInfoGmUserPmd_C) bool {
	self.Info("gm:%s ParseRequestUserInfoGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ReturnUserInfoGmUserPmd_S{})
	return true
}

func (self *GmTask) ParseRequestPunishListGmUserPmd_C(rev *Pmd.RequestPunishListGmUserPmd_C) bool {
	self.Info("gm:%s ParseRequestPunishListGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	if !self.VerifyOk {
		return false
	}
	maxpage, data := checkPunishList(rev.GetGameid(), rev.GetZoneid(), rev.GetCharid(), rev.GetPid(), rev.GetCurpage(), rev.GetPerpage(), rev.GetPtype(), rev.GetState(), rev.GetStarttime(), rev.GetEndtime(), rev.GetPunishvalue())
	send := &Pmd.ReturnPunishListGmUserPmd_S{}
	send.Data = data
	send.Perpage = proto.Uint32(rev.GetPerpage())
	send.Maxpage = proto.Uint32(maxpage)
	send.Curpage = proto.Uint32(rev.GetCurpage())
	self.SendCmd(send)
	return true
}

func (self *GmTask) ParseRequestExecGmCommandGmSmd_SC(rev *Smd.RequestExecGmCommandGmSmd_SC) bool {
	if !self.VerifyOk {
		return false
	}
	zoneTask := zoneTaskManager.GetZoneTaskById(rev.GetGamezoneid())
	rev.Gmuserid = proto.Uint64(self.GetId())
	self.Info("gm:%s ParseRequestExecGmCommandGmSmd_SC data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	if zoneTask != nil {
		zoneTask.SendCmd(rev)
	}
	return true
}

func (self *GmTask) ParseRequestFeedbackListGmUserPmd_C(rev *Pmd.RequestFeedbackListGmUserPmd_C) bool {
	gamekey, _ := get_game_gmlink_and_key(rev.GetGameid(), rev.GetZoneid())
	if gamekey != "" {
		rev.Gmid = proto.Uint32(uint32(self.GetId()))
		self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), false, &Pmd.FeedbackGmUserPmd_CS{})
		return true
	}
	var maxpage uint32
	send := &Pmd.FeedbackGmUserPmd_CS{}
	maxpage, send.Data = checkFeedbackList(rev.GetGameid(), rev.GetZoneid(), rev.GetPlatid(), rev.GetFeedbacktype(), rev.GetState(),
		rev.GetCharid(), rev.GetCharname(), rev.GetStarttime(), rev.GetEndtime(), rev.GetCurpage(), rev.GetPerpage())
	send.Curpage = proto.Uint32(rev.GetCurpage())
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(rev.GetPerpage())
	self.SendCmd(send)
	self.Info("gm:%s ParseRequestFeedbackListGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	return true
}

func (self *GmTask) ParseHttpGmCommandLoginSmd_SC(rev *Smd.HttpGmCommandLoginSmd_SC) bool {
	gamezoneid := rev.GetGamezoneid()
	zoneTask := zoneTaskManager.GetZoneTaskById(gamezoneid)
	if zoneTask == nil {
		rev.Params = proto.String("server shut down")
		self.SendCmd(rev)
		return false
	}
	zoneTask.SendCmd(rev)
	return true
}

// func (self *GmTask) ParseRequestMailRecordGmUserPmd_CS(rev *Pmd.RequestMailRecordGmUserPmd_CS) bool {
// 	self.Info("gm:%s ParseRequestMailRecordGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
// 	rev.Data = checkoutmails(rev.GetGameid(), rev.GetZoneid(), uint32(rev.GetCharid()), rev.GetRecvid(), rev.GetStarttime(), rev.GetEndtime())
// 	self.SendCmd(rev)
// 	return true
// }

// func (self *GmTask) ParseModBlackWhitelistGmUserPmd_CS(rev *Pmd.ModBlackWhitelistGmUserPmd_CS) bool {
// 	updateBlackWhiteList(rev.GetData().GetId(), uint32(self.GetId()), rev.GetData())
// 	self.Info("gm:%s ParseModBlackWhitelistGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
// 	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ModBlackWhitelistGmUserPmd_CS{})
// 	return true
// }

func (self *GmTask) ParseRequestSendMailGmUserPmd_CS(rev *Pmd.RequestSendMailGmUserPmd_CS) bool {
	rev.Data.Id = proto.Uint64(saveGmMail(rev.GetData(), uint32(self.GetId())))
	self.Info("gm:%s ParseRequestSendMailGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.RequestSendMailGmUserPmd_CS{})
	content := "发送邮件"
	add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	return true
}

func (self *GmTask) ParseActionControlGmUserPmd_CS(rev *Pmd.ActionControlGmUserPmd_CS) bool {
	tblname := get_act_control_table()
	data := rev.GetRdata()
	if data.GetRecordid() == 0 {
		str := fmt.Sprintf("insert into %s(gameid, packid, platid, actid, actname, state, stime, etime, gmid, created) values(?,?,?,?,?,?,?,?,?,?)", tblname)
		result, err := db_gm.Exec(str, rev.GetGameid(), data.GetPackid(), data.GetPlatid(), data.GetActid(), data.GetActname(), data.GetState(), data.GetStime(), data.GetEtime(), rev.GetGmid(), time.Now().Unix())
		var recordid int64
		if err == nil {
			recordid, err = result.LastInsertId()
		}
		if err != nil {
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String(err.Error())
			self.SendCmd(rev)
			return false
		}
		rev.Rdata.Recordid = proto.Uint64(uint64(recordid))
	} else {
		str := fmt.Sprintf("update %s set packid=?, platid=?, actid=?, actname=?, state=?, stime=?, etime=? where id=?", tblname)
		_, err := db_gm.Exec(str, data.GetPackid(), data.GetPlatid(), data.GetActid(), data.GetActname(), data.GetState(), data.GetStime(), data.GetEtime(), data.GetRecordid())
		if err != nil {
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String(err.Error())
			self.SendCmd(rev)
			return false
		}
	}
	self.Info("gm:%s ParseActionControlGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	self.ForwardGmCommand(rev, rev.GetGameid(), 0, true, &Pmd.ActionControlGmUserPmd_CS{})
	return true
}

func (self *GmTask) ParseActionControlSearchGmUserPmd_CS(rev *Pmd.ActionControlSearchGmUserPmd_CS) bool {
	tblname := get_act_control_table()

	var where string
	if rev.GetRecordid() != 0 {
		where = fmt.Sprintf("id=%d", rev.GetRecordid())
	} else {
		var etime uint64 = rev.GetEtime()
		if etime == 0 {
			etime = uint64(time.Now().Unix())
		}
		where = fmt.Sprintf("gameid=%d and (stime<=%d and %d<=etime)", rev.GetGameid(), etime, etime)
	}

	str := fmt.Sprintf("select id, packid, platid, actid, actname, state, stime, etime from %s where %s", tblname, where)
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

func (self *GmTask) ParseRequestSendMailExGmUserPmd_CS(rev *Pmd.RequestSendMailExGmUserPmd_CS) bool {
	extdata := strings.TrimSpace(rev.GetExtdata())
	rev.Extdata = proto.String("")
	rev.Data.Id = proto.Uint64(saveGmMailEx(rev.GetData(), uint32(self.GetId()), extdata))
	self.Info("gm:%s ParseRequestSendMailExGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	if extdata != "" {
		if strings.Contains(extdata, "_") {
			senddict := make(map[int]*Pmd.GmMailInfoEx)
			user_zones := strings.Split(extdata, ",")
			for _, user_zone := range user_zones {
				userid_zoneid := strings.SplitN(strings.TrimSpace(user_zone), "_", 2)
				if len(userid_zoneid) != 2 {
					continue
				}
				userid, err := strconv.ParseUint(userid_zoneid[0], 10, 64)
				if err != nil {
					self.Error("ParseRequestSendMailExGmUserPmd_CS error:%s, %v", err.Error(), userid_zoneid)
					continue
				}
				zoneid, err := strconv.Atoi(userid_zoneid[1])
				if err != nil {
					self.Error("ParseRequestSendMailExGmUserPmd_CS error:%s, %v", err.Error(), userid_zoneid)
					continue
				}
				mailinfo := senddict[zoneid]
				if mailinfo == nil {
					mailinfo = &Pmd.GmMailInfoEx{}
					mailinfo.Id = proto.Uint64(rev.Data.GetId())
					mailinfo.Gameid = proto.Uint32(rev.Data.GetGameid())
					mailinfo.Zoneid = proto.Uint32(uint32(zoneid))
					mailinfo.Type = proto.Uint32(rev.Data.GetType())
					mailinfo.Gmid = proto.Uint32(rev.Data.GetGmid())
					mailinfo.Minlevel = proto.Uint32(rev.Data.GetMinlevel())
					mailinfo.Maxlevel = proto.Uint32(rev.Data.GetMaxlevel())
					mailinfo.Minvip = proto.Uint32(rev.Data.GetMinvip())
					mailinfo.Maxvip = proto.Uint32(rev.Data.GetMaxvip())
					mailinfo.Minlogin = proto.Uint32(rev.Data.GetMinlogin())
					mailinfo.Maxlogin = proto.Uint32(rev.Data.GetMaxlogin())
					mailinfo.Online = proto.Uint32(rev.Data.GetOnline())
					mailinfo.Subject = rev.Data.Subject
					mailinfo.Content = rev.Data.Content
					mailinfo.Attachment = rev.Data.Attachment
					mailinfo.Recordtime = rev.Data.Recordtime
					mailinfo.Gold = rev.Data.Gold
					mailinfo.Goldbind = rev.Data.Goldbind
					mailinfo.Money = rev.Data.Money
					senddict[zoneid] = mailinfo
				}
				mailinfo.Charids = append(mailinfo.Charids, uint64(userid))
			}
			for zoneid, mailinfo := range senddict {
				rev.Zoneid = proto.Uint32(uint32(zoneid))
				rev.Data = mailinfo
				self.ForwardGmCommand(rev, rev.GetGameid(), uint32(zoneid), true, &Pmd.RequestSendMailExGmUserPmd_CS{})
				saveMailExtCharids(rev.Data.GetId(), rev.GetGameid(), mailinfo.GetZoneid(), mailinfo.GetCharids())
			}
		} else {
			zoneids := strings.Split(extdata, ",")
			zoneid_list := make([]uint32, 0)
			for _, tmpzoneid := range zoneids {
				zoneid, err := strconv.Atoi(strings.TrimSpace(tmpzoneid))
				if err != nil {
					self.Error("ParseRequestSendMailExGmUserPmd_CS error:%s, %s", err.Error(), tmpzoneid)
					continue
				}
				rev.Zoneid = proto.Uint32(uint32(zoneid))
				rev.Data.Zoneid = proto.Uint32(uint32(zoneid))
				self.ForwardGmCommand(rev, rev.GetGameid(), uint32(zoneid), true, &Pmd.RequestSendMailExGmUserPmd_CS{})
				zoneid_list = append(zoneid_list, uint32(zoneid))
			}
			saveMailExtZoneids(rev.Data.GetId(), rev.GetGameid(), zoneid_list)
		}
	}
	rev.Retcode = proto.Uint32(0)
	rev.Retdesc = proto.String("发送完毕！(成功与否需要后继查看记录)")
	content := "批量发送邮件"
	add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	self.SendCmd(rev)
	return true
}

func (self *GmTask) ParseCPayGmUserPmd_CS(rev *Pmd.CPayGmUserPmd_CS) bool {
	self.Info("gm:%s ParseCPayGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	data := rev.GetData()
	recordid := data.GetRecordid()
	now := uint32(unitime.Time.Sec())
	tblname := get_cpay_table()
	if recordid == 0 {
		str := fmt.Sprintf("insert into %s(gameid,gmid,packid,platid,payplatid,minlevel,maxlevel,minmoney,maxmoney,stime,etime,state,created_at) values(?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
		result, err := db_gm.Exec(str, rev.GetGameid(), self.GetId(), data.GetPackageid(), data.GetPlatid(), data.GetPayplatid(), data.GetMinlevel(), data.GetMaxlevel(), data.GetMinmoney(), data.GetMaxmoney(), data.GetStime(), data.GetEtime(), data.GetState(), now)
		if err == nil {
			lastid, _ := result.LastInsertId()
			data.Recordid = proto.Uint64(uint64(lastid))
		}
	} else {
		str := fmt.Sprintf("update %s set packid=?,platid=?,payplatid=?,minlevel=?,maxlevel=?,minmoney=?,maxmoney=?,stime=?,etime=?,state=?,updated_at=? where id=?", tblname)
		_, err := db_gm.Exec(str, data.GetPackageid(), data.GetPlatid(), data.GetPayplatid(), data.GetMinlevel(), data.GetMaxlevel(), data.GetMinmoney(), data.GetMaxmoney(), data.GetStime(), data.GetEtime(), data.GetState(), now, recordid)
		if err != nil {
			self.Error("CPay update error:%s, sql:%s, id:%d", err.Error(), str, recordid)
		}
	}
	self.ForwardGmCommand(rev, rev.GetGameid(), 0, false, &Pmd.CPayGmUserPmd_CS{})
	return true
}

func (self *GmTask) ParseCPaylistGmUserPmd_CS(rev *Pmd.CPaylistGmUserPmd_CS) bool {
	self.Debug("ParseCPaylistGmUserPmd_CS data:%s", unibase.GetProtoString(rev.String()))
	tblname := get_cpay_table()
	where := fmt.Sprintf("gameid=%d", rev.GetGameid())
	if rev.GetRecordid() != 0 {
		where += fmt.Sprintf(" AND id=%d ", rev.GetRecordid())
	}
	str := fmt.Sprintf("select id, packid,platid,payplatid,minlevel,maxlevel,minmoney,maxmoney,stime,etime,state from %s where %s", tblname, where)
	rows, err := db_gm.Query(str)
	self.Debug("CPay sql:%s", str)
	if err != nil {
		self.Error("ParseCPaylistGmUserPmd_CS query error:%s", err.Error())
		self.SendCmd(rev)
		return false
	}
	rev.Data = make([]*Pmd.CPayData, 0)
	for rows.Next() {
		data := &Pmd.CPayData{}
		if err := rows.Scan(&data.Recordid, &data.Packageid, &data.Platid, &data.Payplatid, &data.Minlevel, &data.Maxlevel, &data.Minmoney, &data.Maxmoney, &data.Stime, &data.Etime, &data.State); err != nil {
			self.Error("ParseCPaylistGmUserPmd_CS scan error:%s", err.Error())
			continue
		}
		rev.Data = append(rev.Data, data)
	}
	rows.Close()
	self.SendCmd(rev)
	return true
}

func (self *GmTask) ParseCmdDefault(rev proto.Message) bool {
	if cmd, ok := rev.(GmTaskMessage); ok {
		self.ForwardGmCommand(rev, cmd.GetGameid(), cmd.GetZoneid(), true, rev)
		return true
	}
	self.Error("通用函数不支持处理该协议，请手动实现. gm:%s, data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	return false
}

func (self *GmTask) ParseAddNewZoneGmUserPmd_CS(rev *Pmd.AddNewZoneGmUserPmd_CS) bool {
	err := add_game_zone(rev.GetGameid(), rev.GetZoneid(), rev.GetZonename(), rev.GetGmlink())
	if err == nil {
		rev.Retcode = proto.Uint32(0)
		rev.Retdesc = proto.String("success")
		content := fmt.Sprintf("创建大区%s", rev.GetZonename())
		add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	} else {
		rev.Retcode = proto.Uint32(1)
		rev.Retdesc = proto.String("创建失败！")
	}
	self.SendCmd(rev)
	return true
}

// func (self *GmTask) ParseRequestChatMessageGmUserPmd_CS(rev *Pmd.RequestChatMessageGmUserPmd_CS) bool {
// 	cdate := rev.GetChatdate()
// 	now := uint32(unitime.Time.YearMonthDay())
// 	day3 := uint32(unitime.Time.YearMonthDay(-3))
// 	self.Info("gm:%s ParseRequestChatMessageGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
// 	if zoneTaskManager.CheckGameOnline(rev.GetGameid()) {
// 		if cdate == 0 || cdate <= day3 || now < cdate {
// 			cdate = now
// 		}
// 		tblname := get_user_chat_table(rev.GetGameid(), rev.GetZoneid(), cdate)
// 		charid, ctype := rev.GetCharid(), rev.GetType()
// 		where := ""
// 		if ctype != 0 {
// 			where += fmt.Sprintf(" type =%d ", ctype)
// 		}
// 		if charid != 0 {
// 			if where != "" {
// 				where += " AND "
// 			}
// 			where += fmt.Sprintf(" charid = %d ", charid)
// 		}
// 		if where != "" {
// 			where = " WHERE " + where
// 		}
// 		str := fmt.Sprintf("select cpid, platid, charid, charname, type, otherid, othername, content, created_at, accid from %s %s order by id desc limit 200", tblname, where)
// 		data := make([]*Pmd.ChatMessageGmUserPmd_C, 0)
// 		rows, err := db_gm.Query(str)
// 		if err == nil {
// 			for rows.Next() {
// 				d := &Pmd.ChatMessageGmUserPmd_C{}
// 				if err := rows.Scan(&d.Cpid, &d.Platid, &d.Charid, &d.Charname, &d.Type, &d.Otherid, &d.Othername, &d.Content, &d.Createdat, &d.Accid); err != nil {
// 					continue
// 				}
// 				data = append(data, d)
// 			}
// 		}
// 		rev.Data = data
// 		self.SendCmd(rev)
// 	} else {
// 		self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.RequestChatMessageGmUserPmd_CS{})
// 	}

// 	return true
// }

func (self *GmTask) ParseRequestDealFeedbackGmUserPmd_CS(rev *Pmd.RequestDealFeedbackGmUserPmd_CS) bool {
	gamekey, _ := get_game_gmlink_and_key(rev.GetGameid(), rev.GetZoneid())
	self.Info("gm:%s ParseRequestDealFeedbackGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	if gamekey == "" {
		charid, ftype, content := checkFeedback(rev.GetRecordid())
		rev.Charid = proto.Uint64(charid)
		rev.Feedbacktype = proto.Uint32(ftype)
		var slen int = 6
		if len([]rune(content)) < slen {
			slen = len([]rune(content))
		}
		rev.Subject = proto.String("回复-" + string([]rune(content)[:slen]))
	}

	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.RequestDealFeedbackGmUserPmd_CS{})
	rev.Retcode = proto.Uint32(0)
	self.SendCmd(rev)
	updateFeedbackState(rev.GetRecordid(), rev.GetState(), rev.GetReply())
	return true
}

// 转发相应的协议到游戏服
func (self *GmTask) ForwardGmCommand(send proto.Message, gameid, zoneid uint32, wrap bool, retcmd proto.Message) bool {
	if !self.VerifyOk {
		return false
	}
	var sendwrap proto.Message
	if wrap && !zoneTaskManager.CheckZoneTaskBwPort(gameid) {
		data, _ := json.Marshal(send)
		sj := sjson.New()
		sj.Set("data", string(data))
		sj.Set("do", reflect.TypeOf(send).String()[1:])
		data, _ = sj.MarshalJSON()
		send2 := &Pmd.RequestExecGmCommandGmPmd_SC{}
		send2.Gameid = proto.Uint32(gameid)
		send2.Zoneid = proto.Uint32(zoneid)
		send2.Gmid = proto.Uint32(uint32(self.GetId()))
		send2.Msg = proto.String(string(data))
		sendwrap = send2
	} else {
		sendwrap = send
	}

	if zoneTaskManager.CheckGameOnline(gameid) {
		self.Info("gm:%s ForwardGmCommand data:%s", self.GetName(), unibase.GetProtoString(sendwrap.String()))
		if zoneid == 0 {
			zoneTaskManager.BroadcastToGame(gameid, sendwrap)
		} else {
			task := zoneTaskManager.GetZoneTaskById(unibase.GetGameZone(gameid, zoneid))
			fmt.Println("gameid:", gameid)
			fmt.Println("zoneid:", zoneid)

			if task != nil {
				task.SendCmd(sendwrap)
			} else {
				self.Error("gm:%s ForwardGmCommand error, gameid:%d, zoneid:%d, %s", self.GetName(), gameid, zoneid, unibase.GetProtoString(sendwrap.String()))
				res := &Pmd.RequestGameZoneErrorGmPmd_S{}
				res.Retcode = proto.Uint32(1)
				res.Retdesc = proto.String("Gameid or zoneid error")
				res.Gmid = proto.Uint32(uint32(self.GetId()))
				self.SendCmd(res)
			}
		}
	} else {
		gamekey, gmlinks := get_game_gmlink_and_key(gameid, zoneid)
		if len(gmlinks) == 0 {
			res := &Pmd.RequestGameZoneErrorGmPmd_S{}
			res.Retcode = proto.Uint32(1)
			res.Retdesc = proto.String("Gameid or zoneid error11")
			res.Gmid = proto.Uint32(uint32(self.GetId()))
			self.SendCmd(res)
			return false
		}

		client := &http.Client{
			Timeout: time.Second * 10,
		}
		body, err := json.Marshal(send)
		if err != nil {
			self.Error("json Marshal error:%s", err.Error())
			return false
		}
		sign := md5String(string(body) + gamekey)

		for _, gmlink := range gmlinks {
			url := gmlink + fmt.Sprintf("?sign=%s&do=%s", sign, reflect.TypeOf(send).String()[5:])
			iobody := strings.NewReader(string(body))
			req, err := http.NewRequest("POST", url, iobody)
			self.Debug("GmTask Http request, url:%s, body:%s", url, string(body))
			if err != nil {
				self.Error("HttpRequestGet NewRequest err:%s", err.Error())
				return false
			}
			response, err := client.Do(req)
			if err == nil {
				defer response.Body.Close()
				var resdata []byte
				resdata, err = ioutil.ReadAll(response.Body)
				err := json.Unmarshal(resdata, retcmd)
				if err != nil {
					self.Error("json Unmarshal error, send:%s, rev:%s, err:%s", unibase.GetProtoString(send.String()), string(resdata), err.Error())
					return false
				}
				self.SendCmd(retcmd)
				self.Debug("GmTask Http send:%s, response:%s", unibase.GetProtoString(send.String()), unibase.GetProtoString(retcmd.String()))
			} else {
				self.Error("Http Request error:%s", err.Error())
				res := &Pmd.RequestGameZoneErrorGmPmd_S{}
				res.Retcode = proto.Uint32(1)
				res.Retdesc = proto.String(err.Error())
				res.Gmid = proto.Uint32(uint32(self.GetId()))
				self.SendCmd(res)
				return false
			}
		}
	}
	return true
}

func (self *GmTask) ParseServerListGmUserPmd_CS(rev *Pmd.ServerListGmUserPmd_CS) bool {
	self.ForwardGmCommand(rev, rev.GetGameid(), rev.GetZoneid(), true, &Pmd.ServerListGmUserPmd_CS{})
	return true
}

func md5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
