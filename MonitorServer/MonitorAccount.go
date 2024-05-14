package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/unitime"

	sjson "github.com/bitly/go-simplejson"
	"github.com/jordan-wright/email"
)

func HandleAccountCount(data *sjson.Json) error {
	gameid, accid, account, platid := uint32(data.Get("gameid").MustInt()), uint64(data.Get("accid").MustInt64()), data.Get("account").MustString(), data.Get("platid").MustInt()
	if gameid <= 0 || accid <= 0 || account == "" {
		return errors.New("参数错误，游戏ID或账号ID错误")
	}
	return LoginAccount(gameid, accid, account, platid, data.Get("adcode").MustString(), data.Get("adcode").MustString(), 0, data.Get("mac").MustString(), 0, uint32(data.Get("vendorid").MustInt()), "", 1, 0)
}

func GetPlatidByAccid(accid uint64, gameid uint32) (platid uint32) {
	if gameid == 0 || accid == 0 {
		return
	}
	tblname := get_user_account_table(gameid)
	str := fmt.Sprintf("select platid from %s where accid=?", tblname)
	row := db_monitor.QueryRow(str, accid)
	if err := row.Scan(&platid); err != nil {
		logging.Error("GetPlatidByAccid error:%s, gameid:%d, accid:%d", err.Error(), gameid, accid)
	}
	return
}

