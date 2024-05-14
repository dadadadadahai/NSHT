package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	Smd "git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unisocket"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type ZoneTask struct {
	nettask.NetTaskInterFace
	Version        uint32
	serverState    *Pmd.GameZoneServerState
	_timer_one_min *unitime.Timer
	VerifyOk       bool
}

func NewZoneTask(ws *websocket.Conn, parsefunc nettask.HandleForwardFunc) *ZoneTask {
	task := &ZoneTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTask(unisocket.NewUniSocketWs(ws, websocket.BinaryFrame), parsefunc, 50*time.Millisecond, gameTaskManager)
	task.SetEntryName(func() string { return "ZoneTask" })
	task.serverState = &Pmd.GameZoneServerState{
		Gamezone: &Pmd.GameZoneInfo{},
	}
	task._timer_one_min = unitime.NewTimer(unitime.Time.Now(), 60*1000, false)
	return task
}

func NewZoneTaskBw(conn *net.TCPConn) *ZoneTask {
	task := &ZoneTask{
		VerifyOk: false,
	}
	task.NetTaskInterFace = nettask.NewNetTaskBw(unisocket.NewUniSocketBw(conn), nettask.ParseBwReflectCommand, 50*time.Millisecond, gameTaskManager)
	task.SetEntryName(func() string { return "ZoneTaskBw" })
	task.SetTaskVersion(nettask.TaskVersion)
	task.serverState = &Pmd.GameZoneServerState{
		Gamezone: &Pmd.GameZoneInfo{},
	}
	task._timer_one_min = unitime.NewTimer(unitime.Time.Now(), 60*1000, false)
	return task
}

func (self *ZoneTask) GetGameId() uint32 {
	return unibase.GetGameId(self.GetId())
}

func (self *ZoneTask) GetZoneId() uint32 {
	return unibase.GetZoneId(self.GetId())
}

// 计算昨日日新增，次日留存
func (self *ZoneTask) CalcRetention() (retm map[string][]interface{}) {

	retm = CalcRetention(1001, self.GetZoneId())
	return
}

func (self *ZoneTask) CalcRealtimeData() {

	CalcRealtimeData(1001, self.GetZoneId())

}

func (self *ZoneTask) Loop() {

}

func (self *ZoneTask) ClockZero() {

	ClockZero(1001)

}

func (self *ZoneTask) CheckResetDatabase(offset int) {
	create_user_data(1001)
	create_coin_table(1001)
	create_user_daily(1001)
	create_user_imei(1001)
	create_user_daily_plat(1001)
	create_user_daily_roi(1001)
	create_user_pay(1001)
	create_user_cashout(1001)
	create_user_daily_subplat(1001)
	create_user_daily_invite(1001)
	create_user_levelup(1001)
	create_user_account_table(1001)
	create_user_transaction(1001)
	create_shop_transaction(1001)
	create_user_lottery(1001)
	create_user_exchange(1001)

	create_exchange_rate_table(1001)
	create_launch_keys_table(1001)
	create_adjapp_table(1001)
	create_change_registersrc_table(1001)
	create_app_cost_today_table(1001)
	create_app_access_table(1001)

	create_money_change_table(1001)

	create_user_login(1001, uint32(unitime.Time.YearMonthDay(offset-1)))
	create_user_online(1001, uint32(unitime.Time.YearMonthDay(offset-1)))
	create_user_detail(1001, uint32(unitime.Time.YearMonthDay(offset-1)))
	create_user_online(1001, uint32(unitime.Time.YearMonthDay(offset)))
	create_user_detail(1001, uint32(unitime.Time.YearMonthDay(offset)))
	create_user_online_plat(1001, uint32(unitime.Time.YearMonthDay(offset-1)))
	create_user_online_plat(1001, uint32(unitime.Time.YearMonthDay(offset)))

	create_action_record(1001, uint32(unitime.Time.YearMonthDay(offset)))
	create_user_economic(1001, uint32(unitime.Time.YearMonthDay(offset)))
	create_user_item(1001, uint32(unitime.Time.YearMonthDay(offset)))

	create_realtime_data_table(1001, uint32(unitime.Time.YearMonthDay(offset-1)))
	create_realtime_data_table(1001, uint32(unitime.Time.YearMonthDay(offset)))
}

func (self *ZoneTask) ParseStartUpGameRequestMonitorSmd_C(rev *Smd.StartUpGameRequestMonitorSmd_C) bool {
	if rev.GetCompress() != "" {
		self.SetCompress(rev.GetCompress(), 0)
	}
	if rev.GetEncrypt() != "" {
		self.SetEncrypt(rev.GetEncrypt(), rev.GetEncryptkey())
	}
	send := &Smd.StartUpGameReturnMonitorSmd_S{Ret: proto.Bool(false)}
	zone := unibase.CheckZoneInfo(rev.GetKey())
	if rev.GetCompress() != "" {
		self.SetCompress(rev.GetCompress(), 0)
	}
	if rev.GetEncrypt() != "" {
		self.SetEncrypt(rev.GetEncrypt(), rev.GetEncryptkey())
	}
	if rev.GetVersion() != uint32(Smd.Config_Version_Monitor) {
		send := &Pmd.CheckVersionUserPmd_CS{
			Versionserver: proto.Uint32(uint32(Smd.Config_Version_Monitor)),
			Versionclient: rev.Version,
		}
		self.SendCmd(send)
		self.Error("版本号错误:%d,%d", rev.GetVersion(), uint32(Smd.Config_Version_Monitor))
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
		zoneTaskManager := gameTaskManager.GetZoneTaskManagerById(uint64(1001))
		oldTask = zoneTaskManager.GetZoneTaskById(self.GetId())
		if oldTask != nil {
			oldTask.Error("duplicate login:%s,%s,%s", oldTask.GetRemoteAddr(), self.GetRemoteAddr(), unibase.GetProtoString(rev.String()))
			if config.GetConfigStr("debug") != "false" && oldTask.GetRemoteIp() != self.GetRemoteIp() {
				oldTask.Error("被新用户踢下线:%s,%s", oldTask.GetRemoteAddr(), self.GetRemoteAddr())
				if rev.GetLastseq() != 0 {
					self.SetReconnectData(oldTask)
				}
				zoneTaskManager.RemoveZoneTask(oldTask)
				send := &Smd.ReconnectKickoutMonitorSmd_S{}
				send.Desc = proto.String(fmt.Sprintf("被新用户踢下线:%s,%s", oldTask.GetRemoteAddr(), self.GetRemoteAddr()))
				oldTask.SendCmd(send)
				oldTask.SetRemoveMeFunc(nil)
				//oldTask.Close()
			}
			//self.SendCmd(send)
			//self.Close()
			//return false
		}

		if zoneTaskManager.AddZoneTask(self) == true {
			self.SetRemoveMeFunc(func() {
				zoneTaskManager.RemoveZoneTask(self)
				gameTaskManager.ResetServerState(1001)
				serverlist, ok := unibase.GlobalZoneListMap[1001]
				if ok == true {
					monitorTaskManager.Broadcast(serverlist)
				}
			})
			send.Ret = proto.Bool(true)
		} else {
			send.Retdesc = proto.String("duplicate login:" + oldTask.GetRemoteAddr())
			self.Error("duplicate login")
		}
		self.Info("login ok:%s,%s", self.GetRemoteAddr(), unibase.GetProtoString(rev.String()))
	} else {
		send.Zoneinfo = &Pmd.GameZoneInfo{}
		send.Retdesc = proto.String("未找到区信息:" + rev.GetKey())
		self.Error("login err:%s,%s", self.GetRemoteAddr(), unibase.GetProtoString(rev.String()))
	}
	if rev.GetLastseq() == 0 || self.SendReconnectData(self, rev.GetLastseq()) == false { //这里说明不进行断线重连或者断线重连尝试失败,就当重新登陆
		self.SendCmd(send)
	}
	if send.GetRet() == true {
		gameTaskManager.ResetServerState(1001)
		serverlist, ok := unibase.GlobalZoneListMap[1001]
		if ok == true {
			monitorTaskManager.Broadcast(serverlist)
		}
		//TODO refresh serverstate after gateway startok
	}

	self.CheckResetDatabase(0)
	checkDBUpdate(1001)
	return true
}

