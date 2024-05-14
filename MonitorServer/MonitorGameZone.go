package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase/unitime"

	// "github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/tidwall/gjson"
)

var (
	g_timer_ten_min  = unitime.NewTimer(unitime.Time.Now(), 10*60*1000, false) //10-1
	g_timer_five_min = unitime.NewTimer(unitime.Time.Now(), 2*60*1000, false)  //5-2
	g_timer_hour_min = unitime.NewTimer(unitime.Time.Now(), 60*60*1000, false)
	g_clock_zero     = unitime.NewClocker(unitime.Time.Now(), 0, 24*3600)
)

func get_zone_table() string {
	return "monitor_zone_info"
}

func create_zone_table() {
	tblname := get_zone_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned not null default '0',
		zoneid int(10) unsigned not null default '0',
		zonename varchar(128) not null default '',
		newzoneid int(10) unsigned not null default '0',
		remarks varchar(256) not null default '',
		state tinyint(2) unsigned not null default '0',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		primary key (id),
		unique key index_gameid_zoneid (gameid,zoneid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
	if !check_column_exists(tblname, "newzoneid") {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN newzoneid int(10) unsigned not null default '0' after zonename;`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add column err:%s,%s,newzoneid", err.Error(), tblname)
		} else {
			logging.Info("monitorserver add column:%s,newzoneid", tblname)
		}
	}
	if !check_column_exists(tblname, "state") {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN state tinyint(2) unsigned not null default '0' after remarks;`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add column err:%s,%s, state", err.Error(), tblname)
		} else {
			logging.Info("monitorserver add column:%s,state", tblname)
		}
	}
}

func get_game_table() string {
	return "monitor_game_info"
}