func HandleRoleCreate(data *sjson.Json) error {
	gameid, zoneid, platid, charid, launchid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint32(data.Get("platid").MustInt()), uint64(data.Get("charid").MustInt64()), uint32(data.Get("launchid").MustInt())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 || launchid <= 0 {
		return errors.New("参数错误,游戏ID、区服ID、投放渠道ID或角色ID错误")
	}
	accid := uint64(data.Get("accid").MustInt64())
	if platid == 0 && accid != 0 {
		platid = GetPlatidByAccid(accid, gameid)
	}
	tblname := get_user_data_table(gameid)

	if !check_table_exists(tblname) {
		create_user_data(gameid)
	}
	str := fmt.Sprintf(`insert ignore into %s (zoneid,userid,username,accid,account,platid,plataccount,ip,
		imei,adcode,firstmin,lastmin,last_levelup,last_login_time,vendorid,packid,reg_time,launchid , ad_account) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, tblname)
	now := uint32(data.Get("curtime").MustInt())
	if now == 0 {
		now = uint32(unitime.Time.Sec())
	}
	packid, vendorid := uint32(data.Get("packid").MustInt()), uint32(data.Get("vendorid").MustInt())
	result, err := db_monitor.Exec(str, zoneid, charid, data.Get("charname").MustString(), accid, data.Get("account").MustString(), platid,
		data.Get("account").MustString(), Ip2Int(data.Get("ip").MustString()), data.Get("mac").MustString(), data.Get("adcode").MustString(), now/60, now/60, now, now, vendorid, packid, now, launchid, data.Get("ad_account").MustString())
	if err == nil {
		_, err = result.LastInsertId()
	}
	if err == nil {
		_, err = UpdateFirstGametime(gameid, zoneid, accid, charid)
	}
	return err
}

// todo登录登出相关处理
func HandleRoleLoginLogout(data *sjson.Json) error {
	optype := data.Get("optype").MustInt()
	if optype == 1 {
		return HandleRoleLogin(data)
	} else if optype == 2 {
		return HandleRoleLogout(data)
	}
	return nil
}

func HandleRoleLogin(data *sjson.Json) error {
	gameid, zoneid, charid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 {
		return errors.New("参数错误,游戏ID、区服ID或角色ID错误")
	}
	level, viplevel, guid := data.Get("level").MustInt(), data.Get("viplevel").MustInt(), data.Get("guid").MustInt()
	power, gold, goldgive, coin := data.Get("power").MustInt(), data.Get("gold").MustInt(), data.Get("goldgive").MustInt(), data.Get("coin").MustInt()
	ctime := uint32(unitime.Time.Sec())
	cdate := uint32(unitime.Time.YearMonthDay())
	tblname := get_user_detail_table(gameid, cdate)
	if !check_table_exists(tblname) {
		create_user_detail(gameid, cdate)
	}

	str := fmt.Sprintf("insert ignore into %s (zoneid,userid,ip,imei,min,sid) values(?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, zoneid, charid, data.Get("ip").MustString(), data.Get("mac").MustString(), ctime/60, data.Get("sid").MustInt())
	if err != nil {
		return err
	}
	ymd1 := uint32(SubDate(int(cdate), 0, 0, -1))
	ymd3 := uint32(SubDate(int(cdate), 0, 0, -3))
	tblname = get_user_data_table(gameid)

	if !check_table_exists(tblname) {
		create_user_data(gameid)
	}

	//兼容老号的情况
	accid := data.Get("accid").MustInt64()
	platid := uint32(data.Get("platid").MustInt())
	if platid == 0 {
		platid = GetPlatidByAccid(uint64(accid), gameid)
	}

	var charid2 uint32
	str = fmt.Sprintf("select userid from %s where userid=?", tblname)
	row := db_monitor.QueryRow(str, charid)
	if err := row.Scan(&charid2); err != nil {
		//没有数据则插入一条，主要解决角色在monitor启动之前创建的情况
		logging.Debug("HandleRoleLogin 时没有角色，插入一条 charid:%d, charname:%s,%s", charid, data.Get("charname").MustString(), err.Error())
		str = fmt.Sprintf(`insert ignore into %s (zoneid,userid,username,accid,account,platid,plataccount,ip,
			imei,adcode,firstmin,lastmin,last_levelup,last_login_time,vendorid,packid) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, tblname)
		now := ctime
		now = uint32(unitime.Time.Sec())
		packid, vendorid := uint32(data.Get("packid").MustInt()), uint32(data.Get("vendorid").MustInt())
		result, err := db_monitor.Exec(str, zoneid, charid, data.Get("charname").MustString(), accid, data.Get("account").MustString(), platid,
			data.Get("account").MustString(), Ip2Int(data.Get("ip").MustString()), data.Get("mac").MustString(), data.Get("adcode").MustString(), now/60, now/60, now, now, vendorid, packid)
		if err == nil {
			_, err = result.LastInsertId()
		}
		if err == nil {
			_, err = UpdateFirstGametime(gameid, zoneid, uint64(accid), charid)
		}
	}

	str = fmt.Sprintf(`
		update %s set logindays=(case from_unixtime(last_login_time, '%%Y%%m%%d') when ? then logindays+1 when ? then logindays else 1 end),
		logintimes=logintimes+1,flag=(case when from_unixtime(last_login_time, '%%Y%%m%%d')<? then 1 else 0 end),
		isonline=1, lastmin=?, isguid=?, userlevel=?, viplevel=?, power=?, gold=?, goldgive=?, money=?, last_login_time=? where zoneid=? and userid=?`, tblname)
	result, err := db_monitor.Exec(str, ymd1, cdate, ymd3, ctime/60, guid, level, viplevel, power, gold, goldgive, coin, ctime, zoneid, charid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func CalcOnlinetime(gameid, zoneid, logouttime uint32, charid uint64) (onlinetime uint32) {
	tblname := get_user_data_table(gameid)
	str := fmt.Sprintf("select last_login_time, last_logout_time, onlinetime from %s where zoneid=? and userid=?", tblname)
	row := db_monitor.QueryRow(str, zoneid, charid)
	var lastlogin, lastlogout, online uint32
	if err := row.Scan(&lastlogin, &lastlogout, &online); err != nil {
		return
	}
	tmptime := MaxUint32(lastlogin, lastlogout)
	if logouttime < tmptime {
		onlinetime = online
	} else {
		onlinetime = ((logouttime - tmptime) / 60) + online
	}
	return
}

func HandleRoleLogout(data *sjson.Json) error {
	gameid, zoneid, charid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 {
		return errors.New("参数错误,游戏ID、区服ID或角色ID错误")
	}
	ctime := uint32(unitime.Time.Sec())
	cdate := uint32(unitime.Time.YearMonthDay())
	onlinetime := uint32(data.Get("onlinetime").MustInt()) / 60
	if onlinetime == 0 {
		onlinetime = CalcOnlinetime(gameid, zoneid, ctime, charid)
	}
	level, viplevel, guid := data.Get("level").MustInt(), data.Get("viplevel").MustInt(), data.Get("guid").MustInt()
	power, gold, coin := data.Get("power").MustInt(), data.Get("gold").MustInt(), data.Get("coin").MustInt()
	tblname := get_user_data_table(gameid)
	str := fmt.Sprintf(`update %s set isonline=0, last_logout_time=?, onlinemin=?, isguid=?, userlevel=?, viplevel=?,
		power=?, gold=?, money=? where zoneid=? and userid=?`, tblname)
	_, err := db_monitor.Exec(str, ctime, onlinetime, guid, level, viplevel, power, gold, coin, zoneid, charid)
	if err != nil {
		return fmt.Errorf("HandleRoleLogout update user_data error:%s, gameid:%d, zoneid:%d, charid:%d", err.Error(), gameid, zoneid, charid)
	}
	tblname = get_user_detail_table(gameid, cdate)
	if !check_table_exists(tblname) {
		tblname = get_user_detail_table(gameid, uint32(SubDate(int(cdate), 0, 0, -1)))
	}
	str = fmt.Sprintf("select id, logoutmin from %s where zoneid=? and userid=? order by id desc limit 1", tblname)
	row := db_monitor.QueryRow(str, zoneid, charid)
	var id, logoutmin int64
	err = row.Scan(&id, &logoutmin)
	if err == nil && id != 0 && logoutmin == 0 {
		str = fmt.Sprintf("update %s set logoutmin=?, onlinetime=(logoutmin-min), sceneid=?, taskid=?, level=? where id=? ", tblname)
		_, err = db_monitor.Exec(str, ctime/60, 0, 0, data.Get("level").MustInt(), id)
	}
	packid := uint32(data.Get("packid").MustInt())
	LogoutAccount(gameid, uint64(data.Get("accid").MustInt64()), charid, packid)
	if err != nil {
		logging.Error("HandleRoleLogout error:%s, sql:%s, gameid:%d, zoneid:%d, charid:%d, id:%d", str, err.Error(), gameid, zoneid, charid, id)
	}
	return nil
}

func CalcLeveltime(gameid, zoneid, leveltime uint32, charid uint64) (usetime uint32) {
	tblname := get_user_data_table(gameid)
	str := fmt.Sprintf("select last_levelup from %s where zoneid=? and userid=?", tblname)
	row := db_monitor.QueryRow(str, zoneid, charid)
	var tmptime uint32
	if err := row.Scan(&tmptime); err != nil {
		logging.Error("CalcLeveltime error:%s, sql:%s, zoneid:%d, charid:%d", err.Error(), zoneid, charid)
		return
	}
	str = fmt.Sprintf("update %s set last_levelup=? where zoneid=? and userid=?", tblname)
	db_monitor.Exec(str, leveltime, zoneid, charid)
	usetime = leveltime - tmptime
	return
}

func HandleRoleLevelup(data *sjson.Json) error {
	gameid, zoneid, charid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64())
	oldlevel, newlevel, leveltime, leveltype, typename := uint32(data.Get("oldlevel").MustInt()), uint32(data.Get("newlevel").MustInt()), uint32(data.Get("duration").MustInt()), uint64(data.Get("actid").MustInt64()), data.Get("actname").MustString()
	if gameid <= 0 || zoneid <= 0 || charid <= 0 || leveltype <= 0 || newlevel <= 0 {
		return errors.New("参数错误")
	}
	tblname := get_user_levelup_table(gameid)
	if !check_table_exists(tblname) {
		create_user_levelup(gameid)
	}
	ctime, daynum := uint32(data.Get("curtime").MustInt()), uint32(data.Get("curdate").MustInt())
	if daynum == 0 {
		daynum = uint32(unitime.Time.YearMonthDay())
	}
	if leveltime == 0 {
		leveltime = CalcLeveltime(gameid, zoneid, ctime, charid)
	}
	str := fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, oldlevel, newlevel, leveltime, leveltype, typename) values(?,?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, zoneid, daynum, charid, oldlevel, newlevel, leveltime, leveltype, typename)

	//更新一下user_data
	tblname = get_user_data_table(gameid)
	str = fmt.Sprintf("update %s set userlevel=? where zoneid=? and userid=?", tblname)
	_, err = db_monitor.Exec(str, newlevel, zoneid, charid)
	if err != nil {
		logging.Error("HandleRoleLevelup update error:%s, sql:%s, zoneid:%d, charid:%d", err.Error(), str, gameid, zoneid)
	}

	return err
}