func (self *ZoneTask) ParseReturnOnlineNumMonitorSmd_C(rev *Smd.ReturnOnlineNumMonitorSmd_C) bool {
	// ClockZero(1001)
	platinfo := rev.GetPlatonlineinfo()
	var tmpinfo = make(map[string]int)
	onlinenum := rev.GetOnlinenum()
	if onlinenum == 0 {
		for _, p := range platinfo {
			tmpinfo[strconv.Itoa(int(p.GetPlatid()))] = int(p.GetOnlinenum())
			onlinenum += p.GetOnlinenum()
		}
	}
	oninleinfo, _ := json.Marshal(tmpinfo)
	zoneid, timemin := self.GetZoneId(), uint32(rev.GetTimestamp()/60)
	tblname := get_user_online_table(1001, uint32(unitime.Time.YearMonthDay()))
	str := fmt.Sprintf("insert into %s (zoneid,onlinenum,onlineinfo,timestamp_min) values(?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, zoneid, onlinenum, string(oninleinfo), timemin)
	if err != nil && strings.Split(err.Error(), ":")[0] != "Error 1062" {
		self.Error("ParseReturnOnlineNumMonitorSmd_C insert err:%s", err.Error())
	}
	if len(platinfo) > 0 {
		params := strings.Repeat("(?,?,?,?),", len(platinfo))
		values := make([]interface{}, 0)
		for _, p := range platinfo {
			values = append(values, zoneid, p.GetPlatid(), p.GetOnlinenum(), timemin)
		}
		tblname = get_user_online_plat_table(1001, uint32(unitime.Time.YearMonthDay()))
		str = fmt.Sprintf("insert into %s (zoneid, platid, onlinenum, timestamp_min) values %s", tblname, params[:len(params)-1])
		_, err = db_monitor.Exec(str, values...)
		if err != nil {
			self.Error("ParseReturnOnlineNumMonitorSmd_C error:%s", err.Error())
		}
	}
	self.Debug("ParseReturnOnlineNumMonitorSmd_C:%s, zoneid: %d ", unibase.GetProtoString(rev.String()), zoneid)
	send := &Smd.ForwardOnlineNumMonitorSmd_S{}
	send.Gamezone = self.serverState.Gamezone
	send.Onlinenum = proto.Uint32(onlinenum)
	monitorTaskManager.Broadcast(send)
	return true
}

// 玩家登陆的扩展信息，包含子平台、邀请者openid;主要用在登陆和刷新玩家数据，故定义在此处
type UserLoginExtData struct {
	Subplat string `json:"subplat"` //对应数据库
	Inviter string `json:"inviter"`
}

func (self *ZoneTask) CheckoutSubplatAndInviter(extdata string) (subplatid int, inviterid int) {
	extdata = strings.TrimSpace(extdata)
	if extdata == "" {
		return 0, 0
	}
	var tmpExtdata = &UserLoginExtData{}
	data, err := strconv.Unquote(extdata)
	if err != nil {
		data = extdata
	}
	err = json.Unmarshal([]byte(data), tmpExtdata)
	if err != nil {
		self.Error("CheckoutSubplatAndInviter error:%s extdata:%s", err.Error(), data)
		return 0, 0
	}
	subplat := strings.TrimSpace(tmpExtdata.Subplat)
	if subplat != "" {
		tblname := get_subplat_record_table()
		str := fmt.Sprintf("select `id` from %s where subplat=? and gameid=?", tblname)
		row := db_monitor.QueryRow(str, subplat, 1001)
		if err = row.Scan(&subplatid); err != nil {
			str = fmt.Sprintf("insert into %s (subplat, gameid, createtime) values(?,?,?)", tblname)
			result, err1 := db_monitor.Exec(str, subplat, 1001, uint32(unitime.Time.Sec()/60))
			if err1 != nil {
				self.Error("CheckoutSubplatAndInviter insert into %s error: %s, extdata:%s", tblname, err1.Error(), extdata)
				return 0, 0
			}
			tmpId, _ := result.LastInsertId()
			subplatid = int(tmpId)
		}
	}
	inviter := strings.TrimSpace(tmpExtdata.Inviter)
	if inviter != "" {
		tblname := get_invite_record_table()
		str := fmt.Sprintf("select `id` from %s where inviter=? and gameid=?", tblname)
		row := db_monitor.QueryRow(str, inviter, 1001)
		if err = row.Scan(&inviterid); err != nil {
			str = fmt.Sprintf("insert into %s (inviter, gameid, createtime) values(?,?,?)", tblname)
			result, err1 := db_monitor.Exec(str, inviter, 1001, uint32(unitime.Time.Sec()/60))
			if err1 != nil {
				self.Error("CheckoutSubplatAndInviter insert into %s error: %s, extdata:%s", tblname, err1.Error(), extdata)
				return 0, 0
			}
			tmpId, _ := result.LastInsertId()
			inviterid = int(tmpId)
		}
	}
	return
}

func (self *ZoneTask) ParseRefreshUserDataMonitorSmd_C(rev *Smd.RefreshUserDataMonitorSmd_C) bool {
	tblname := get_user_data_table(1001)
	str := fmt.Sprintf("update %s set userlevel=?, viplevel=?, power=?, gold=?, goldgive=?, money=?, isguid=? where zoneid=? and userid=?", tblname)
	_, err := db_monitor.Exec(str, rev.GetUserlevel(), rev.GetViplevel(), rev.GetPower(), rev.GetGold(), rev.GetGoldgive(), rev.GetMoney(), rev.GetIsguid(), self.GetZoneId(), rev.GetData().GetUserid())
	if err != nil {
		self.Error("ParseRefreshUserDataMonitorSmd_C error:%s", err.Error())
	}
	//self.Info("ParseRefreshUserDataMonitorSmd_C:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseLoginUserDataMonitorSmd_C(rev *Smd.LoginUserDataMonitorSmd_C) bool {

	self.Info("imei:%d", rev.GetImei())
	tblname := get_user_detail_table(1001, uint32(unitime.Time.YearMonthDay()))
	// tblnamedaily := get_user_daily_plat_table(1001)
	str := fmt.Sprintf("insert ignore into %s (zoneid,userid,ip,imei,min,sid) values(?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), rev.GetData().GetUserid(), rev.GetIp(), rev.GetImei(), uint32(unitime.Time.Sec()/60), rev.GetSid())
	if err != nil {
		self.Error("ParseLoginUserDataMonitorSmd_C insert err:%s", err.Error())
	}

	tblname = get_user_data_table(1001)
	data := rev.GetData()
	now := uint32(unitime.Time.Sec())
	ymd := uint32(unitime.Time.YearMonthDay())
	ymd1 := uint32(unitime.Time.YearMonthDay(-1))
	ymd7 := uint32(unitime.Time.YearMonthDay(-7))
	//
	selstr := fmt.Sprintf(`select count(*),launchid , curr_platid , curr_launchid , curr_ad_account , FROM_UNIXTIME(reg_time , "%%Y%%m%%d")  from %s where zoneid=? and userid = ? `, tblname)

	rowsel := db_monitor.QueryRow(selstr, self.GetZoneId(), data.GetUserid())

	var count, oldplatid, regtime uint32
	var launchid, ad_account, reglaunchid string

	update := 0

	rowsel.Scan(&count, &reglaunchid, &oldplatid, &launchid, &ad_account, &regtime)

	if count > 0 && (launchid != rev.GetAdcode() || ad_account != rev.GetExtdata() || oldplatid != uint32(data.GetPlatid())) {

		//添加渠道变化信息

		tblchan := get_change_registersrc_table(1001)

		selstr := fmt.Sprintf(`insert into %s (userid,zoneid,sid,befor_platid,befor_launchid,befor_ad_account,after_platid,after_launchid,after_ad_account) values(?,?,?,?,?,?,?,?,?) `, tblchan)

		_, err := db_monitor.Exec(selstr, data.GetUserid(), self.GetZoneId(), rev.GetSid(), oldplatid, launchid, ad_account, uint32(data.GetPlatid()), rev.GetAdcode(), rev.GetExtdata())

		if err != nil {
			self.Error("insert change_registersrc err:%s", err.Error())
		}
	}

	packid := data.GetPackageid()
	platid, platacc := int(data.GetPlatid()), data.GetAccountname()
	accid := data.GetAccountid()
	if accid == 0 {
		accid = data.GetUserid()
	}
	str = fmt.Sprintf(`
		update %s set logindays=(case from_unixtime(last_login_time, '%%Y%%m%%d') when ? then logindays+1 when ? then logindays else 1 end),
		logintimes=logintimes+1,flag=(case when from_unixtime(last_login_time, '%%Y%%m%%d')<? then 1 else 0 end), packid=?,
        isonline=1, lastmin=?, username=?, isguid=?, userlevel=?, viplevel=?, power=?, gold=?, goldgive=?, money=?,  last_login_time=?, ip = ? , mobile = ? , curr_launchid = ?,curr_ad_account = ? , curr_platid = ? , imei=? , account = ? , plataccount = ? , platid = (case when platid = 0 then ? else platid end) , launchid = (case when launchid = "" then ? else launchid end) , reg_time = (case when reg_time = 0 then ? else reg_time end) where zoneid=? and userid=? `, tblname)
	result, err := db_monitor.Exec(str, ymd1, ymd, ymd7, packid, now/60, data.GetUsername(), rev.GetIsguid(), rev.GetUserlevel(), rev.GetViplevel(), rev.GetPower(), rev.GetGold(), rev.GetGoldgive(), rev.GetMoney(), now, rev.GetIp(), data.GetMobilenum(), rev.GetAdcode(), rev.GetExtdata(), platid, rev.GetImei(), platacc, platacc, platid, rev.GetAdcode(), rev.GetCreatetime(), self.GetZoneId(), data.GetUserid())

	// fmt.Println("rev.GetAdcode(), rev.GetExtdata(), int(data.GetPlatid()):", rev.GetAdcode(), rev.GetExtdata(), int(data.GetPlatid()))
	var row int64
	if err == nil {
		row, err = result.RowsAffected()
	} else {
		self.Error("ParseLoginUserDataMonitorSmd_C update err:%s", err.Error())
	}

	if err != nil || row == 0 {
		//subplatid, inviterid := self.CheckoutSubplatAndInviter(rev.GetExtdata())
		subplatid, inviterid := 0, 0
		str := fmt.Sprintf(`insert ignore into %s (zoneid,userid,username,accid,account,platid,plataccount,ip,imei,firstmin,
			lastmin,initsubplat,cursubplat,inviterid,launchid,sid,isguid,isonline,last_login_time,logindays,logintimes, packid,ad_account,mobile,curr_launchid,curr_ad_account, curr_platid)
			values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,1,1,?,?,?,?,?,?)`, tblname)
		_, err := db_monitor.Exec(str, self.GetZoneId(), data.GetUserid(), data.GetUsername(), accid, platacc,
			platid, platacc, rev.GetIp(), rev.GetImei(), now/60, now/60, subplatid, subplatid, inviterid, rev.GetAdcode(), rev.GetSid(), rev.GetIsguid(), now, packid, rev.GetExtdata(), data.GetMobilenum(), rev.GetAdcode(), rev.GetExtdata(), platid)

		if err != nil {
			self.Error("ParseLoginUserDataMonitorSmd_C insert err:%s", err.Error())
		}
	}
	str = fmt.Sprintf("update %s set reg_ip=?, reg_time=? where userid=? and reg_ip=0 and zoneid=? ", tblname)
	_, err = db_monitor.Exec(str, rev.GetIp(), rev.GetCreatetime(), data.GetUserid(), self.GetZoneId())
	if err != nil {
		self.Error("ParseLoginUserDataMonitorSmd_C insert reg_ip err:%s", err.Error())
	}

	LoginAccount(1001, accid, platacc, platid, rev.GetExtdata(), rev.GetAdcode(), rev.GetIp(), rev.GetImei(), self.GetZoneId(), 0, data.GetMobilenum(), rev.GetSid(), uint32(update))
	UpdateFirstGametime(1001, self.GetZoneId(), accid, data.GetUserid())
	self.Info("ParseLoginUserDataMonitorSmd_C:%s,zoneid:%d", unibase.GetProtoString(rev.String()), self.GetZoneId())
	return true
}

func (self *ZoneTask) ParseLogoutUserDataMonitorSmd_C(rev *Smd.LogoutUserDataMonitorSmd_C) bool {

	self.Info("ParseLogoutUserDataMonitorSmd_C:%s", unibase.GetProtoString(rev.String()))
	nowmin := uint32(unitime.Time.Sec() / 60)
	tblname := get_user_data_table(1001)
	str := fmt.Sprintf("update %s set isonline=0, last_logout_time=?, isguid=?, userlevel=?, viplevel=?, power=?, gold=?, goldgive=?, money=?, onlinemin= onlinemin + ? , plataccount=? where userid=?", tblname)
	_, err := db_monitor.Exec(str, uint32(unitime.Time.Sec()), rev.GetIsguid(), rev.GetLevel(), rev.GetViplevel(), rev.GetPower(), rev.GetGold(), rev.GetGoldgive(), rev.GetMoney(), rev.GetOnlinetime()/60, rev.GetData().GetAccountname(), rev.GetData().GetUserid())
	if err != nil {
		self.Error("ParseLogoutUserDataMonitorSmd_C update err:%s", err.Error())
	}
	tblname = get_user_detail_table(1001, uint32(unitime.Time.YearMonthDay()))
	if !check_table_exists(tblname) {
		tblname = get_user_detail_table(1001, uint32(unitime.Time.YearMonthDay(-1)))
	}
	str = fmt.Sprintf("select id, logoutmin from %s where zoneid=? and userid=? order by id desc limit 1", tblname)
	row := db_monitor.QueryRow(str, self.GetZoneId(), rev.GetData().GetUserid())
	var id, logoutmin int64
	err = row.Scan(&id, &logoutmin)
	if err == nil && id != 0 && logoutmin == 0 {
		str = fmt.Sprintf("update %s set logoutmin=?, onlinetime=(?-min), sceneid=?, taskid=?, level=? where id=? ", tblname)
		_, err = db_monitor.Exec(str, nowmin, nowmin, rev.GetSceneid(), rev.GetTaskid(), rev.GetLevel(), id)
	} else {
		str = fmt.Sprintf("select last_login_time from %s where zoneid=? and userid=?", get_user_data_table(1001))
		db_monitor.QueryRow(str, self.GetZoneId(), rev.GetData().GetUserid()).Scan(&logoutmin)
		str = fmt.Sprintf("insert ignore into %s (zoneid,userid,min,logoutmin,onlinetime,sceneid,taskid,level) values(?,?,?,?,?,?,?,?)", tblname)
		_, err = db_monitor.Exec(str, self.GetZoneId(), rev.GetData().GetUserid(), int(logoutmin/60), nowmin, (nowmin - uint32(logoutmin/60)), rev.GetSceneid(), rev.GetTaskid(), rev.GetLevel())
	}
	if err != nil {
		self.Error("ParseLogoutUserDataMonitorSmd_C err:%s, sql:%s, zoneid:%d, roleid:%d", err.Error(), str, self.GetZoneId(), rev.GetData().GetUserid())
	}
	self.Info("ParseLogoutUserDataMonitorSmd_C:%s", unibase.GetProtoString(rev.String()))
	LogoutAccount(1001, rev.GetData().GetAccountid(), rev.GetData().GetUserid(), rev.GetData().GetPackageid())
	return true
}

func (self *ZoneTask) ParseLevelUpUserDataMonitorSmd_C(rev *Smd.LevelUpUserDataMonitorSmd_C) bool {
	return true //暂不处理
	tblname := get_user_levelup_table(1001)
	daynum := uint32(unitime.Time.YearMonthDay())
	str := fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, oldlevel, newlevel, leveltime, leveltype, typename) values(?,?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), daynum, rev.GetData().GetUserid(), rev.GetOldlevel(), rev.GetNewlevel(), rev.GetLeveltime(), rev.GetLeveltype(), rev.GetTypename())
	if err != nil {
		self.Error("ParseLevelUpUserDataMonitorSmd_C insert error:%s", err.Error())
	}
	//self.Info("ParseLevelUpUserDataMonitorSmd_C:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseEconomicProduceConsumeMonitorSmd_C(rev *Smd.EconomicProduceConsumeMonitorSmd_C) bool {
	/*
		coinid := rev.GetCoinid()
		if coinid != 2 && coinid != 8 {
			return true
		}
	*/

	daynum := uint32(unitime.Time.YearMonthDay())
	tblname := get_user_economic_table(1001, daynum)
	str := fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, coinid, coincount, actionid, actioncount, type, level, actionname, coinname, curcoin) values(?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), daynum, rev.GetData().GetUserid(), rev.GetCoinid(), rev.GetCoincount(), rev.GetActionid(), rev.GetActioncount(), rev.GetType(), rev.GetLevel(), rev.GetActionname(), rev.GetCoinname(), rev.GetCurcoin())
	if err != nil {
		self.Error("ParseEconomicProduceConsumeMonitorSmd_C insert error:%s", err.Error())
	}
	//self.Info("ParseEconomicProduceConsumeMonitorSmd_C:%s", unibase.GetProtoString(rev.String()))
	//更新下玩家身上的数据

	bUpdate := false
	tbdata := get_user_data_table(1001)
	//金币
	if rev.GetCoinid() == 1 {
		str = fmt.Sprintf("update %s set money=? where userid=? and zoneid=?", tbdata)
		bUpdate = true
	}
	//钻石
	if rev.GetCoinid() == 8 {
		str = fmt.Sprintf("update %s set gold=? where userid=? and zoneid=?", tbdata)
		bUpdate = true
	}

	if bUpdate == true {
		_, err = db_monitor.Exec(str, rev.GetCurcoin(), rev.GetData().GetUserid(), self.GetZoneId())
		if err != nil {
			self.Error("ParseEconomicProduceConsumeMonitorSmd_C update err:%s", err.Error())
		}
	}

	return true
}