func create_game_table() {
	tblname := get_game_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned not null default '0',
		gamename varchar(128) not null default '',
		gamekey varchar(64) not null default '',
		config varchar(512) not null default '',
		logolink varchar(256) not null default '',
		remarks varchar(256) not null default '',
		type int(10) unsigned not null default '0',
		conntype tinyint(2) unsigned not null default '1',
		state tinyint(2) unsigned not null default '0',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		primary key (id),
		unique key index_gameid (gameid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
	if !check_column_exists(tblname, "gamekey") {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN gamekey varchar(64) not null default '' after gamename;`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add column err:%s,%s,gamekey", err.Error(), tblname)
		} else {
			logging.Info("monitorserver add column:%s,gamekey", tblname)
		}
	}
	if !check_column_exists(tblname, "conntype") {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN conntype tinyint(2) unsigned not null default '1' after type;`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add column err:%s,%s,conntype", err.Error(), tblname)
		} else {
			logging.Info("monitorserver add column:%s,conntype", tblname)
		}
	}

	if !check_column_exists(tblname, "config") {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN config varchar(512) not null default '' after gamekey;`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add column err:%s,%s,config", err.Error(), tblname)
		} else {
			logging.Info("monitorserver add column:%s,config", tblname)
		}
	}
}

func get_all_gameid() (retl []uint32) {
	table := get_game_table()
	str := fmt.Sprintf("select gameid from %s where state=0 and conntype=2", table)
	rows, err := db_monitor.Query(str)
	if err != nil {
		logging.Error("查询联运游戏ID失败,err:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var gameid int64
		if err = rows.Scan(&gameid); err != nil {
			continue
		}
		retl = append(retl, uint32(gameid))
	}
	return
}

func get_all_zoneid(gameid uint32) (retl []uint32) {
	table := get_zone_table()
	str := fmt.Sprintf("select zoneid from %s where gameid=? and state=0", table)
	rows, err := db_monitor.Query(str, gameid)
	if err != nil {
		logging.Error("查询联运游戏ID失败,err:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var zoneid int64
		if err = rows.Scan(&zoneid); err != nil {
			continue
		}
		retl = append(retl, uint32(zoneid))
	}
	return
}

func get_all_zoneid_by_userdata(gameid uint32) (retl []uint32) {
	tblname := get_user_data_table(gameid)
	str := fmt.Sprintf("select distinct zoneid from %s ", tblname)
	rows, err := db_monitor.Query(str)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var zoneid int64
		if err = rows.Scan(&zoneid); err != nil {
			continue
		}
		retl = append(retl, uint32(zoneid))
	}
	return
}

func initGameTable() {
	gameids := get_all_gameid()
	for _, gameid := range gameids {
		create_user_data(gameid)
		create_user_daily(gameid)
		create_user_imei(gameid)
		create_user_daily_plat(gameid)
		create_user_daily_roi(gameid)
		create_user_pay(gameid)
		create_user_cashout(gameid)
		create_user_daily_subplat(gameid)
		create_user_daily_invite(gameid)
		create_user_levelup(gameid)
		create_user_account_table(gameid)
		create_user_transaction(gameid)
		create_shop_transaction(gameid)
		create_user_lottery(gameid)

		create_user_login(gameid, uint32(unitime.Time.YearMonthDay(-1)))
		create_user_online(gameid, uint32(unitime.Time.YearMonthDay(-1)))
		create_user_detail(gameid, uint32(unitime.Time.YearMonthDay(-1)))
		create_user_online(gameid, uint32(unitime.Time.YearMonthDay()))
		create_user_detail(gameid, uint32(unitime.Time.YearMonthDay()))
		create_realtime_data_table(gameid, uint32(unitime.Time.YearMonthDay(-1)))
		create_realtime_data_table(gameid, uint32(unitime.Time.YearMonthDay()))
		create_user_online_plat(gameid, uint32(unitime.Time.YearMonthDay(-1)))
		create_user_online_plat(gameid, uint32(unitime.Time.YearMonthDay()))

		create_game_device_table(gameid)

		create_action_record(gameid, uint32(unitime.Time.YearMonthDay()))
		create_user_economic(gameid, uint32(unitime.Time.YearMonthDay()))
		create_user_item(gameid, uint32(unitime.Time.YearMonthDay()))

		create_exchange_rate_table(gameid)
		create_launch_keys_table(gameid)
		create_adjapp_table(gameid)
		create_change_registersrc_table(gameid)
		create_app_cost_today_table(gameid)

		create_app_access_table(gameid)

		create_money_change_table(gameid)

		checkDBUpdate(gameid)
	}
}

func HttpLoop() {
	//每分钟计算一次数据
	if g_timer_five_min.Check(unitime.Time.Now()) == true {
		logging.Info("Http接口计算实时数据")
		go func() {
			initGameTable()
			CalcData()
			Warning()

		}()
	}
	if g_timer_ten_min.Check(unitime.Time.Now()) == true {
		go func() {
			// tengetdata()
		}()
	}
	if g_clock_zero.Check(unitime.Time.Now()) == true {
		go func() {
			initGameTable()
			ClockZeroCalc()
		}()
		logging.Info("Http接口零点重置")
	}

	//每小时获取一次广告消耗
	if g_timer_hour_min.Check(unitime.Time.Now()) == true {
		go func() {
			logging.Info("每小时获取一次广告消耗")
			// hourgetadjustreport()
		}()
	}
}

func CalcData() {
	gameids := get_all_gameid()
	for _, gameid := range gameids {
		// zoneids := get_all_zoneid_by_userdata(gameid)
		// for _, zoneid := range zoneids {
		// 	go CalcRealtimeData(gameid, zoneid)
		// }
		go CalcGameRealtimeData(gameid)
		// go CalculateRetention(gameid)
	}
}

func ClockZeroCalc() {
	gameids := get_all_gameid()
	for _, gameid := range gameids {
		go ClockZero(gameid)
		go getExchangeRate(gameid)

		// go getfacebookcost(gameid)
	}
}
func Warning() {
	gameids := get_all_gameid()
	for _, gameid := range gameids {
		go earlyWarning(gameid)
		go getfacebookcost(gameid)

	}

}
func hourgetadjustreport() {
	// gameids := get_all_gameid()
	// for _, gameid := range gameids {
	// 	go getadjustreport(gameid)
	// 	// go costWarning(gameid)
	// }
}

// 计算昨日日新增，次日留存
func CalcRetention(gameid, zoneid uint32) (retm map[string][]interface{}) {
	tblname := get_user_data_table(gameid)
	ymd := uint32(unitime.Time.YearMonthDay())
	ymd2 := uint32(unitime.Time.YearMonthDay(-1))

	str := fmt.Sprintf("select platid, count(distinct userid) from %s where zoneid=? and from_unixtime(firstmin*60, '%%Y%%m%%d')=? group by platid", tblname)
	rows, err := db_monitor.Query(str, zoneid, ymd2)
	if err != nil {
		logging.Error("CalcRetention error:%s, sql:%s, ymd:%d", err.Error(), str, ymd)
		return
	}
	retm = make(map[string][]interface{})
	for rows.Next() {
		var day2 uint32
		var platid string
		if err = rows.Scan(&platid, &day2); err != nil {
			logging.Error("CalcRetention Scan error:%s", err.Error())
			continue
		}
		retm[platid] = []interface{}{day2, 0}
	}
	rows.Close()
	//获取留存数
	str = fmt.Sprintf(`select platid, count(distinct userid) from %s where zoneid=%d
		and from_unixtime(firstmin*60, '%%Y%%m%%d')=? and last_login_time>unix_timestamp(?) group by platid`,
		tblname, zoneid)
	rows, err = db_monitor.Query(str, ymd2, ymd)
	if err != nil {
		logging.Error("CalcRetention error:%s, sql:%s, ymd:%d", err.Error(), str, ymd)
		return
	}
	for rows.Next() {
		var day2 uint32
		var platid string
		if err = rows.Scan(&platid, &day2); err != nil {
			logging.Error("CalcRetention Scan error:%s", err.Error())
			continue
		}
		if _, ok := retm[platid]; !ok {
			retm[platid] = []interface{}{day2, day2}
		} else {
			retm[platid][1] = day2
		}
	}
	rows.Close()
	return
}

func CalcRealtimeData(gameid, zoneid uint32) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("CalcRealtimeData error:%v, gameid:%d, zoneid:%d, time:%d", err, gameid, zoneid, time.Now().Unix())
		}
	}()
	t1 := time.Now()
	//logging.Info("CalcRealtimeData doing, gameid:%d, zoneid:%d, time:%d", gameid, logging.GetZoneId(), t1)
	tblname := get_user_data_table(gameid)
	ymd, now := uint32(unitime.Time.YearMonthDay()), uint32((unitime.Time.Sec()/300)*300)
	//计算各渠道的总人数
	str := fmt.Sprintf("select platid, count(distinct userid) from %s where zoneid=%d group by platid", tblname, zoneid)
	rows, err := db_monitor.Query(str)
	if err != nil {
		logging.Error("CalcRealtimeData error:%s, sql:%s", err.Error(), str)
		return
	}
	retm := make(map[string][]interface{})
	for rows.Next() {
		var platid string
		var total uint32
		if err = rows.Scan(&platid, &total); err != nil {
			logging.Error("CalcRealtimeData error:%s", err.Error())
			continue
		}
		retm[platid] = []interface{}{now, total}
	}
	rows.Close()
	//查询当前在线玩家
	tbonline := get_user_online_table(gameid, ymd)
	str = fmt.Sprintf("select onlinenum from %s where zoneid=%d order by id desc limit 1", tbonline, zoneid)
	row := db_monitor.QueryRow(str)
	var olstr uint32
	retm1 := make(map[string]uint32)
	if err = row.Scan(&olstr); err == nil {
		//err = json.Unmarshal([]byte(olstr), &retm1)
		for k, _ := range retm {
			retm1[k] += olstr
		}
	}
	if err != nil {
		logging.Error("CalcRealtimeData1 check online num error:%s", err.Error())
	}
	//计算各渠道的新增角色，活跃，充值等
	str = fmt.Sprintf(`select platid,count(distinct userid), sum(case pay_first_day when ? then pay_first else 0 end),
		sum(case pay_last_day when ? then pay_last else 0 end) from %s where zoneid=%d and last_login_time>unix_timestamp(?) group by platid`, tblname, zoneid)
	rows, err = db_monitor.Query(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("CalcRealtimeData error:%s, sql:%s, ymd:%d", err.Error(), str, ymd)
		return
	}
	retm2 := make(map[string][]interface{})
	for rows.Next() {
		var dau, paytotal, payfirst uint32
		var platid string
		if err = rows.Scan(&platid, &dau, &payfirst, &paytotal); err != nil {
			logging.Error("CalcRealtimeData Scan error:%s", err.Error())
			continue
		}
		retm2[platid] = []interface{}{dau, paytotal, payfirst}
	}
	rows.Close()
	str = fmt.Sprintf("select platid, count(distinct userid) from %s where zoneid=? and firstmin>(unix_timestamp(?)/60) group by platid", tblname)
	rows, err = db_monitor.Query(str, zoneid, ymd)
	if err != nil {
		logging.Error("CalcRealtimeData error:%s, sql:%s, zoneid:%d", err.Error(), str, zoneid)
		return
	}
	retm4 := make(map[string]int)
	for rows.Next() {
		var platid string
		var newuser int
		if err = rows.Scan(&platid, &newuser); err != nil {
			continue
		}
		retm4[platid] = newuser
	}
	rows.Close()
	str = fmt.Sprintf("select platid, count(distinct userid) from %s where zoneid=? and pay_last_day=? group by platid", tblname)
	rows, err = db_monitor.Query(str, zoneid, ymd)
	if err != nil {
		logging.Error("CalcRealtimeData error:%s, sql:%s, zoneid:%d", err.Error(), str, zoneid)
		return
	}
	retm5 := make(map[string]int)
	for rows.Next() {
		var platid string
		var payuser int
		if err = rows.Scan(&platid, &payuser); err != nil {
			continue
		}
		retm5[platid] = payuser
	}
	rows.Close()

	retm3 := CalcRetention(gameid, zoneid)
	for platid, data := range retm {
		//添加在线人数
		data = append(data, retm1[platid])
		newuser := retm4[platid]
		payuser := retm5[platid]
		//添加新增角色。活跃
		if data2, ok := retm2[platid]; ok {
			data = append(data, newuser, data2[0], payuser, data2[1], data2[2])
		} else {
			data = append(data, newuser, 0, payuser, 0, 0)
		}
		//添加留存数据
		if data3, ok := retm3[platid]; ok {
			data = append(data, data3...)
		} else {
			data = append(data, 0, 0)
		}
		retm[platid] = data
	}
	//platid -- [时间，总玩家数、在线人数、新角色数、DAU、付费人数、付费总额、新增用户付费、昨日新增、次日留存]
	data, err := json.Marshal(retm)
	if err != nil {
		logging.Error("CalcRealtimeData error:%s", err.Error())
		retm, retm1, retm2, retm3 = nil, nil, nil, nil
		return
	}
	d1 := time.Now().Sub(t1)
	logging.Debug("查询区服各平台实时数据用时：%d", d1)

	retm, retm1, retm2, retm3 = nil, nil, nil, nil
	datakey := fmt.Sprintf("RD%d_%d_%d_%d", gameid, zoneid, ymd, now)
	redis_handle.Set(datakey, string(data), 60*60*24*8, 0, false, false)
}

func CalcGameRealtimeData(gameid uint32) {

	defer func() {
		if err := recover(); err != nil {
			logging.Error("CalcRealtimeData error:%v, gameid:%d, zoneid:0, time:%d", err, gameid, time.Now().Unix())
		}
	}()
	tblaccount := get_user_account_table(gameid)
	tblpay := get_user_pay_table(gameid)
	tblcashout := get_user_cashout_table(gameid)

	hours := time.Now().Hour()

	var hour string
	if hours < 10 {
		hour = fmt.Sprintf("0%d", hours)
	} else {
		hour = fmt.Sprintf("%d", hours)
	}

	ymd := time.Now().Format("2006-01-02")
	ymd2 := uint32(unitime.Time.YearMonthDay(0))

	//总玩家
	str := fmt.Sprintf("select platid , launchid , adcode , sid , sum(case  when login>= unix_timestamp(?) and login <= unix_timestamp(? ) then 1 else 0 end) as total , sum(case  when created_at >=  unix_timestamp(?) and created_at <=  unix_timestamp(?) then 1 else 0 end) as newuser, sum(case  when login>= unix_timestamp(?) and login <= unix_timestamp(?) then 1 else 0 end) as dau from %s group by platid , launchid , adcode , sid ", tblaccount)

	rows, _ := db_monitor.Query(str, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", ymd+" "+hour+":00:00", ymd+" "+hour+":59:59")

	realtab := get_realtime_data_table(gameid, ymd2)

	for rows.Next() {

		var platid, total, newuser, dau, sid int32

		var launchid, adcode string

		rows.Scan(&platid, &launchid, &adcode, &sid, &total, &newuser, &dau)

		str1 := fmt.Sprintf("select id from %s where platid = ? and launchid = ? and adcode = ? and sid = ? and type = 0 and hour = ?", realtab)

		id := 0

		idr := db_monitor.QueryRow(str1, platid, launchid, adcode, sid, hour)

		idr.Scan(&id)

		if id > 0 {

			olduser := int32(0)

			if total > newuser {
				olduser = total - newuser
			}

			str2 := fmt.Sprintf("update %s set totaluser=?, newuser=?, olduser=?, dau=? where id = ?", realtab)

			_, err := db_monitor.Exec(str2, total, newuser, olduser, dau, id)

			if err != nil {
				logging.Error("update realtab error:%s, sql:%s", err.Error(), str2)
			}
		} else {

			olduser := int32(0)

			if total > newuser {
				olduser = total - newuser
			}

			str2 := fmt.Sprintf("insert ignore into %s (platid,launchid,adcode,sid,type,totaluser,newuser,olduser,dau , hour) values(?,?,?,?,0,?,?,?,?,?)", realtab)

			_, err := db_monitor.Exec(str2, platid, launchid, adcode, sid, total, newuser, olduser, dau, hour)

			if err != nil {
				logging.Error("update realtab error:%s, sql:%s", err.Error(), str2)
			}
		}
	}

	//充值人数

	strorder := fmt.Sprintf("update %s as a set a.paynum=IFNULL((select paynum from (select  c.platid, count(distinct c.accid) as paynum ,c.launchid ,c.adcode,c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and b.type=0 group by platid, c.launchid , c.adcode,c.sid ) as d where a.platid = d.platid and a.launchid = d.launchid and a.adcode = d.adcode and a.sid = d.sid and a.hour=?) ,0 ) where a.hour=? and a.type =0", realtab, tblpay, tblaccount)

	_, err := db_monitor.Exec(strorder, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update paynum err:%s,%s", str, err.Error())

	}
	//充值金额
	strordermoney := fmt.Sprintf("update %s as a set a.paytotal=IFNULL((select money from (select c.platid, sum(b.money) as money , c.launchid ,c.adcode ,c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and b.type=0 group by platid, c.launchid , c.adcode , c.sid ) as d where a.platid = d.platid and a.hour=? and a.launchid = d.launchid and a.adcode = d.adcode  and a.sid = d.sid  ),0) where a.hour=? and a.type =0", realtab, tblpay, tblaccount)

	_, err = db_monitor.Exec(strordermoney, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update paytotal err:%s,%s", str, err.Error())

	}
	//提现
	strcash := fmt.Sprintf("update %s as a set a.cashnum=IFNULL((select paynum from (select  c.platid, count(distinct c.accid) as paynum ,c.launchid ,c.adcode ,  c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? group by platid, c.launchid , c.adcode , c.sid ) as d where a.platid = d.platid and a.launchid = d.launchid and a.adcode = d.adcode  and a.sid = d.sid  and a.hour=?) ,0 ) where a.hour=?  and a.type =0", realtab, tblcashout, tblaccount)

	_, err = db_monitor.Exec(strcash, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update cashnum1 err:%s,%s", str, err.Error())

	}
	//提现金额
	strcashmoney := fmt.Sprintf("update %s as a set a.cashtotal=IFNULL((select money from (select c.platid, sum(b.money) as money , c.launchid ,c.adcode , c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? group by platid, c.launchid , c.adcode , c.sid ) as d where a.platid = d.platid and a.hour=? and a.launchid = d.launchid and a.adcode = d.adcode and a.sid = d.sid  ),0) where a.hour=?  and a.type =0", realtab, tblcashout, tblaccount)

	_, err = db_monitor.Exec(strcashmoney, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)

	if err != nil {
		logging.Error("CalcGameRealtimeData update cashtotal err:%s,%s", str, err.Error())

	}
	rows.Close()
	//新玩家
	strnew := fmt.Sprintf("select platid , launchid , adcode, sid , sum(case  when login>= unix_timestamp(?) and login <= unix_timestamp(?) then 1 else 0 end) as total , sum(case  when created_at >=  unix_timestamp(?) and created_at <=  unix_timestamp(?)then 1 else 0 end) as newuser, sum(case  when login>= unix_timestamp(?) and login <= unix_timestamp(?) and created_at >=unix_timestamp(?) then 1 else 0 end) as dau from %s group by platid , launchid , adcode , sid", tblaccount)

	rowsnew, _ := db_monitor.Query(strnew, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", ymd+" 00:00:00")

	if err != nil {
		logging.Error("CalcGameRealtimeData select newuser err:%s,%s", str, err.Error())

	}

	for rowsnew.Next() {

		var platid, total, newuser, dau, sid int32

		var launchid, adcode string

		rowsnew.Scan(&platid, &launchid, &adcode, &sid, &total, &newuser, &dau)

		strnew1 := fmt.Sprintf("select id from %s where platid = ? and launchid = ? and adcode = ? and type = 1 and hour = ? and sid = ?", realtab)

		id := 0

		idr := db_monitor.QueryRow(strnew1, platid, launchid, adcode, hour, sid)

		idr.Scan(&id)

		if id > 0 {

			olduser := int32(0)

			if total > newuser {
				olduser = total - newuser
			}

			str2 := fmt.Sprintf("update %s set totaluser=?, newuser=?, olduser=?, dau=? where id = ?", realtab)

			_, err := db_monitor.Exec(str2, total, newuser, olduser, dau, id)

			if err != nil {
				logging.Error("update realtab error:%s, sql:%s", err.Error(), str2)
			}
		} else {

			olduser := int32(0)

			if total > newuser {
				olduser = total - newuser
			}
			str2 := fmt.Sprintf("insert ignore into %s (platid,launchid,adcode,type,totaluser,newuser,olduser,dau , hour , sid) values(?,?,?,1,?,?,?,?,? , ?)", realtab)

			_, err := db_monitor.Exec(str2, platid, launchid, adcode, total, newuser, olduser, dau, hour, sid)

			if err != nil {
				logging.Error("update realtab error:%s, sql:%s", err.Error(), str2)
			}
		}
	}

	//充值人数

	strneworder := fmt.Sprintf("update %s as a set a.paynum=IFNULL((select paynum from (select c.platid, count(distinct c.accid) as paynum ,c.launchid ,c.adcode , c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and b.type=0 and c.created_at >= unix_timestamp('%s') group by platid, c.launchid , c.adcode ,c.sid ) as d where a.platid = d.platid and a.launchid = d.launchid and a.adcode = d.adcode  and a.sid = d.sid and a.hour=?) ,0 ) where hour=?  and type =1", realtab, tblpay, tblaccount, ymd+" 00:00:00")

	_, err = db_monitor.Exec(strneworder, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update paynum err:%s,%s", str, err.Error())

	}
	//充值金额
	strnewordermoney := fmt.Sprintf("update %s as a set a.paytotal=IFNULL((select money from (select c.platid, sum(b.money) as money , c.launchid ,c.adcode ,c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and b.type=0 and c.created_at >= unix_timestamp('%s') group by platid, c.launchid , c.adcode,c.sid ) as d where a.platid = d.platid and a.hour=? and a.launchid = d.launchid and a.adcode = d.adcode and a.sid = d.sid ),0) where hour=?  and type =1", realtab, tblpay, tblaccount, ymd+" 00:00:00")

	_, err = db_monitor.Exec(strnewordermoney, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update paytotal err:%s,%s", str, err.Error())

	}
	//提现
	strnewcash := fmt.Sprintf("update %s as a set a.cashnum=IFNULL((select paynum from (select  c.platid, count(distinct c.accid) as paynum ,c.launchid ,c.adcode , c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and c.created_at >= unix_timestamp('%s') group by platid, c.launchid , c.adcode,c.sid ) as d where a.platid = d.platid and a.launchid = d.launchid and a.adcode = d.adcode and a.sid = d.sid and a.hour=?) ,0 ) where hour=? and type =1", realtab, tblcashout, tblaccount, ymd+" 00:00:00")

	_, err = db_monitor.Exec(strnewcash, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update cashnum2 err:%s,%s", str, err.Error())

	}
	//提现金额
	strnewcashmoney := fmt.Sprintf("update %s as a set a.cashtotal=IFNULL((select money from (select c.platid, sum(b.money) as money , c.launchid ,c.adcode ,c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and c.created_at >= unix_timestamp('%s') group by platid, c.launchid , c.adcode , c.sid ) as d where a.platid = d.platid and a.hour=? and a.launchid = d.launchid and a.adcode = d.adcode and a.sid = d.sid ),0) where hour=? and type =1 ", realtab, tblcashout, tblaccount, ymd+" 00:00:00")
	_, err = db_monitor.Exec(strnewcashmoney, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)

	if err != nil {
		logging.Error("CalcGameRealtimeData update cashtotal err:%s,%s", str, err.Error())

	}
	rowsnew.Close()

	//老玩家

	strold := fmt.Sprintf("select platid , launchid , adcode,sid , sum(case  when login>= unix_timestamp(?) and login <= unix_timestamp(?) then 1 else 0 end) as total , sum(case  when created_at >=  unix_timestamp(?) and created_at <=  unix_timestamp(?) then 1 else 0 end) as newuser, sum(case  when login>= unix_timestamp(?) and login <= unix_timestamp(?) and created_at < unix_timestamp(?) then 1 else 0 end) as dau from %s group by platid , launchid , adcode , sid", tblaccount)

	rowsold, _ := db_monitor.Query(strold, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", ymd+" 00:00:00")

	for rowsold.Next() {

		var platid, total, newuser, dau, sid int32

		var launchid, adcode string

		rowsold.Scan(&platid, &launchid, &adcode, &sid, &total, &newuser, &dau)

		strold1 := fmt.Sprintf("select id from %s where platid = ? and launchid = ? and adcode = ? and type = 2 and hour = ? and sid = ?", realtab)

		id := 0

		idr := db_monitor.QueryRow(strold1, platid, launchid, adcode, hour, sid)

		idr.Scan(&id)

		if id > 0 {
			olduser := int32(0)

			if total > newuser {
				olduser = total - newuser
			}

			str2 := fmt.Sprintf("update %s set totaluser=?, newuser=?, olduser=?, dau=? where id = ?", realtab)

			_, err := db_monitor.Exec(str2, total, newuser, olduser, dau, id)

			if err != nil {
				logging.Error("update realtab error:%s, sql:%s", err.Error(), str2)
			}
		} else {

			olduser := int32(0)

			if total > newuser {
				olduser = total - newuser
			}
			str2 := fmt.Sprintf("insert ignore into %s (platid,launchid,adcode,type,totaluser,newuser,olduser,dau , hour,sid) values(?,?,?,2,?,?,?,?,?,?)", realtab)

			_, err := db_monitor.Exec(str2, platid, launchid, adcode, total, newuser, olduser, dau, hour, sid)

			if err != nil {
				logging.Error("update realtab error:%s, sql:%s", err.Error(), str2)
			}
		}
	}

	//充值人数

	stroldorder := fmt.Sprintf("update %s as a set a.paynum=IFNULL((select paynum from (select  c.platid, count(distinct c.accid) as paynum ,c.launchid ,c.adcode,c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and b.type=0 and c.created_at < unix_timestamp('%s') group by platid, c.launchid , c.adcode ,c.sid ) as d where a.platid = d.platid and a.launchid = d.launchid and a.adcode = d.adcode and a.hour=? and a.sid = d.sid) ,0 ) where hour=?  and type =2", realtab, tblpay, tblaccount, ymd+" 00:00:00")

	_, err = db_monitor.Exec(stroldorder, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update paynum err:%s,%s", str, err.Error())

	}
	//充值金额
	stroldordermoney := fmt.Sprintf("update %s as a set a.paytotal=IFNULL((select money from (select c.platid, sum(b.money) as money , c.launchid ,c.adcode,c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and b.type=0 and c.created_at < unix_timestamp('%s') group by platid, c.launchid , c.adcode,c.sid ) as d where a.platid = d.platid and a.hour=? and a.launchid = d.launchid and a.adcode = d.adcode and a.sid = d.sid ),0) where hour=?  and type =2", realtab, tblpay, tblaccount, ymd+" 00:00:00")

	_, err = db_monitor.Exec(stroldordermoney, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update paytotal err:%s,%s", str, err.Error())

	}
	//提现
	stroldcash := fmt.Sprintf("update %s as a set a.cashnum=IFNULL((select paynum from (select  c.platid, count(distinct c.accid) as paynum ,c.launchid ,c.adcode ,c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and c.created_at < unix_timestamp('%s') group by platid, c.launchid , c.adcode,c.sid ) as d where a.platid = d.platid and a.launchid = d.launchid and a.adcode = d.adcode and a.hour=? and a.sid = d.sid) ,0 ) where hour=? and type =2", realtab, tblcashout, tblaccount, ymd+" 00:00:00")

	_, err = db_monitor.Exec(stroldcash, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)
	if err != nil {
		logging.Error("CalcGameRealtimeData update cashnum3 err:%s,%s", str, err.Error())

	}
	//充值金额
	stroldcashmoney := fmt.Sprintf("update %s as a set a.cashtotal=IFNULL((select money from (select  c.platid, sum(b.money) as money , c.launchid ,c.adcode,c.sid from %s as b inner join %s as c on (b.userid=c.accid) where b.created_at >= ? and  b.created_at <= ? and c.created_at < unix_timestamp('%s') group by platid, c.launchid , c.adcode,c.sid ) as d where a.platid = d.platid and a.hour=? and a.launchid = d.launchid and a.adcode = d.adcode and a.sid=d.sid ),0) where hour=? and type =2 ", realtab, tblcashout, tblaccount, ymd+" 00:00:00")

	_, err = db_monitor.Exec(stroldcashmoney, ymd+" "+hour+":00:00", ymd+" "+hour+":59:59", hour, hour)

	if err != nil {
		logging.Error("CalcGameRealtimeData update cashtotal err:%s,%s", str, err.Error())

	}

	rowsold.Close()

}
func CalculateRetention(gameid uint32) {
	// now := uint32(unitime.Time.YearMonthDay())
	ymd := uint32(unitime.Time.YearMonthDay(-1))
	ymd2 := uint32(unitime.Time.YearMonthDay(-2))
	ymd3 := uint32(unitime.Time.YearMonthDay(-3))
	ymd4 := uint32(unitime.Time.YearMonthDay(-4))
	ymd5 := uint32(unitime.Time.YearMonthDay(-5))
	ymd6 := uint32(unitime.Time.YearMonthDay(-6))
	ymd7 := uint32(unitime.Time.YearMonthDay(-7))
	ymd8 := uint32(unitime.Time.YearMonthDay(-8))
	ymd9 := uint32(unitime.Time.YearMonthDay(-9))
	ymd10 := uint32(unitime.Time.YearMonthDay(-10))
	ymd11 := uint32(unitime.Time.YearMonthDay(-11))
	ymd12 := uint32(unitime.Time.YearMonthDay(-12))
	ymd13 := uint32(unitime.Time.YearMonthDay(-13))
	ymd29 := uint32(unitime.Time.YearMonthDay(-29))
	ymd59 := uint32(unitime.Time.YearMonthDay(-59))
	ymd89 := uint32(unitime.Time.YearMonthDay(-89))
	ymd179 := uint32(unitime.Time.YearMonthDay(-179))

	tbldataname := get_user_data_table(gameid)
	tbldailyplatname := get_user_daily_plat_table(gameid)

	sqlstr := "update %s as a set a.%s=IFNULL((SELECT count from (SELECT count(*) as count , platid , zoneid , launchid , ad_account ,sid from %s where (last_login_time -  reg_time) >= ?*24*60*60 and (last_login_time -  reg_time) < ?*24*60*60 and FROM_UNIXTIME(reg_time , '%%Y%%m%%d') = ? GROUP BY platid , zoneid , launchid , ad_account ,sid ) b where a.platid=b.platid and a.zoneid !=0 and a.zoneid=b.zoneid and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid and a.daynum = ? group by b.platid, b.zoneid,b.launchid,b.ad_account,b.sid),0) where daynum=? and zoneid!=0"
	fdicts := map[string][]interface{}{
		// "day_1":   {0, 1, ymd, ymd, ymd},
		"day_2":   {1, 2, ymd, ymd, ymd},
		"day_3":   {2, 3, ymd2, ymd2, ymd2},
		"day_4":   {3, 4, ymd3, ymd3, ymd3},
		"day_5":   {4, 5, ymd4, ymd4, ymd4},
		"day_6":   {5, 6, ymd5, ymd5, ymd5},
		"day_7":   {6, 7, ymd6, ymd6, ymd6},
		"day_8":   {7, 8, ymd7, ymd7, ymd7},
		"day_9":   {8, 9, ymd8, ymd8, ymd8},
		"day_10":  {9, 10, ymd9, ymd9, ymd9},
		"day_11":  {10, 11, ymd10, ymd10, ymd10},
		"day_12":  {11, 12, ymd11, ymd11, ymd11},
		"day_13":  {12, 13, ymd12, ymd12, ymd12},
		"day_14":  {13, 14, ymd13, ymd13, ymd13},
		"day_30":  {29, 30, ymd29, ymd29, ymd29},
		"day_60":  {59, 60, ymd59, ymd59, ymd59},
		"day_90":  {89, 90, ymd89, ymd89, ymd89},
		"day_180": {179, 180, ymd179, ymd179, ymd179},
	}
	for field, args := range fdicts {
		str := fmt.Sprintf(sqlstr, tbldailyplatname, field, tbldataname)
		_, err := db_monitor.Exec(str, args...)
		if err != nil {
			logging.Error("ClockTen update %s err:%s,%s", field, str, err.Error())
		}
	}
}

func ClockZero(gameid uint32) {
	tblpay := get_user_pay_table(gameid)
	tblcashout := get_user_cashout_table(gameid)
	tbldataname := get_user_data_table(gameid)
	// tblaccname := get_user_account_table(gameid)
	tbldailyplatname := get_user_daily_plat_table(gameid)
	tbldailyroiname := get_user_daily_roi_table(gameid)
	now := uint32(unitime.Time.YearMonthDay())
	ymd := uint32(unitime.Time.YearMonthDay(-1))
	ymd2 := uint32(unitime.Time.YearMonthDay(-2))
	ymd3 := uint32(unitime.Time.YearMonthDay(-3))
	ymd4 := uint32(unitime.Time.YearMonthDay(-4))
	ymd5 := uint32(unitime.Time.YearMonthDay(-5))
	ymd6 := uint32(unitime.Time.YearMonthDay(-6))
	ymd7 := uint32(unitime.Time.YearMonthDay(-7))
	ymd8 := uint32(unitime.Time.YearMonthDay(-8))
	ymd9 := uint32(unitime.Time.YearMonthDay(-9))
	ymd10 := uint32(unitime.Time.YearMonthDay(-10))
	ymd11 := uint32(unitime.Time.YearMonthDay(-11))
	ymd12 := uint32(unitime.Time.YearMonthDay(-12))
	ymd13 := uint32(unitime.Time.YearMonthDay(-13))
	ymd14 := uint32(unitime.Time.YearMonthDay(-14))
	ymd20 := uint32(unitime.Time.YearMonthDay(-20))
	ymd21 := uint32(unitime.Time.YearMonthDay(-21))
	ymd29 := uint32(unitime.Time.YearMonthDay(-29))
	ymd30 := uint32(unitime.Time.YearMonthDay(-30))
	ymd44 := uint32(unitime.Time.YearMonthDay(-44))
	ymd45 := uint32(unitime.Time.YearMonthDay(-45))
	ymd59 := uint32(unitime.Time.YearMonthDay(-59))
	ymd60 := uint32(unitime.Time.YearMonthDay(-60))
	ymd89 := uint32(unitime.Time.YearMonthDay(-89))
	ymd90 := uint32(unitime.Time.YearMonthDay(-90))
	ymd179 := uint32(unitime.Time.YearMonthDay(-179))
	ymd180 := uint32(unitime.Time.YearMonthDay(-180))
	tbldetail := get_user_detail_table(gameid, ymd)
	tblonlinename := get_user_online_table(gameid, ymd)

	logging.Debug("ClockZero start at ymd:%d, time:%d", ymd, unitime.Time.Sec())
	t1 := time.Now()
	//按角色统计
	str := fmt.Sprintf("insert ignore into %s (platid,zoneid,daynum,launchid,ad_account,sid)  select platid, zoneid, ? ,launchid , ad_account,sid from %s  group by platid,zoneid ,launchid, ad_account,sid", tbldailyplatname, tbldataname)
	_, err := db_monitor.Exec(str, ymd)
	if err != nil {
		logging.Error("ClockZero insert err:%s,%s", str, err.Error())
	}

	str = fmt.Sprintf("update %s as a set dau=IFNULL((select count(userid) from %s as b where b.lastmin>=cast(unix_timestamp(?)/60 as signed) and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by zoneid,platid,launchid,ad_account,sid),0) where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update dau err:%s,%s", str, err.Error())
	}
	str = fmt.Sprintf("update %s as a set pcu=(select max(onlinenum) from %s as b where a.zoneid=b.zoneid and a.daynum=? group by b.zoneid) where daynum=?", tbldailyplatname, tblonlinename)
	_, err = db_monitor.Exec(str, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pcu err:%s,%s", str, err.Error())
	}
	//首日充值
	str = fmt.Sprintf("update %s as a set a.pay_first=IFNULL((select sum(pay_all) from %s as b where pay_first_day = ? and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid and FROM_UNIXTIME(b.reg_time,'%%Y%%m%%d') = ? group by b.platid,b.zoneid,launchid,ad_account,sid), 0) where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pay_first err:%s,%s", str, err.Error())
	}
	//首冲人数
	str = fmt.Sprintf("update %s as a set a.pay_first_num=IFNULL((select count(userid) from %s as b where pay_first_day = ? and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid  and FROM_UNIXTIME(b.reg_time,'%%Y%%m%%d') = ? group by b.platid,b.zoneid,launchid,ad_account , sid), 0) where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pay_first_num err:%s,%s", str, err.Error())
	}
	//今日充值
	str = fmt.Sprintf("update %s as a set a.pay_today=IFNULL((select money from (select c.zoneid, c.platid, sum(b.money) as money , c.launchid ,c.ad_account ,c.sid from %s as b inner join %s as c on (b.userid=c.userid and b.zoneid=c.zoneid) where b.daynum=? and b.type=0 group by platid, zoneid , c.launchid , c.ad_account , c.sid ) as d where a.platid = d.platid and a.zoneid=d.zoneid and a.daynum=? and a.launchid = d.launchid and a.ad_account = d.ad_account and a.sid = d.sid ),0) where daynum=? and zoneid!=0", tbldailyplatname, tblpay, tbldataname)
	//str = fmt.Sprintf("update %s as a set a.pay_today=(select sum(pay_last) from %s as b where pay_last_day = ? and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? group by b.platid,b.zoneid) where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pay_today err:%s,%s", str, err.Error())
		//如果出错，再执行一遍老的算法
		str = fmt.Sprintf("update %s as a set a.pay_today=(select sum(pay_last) from %s as b where pay_last_day = ? and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid,b.launchid,b.ad_account , b.sid) where daynum=?", tbldailyplatname, tbldataname)
		_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	}
	//今天充值人数
	str = fmt.Sprintf("update %s as a set a.pay_today_num=IFNULL((select paynum from (select c.zoneid, c.platid, count(distinct c.userid) as paynum ,c.launchid ,c.ad_account , c.sid from %s as b inner join %s as c on (b.userid=c.userid and b.zoneid=c.zoneid) where b.daynum=? and b.type=0 group by platid, zoneid , c.launchid , c.ad_account , c.sid ) as d where a.platid = d.platid and a.zoneid=d.zoneid and a.daynum=? and a.launchid = d.launchid and a.ad_account = d.ad_account and a.sid = d.sid) , 0) where daynum=? and zoneid!=0", tbldailyplatname, tblpay, tbldataname)
	//str = fmt.Sprintf("update %s as a set a.pay_today_num=(select count(userid) from %s as b where pay_last_day = ? and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? group by b.platid,b.zoneid) where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pay_today_num err:%s,%s", str, err.Error())
		str = fmt.Sprintf("update %s as a set a.pay_today_num=(select count(userid) from %s as b where pay_last_day = ? and a.platid = b.platid and a.zoneid=b.zoneid and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid and a.daynum=? group by b.platid,b.zoneid,b.launchid,b.ad_account , b.sid ) where daynum=?", tbldailyplatname, tbldataname)
		_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	}
	d1 := time.Now().Sub(t1)
	logging.Info("ClockZero game:%d part 0:%d", gameid, d1)

	//今日提现
	str = fmt.Sprintf("update %s as a set a.pay_cashout=IFNULL((select money from (select c.zoneid, c.platid, sum(b.money) as money , c.launchid ,c.ad_account , c.sid from %s as b inner join %s as c on (b.userid=c.userid and b.zoneid=c.zoneid) where b.daynum=? group by platid, zoneid , c.launchid , c.ad_account , c.sid ) as d where a.platid = d.platid and a.zoneid=d.zoneid and a.daynum=? and a.launchid = d.launchid and a.ad_account = d.ad_account and a.sid = d.sid ),0) where daynum=? and zoneid!=0", tbldailyplatname, tblcashout, tbldataname)

	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pay_today err:%s,%s", str, err.Error())

	}
	//今日提现人数
	str = fmt.Sprintf("update %s as a set a.pay_cashout_num=IFNULL((select paynum from (select c.zoneid, c.platid, count(distinct c.userid) as paynum ,c.launchid ,c.ad_account , c.sid from %s as b inner join %s as c on (b.userid=c.userid and b.zoneid=c.zoneid) where b.daynum=? group by platid, zoneid , c.launchid , c.ad_account , c.sid ) as d where a.platid = d.platid and a.zoneid=d.zoneid and a.daynum=? and  a.launchid = d.launchid and a.ad_account = d.ad_account and a.sid = d.sid) ,0 )where daynum=? and zoneid!=0", tbldailyplatname, tblcashout, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pay_today_num err:%s,%s", str, err.Error())

	}
	//记录平均登陆次数，平均在线时间等
	str = fmt.Sprintf("update %s as a set a.ol3d=IFNULL((select sum(case when logindays>=3 then 1 else 0 end) from %s as b where lastmin>cast(unix_timestamp(?)/60 as signed) and a.platid = b.platid and a.zoneid=b.zoneid and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid and a.daynum=? group by b.platid,b.zoneid,b.launchid,b.ad_account,b.sid) , 0)where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update ol3d err:%s,%s", str, err.Error())
	}

	str = fmt.Sprintf("update %s as a set a.avglogintimes=IFNULL((select avg(logintimes) from %s as b where lastmin>cast(unix_timestamp(?)/60 as signed) and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid,b.launchid,b.ad_account , b.sid),0) where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update avglogintimes err:%s,%s", str, err.Error())
	}
	str = fmt.Sprintf("update %s as a set a.avgonlinemin=IFNULL((select avg(onlinemin) from %s as b where lastmin>cast(unix_timestamp(?)/60 as signed) and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid,b.launchid,b.ad_account , b.sid),0) where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update avgonlinemin err:%s,%s", str, err.Error())
	}
	str = fmt.Sprintf("update %s as a set onlinemin= 0 ", tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update onlinemin err:%s,%s", str, err.Error())
	}
	str = fmt.Sprintf("update %s as a set a.refluxnum=IFNULL((select sum(flag) from %s as b where lastmin>cast(unix_timestamp(?)/60 as signed) and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid,b.launchid,b.ad_account , b.sid),0) where daynum=?", tbldailyplatname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update refluxnum err:%s,%s", str, err.Error())
	}
	str = fmt.Sprintf("update %s as a set a.day_1=IFNULL((select num from (select d.zoneid, count(*) as num, d.platid , d.launchid , d.ad_account , d.sid from %s as d where FROM_UNIXTIME(d.reg_time , '%%Y%%m%%d') = ? group by platid, zoneid,launchid,ad_account,sid) as b where a.platid=b.platid and a.zoneid!=0 and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by platid, zoneid,launchid,ad_account,sid),0) where daynum=? and zoneid!=0", tbldailyplatname, tbldataname)

	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update day_1 err:%s,%s", str, err.Error())
	}
	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero game:%d part 1:%d", gameid, d1)
	sqlstr := "update %s as a set a.%s=IFNULL((select count(distinct userid) from (select c.zoneid, c.userid, d.platid , d.launchid , d.ad_account , d.sid from %s as c inner join %s as d on (c.zoneid=d.zoneid and c.userid=d.userid) where d.firstmin between cast(unix_timestamp(?)/60 as signed) and cast(unix_timestamp(?)/60 as signed)) as b where a.platid=b.platid and a.zoneid!=0 and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by platid, zoneid,launchid,ad_account,sid),0) where daynum=? and zoneid!=0"
	fdicts := map[string][]interface{}{
		// "day_1":   {ymd, now, ymd, ymd},
		"day_2":   {ymd2, ymd, ymd2, ymd2},
		"day_3":   {ymd3, ymd2, ymd3, ymd3},
		"day_4":   {ymd4, ymd3, ymd4, ymd4},
		"day_5":   {ymd5, ymd4, ymd5, ymd5},
		"day_6":   {ymd6, ymd5, ymd6, ymd6},
		"day_7":   {ymd7, ymd6, ymd7, ymd7},
		"day_8":   {ymd8, ymd7, ymd8, ymd8},
		"day_9":   {ymd9, ymd8, ymd9, ymd9},
		"day_10":  {ymd10, ymd9, ymd10, ymd10},
		"day_11":  {ymd11, ymd10, ymd11, ymd11},
		"day_12":  {ymd12, ymd11, ymd12, ymd12},
		"day_13":  {ymd13, ymd12, ymd13, ymd13},
		"day_14":  {ymd14, ymd13, ymd14, ymd14},
		"day_30":  {ymd30, ymd29, ymd30, ymd30},
		"day_60":  {ymd60, ymd59, ymd60, ymd60},
		"day_90":  {ymd90, ymd89, ymd90, ymd90},
		"day_180": {ymd180, ymd179, ymd180, ymd180},
	}
	for field, args := range fdicts {
		str = fmt.Sprintf(sqlstr, tbldailyplatname, field, tbldetail, tbldataname)
		_, err = db_monitor.Exec(str, args...)
		if err != nil {
			logging.Error("ClockZero update %s err:%s,%s", field, str, err.Error())
		}
	}
	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero game:%d part 2:%d", gameid, d1)

	//统计充值
	sqlstr = "update %s as a set a.%s=IFNULL((select sum(pay_all) from %s as b where (b.firstmin between cast(unix_timestamp(?)/60 as signed) and cast(unix_timestamp(?)/60 as signed)) and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid , b.launchid,b.ad_account , b.sid),0) where daynum=? and zoneid!=0"
	fdicts1 := map[string][]interface{}{
		"pay_1":   {ymd, now, ymd, ymd},
		"pay_2":   {ymd2, ymd, ymd2, ymd2},
		"pay_3":   {ymd3, ymd2, ymd3, ymd3},
		"pay_4":   {ymd4, ymd3, ymd4, ymd4},
		"pay_5":   {ymd5, ymd4, ymd5, ymd5},
		"pay_6":   {ymd6, ymd5, ymd6, ymd6},
		"pay_7":   {ymd7, ymd6, ymd7, ymd7},
		"pay_8":   {ymd8, ymd7, ymd8, ymd8},
		"pay_9":   {ymd9, ymd8, ymd9, ymd9},
		"pay_10":  {ymd10, ymd9, ymd10, ymd10},
		"pay_11":  {ymd11, ymd10, ymd11, ymd11},
		"pay_12":  {ymd12, ymd11, ymd12, ymd12},
		"pay_13":  {ymd13, ymd12, ymd13, ymd13},
		"pay_14":  {ymd14, ymd13, ymd14, ymd14},
		"pay_21":  {ymd21, ymd20, ymd21, ymd21},
		"pay_30":  {ymd30, ymd29, ymd30, ymd30},
		"pay_45":  {ymd45, ymd44, ymd45, ymd45},
		"pay_60":  {ymd60, ymd59, ymd60, ymd60},
		"pay_90":  {ymd90, ymd89, ymd90, ymd90},
		"pay_180": {ymd180, ymd179, ymd180, ymd180},
	}
	for field, args := range fdicts1 {
		str = fmt.Sprintf(sqlstr, tbldailyplatname, field, tbldataname)
		_, err = db_monitor.Exec(str, args...)
		if err != nil {
			logging.Error("ClockZero update %s err:%s,%s", field, str, err.Error())
		}
	}
	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero game:%d part 3:%d", gameid, d1)

	//统计提现
	sqlstr = "update %s as a set a.%s=IFNULL((select sum(cash_out_all) from %s as b where (b.firstmin between cast(unix_timestamp(?)/60 as signed) and cast(unix_timestamp(?)/60 as signed)) and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid , b.launchid,b.ad_account , b.sid),0) where daynum=? and zoneid!=0"
	fdicts2 := map[string][]interface{}{
		"cashout_1":   {ymd, now, ymd, ymd},
		"cashout_2":   {ymd2, ymd, ymd2, ymd2},
		"cashout_3":   {ymd3, ymd2, ymd3, ymd3},
		"cashout_4":   {ymd4, ymd3, ymd4, ymd4},
		"cashout_5":   {ymd5, ymd4, ymd5, ymd5},
		"cashout_6":   {ymd6, ymd5, ymd6, ymd6},
		"cashout_7":   {ymd7, ymd6, ymd7, ymd7},
		"cashout_8":   {ymd8, ymd7, ymd8, ymd8},
		"cashout_9":   {ymd9, ymd8, ymd9, ymd9},
		"cashout_10":  {ymd10, ymd9, ymd10, ymd10},
		"cashout_11":  {ymd11, ymd10, ymd11, ymd11},
		"cashout_12":  {ymd12, ymd11, ymd12, ymd12},
		"cashout_13":  {ymd13, ymd12, ymd13, ymd13},
		"cashout_14":  {ymd14, ymd13, ymd14, ymd14},
		"cashout_21":  {ymd21, ymd20, ymd21, ymd21},
		"cashout_30":  {ymd30, ymd29, ymd30, ymd30},
		"cashout_45":  {ymd45, ymd44, ymd45, ymd45},
		"cashout_60":  {ymd60, ymd59, ymd60, ymd60},
		"cashout_90":  {ymd90, ymd89, ymd90, ymd90},
		"cashout_180": {ymd180, ymd179, ymd180, ymd180},
	}
	for field, args := range fdicts2 {
		str = fmt.Sprintf(sqlstr, tbldailyplatname, field, tbldataname)
		_, err = db_monitor.Exec(str, args...)
		if err != nil {
			logging.Error("ClockZero update %s err:%s,%s", field, str, err.Error())
		}
	}
	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero game:%d fdicts2 :%d", gameid, d1)

	//roi数据
	//按角色统计
	str = fmt.Sprintf("insert ignore into %s (platid,zoneid,daynum,launchid,ad_account,sid)  select platid, zoneid, ? ,launchid , ad_account,sid from %s pay  group by platid,zoneid ,launchid, ad_account,sid ", tbldailyroiname, tbldataname)
	_, err = db_monitor.Exec(str, ymd)
	if err != nil {
		logging.Error("ClockZero insert err:%s,%s", str, err.Error())
	}

	str = fmt.Sprintf("update %s as a set a.new_user=IFNULL((select num from (select d.zoneid, count(*) as num, d.platid , d.launchid , d.ad_account , d.sid from %s as d where FROM_UNIXTIME(d.reg_time , '%%Y%%m%%d') = ? group by platid, zoneid,launchid,ad_account,sid) as b where a.platid=b.platid and a.zoneid!=0 and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by platid, zoneid,launchid,ad_account,sid),0) where daynum=? and zoneid!=0", tbldailyroiname, tbldataname)

	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update day_1 err:%s,%s", str, err.Error())
	}
	str = fmt.Sprintf("update %s as a set a.new_user=IFNULL((select num from (select d.zoneid, count(*) as num, d.platid , d.launchid , d.ad_account , d.sid from %s as d where FROM_UNIXTIME(d.reg_time , '%%Y%%m%%d') = ? group by platid, zoneid,launchid,ad_account,sid) as b where a.platid=b.platid and a.zoneid!=0 and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by platid, zoneid,launchid,ad_account,sid),0) where daynum=? and zoneid!=0", tbldailyroiname, tbldataname)

	_, err = db_monitor.Exec(str, ymd2, ymd2, ymd2)
	if err != nil {
		logging.Error("ClockZero update day_1 err:%s,%s", str, err.Error())
	}
	//首日充值
	str = fmt.Sprintf("update %s as a set a.pay_first=IFNULL((select sum(pay_all) from %s as b where pay_first_day = ? and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid,launchid,ad_account,sid), 0) where daynum=?", tbldailyroiname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pay_first err:%s,%s", str, err.Error())
	}
	//首冲人数
	str = fmt.Sprintf("update %s as a set a.pay_first_num=IFNULL((select count(userid) from %s as b where pay_first_day = ? and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid,launchid,ad_account , sid), 0) where daynum=?", tbldailyroiname, tbldataname)
	_, err = db_monitor.Exec(str, ymd, ymd, ymd)
	if err != nil {
		logging.Error("ClockZero update pay_first_num err:%s,%s", str, err.Error())
	}

	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero game:%d part 2:%d", gameid, d1)

	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero tbldailyroiname game:%d part 7:%d", gameid, d1)
	sqlstr = "update %s as a set a.%s=IFNULL((select count(distinct userid) from (select c.zoneid, c.userid, d.platid , d.launchid , d.ad_account , d.sid from %s as c inner join %s as d on (c.zoneid=d.zoneid and c.userid=d.userid) where d.pay_first_day=?) as b where a.platid=b.platid and a.zoneid!=0 and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by platid, zoneid,launchid,ad_account,sid),0) where daynum= ? and zoneid!=0"
	fdicts = map[string][]interface{}{
		// "day_1":   {ymd, now, ymd, ymd},
		"day_2":   {ymd2, ymd2, ymd2},
		"day_3":   {ymd3, ymd3, ymd3},
		"day_4":   {ymd4, ymd4, ymd4},
		"day_5":   {ymd5, ymd5, ymd5},
		"day_6":   {ymd6, ymd6, ymd6},
		"day_7":   {ymd7, ymd7, ymd7},
		"day_8":   {ymd8, ymd8, ymd8},
		"day_9":   {ymd9, ymd9, ymd9},
		"day_10":  {ymd10, ymd10, ymd10},
		"day_11":  {ymd11, ymd11, ymd11},
		"day_12":  {ymd12, ymd12, ymd12},
		"day_13":  {ymd13, ymd13, ymd13},
		"day_14":  {ymd14, ymd14, ymd14},
		"day_30":  {ymd30, ymd30, ymd30},
		"day_60":  {ymd60, ymd60, ymd60},
		"day_90":  {ymd90, ymd90, ymd90},
		"day_180": {ymd180, ymd180, ymd180},
	}
	for field, args := range fdicts {
		str = fmt.Sprintf(sqlstr, tbldailyroiname, field, tbldetail, tbldataname)
		// fmt.Println("str:", str)
		_, err = db_monitor.Exec(str, args...)
		if err != nil {
			logging.Error("ClockZero update %s err:%s,%s", field, str, err.Error())
		}
	}
	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero game:%d part 2:%d", gameid, d1)
	//统计充值
	sqlstr = "update %s as a set a.%s=IFNULL((select sum(pay_all) from %s as b where b.pay_first_day = ?  and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid , b.launchid,b.ad_account , b.sid),0) where daynum=? and zoneid!=0"
	fdicts1 = map[string][]interface{}{
		"pay_1":   {ymd, ymd, ymd},
		"pay_2":   {ymd2, ymd2, ymd2},
		"pay_3":   {ymd3, ymd3, ymd3},
		"pay_4":   {ymd4, ymd4, ymd4},
		"pay_5":   {ymd5, ymd5, ymd5},
		"pay_6":   {ymd6, ymd6, ymd6},
		"pay_7":   {ymd7, ymd7, ymd7},
		"pay_8":   {ymd8, ymd8, ymd8},
		"pay_9":   {ymd9, ymd9, ymd9},
		"pay_10":  {ymd10, ymd10, ymd10},
		"pay_11":  {ymd11, ymd11, ymd11},
		"pay_12":  {ymd12, ymd12, ymd12},
		"pay_13":  {ymd13, ymd13, ymd13},
		"pay_14":  {ymd14, ymd14, ymd14},
		"pay_21":  {ymd21, ymd21, ymd21},
		"pay_30":  {ymd30, ymd30, ymd30},
		"pay_45":  {ymd45, ymd45, ymd45},
		"pay_60":  {ymd60, ymd60, ymd60},
		"pay_90":  {ymd90, ymd90, ymd90},
		"pay_180": {ymd180, ymd180, ymd180},
	}
	for field, args := range fdicts1 {
		str = fmt.Sprintf(sqlstr, tbldailyroiname, field, tbldataname)
		_, err = db_monitor.Exec(str, args...)
		if err != nil {
			logging.Error("ClockZero update %s err:%s,%s", field, str, err.Error())
		}
	}
	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero game:%d part 4:%d", gameid, d1)

	//统计提现
	sqlstr = "update %s as a set a.%s=IFNULL((select sum(cash_out_all) from %s as b where b.pay_first_day = ?  and a.platid = b.platid and a.zoneid=b.zoneid and a.daynum=? and a.launchid = b.launchid and a.ad_account = b.ad_account and a.sid = b.sid group by b.platid,b.zoneid , b.launchid,b.ad_account , b.sid),0) where daynum=? and zoneid!=0"
	fdicts1 = map[string][]interface{}{
		"cashout_1":   {ymd, ymd, ymd},
		"cashout_2":   {ymd2, ymd2, ymd2},
		"cashout_3":   {ymd3, ymd3, ymd3},
		"cashout_4":   {ymd4, ymd4, ymd4},
		"cashout_5":   {ymd5, ymd5, ymd5},
		"cashout_6":   {ymd6, ymd6, ymd6},
		"cashout_7":   {ymd7, ymd7, ymd7},
		"cashout_8":   {ymd8, ymd8, ymd8},
		"cashout_9":   {ymd9, ymd9, ymd9},
		"cashout_10":  {ymd10, ymd10, ymd10},
		"cashout_11":  {ymd11, ymd11, ymd11},
		"cashout_12":  {ymd12, ymd12, ymd12},
		"cashout_13":  {ymd13, ymd13, ymd13},
		"cashout_14":  {ymd14, ymd14, ymd14},
		"cashout_21":  {ymd21, ymd21, ymd21},
		"cashout_30":  {ymd30, ymd30, ymd30},
		"cashout_45":  {ymd45, ymd45, ymd45},
		"cashout_60":  {ymd60, ymd60, ymd60},
		"cashout_90":  {ymd90, ymd90, ymd90},
		"cashout_180": {ymd180, ymd180, ymd180},
	}
	for field, args := range fdicts1 {
		str = fmt.Sprintf(sqlstr, tbldailyroiname, field, tbldataname)
		_, err = db_monitor.Exec(str, args...)
		if err != nil {
			logging.Error("ClockZero update %s err:%s,%s", field, str, err.Error())
		}
	}
	d1 = time.Now().Sub(t1)
	logging.Info("ClockZero game:%d part 5:%d", gameid, d1)

	logging.Info("ClockZero game:%d part 6:%d", gameid, time.Now().Sub(t1))
	logging.Debug("ClockZero end at ymd:%d, time:%d", ymd, unitime.Time.Sec())
}
func getadjustreport(gameid uint32) {

	logging.Error("开始获取消耗数据")

	h3, _ := time.ParseDuration("-96h")
	h1, _ := time.ParseDuration("-24h")
	sdate := time.Now().Add(h3).Format("2006-01-02")
	edate := time.Now().Add(h1).Format("2006-01-02")
	client := &http.Client{
		Timeout: time.Second * 600,
	}

	url := fmt.Sprintf("https://dash.adjust.com/control-center/reports-service/report?utc_offset=-03:00&dimensions=day,app,campaign_network&metrics=cost,installs&date_period=%s:%s&ad_spend_mode=network", sdate, edate)

	var app_token__in string

	tbladj := get_adjapp_table(gameid)

	str := fmt.Sprintf("select content from %s where id = 1000", tbladj)

	res := db_monitor.QueryRow(str)

	res.Scan(&app_token__in)

	url += "&app_token__in=" + app_token__in

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", "Bearer Mrzx6vznfuPSSgce_9Ru")

	resp, err := client.Do(req)

	if err != nil {
		logging.Error("getadjustreport err:%s", err.Error())
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err == nil {

		tblnamep := get_user_daily_plat_table(gameid)
		tblnamer := get_user_daily_roi_table(gameid)
		result := gjson.Get(string(body), "rows")

		result.ForEach(func(key, value gjson.Result) bool {

			// network := gjson.Get(value.String(), "network")

			// var campaign_network gjson.Result

			// if network.String() == "Kwai for Business" {
			// 	campaign_network = gjson.Get(value.String(), "campaign")

			// } else {
			campaign_network := gjson.Get(value.String(), "campaign_network")
			// }

			cost := gjson.Get(value.String(), "cost")
			day := gjson.Get(value.String(), "day")
			sdate1 := strings.Replace(day.String(), "-", "", -1)

			logging.Error("key : %d , campaign_network:%s ,cost:%.2f , day:%s , sdate1:%d ", key, campaign_network, cost.Float(), day.String(), sdate1)

			if campaign_network.String() != "unknown" && cost.Float() > 0 {

				str1 := fmt.Sprintf(`update %s  set cost= ? where ad_account = "%s" and  daynum = ? and is_examination != 1 and platid < 100 `, tblnamep, campaign_network.String())
				str2 := fmt.Sprintf(`update %s  set cost= ? where ad_account = "%s" and  daynum = ? and is_examination != 1 and platid < 100 `, tblnamer, campaign_network.String())

				_, err = db_monitor.Exec(str1, cost.Float(), sdate1)
				if err != nil {
					logging.Error("getadjustreport %s err:%s,%s", campaign_network.String(), str, err.Error())
				}
				_, err = db_monitor.Exec(str2, cost.Float(), sdate1)
				if err != nil {
					logging.Error("getadjustreport %s err:%s,%s", campaign_network.String(), str, err.Error())
				}
			}
			return true
		})

	}
}

// 拉取当日实时消耗
func costWarning(gameid uint32) {

	fmt.Println("adjust拉取当日实时消耗")
	edate := time.Now().Format("2006-01-02")
	client := &http.Client{
		Timeout: time.Second * 600,
	}

	url := fmt.Sprintf("https://dash.adjust.com/control-center/reports-service/report?utc_offset=-03:00&dimensions=day,app,partner&metrics=cost,installs&date_period=%s:%s&ad_spend_mode=network", edate, edate)

	var app_token__in string

	tbladj := get_adjapp_table(gameid)
	tbldata := get_app_cost_today_table(gameid)

	str := fmt.Sprintf("select content from %s where id = 1000", tbladj)

	res := db_monitor.QueryRow(str)

	res.Scan(&app_token__in)

	url += "&app_token__in=" + app_token__in

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", "Bearer Mrzx6vznfuPSSgce_9Ru")

	resp, err := client.Do(req)

	if err != nil {
		logging.Error("costWarning err:%s", err.Error())
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err == nil {
		result := gjson.Get(string(body), "rows")

		result.ForEach(func(key, value gjson.Result) bool {
			cost := gjson.Get(value.String(), "cost")
			sdate1 := strings.Replace(edate, "-", "", -1)
			partner := gjson.Get(value.String(), "partner")
			app := gjson.Get(value.String(), "app")

			// logging.Info("app : %d , costWarning:%s ,cost:%.2f , sdate1:%d ", app.String(), partner.String(), cost.Float(), sdate1)
			launchid := ""

			if partner.String() == "facebook" {
				launchid = "unattributed"
			} else if partner.String() == "adwords" {
				launchid = "Google Ads ACI"
			} else if partner.String() == "kuaishou_global" {
				launchid = "Kwai for Business"
			}
			var id uint32
			str := fmt.Sprintf("select id from %s where appname = ?", tbladj)

			res := db_monitor.QueryRow(str, app.String())

			res.Scan(&id)

			if launchid != "" {

				str := fmt.Sprintf(`select id from  %s  where launchid = "%s" and  daynum = ? and appid = ? `, tbldata, launchid)
				count := 0
				res := db_monitor.QueryRow(str, sdate1, id)
				res.Scan(&count)
				var str1 string
				if count > 0 {
					str1 = fmt.Sprintf(`update %s set cost= ? where launchid = "%s" and  daynum = ? and appid = ? `, tbldata, launchid)
					_, err = db_monitor.Exec(str1, cost.Float(), sdate1, id)
					if err != nil {
						logging.Error("costWarningupdate %s err:%s,%s", launchid, str, err.Error())
					}
				} else {
					str1 = fmt.Sprintf("insert into %s (appid,launchid,daynum,cost) values(?,?,?,?)", tbldata)

					_, err := db_monitor.Exec(str1, id, launchid, sdate1, cost.Float())

					if err != nil {
						logging.Error("costWarningninsert %s err:%s,%s", launchid, str, err.Error())
					}
				}
			}
			return true
		})

	}

}
func getfacebookcost(gameid uint32) {
	logging.Error("fb实时消耗获取")
	tbldata := get_app_cost_today_table(gameid)
	ymd := uint32(unitime.Time.YearMonthDay())
	date := time.Now().Format("2006-01-02")

	tblname := get_launch_keys_table(gameid)
	tblrate := get_exchange_rate_table(gameid)

	strrate := fmt.Sprintf(`select exchange_rate, currency  from  %s  where time= "%s" `, tblrate, date)

	rateinfo, _ := db_monitor.Query(strrate)

	defer rateinfo.Close()

	var currencyarr = make(map[uint32]float64)

	for rateinfo.Next() {
		var exchange_rate float64
		var currenc uint32

		rateinfo.Scan(&exchange_rate, &currenc)
		if currenc == 1 {
			currencyarr[1] = 1
		} else {
			currencyarr[currenc] = exchange_rate
		}

	}

	str := fmt.Sprintf(`select ad_account_id , token , agent_id , appid , currency  from  %s  where status = 1 and launchid = "unattributed" `, tblname)

	actinfo, _ := db_monitor.Query(str)

	defer actinfo.Close()

	for actinfo.Next() {

		var ad_account_id, token, agent_id, appid string
		var currency uint32

		actinfo.Scan(&ad_account_id, &token, &agent_id, &appid, &currency)
		client := &http.Client{
			Timeout: time.Second * 600,
		}

		url := fmt.Sprintf("https://graph.facebook.com/v18.0/act_%s/insights?time_range={'since':'%s','until':'%s'}&access_token=%s", ad_account_id, date, date, token)
		req, _ := http.NewRequest("GET", url, nil)

		resp, err := client.Do(req)

		if err != nil {
			logging.Error("costWarning err:%s", err.Error())
			return
		}

		body, err := ioutil.ReadAll(resp.Body)

		defer resp.Body.Close()

		if err == nil {
			result := gjson.Get(string(body), "data.#.spend")
			cost := 0.00
			for _, value := range result.Array() {

				cost += value.Float()
				logging.Error("ad_account_id:%s,fb实时消耗获取cost:", ad_account_id, cost)
			}
			strs := fmt.Sprintf(`select count(*) from %s where launchid = "unattributed" and ad_account_id= ? and daynum=?  `, tbldata)

			couns := db_monitor.QueryRow(strs, ad_account_id, ymd)

			var count uint32

			couns.Scan(&count)

			var costsource = cost

			if currency != 1 {
				if _, ok := currencyarr[currency]; ok {

					costsource = cost * float64(currencyarr[currency])
				} else {
					currencyarr[currency] = 1
				}

			} else {
				currencyarr[currency] = 1
			}

			if count > 0 {

				str := fmt.Sprintf(`update %s set cost = ? , currency = ? , rate = ? , appid = ? where launchid = "unattributed" and ad_account_id= ? and daynum=? `, tbldata)
				_, err := db_monitor.Exec(str, costsource, cost, currencyarr[currency], appid, ad_account_id, ymd)
				if err != nil {
					logging.Error("costWarning err:%s", err.Error())
				}

			} else {

				str := fmt.Sprintf("insert into %s set appid = ? , launchid= 'unattributed' , daynum= ? , cost = ? , agent_id = ? , ad_account_id=?, currency = ? , rate = ?", tbldata)
				_, err := db_monitor.Exec(str, appid, ymd, costsource, agent_id, ad_account_id, costsource, currencyarr[currency])
				if err != nil {
					logging.Error("costWarning err:%s", err.Error())
				}

			}
		}
	}
}

func earlyWarning(gameid uint32) bool {

	logging.Error("预警监测")
	tbldata := get_user_data_table(gameid)
	tblname := get_account_table()
	tbladjapp := get_adjapp_table(gameid)

	str1 := fmt.Sprintf(`SELECT id , appsimname , minute , rate FROM %s where status =  1 and id != 1000`, tbladjapp)

	res1, err := db_monitor.Query(str1)

	if err != nil {
		logging.Error("earlyWarning  error:%s", err.Error())
		return false
	}
	defer res1.Close()
	for res1.Next() {
		var appname string

		var id, minute, count1, count2, count3 uint32

		var rate float32

		res1.Scan(&id, &appname, &minute, &rate)

		logging.Error("预警监测 id : %d name: %s", id, appname)

		str2 := fmt.Sprintf(`SELECT sum( case when  reg_time BETWEEN  UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL ? MINUTE)) and UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL ? MINUTE))  then 1 else 0 end) as count1,sum( case when  reg_time BETWEEN  UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL ? MINUTE)) and UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL ? MINUTE))  then 1 else 0 end) as count2,sum( case when  reg_time >=  UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL ? MINUTE))  then 1 else 0 end) as count3 FROM %s where reg_time >=  UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL ? MINUTE)) and launchid in ("Unattributed" , "Google Ads ACI" , "Kwai for Business") and platid = ?`, tbldata)

		res2 := db_monitor.QueryRow(str2, 3*minute, 2*minute, 2*minute, minute, minute, 3*minute, id)

		res2.Scan(&count1, &count2, &count3)

		logging.Error("预警监测 count1 : %d count2: %d count3 : %d", count1, count2, count3)
		if float32(count1)*rate > float32(count3) || float32(count2)*rate > float32(count3) {

			str := fmt.Sprintf("select name from %s ", tblname)

			res, err := db_monitor.Query(str)

			if err != nil {
				logging.Error("earlyWarning  error:%s", err.Error())
				return false
			}

			for res.Next() {
				var phone string

				res.Scan(&phone)

				if phone != "" {
					sendsms(phone, appname)
				}
			}
			defer res.Close()

		}
	}
	return true

}

func sendsms(phone string, appname string) bool {
	logging.Error("发送短信 phone : %s appname: %s ", phone, appname)
	resp, err := http.Get("http://sendsmsapi.lxcyco.com:7050/?appname=" + appname + "&phone=" + phone)
	if err != nil {
		logging.Error("earlyWarning  error:%s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logging.Error("earlyWarning  error:%s", err.Error())
	}

	fmt.Println(string(body))
	return true
}
func getExchangeRate(gameid uint32) {
	daynum := uint32(unitime.Time.YearMonthDay())
	resp, err := http.Get("https://api.exchangerate-api.com/v4/latest/USD")
	if err != nil {
		logging.Error("getExchangeRate  error:%s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logging.Error("getExchangeRate  error:%s", err.Error())
	}

	BRL := gjson.Get(string(body), "rates.BRL")
	EUR := gjson.Get(string(body), "rates.EUR")

	eur_usd, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", float64(1)/float64(EUR.Float())), 32)
	brl_usd, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", float64(1)/float64(BRL.Float())), 32)

	tblname := get_exchange_rate_table(gameid)

	str1 := fmt.Sprintf(`insert into %s set time=?,exchange_rate=?,currency=1 `, tblname)

	_, err = db_monitor.Exec(str1, daynum, BRL.Float())

	str2 := fmt.Sprintf(`insert into %s set time=?,exchange_rate=?,currency=2 `, tblname)

	_, err = db_monitor.Exec(str2, daynum, eur_usd)

	str3 := fmt.Sprintf(`insert into %s set time=?,exchange_rate=?,currency=3 `, tblname)

	_, err = db_monitor.Exec(str3, daynum, brl_usd)

	if err != nil {
		logging.Error("getExchangeRate  error:%s", err.Error())
	}
}