func HandleRoleRecharge(data *sjson.Json) error {
	gameid, zoneid, charid, gameorder, amount := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64()), data.Get("gameorder").MustString(), uint32(data.Get("amount").MustInt())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 || amount <= 0 {
		return errors.New("参数错误")
	}
	ymd := uint32(data.Get("curdate").MustInt())
	if ymd == 0 {
		ymd = uint32(unitime.Time.YearMonthDay())
	}
	tbdata := get_user_data_table(gameid)
	tbpay := get_user_pay_table(gameid)

	var id, accid uint64
	var user_level uint32
	str := fmt.Sprintf("select id, accid, userlevel from %s where zoneid=? and userid=?", tbdata)
	row := db_monitor.QueryRow(str, zoneid, charid)
	err := row.Scan(&id, &accid, &user_level)
	if err != nil {
		return err
	}

	var isfirst uint32
	str = fmt.Sprintf("select isfirst from %s where zoneid=? and userid=?", tbpay)
	row = db_monitor.QueryRow(str, zoneid, charid)
	err = row.Scan(&isfirst)
	//没有找到，说明是首充
	if err != nil {
		isfirst = 1
	} else {
		isfirst = 0
	}

	str = fmt.Sprintf("insert ignore into %s (zoneid,daynum,userid,gameorder,money,goodid,goodnum,state,platorder, curlevel, isfirst) values(?,?,?,?,?,?,?,1,?,?,?)", tbpay)
	_, err = db_monitor.Exec(str, zoneid, ymd, charid, gameorder, amount, data.Get("goodsid").MustInt(), 1, data.Get("platorder").MustString(), user_level, isfirst)
	if err != nil {
		logging.Error("HandleRoleRecharge insert err:%s", err.Error())
		return err
	}

	gold, goldgive := uint32(data.Get("gold").MustInt()), uint32(data.Get("goldgive").MustInt())
	str = fmt.Sprintf("update %s set gold=gold+?,goldgive=goldgive+?,pay_last=(case when pay_last_day=? then pay_last + ? else ? end), pay_first_level=case when pay_first_level=0 then userlevel else pay_first_level end, pay_all=pay_all + ?, pay_last_day=? where id=?", tbdata)
	_, err = db_monitor.Exec(str, gold, goldgive, ymd, amount, amount, amount, ymd, id)
	if err != nil {
		logging.Error("HandleRoleRecharge update err:%s", err.Error())
	}
	str = fmt.Sprintf("update %s set pay_first=pay_first + ?, pay_first_day=? where id=? and from_unixtime(firstmin*60, '%%Y%%m%%d')= ? ", tbdata)
	_, err = db_monitor.Exec(str, amount, ymd, id, ymd)
	if err == nil {
		UpdateFirstPaytime(gameid, zoneid, user_level, accid, charid)
	}
	tblmoney := get_money_change_table(gameid)

	str = fmt.Sprintf("select count(*) form %s where daynum = %d and userid = %d", tblmoney, ymd, charid)

	var count = 0

	row = db_monitor.QueryRow(str, zoneid, charid)
	err = row.Scan(&count)

	if count == 0 {
		str = fmt.Sprintf("insert ignore into %s (daynum,userid,pay_num,pay_all) values(?,?,?,?)", tblmoney)
		_, err = db_monitor.Exec(str, ymd, charid, 1, amount)
	} else {
		str = fmt.Sprintf("update %s set pay_num=pay_num+1,pay_all=pay_all+? where daynum = %d and userid = %d", tblmoney, ymd, charid)
		_, err = db_monitor.Exec(str, amount)
	}

	if err != nil {
		logging.Error("money_change update err:%s", err.Error())
	}

	return err
}
func HandleCashout(data *sjson.Json) error {
	gameid, zoneid, charid, gameorder, amount := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64()), data.Get("gameorder").MustString(), uint32(data.Get("amount").MustInt())
	ymd := uint32(data.Get("curdate").MustInt())
	if ymd == 0 {
		ymd = uint32(unitime.Time.YearMonthDay())
	}

	tbdata := get_user_data_table(gameid)
	tblcashout := get_user_cashout_table(gameid)

	var id, accid uint64
	var user_level uint32
	str := fmt.Sprintf("select id, accid, userlevel from %s where zoneid=? and userid=?", tbdata)
	row := db_monitor.QueryRow(str, zoneid, charid)
	err := row.Scan(&id, &accid, &user_level)
	if err != nil {
		return err
	}

	str = fmt.Sprintf("insert ignore into %s (zoneid,daynum,userid,gameorder,money,state,platorder) values(?,?,?,?,?,?,?)", tblcashout)
	_, err = db_monitor.Exec(str, zoneid, ymd, charid, gameorder, amount, 1, data.Get("platorder").MustString())
	if err != nil {
		logging.Error("HandleCashout insert err:%s", err.Error())
		return err
	}
	gold := uint32(data.Get("gold").MustInt())
	str = fmt.Sprintf("update %s set gold=gold-?, cash_out_all=cash_out_all + ?, cash_out_num=cash_out_num +1 where id=?", tbdata)
	_, err = db_monitor.Exec(str, gold, amount, id)
	if err != nil {
		logging.Error("HandleCashout update err:%s", err.Error())
	}
	tblmoney := get_money_change_table(gameid)

	str = fmt.Sprintf("select count(*) form %s where daynum = %d and userid = %d", tblmoney, ymd, charid)

	var count = 0

	row = db_monitor.QueryRow(str, zoneid, charid)
	err = row.Scan(&count)

	if count == 0 {
		str = fmt.Sprintf("insert ignore into %s (daynum,userid,cash_num,cash_all) values(?,?,?,?)", tblmoney)
		_, err = db_monitor.Exec(str, ymd, charid, 1, amount)
	} else {
		str = fmt.Sprintf("update %s set cash_num=cash_num+1,cash_all=cash_all+? where daynum = %d and userid = %d", tblmoney, ymd, charid)
		_, err = db_monitor.Exec(str, amount)
	}
	return err
}
func HandleOnlineNum(data *sjson.Json) error {
	gameid, zoneid, onlinenum := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), data.Get("onlinenum").MustInt()
	ext := data.Get("ext").MustString()
	if gameid <= 0 || zoneid <= 0 {
		return errors.New("游戏或区服ID错误")
	}
	var platinfo map[string]interface{}
	if ext != "" {
		tmp, err := sjson.NewJson([]byte(ext))
		if err == nil {
			platinfo = tmp.MustMap()
		}
	}

	curtime, curdate := uint32(data.Get("curtime").MustInt()), uint32(data.Get("curdate").MustInt())

	if curtime == 0 {
		curtime = uint32(unitime.Time.Sec())
	}
	if curdate == 0 {
		curdate = uint32(unitime.Time.YearMonthDay())
	}

	timemin, cdate := uint32(curtime/60), curdate
	if onlinenum == 0 && platinfo != nil {
		for _, num := range platinfo {
			onlinenum += int(num.(float64))
		}
	}
	oninleinfo, _ := json.Marshal(platinfo)
	var number int
	tblname := get_user_online_table(gameid, cdate)
	str1 := fmt.Sprintf("select onlinenum from %s where timestamp_min=? and zoneid=?", tblname)
	row := db_monitor.QueryRow(str1, timemin, zoneid)
	err1 := row.Scan(&number)
	if err1 == sql.ErrNoRows { //查询返回结果集中没有值  需要插入新数据
		str := fmt.Sprintf("insert into %s (zoneid,onlinenum,onlineinfo,timestamp_min) values(?,?,?,?)", tblname)
		_, err := db_monitor.Exec(str, zoneid, onlinenum, string(oninleinfo), timemin)
		if err != nil {
			logging.Error("HandleOnlineNum insert err:%s", err.Error())
		}
		if len(platinfo) > 0 {
			params := strings.Repeat("(?,?,?,?),", len(platinfo))
			values := make([]interface{}, 0)
			for pid, pnum := range platinfo {
				platid, num := uint32(unibase.Atoi(pid, 0)), uint32(unibase.Atoi(pnum.(json.Number).String(), 0))
				values = append(values, zoneid, platid, num, timemin)
			}
			tblname = get_user_online_plat_table(gameid, cdate)
			str = fmt.Sprintf("insert into %s (zoneid, platid, onlinenum, timestamp_min) values %s", tblname, params[:len(params)-1])
			_, err = db_monitor.Exec(str, values...)
			if err != nil {
				logging.Error("HandleOnlineNum error:%s, sql:%s, args:%v", err.Error(), str, values)
			}
		}
		err1 = nil
	} else { //查询返回结果集中有值  更新在线人数就行
		str := fmt.Sprintf("update %s set onlinenum=? where timestamp_min=? and zoneid=?", tblname)
		_, err := db_monitor.Exec(str, onlinenum, timemin, zoneid)
		if err != nil {
			logging.Error("HandleOnlineNum update err:%s", err.Error())
		}
	}

	return err1
}