func (self *ZoneTask) ParseItemProduceConsumeMonitorSmd_C(rev *Smd.ItemProduceConsumeMonitorSmd_C) bool {
	daynum := uint32(unitime.Time.YearMonthDay())
	tblname := get_user_item_table(1001, daynum)
	str := fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, itemtype, itemid, itemname, itemcount, actionid, type, level, actionname, gold, curnum, extdata) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), daynum, rev.GetData().GetUserid(), rev.GetItemtype(), rev.GetItemid(), rev.GetItemname(), rev.GetItemcount(), rev.GetActionid(), rev.GetType(), rev.GetLevel(), rev.GetActionname(), rev.GetGold(), rev.GetCuritem(), rev.GetExtdata())
	if err != nil {
		self.Error("ParseItemProduceConsumeMonitorSmd_C insert error:%s", err.Error())
	}
	//self.Info("ParseItemProduceConsumeMonitorSmd_C:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseActionRecordMonitorSmd_C(rev *Smd.ActionRecordMonitorSmd_C) bool {
	tblname := get_action_table(1001, rev.GetType(), uint32(unitime.Time.YearMonthDay()))
	if len(tblname) == 0 {
		self.Error("ParseActionRecordMonitorSmd_C type error:%d", rev.GetType())
		return true
	}
	daynum := uint32(unitime.Time.YearMonthDay())
	str := fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, actionid, actionname, starttime, duration, endtime, state, power, acttype, acttypename, level, viplevel, extdata) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), daynum, rev.GetData().GetUserid(), rev.GetActionid(), rev.GetActionname(), rev.GetStarttime(), rev.GetDuration(), rev.GetStarttime()+rev.GetDuration(), rev.GetState(), rev.GetPower(), rev.GetActtype(), rev.GetActtypename(), rev.GetLevel(), rev.GetViplevel(), rev.GetExtdata())
	if err != nil {
		self.Error("ParseActionRecordMonitorSmd_C error:%s, %s", err.Error(), rev.String())
	}
	//self.Info("ParseActionRecordMonitorSmd_C:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseActioniListRecordMonitorSmd_C(rev *Smd.ActionRecordListMonitorSmd_C) bool {
	tblname := get_action_table(1001, rev.GetType(), uint32(unitime.Time.YearMonthDay()))
	datas := rev.GetData()
	if len(tblname) == 0 || len(datas) == 0 {
		self.Error("ParseActionRecordMonitorSmd_C error:%d, %d", rev.GetType(), len(datas))
		return true
	}
	daynum := uint32(unitime.Time.YearMonthDay())
	values, params, count, str := make([]interface{}, 0), "", 0, ""
	for _, data := range datas {
		count += 1
		values = append(values, self.GetZoneId(), daynum, data.GetData().GetUserid(), data.GetActionid(), data.GetActionname(), data.GetStarttime(), data.GetDuration(), data.GetStarttime()+data.GetDuration(), data.GetState(), data.GetPower(), data.GetActtype(), data.GetActtypename(), data.GetLevel(), data.GetViplevel())
		if count >= 100 {
			params = strings.Repeat("(?,?,?,?,?,?,?,?,?,?,?,?,?,?),", count)
			str = fmt.Sprintf("insert ignore into %s (zoneid, daynum, userid, actionid, actionname, starttime, duration, endtime, state, power, acttype, acttypename, level, viplevel) values %s", tblname, params[:len(params)-1])
			_, err := db_monitor.Exec(str, values...)
			if err != nil {
				self.Error("ParseActioniListRecordMonitorSmd_C Error:%s", err.Error())
			}
			count = 0
			values = values[:0]
		}
	}
	params = strings.Repeat("(?,?,?,?,?,?,?,?,?,?,?,?,?,?),", count)
	str = fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, actionid, actionname, starttime, duration, endtime, state, power, acttype, acttypename, level, viplevel) values %s", tblname, params[:len(params)-1])
	_, err := db_monitor.Exec(str, values...)
	if err != nil {
		self.Error("ParseActionRecordMonitorSmd_C error:%s", err.Error())
	}
	//self.Info("ParseActioniListRecordMonitorSmd_C:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseReturnRedeemPlatPointSdkPmd_S(rev *Pmd.ReturnRedeemPlatPointSdkPmd_S) bool {
	tbpay := get_user_pay_table(1001)
	ymd := uint32(unitime.Time.YearMonthDay())
	str := fmt.Sprintf("insert ignore into %s (zoneid,daynum,userid,gameorder,money,state,type) values(?,?,?,?,?,1,2)", tbpay)
	_, err := db_monitor.Exec(str, self.GetZoneId(), ymd, rev.GetData().GetMyaccid(), rev.GetGameorder(), int(rev.GetMoney()))
	if err != nil {
		self.Error("ParseReturnRedeemPlatPointSdkPmd_S insert err:%s", err.Error())
	}
	//self.Info("ParseReturnRedeemPlatPointSdkPmd_S:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseReturnRedeemBackPlatPointSdkPmd_S(rev *Pmd.ReturnRedeemBackPlatPointSdkPmd_S) bool {
	tbpay := get_user_pay_table(1001)
	ymd := uint32(unitime.Time.YearMonthDay())
	str := fmt.Sprintf("insert ignore into %s (zoneid,daynum,userid,gameorder,money,state,type) values(?,?,?,?,?,1,3)", tbpay)
	_, err := db_monitor.Exec(str, self.GetZoneId(), ymd, rev.GetData().GetMyaccid(), rev.GetGameorder(), int(rev.GetMoney()))
	if err != nil {
		self.Error("ParseReturnRedeemPlatPointSdkPmd_S insert err:%s", err.Error())
	}
	//self.Info("ParseReturnRedeemBackPlatPointSdkPmd_S:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseCreatePlatOrderReturnSdkPmd_S(rev *Pmd.CreatePlatOrderReturnSdkPmd_S) bool {
	tblname := get_user_pay_table(1001)
	ymd := uint32(unitime.Time.YearMonthDay())
	str := fmt.Sprintf("insert ignore into %s (zoneid,daynum,userid,gameorder,money,goodid,goodnum,platorder,type) values(?,?,?,?,?,?,?,?,1)", tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), ymd, rev.GetRoleid(), rev.GetGameorder(), int(rev.GetOrdermoney()), rev.GetGoodid(), rev.GetGoodnum(), rev.GetPlatorder())
	if err != nil {
		self.Error("ParseCreatePlatOrderReturnSdkPmd_S insert err:%s", err.Error())
	}
	//self.Info("ParseCreatePlatOrderReturnSdkPmd_S:%s", unibase.GetProtoString(rev.String()))
	return true
}

//提现

func (self *ZoneTask) ParseStUserWithDrawMonitorSmd_CS(rev *Smd.StUserWithDrawMonitorSmd_CS) bool {

	tbdata := get_user_data_table(1001)
	tblcashout := get_user_cashout_table(1001)

	var id, accid, count uint64
	var user_level uint32
	str := fmt.Sprintf("select id, accid, userlevel from %s where userid=?", tbdata)
	row := db_monitor.QueryRow(str, rev.GetData().Userid)
	err := row.Scan(&id, &accid, &user_level)
	if err != nil {
		return false
	}
	ymd := uint32(unitime.Time.YearMonthDay())

	str = fmt.Sprintf("select count(*) from %s where gameorder=? and platorder=? and zoneid = ? and userid = ? ", tblcashout)

	row = db_monitor.QueryRow(str, rev.GetGameorder(), rev.GetPlatorder(), self.GetZoneId(), rev.GetData().Userid)
	row.Scan(&count)
	if count > 0 {
		self.Error("HandleCashout Already exists:%s", rev.GetPlatorder())
		return false
	}

	str = fmt.Sprintf("insert ignore into %s (zoneid,daynum,userid,gameorder,money,state,platorder) values(?,?,?,?,?,?,?)", tblcashout)
	_, err = db_monitor.Exec(str, self.GetZoneId(), ymd, rev.GetData().Userid, rev.GetGameorder(), rev.GetMoney(), 1, rev.GetPlatorder())
	if err != nil {
		self.Error("HandleCashout insert err:%s", err.Error())
		return false
	}

	str = fmt.Sprintf("update %s set  cash_out_all=cash_out_all + ?, cash_out_num=cash_out_num +1 where id=?", tbdata)
	_, err = db_monitor.Exec(str, rev.GetMoney(), id)
	if err != nil {
		logging.Error("HandleCashout update err:%s", err.Error())
		return false
	}
	self.Info("ParseStUserWithDrawMonitorSmd_CS:%s", unibase.GetProtoString(rev.String()))
	// tblmoney := get_money_change_table(1001)

	// str = fmt.Sprintf("select count(*) form %s where daynum = %d and userid = %d", tblmoney, ymd, rev.GetRoleid())

	// var count = 0

	// row = db_monitor.QueryRow(str, self.GetZoneId(), rev.GetRoleid())
	// err = row.Scan(&count)

	// if count == 0 {
	// 	str = fmt.Sprintf("insert ignore into %s (daynum,userid,cash_num,cash_all) values(?,?,?,?)", tblmoney)
	// 	_, err = db_monitor.Exec(str, ymd, rev.GetRoleid(), 1, rev.GetMoney())
	// } else {
	// 	str = fmt.Sprintf("update %s set cash_num=cash_num+1,cash_all=cash_all+? where daynum = %d and userid = %d", tblmoney, ymd, rev.GetRoleid())
	// 	_, err = db_monitor.Exec(str, rev.GetMoney())
	// }
	// if err != nil {
	// 	return false
	// }
	return true

}
func (self *ZoneTask) ParseNotifyRechargeRequestSdkPmd_S(rev *Pmd.NotifyRechargeRequestSdkPmd_S) bool {
	self.Info("ParseNotifyRechargeRequestSdkPmd_S:%s", unibase.GetProtoString(rev.String()))
	if rev.GetPlatorder() == "123456" {
		//测试充值订单不做记录
		return true
	}
	var id, userlevel, isfirst uint32
	tbdata := get_user_data_table(1001)
	str := fmt.Sprintf("select id, userlevel, pay_first_level=0 from %s where userid=?", tbdata)
	row := db_monitor.QueryRow(str, rev.GetRoleid())
	err := row.Scan(&id, &userlevel, &isfirst)
	now := uint32(unitime.Time.Sec())
	if err != nil {
		self.Error("ParseNotifyRechargeRequestSdkPmd_S error:%s", err.Error())
		return false
	}
	if rev.GetRolelevel() != 0 {
		userlevel = rev.GetRolelevel()
	}

	tbpay := get_user_pay_table(1001)
	//daynum := uint32(unitime.Time.Sec() / 86400)
	ymd := uint32(unitime.Time.YearMonthDay())

	str = fmt.Sprintf("insert ignore into %s (zoneid,daynum,userid,gameorder,money,goodid,goodnum,state,platorder, curlevel, isfirst, type) values(?,?,?,?,?,?,?,1,?,?,?,?)", tbpay)
	_, err = db_monitor.Exec(str, self.GetZoneId(), ymd, rev.GetRoleid(), rev.GetGameorder(), int(rev.GetOrdermoney()), rev.GetGoodid(), rev.GetGoodnum(), rev.GetPlatorder(), userlevel, isfirst, rev.GetType())
	if err != nil {
		self.Error("ParseNotifyRechargeRequestSdkPmd_S insert err:%s", err.Error())
	}

	if rev.GetType() == 4 {
		return true //沙箱充值不计入真实充值中
	}

	str = fmt.Sprintf("update %s set pay_last=(case when pay_last_day=? then ? else ? end),  pay_first_level=case when pay_first_level=0 then userlevel else pay_first_level end, pay_all=pay_all + ?,pay_all_num=pay_all_num+1, pay_last_day=? ,pay_last_time=? where id=?", tbdata)
	_, err = db_monitor.Exec(str, ymd, int(rev.GetOrdermoney()), int(rev.GetOrdermoney()), int(rev.GetOrdermoney()), ymd, now, id)
	if err != nil {
		self.Error("ParseNotifyRechargeRequestSdkPmd_S update err:%s", err.Error())
	}

	//str = fmt.Sprintf("update %s set pay_first=pay_first ?, pay_first_day=? where id=? and from_unixtime(firstmin*60, '%%Y%%m%%d')= ? and pay_first=0 and pay_first_day=0", tbdata)
	str = fmt.Sprintf("update %s set pay_first=?, pay_first_day=?, pay_first_time=? where id=? and pay_first=0 and pay_first_day=0", tbdata)
	_, err = db_monitor.Exec(str, int(rev.GetOrdermoney()), ymd, now, id)
	if err != nil {
		self.Error("ParseNotifyRechargeRequestSdkPmd_S update err:%s", err.Error())
	}

	UpdateFirstPaytime(1001, self.GetZoneId(), userlevel, rev.GetData().GetMyaccid(), rev.GetRoleid())

	self.Info("ParseNotifyRechargeRequestSdkPmd_S:%s", unibase.GetProtoString(rev.String()))
	//str = fmt.Sprintf("update %s set pay_last=(case when pay_last_day=? then pay_last + ? else ? end), pay_first_level=case when pay_first_level=0 then userlevel else pay_first_level end, pay_all=pay_all + ?, pay_last_day=? where id=?", tbdata)

	return true
}