func HandleEconomicProduceConsume(data *sjson.Json) error {
	gameid, zoneid, charid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 {
		return errors.New("游戏、区服或角色ID错误")
	}
	daynum := uint32(data.Get("curdate").MustInt())

	if daynum == 0 {
		daynum = uint32(unitime.Time.YearMonthDay())
	}
	tblname := get_user_economic_table(gameid, daynum)
	str := fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, coinid, coincount, actionid, actioncount, type, level, actionname, coinname, curcoin) values(?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, zoneid, daynum, charid, data.Get("coinid").MustInt(), data.Get("coinnum").MustInt(), data.Get("actid").MustInt(), 1, data.Get("optype").MustInt()-1, data.Get("level").MustInt(), data.Get("actname").MustString(), data.Get("coinname").MustString(), data.Get("curcoin").MustInt())
	return err
}

func HandleItemProduceConsume(data *sjson.Json) error {
	gameid, zoneid, charid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 {
		return errors.New("游戏、区服或角色ID错误")
	}
	daynum := uint32(data.Get("curdate").MustInt())
	tblname := get_user_item_table(gameid, daynum)
	str := fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, itemtype, itemid, itemcount, actionid, type, level, actionname, itemname, gold, curnum, extdata) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, zoneid, daynum, charid, data.Get("goodstype").MustInt(), data.Get("goodsid").MustInt(), data.Get("goodsnum").MustInt(), data.Get("actid").MustInt(), data.Get("optype").MustInt()-1, data.Get("level").MustInt(), data.Get("actname").MustString(), data.Get("itemname").MustString(), data.Get("gold").MustInt(), data.Get("curnum").MustInt(), data.Get("ext").MustString())
	return err
}