func (self *ZoneTask) ParseRefreshServerStateMonitorPmd_CSC(rev *Pmd.RefreshServerStateMonitorPmd_CSC) bool {
	rev.GetState().Gamezone = self.serverState.Gamezone
	//if rev.GetReset() == true {
	if true {
		self.serverState = rev.State
	} else {
		//WHJ TODO 暂时不支持
		self.serverState = rev.State
	}
	callback := func(v entry.EntryInterface) bool {
		task := v.(*MonitorTask)
		//WHJ 如果有指定游戏区，则必须匹配
		if len(task.gameMap) != 0 {
			if _, ok := task.gameMap[1001]; ok == false {
				return true
			}
		}
		self.SendCmd(rev)
		return true
	}
	monitorTaskManager.ExecEvery(callback)
	//self.Info("ParseRefreshServerStateMonitorPmd_CSC:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseStErrorLogMonitorUserCmd_S(rev *Pmd.StErrorLogMonitorUserCmd_S) bool {
	self.Error(rev.GetLogger())
	rev.Gamezone = self.serverState.Gamezone
	rev.Remoteaddr = proto.String(self.GetRemoteAddr())
	monitorTaskManager.Broadcast(rev)
	return true
}

func (self *ZoneTask) ParseStServerShutdownMonitorUserCmd_S(rev *Pmd.StServerShutdownMonitorUserCmd_S) bool {
	self.Error("ParseStServerShutdownMonitorUserCmd_S data:%s", unibase.GetProtoString(rev.String()))
	send := &Pmd.StErrorLogMonitorUserCmd_S{}
	send.Gamezone = self.serverState.Gamezone
	send.Logger = proto.String("服务器宕机")
	send.Remoteaddr = proto.String(self.GetRemoteAddr())
	monitorTaskManager.Broadcast(send)
	return true
}

func (self *ZoneTask) ParseGameOrderListGmUserPmd_CS(rev *Pmd.GameOrderListGmUserPmd_CS) bool {
	//self.Info("ParseGameOrderListGmUserPmd_CS rev:%s", unibase.GetProtoString(rev.String()))
	tblname := get_user_pay_table(1001)
	where := fmt.Sprintf("zoneid=%d and created_at>from_unixtime(%d) and type=0", rev.GetZoneid(), rev.GetTimestamp())
	// where := fmt.Sprintf("zoneid=%d and type=0", rev.GetZoneid())
	charid := rev.GetCharid()
	if charid != 0 {
		where += fmt.Sprintf(" and userid=%d", charid)
	}
	str := fmt.Sprintf("select id, zoneid, userid, gameorder, money, unix_timestamp(`created_at`) from %s where %s", tblname, where)
	self.Debug("sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		self.Error("ParseGameOrderListGmUserPmd_CS query error:%s,sql:%s", err.Error(), str)
		return false
	}
	defer rows.Close()
	orders := make([]*Pmd.OrderInfo, 0)
	for rows.Next() {
		data := &Pmd.OrderInfo{}
		if err := rows.Scan(&data.Id, &data.Zoneid, &data.Charid, &data.Gameorder, &data.Money, &data.Createtime); err != nil {
			self.Error("ParseGameOrderListGmUserPmd_CS error:%s", err.Error())
			continue
		}
		orders = append(orders, data)
	}
	rev.Data = orders
	self.SendCmd(rev)
	self.Debug(unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseLotteryRecordMonitorSmd_C(rev *Smd.LotteryRecordMonitorSmd_C) bool {
	tblname := get_user_lottery_table(1001)
	str := fmt.Sprintf(`insert into %s(zoneid,daynum,userid,type,name,itemtype,itemid,itemcount,itemname,
		optype,opid,opcount,opname,remaincount,created_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), unitime.Time.YearMonthDay(), rev.GetData().GetUserid(),
		rev.GetType(), rev.GetName(), rev.GetItemtype(), rev.GetItemid(), rev.GetItemcount(), rev.GetItemname(),
		rev.GetOptype(), rev.GetOpid(), rev.GetOpcount(), rev.GetOpname(), rev.GetRemaincount(), unitime.Time.Sec())
	if err != nil {
		self.Error("ParseLotteryRecordMonitorSmd_C error:%s, rev:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	return true
}

func (self *ZoneTask) ParseTransactionRecordMonitorSmd_C(rev *Smd.TransactionRecordMonitorSmd_C) bool {
	tblname := get_shop_transaction_table(1001)
	str := fmt.Sprintf(`insert into %s(zoneid,daynum,userid,type,name,itemtype,itemid,itemcount,itemname,
		optype,opid,opcount,opname,remaincount,created_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), unitime.Time.YearMonthDay(), rev.GetData().GetUserid(),
		rev.GetType(), rev.GetName(), rev.GetItemtype(), rev.GetItemid(), rev.GetItemcount(), rev.GetItemname(),
		rev.GetOptype(), rev.GetOpid(), rev.GetOpcount(), rev.GetOpname(), rev.GetRemaincount(), unitime.Time.Sec())
	if err != nil {
		self.Error("ParseTransactionRecordMonitorSmd_C error:%s, rev:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	return true
}

func (self *ZoneTask) ParseRoleTransactionRecordMonitorSmd_C(rev *Smd.RoleTransactionRecordMonitorSmd_C) bool {
	tblname := get_user_transaction_table(1001)
	str := fmt.Sprintf(`insert into %s(zoneid,daynum,userid,sellerid,actid, actname,itemtype,itemid,itemcount,itemname,
		coinid,coincount,coinname,curnum, sellercurnum,created_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId(), unitime.Time.YearMonthDay(), rev.GetData().GetUserid(),
		rev.GetSellerid(), rev.GetActid(), rev.GetActname(), rev.GetItemtype(), rev.GetItemid(), rev.GetItemcount(), rev.GetItemname(),
		rev.GetCoinid(), rev.GetCoinnum(), rev.GetCoinname(), rev.GetCurnum(), rev.GetSellercurnum(), unitime.Time.Sec())
	if err != nil {
		self.Error("ParseRoleTransactionRecordMonitorSmd_C error:%s, rev:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	return true
}

func (self *ZoneTask) ParseExchangeInfoRecordMonitorSmd_C(rev *Smd.ExchangeInfoRecordMonitorSmd_C) bool {
	tblname := get_exchange_info_table(1001)
	recved_at := rev.GetRecvtime()
	if recved_at == 0 {
		recved_at = uint64(unitime.Time.Sec())
	}
	ymd := uint32(unitime.Time.YearMonthDay())
	str := fmt.Sprintf(`insert into %s(zoneid, daynum, charid, charname, recveruid, recvername, amount, created_at, recved_at, agent)
		values(?,?,?,?,?,?,?,?,?,?)`, tblname)
	if _, err := db_monitor.Exec(str, self.GetZoneId(), ymd, rev.GetCharid(), rev.GetCharname(), rev.GetRecveruid(), rev.GetRecvername(),
		rev.GetAmount(), rev.GetCreatetime(), recved_at, rev.GetIsagent()); err != nil {
		self.Error("ParseExchangeInfoRecordMonitorSmd_C error:%s", err.Error())
	}
	return true
}

func (self *ZoneTask) ParseMahjongRecordMonitorSmd_C(rev *Smd.MahjongRecordMonitorSmd_C) bool {
	ymd, now := uint32(unitime.Time.YearMonthDay()), uint32(unitime.Time.Sec())
	tblname := get_mahjong_table(1001, ymd)
	if !check_table_exists(tblname) {
		create_user_mahjong(1001, ymd)
	}
	gameid := rev.GetGameid()
	if gameid == 0 {
		gameid = 1001
	}
	meminfo, _ := json.Marshal(rev.GetRdata())
	str := fmt.Sprintf("insert into %s(daynum, zoneid, charid, charname, platid, charnum, repnum, roomid, groomid, optype, realnum, extdata, created_at, diamond, meminfo) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	if _, err := db_monitor.Exec(str, ymd, gameid, rev.GetData().GetUserid(), rev.GetData().GetUsername(), rev.GetData().GetPlatid(), rev.GetCharnum(),
		rev.GetRepnum(), rev.GetRoomid(), rev.GetGroomid(), rev.GetType(), rev.GetRealnum(), rev.GetExtdata(), now, rev.GetDiamond(), string(meminfo)); err != nil {
		self.Error("ParseMahjongRecordMonitorSmd_C err:%s, data:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	tblname = get_mjpoint_table(ymd)
	if !check_table_exists(tblname) {
		create_user_mjpoint(ymd)
	}
	rounds := rev.GetRealnum()
	for _, data := range rev.GetRdata() {
		rooms, diamond := 0, 0.0
		if data.GetCharid() == rev.GetData().GetUserid() {
			rooms = 1
			diamond = rev.GetDiamond()
		}
		str = fmt.Sprintf(`insert into %s(daynum, gameid, subgameid, charid, point, rooms, rounds, diamond) values(?,?,?,?,?,?,?,?) on DUPLICATE KEY UPDATE point=point+?, rooms=rooms+?, rounds=rounds+?, diamond=diamond+? `, tblname)
		_, err := db_monitor.Exec(str, ymd, 1001, gameid, data.GetCharid(), data.GetPoint(), rooms, rounds, diamond, data.GetPoint(), rooms, rounds, diamond)
		if err != nil {
			self.Error("error:%s, sql:%s, daynum:%d, gameid:%d, subgameid:%d, charid:%d, point:%d", err.Error(), str, ymd, 1001, gameid, data.GetCharid(), data.GetPoint())
		}
	}
	self.Debug("ParseMahjongRecordMonitorSmd_C data:%s", unibase.GetProtoString(rev.String()))
	return true
}

func (self *ZoneTask) ParseRedpackCodeRecordMonitorSmd_C(rev *Smd.RedpackCodeRecordMonitorSmd_C) bool {
	tblname := get_mjredpack_code_table()
	ymd, now := uint32(unitime.Time.YearMonthDay()), uint32(unitime.Time.Sec())
	subgameid := rev.GetGameid()
	if subgameid == 0 {
		subgameid = 1001
	}
	str := fmt.Sprintf("insert into %s(daynum, gameid, subgameid, charid, code, money, created) values(?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, ymd, 1001, subgameid, rev.GetCharid(), rev.GetCode(), rev.GetMoney(), now)
	if err != nil {
		self.Error("ParseRedpackCodeRecordMonitorSmd_C error:%s, data:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	return true
}

func (self *ZoneTask) ParseReturnCountryOnlineNumForwardMonitorSessionCmd(rev *Smd.ReturnCountryOnlineNumForwardMonitorSessionCmd) bool {
	return true
}

func (self *ZoneTask) ParseLevelUpForwardMonitorSessionCmd(rev *Smd.LevelUpForwardMonitorSessionCmd) bool {
	return true
}

func (self *ZoneTask) ParseStUserCmdToAllMonitorForwardMonitorUserCmd(rev *Smd.StUserCmdToAllMonitorForwardMonitorUserCmd) bool {
	return true
}

// 建群统计
func (self *ZoneTask) ParseGroupRecordMonitorSmd_C(rev *Smd.GroupRecordMonitorSmd_C) bool {
	ymd := uint32(unitime.Time.YearMonthDay())
	tblname := get_user_group_table(1001, ymd)
	if !check_table_exists(tblname) {
		create_user_group(1001, ymd)
	}
	optype := rev.GetOptype()
	if optype == 1 {
		str := fmt.Sprintf("insert into %s(zoneid, daynum, userid, groupid, groupname, updated_at) values(?, ?, ?, ?, ?, ?)", tblname)
		if _, err := db_monitor.Exec(str, self.GetZoneId(), ymd, rev.GetData().GetUserid(), rev.GetGroupid(), rev.GetGroupname(), unitime.Time.Sec()); err != nil {
			self.Error("ParseGroupRecordMonitorSmd_C error:%s, data:%s", err.Error(), unibase.GetProtoString(rev.String()))
		}
	} else {
		str := fmt.Sprintf("update %s set state=2 where groupid=%d", tblname, rev.GetGroupid())
		if _, err := db_monitor.Exec(str); err != nil {
			self.Error("ParseGroupRecordMonitorSmd_C error:%s, data:%s", err.Error(), unibase.GetProtoString(rev.String()))
		}
	}
	return true
}

// 红包统计
func (self *ZoneTask) ParseRedpackRecordMonitorSmd_C(rev *Smd.RedpackRecordMonitorSmd_C) bool {
	ymd := uint32(unitime.Time.YearMonthDay())
	tblname := get_user_redpack_table(1001, ymd)
	if !check_table_exists(tblname) {
		create_user_redpack(1001, ymd)
	}
	str := fmt.Sprintf("insert into %s(zoneid,userid,groupid,packid,packtype,coin,packnum,thundernum,created_at) values(?,?,?,?,?,?,?,?,?)", tblname)
	if _, err := db_monitor.Exec(str, self.GetZoneId(), rev.GetData().GetUserid(), rev.GetGroupid(), rev.GetPackid(), rev.GetPacktype(), rev.GetCoin(), rev.GetPacknum(), rev.GetThundernum(), unitime.Time.Sec()); err != nil {
		self.Error("ParseRedpackRecordMonitorSmd_C error:%s, data:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	return true
}

// 分享统计
func (self *ZoneTask) ParseShareRecordMonitorSmd_C(rev *Smd.ShareRecordMonitorSmd_C) bool {
	ymd := uint32(unitime.Time.YearMonthDay())
	tblname := get_user_share_table(1001, ymd)
	if !check_table_exists(tblname) {
		create_user_share(1001, ymd)
	}
	str := fmt.Sprintf("insert into %s(zoneid,daynum,userid,created_at) values(?,?,?,?)", tblname)
	if _, err := db_monitor.Exec(str, self.GetZoneId(), ymd, rev.GetData().GetUserid(), unitime.Time.Sec()); err != nil {
		self.Error("ParseShareRecordMonitorSmd_C error:%s, data:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	return true
}

func (self *ZoneTask) ParseUserCoinOutputConsumeMonitorSmd_C(rev *Smd.UserCoinOutputConsumeMonitorSmd_C) bool {
	self.Info("ParseUserCoinOutputConsumeMonitorSmd_C data:%s", unibase.GetProtoString(rev.String()))
	tblname, ymd := "", uint32(unitime.Time.YearMonthDay())
	if rev.GetType() == 0 {
		tblname = get_output_table(1001, ymd)
		if !check_table_exists(tblname) {
			create_output_table(1001, ymd)
		}
	} else {
		tblname = get_consume_table(1001, ymd)
		if !check_table_exists(tblname) {
			create_consume_table(1001, ymd)
		}
	}
	str := fmt.Sprintf("insert into %s(daynum, subgameid, userid, acttype, typename, actid, actname, actnum, coinid, coinnum, coinleft, level, viplevel, created) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	if _, err := db_monitor.Exec(str, ymd, rev.GetSubgameid(), rev.GetData().GetUserid(), rev.GetActtype(), rev.GetTypename(), rev.GetActid(), rev.GetActname(), rev.GetActnum(), rev.GetCoinid(), rev.GetCoinnum(), rev.GetCoinleft(), rev.GetLevel(), rev.GetViplevel(), rev.GetTime()); err != nil {
		self.Error("ParseUserCoinOutputConsumeMonitorSmd_C error:%s, data:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	return true
}

func (self *ZoneTask) ParseUserChipsMonitorSmd_C(rev *Smd.UserChipsMonitorSmd_C) bool {
	self.Info("ParseUserChipsMonitorSmd_C data:%s", unibase.GetProtoString(rev.String()))
	ymd := uint32(unitime.Time.YearMonthDay())
	tblname := get_chips_table(1001, ymd)
	if !check_table_exists(tblname) {
		create_chips_table(1001, ymd)
	}

	str := fmt.Sprintf("insert into %s(daynum, userid, subgameid, roomtype, roomid, coinid, coinbet, coinnum, coinleft, win, wincard, owncard, point, created) values ", tblname)

	count, params, values := 0, "", make([]interface{}, 0)
	for _, data := range rev.GetData() {
		count += 1
		values = append(values, ymd, data.GetCharid(), rev.GetSubgameid(), rev.GetRoomtype(), rev.GetRoomid(), data.GetCoinid(), data.GetCoinnum(), data.GetCoinwin(), data.GetCoinleft(), data.GetWin(), data.GetWincard(), data.GetOwncard(), data.GetPoint(), rev.GetTime())
		if count >= 100 {
			params = strings.Repeat("(?,?,?,?,?,?,?,?,?,?,?,?,?,?),", count)
			_, err := db_monitor.Exec(fmt.Sprintf("%s %s", str, params[:len(params)-1]), values...)
			if err != nil {
				self.Error("ParseUserChipsMonitorSmd_C Error:%s", err.Error())
			}
			count = 0
			values = values[:0]
		}
	}
	params = strings.Repeat("(?,?,?,?,?,?,?,?,?,?,?,?,?,?),", count)
	_, err := db_monitor.Exec(fmt.Sprintf("%s %s", str, params[:len(params)-1]), values...)
	if err != nil {
		self.Error("ParseUserChipsMonitorSmd_C Error:%s", err.Error())
	}
	params, values = "", nil
	return true
}

func (self *ZoneTask) ParseUserCoinUpdateMonitorSmd_C(rev *Smd.UserCoinUpdateMonitorSmd_C) bool {

	// self.Info("ParseUserCoinUpdateMonitorSmd_C data:%s", unibase.GetProtoString(rev.String()))
	str := fmt.Sprintf("replace into %s(userid, coinid, curnum, updated) values ", get_coin_table(1001))
	userid, params, values := rev.GetData().GetUserid(), strings.Repeat("(?,?,?,?),", len(rev.GetRdata())), make([]interface{}, 0)
	for _, data := range rev.GetRdata() {
		values = append(values, userid, data.GetCoinid(), data.GetCoinnum(), time.Now().Unix())
	}
	if _, err := db_monitor.Exec(fmt.Sprintf("%s %s", str, params[:len(params)-1]), values...); err != nil {
		self.Error("ParseUserCoinUpdateMonitorSmd_C error:%s, data:%s", err.Error(), unibase.GetProtoString(rev.String()))
	}
	return true
}

// 设置所有玩家的离线状态
func (self *ZoneTask) ResetAllUserShutdownOnline() bool {
	tblname := get_user_data_table(1001)
	str := fmt.Sprintf("update %s set isonline=0  where zoneid=?", tblname)
	_, err := db_monitor.Exec(str, self.GetZoneId())
	if err != nil {
		self.Error("ResetAllUserDownOnline update err:%s", err.Error())
	}
	return true
}