func HandleTransactionRecord(data *sjson.Json) error {
	gameid, zoneid, charid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 {
		return errors.New("游戏、区服或角色ID错误")
	}
	ctime, daynum := uint32(data.Get("curtime").MustInt()), uint32(data.Get("curdate").MustInt())

	if ctime == 0 {
		ctime = uint32(unitime.Time.Sec())
	}
	if daynum == 0 {
		daynum = uint32(unitime.Time.YearMonthDay())
	}

	tblname := get_shop_transaction_table(gameid)
	str := fmt.Sprintf(`insert into %s(zoneid,daynum,userid,type,name,itemtype,itemid,itemcount,itemname,
		optype,opid,opcount,opname,remaincount,created_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, tblname)
	_, err := db_monitor.Exec(str, zoneid, daynum, charid,
		data.Get("actid").MustInt(), data.Get("actname").MustString(), data.Get("goodstype").MustInt(), data.Get("goodsid").MustInt(), data.Get("goodsnum").MustInt(), data.Get("goodsname").MustString(),
		0, data.Get("coinid").MustInt(), data.Get("coinnum").MustInt(), data.Get("coinname").MustString(), data.Get("curnum").MustInt(), ctime)
	return err
}

func HandleActionRecord(data *sjson.Json) error {
	gameid, zoneid, charid, optype := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64()), uint32(data.Get("optype").MustInt())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 || optype <= 0 {
		return errors.New("游戏、区服、角色ID或玩法类型错误")
	}
	daynum := uint32(data.Get("curdate").MustInt())

	if daynum == 0 {
		daynum = uint32(unitime.Time.YearMonthDay())
	}

	tblname := get_action_table(gameid, optype, daynum)
	if len(tblname) == 0 {
		return errors.New("玩法类型错误")
	}
	acttypename := data.Get("typename").MustString()
	if acttypename == "" {
		acttypename = data.Get("actname").MustString()
	}
	starttime, duration := uint32(data.Get("starttime").MustInt()), uint32(data.Get("duration").MustInt())
	str := fmt.Sprintf("insert ignore into %s(zoneid, daynum, userid, actionid, actionname, starttime, duration, endtime, state, power, acttype, acttypename, level, viplevel) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	_, err := db_monitor.Exec(str, zoneid, daynum, charid, data.Get("actid").MustInt(), data.Get("actname").MustString(), starttime, duration, starttime+duration, uint32(data.Get("state").MustInt()), data.Get("power").MustInt(), data.Get("acttype").MustInt(), acttypename, data.Get("level").MustInt(), data.Get("viplevel").MustInt())
	return err
}

func HandleUserTransactionRecord(data *sjson.Json) error {
	gameid, zoneid, charid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt()), uint64(data.Get("charid").MustInt64())
	if gameid <= 0 || zoneid <= 0 || charid <= 0 {
		return errors.New("游戏、区服或角色ID错误")
	}
	ctime, daynum := uint32(data.Get("curtime").MustInt()), uint32(data.Get("curdate").MustInt())
	tblname := get_user_transaction_table(gameid)
	str := fmt.Sprintf(`insert into %s(zoneid,daynum,userid,sellerid,actid,actname,itemtype,itemid,itemcount,itemname,
		coinid,coincount,coinname,curnum, sellercurnum,created_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, tblname)
	_, err := db_monitor.Exec(str, zoneid, daynum, charid, data.Get("sellerid").MustInt(),
		data.Get("actid").MustInt(), data.Get("actname").MustString(), data.Get("goodstype").MustInt(), data.Get("goodsid").MustInt(), data.Get("goodsnum").MustInt(), data.Get("goodsname").MustString(),
		0, data.Get("coinid").MustInt(), data.Get("coinnum").MustInt(), data.Get("coinname").MustString(), data.Get("curnum").MustInt(), data.Get("sellercurnum").MustInt(), ctime)
	return err
}

func HandleGameServerCrash(data *sjson.Json) error {
	gameid, zoneid := uint32(data.Get("gameid").MustInt()), uint32(data.Get("zoneid").MustInt())
	if gameid <= 0 || zoneid <= 0 {
		return errors.New("游戏、区服、角色ID或玩法类型错误")
	}
	daynum := uint32(data.Get("curdate").MustInt())

	if daynum == 0 {
		daynum = uint32(unitime.Time.YearMonthDay())
	}

	tblname := get_game_table()
	str := fmt.Sprintf("select config,gamename from %s where gameid = ?", tblname)
	row := db_monitor.QueryRow(str, gameid)
	config := ""
	game_name := ""
	if err := row.Scan(&config, &game_name); err != nil {
		logging.Error("get game config error:%s", err.Error())
		return err
	}

	//邮件格式{"mail_from":"xxx@qq.com", "mail_key":"rwjoivpqhrkwcbdg", "mail_smtp":"smtp.qq.com", "mail_to_list":"aa@qq.com,bb@qq.com"}

	map_config := make(map[string]string)

	err := json.Unmarshal([]byte(config), &map_config)
	if err != nil {
		logging.Error("get game config error2:%s", err.Error())
		return err
	}

	mail_to_list := strings.Split(string(map_config["mail_to_list"]), ",")
	mail_from := map_config["mail_from"]
	mail_key := map_config["mail_key"]
	mail_smtp := map_config["mail_smtp"]
	mail_content := fmt.Sprintf("游戏:(%s) GameId:%d, ZoneID:%d, 发生不正常关闭，请检查服务器", game_name, int32(gameid), int32(zoneid))

	for _, mail_to := range mail_to_list {

		em := email.NewEmail()
		// 设置 sender 发送方 的邮箱 ， 此处可以填写自己的邮箱
		em.From = mail_from
		// 设置 receiver 接收方 的邮箱  此处也可以填写自己的邮箱， 就是自己发邮件给自己
		em.To = []string{string(mail_to)}
		// 设置主题
		em.Subject = mail_content
		// 简单设置文件发送的内容，暂时设置成纯文本
		em.Text = []byte(mail_content)
		//设置服务器相关的配置
		err = em.Send(mail_smtp+":25", smtp.PlainAuth("", mail_from, mail_key, mail_smtp))
		if err != nil {
			logging.Error("send mail error:%s", err.Error())
		}
	}

	return nil
}
