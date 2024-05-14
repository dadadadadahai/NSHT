package main

import (
	"fmt"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

func initTable() {
	create_subplat_record()
	create_invite_record()
	create_device_table()
	create_game_table()
	create_zone_table()
	create_user_mjredpack_code()
	initGameTable()

}

func checkDBUpdate(gameid uint32) {
	tblname := get_user_data_table(gameid)
	if check_table_exists(tblname) {
		add_uint_column(tblname, "power")
		add_uint_column(tblname, "viplevel")
		add_uint_column(tblname, "goldgive")
		add_uint_column(tblname, "pay_first_level")
		add_uint_column(tblname, "vendorid")
		add_index(tblname, "vendorid")
		add_uint_column(tblname, "packid")
		add_uint_column(tblname, "reg_ip")
		add_uint_column(tblname, "reg_time")
		add_uint_column(tblname, "pay_first_time")
		add_uint_column(tblname, "pay_last_time")

		colnames := []string{"last_levelup", "last_login_time", "last_logout_time", "logintimes", "logindays", "flag"}
		for _, colname := range colnames {
			add_uint_column(tblname, colname)
		}

		indexs := []string{"pay_last_day", "pay_first_day", "accid", "packid"}
		for _, indexname := range indexs {
			add_index(tblname, indexname)
		}

		//新增当前渠道
		add_uint_column(tblname, "curr_platid")
		add_char_column(tblname, "curr_launchid", 64)
		add_char_column(tblname, "curr_ad_account", 64)
		add_uint_column(tblname, "pay_all_num")
	}
	tblname = get_coin_table(gameid)
	if check_table_exists(tblname) {
		add_uint_column(tblname, "updated")
	}

	ymd := uint32(unitime.Time.YearMonthDay())
	tblnames := []string{get_checkpoint_table(gameid, ymd), get_activity_table(gameid, ymd), get_task_table(gameid, ymd), get_battle_table(gameid, ymd)}
	for _, tblname = range tblnames {
		colname := "acttype"
		if check_table_exists(tblname) && !check_column_exists(tblname, colname) {
			str := fmt.Sprintf(`ALTER TABLE %s ADD index index_type (%s) ;`, tblname, colname)
			_, err := db_monitor.Exec(str)
			if err != nil {
				logging.Error("monitorserver add column err:%s,%s,%s", err.Error(), tblname, colname)
			} else {
				logging.Info("monitorserver add column:%s,%s", tblname, colname)
			}
		}
		if check_table_exists(tblname) {
			add_char_column(tblname, "acttypename", 64)
			add_uint_column(tblname, "level")
			add_uint_column(tblname, "viplevel")
			add_uint_column(tblname, "power")
			add_uint_column(tblname, colname)
		}
	}

	tblname = get_mahjong_table(gameid, uint32(unitime.Time.YearMonthDay()))
	if check_table_exists(tblname) {
		add_decimal_column(tblname, "diamond")
		add_bigint_column(tblname, "groomid")
		add_char_column(tblname, "meminfo", 500)
		add_index(tblname, "roomid")
	}

	tblname = get_user_pay_table(gameid)
	if check_table_exists(tblname) {
		add_char_column(tblname, "platorder", 64)

		colnames := []string{"state", "curlevel", "isfirst"}
		for _, colname := range colnames {
			add_uint_column(tblname, colname)
		}
	}

	tblname = get_mjpoint_table(ymd)
	if check_table_exists(tblname) {
		add_uint_column(tblname, "rooms")
		add_uint_column(tblname, "rounds")
		add_decimal_column(tblname, "diamond")
		add_index(tblname, "daynum")
	}

	colnames := []string{"ol3d", "avglogintimes", "avgonlinemin", "refluxnum", "day_4", "day_5", "day_6", "day_14",
		"day_21", "day_60", "day_90", "pay_1", "pay_2", "pay_3", "pay_4", "pay_5", "pay_6", "pay_7", "pay_14",
		"pay_21", "pay_30", "pay_60", "pay_90", "pay_180", "cashout_1", "cashout_2", "cashout_3", "cashout_4", "cashout_5", "cashout_6", "cashout_7",
		"cashout_8", "cashout_9", "cashout_10", "cashout_11", "cashout_12", "cashout_13", "cashout_14", "cashout_21", "cashout_30", "cashout_45", "cashout_60", "cashout_90", "cashout_180",
	}

	tblname = get_user_daily_plat_table(gameid)
	add_decimal_column(tblname, "cost")
	add_int_column(tblname, "is_examination")
	for _, colname := range colnames {
		add_uint_column(tblname, colname)
	}

	tblname = get_user_daily_subplat_table(gameid)
	for _, colname := range colnames {
		add_uint_column(tblname, colname)
	}

	tblname = get_user_levelup_table(gameid)
	if check_table_exists(tblname) {
		add_uint_column(tblname, "leveltype")
		add_char_column(tblname, "typename", 64)

		if !check_column_exists(tblname, "leveltype") {
			str := fmt.Sprintf(`ALTER TABLE %s ADD index index_leveltype (leveltype);`, tblname)
			_, err := db_monitor.Exec(str)
			if err != nil {
				logging.Error("monitor add column err:%s, %s, leveltype", err.Error(), tblname)
			} else {
				logging.Info("monitor add column:%s, leveltype", tblname)
			}
		}
	}

	tblname = get_user_economic_table(gameid, ymd)
	if check_table_exists(tblname) {
		add_char_column(tblname, "actionname", 64)
		add_char_column(tblname, "coinname", 64)
		add_uint_column(tblname, "curcoin")
	}
	tblname = get_user_item_table(gameid, ymd)
	if check_table_exists(tblname) {
		add_char_column(tblname, "actionname", 64)
		add_char_column(tblname, "itemname", 64)
		add_uint_column(tblname, "gold")
		add_uint_column(tblname, "curnum")
		add_char_column(tblname, "extdata", 126)
	}
	tblname = get_user_account_table(gameid)
	if check_table_exists(tblname) {
		add_uint_column(tblname, "firstgamezone")
		add_uint_column(tblname, "vendorid")
		add_index(tblname, "vendorid")
		add_uint_column(tblname, "packid")
		add_index(tblname, "packid")
		add_uint_column(tblname, "firstpayzone")
		add_index(tblname, "login")
	}
	tblname = get_exchange_info_table(gameid)
	if check_table_exists(tblname) {
		if !check_column_exists(tblname, "agent") {
			str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN agent tinyint(2) unsigned not null default '0';`, tblname)
			_, err := db_monitor.Exec(str)
			if err != nil {
				logging.Error("monitor add column err:%s, %s, agent", err.Error(), tblname)
			} else {
				logging.Info("monitor add column:%s, agent", tblname)
			}
			str = fmt.Sprintf(`ALTER TABLE %s ADD index index_agent (agent);`, tblname)
			_, err = db_monitor.Exec(str)
			if err != nil {
				logging.Error("monitor add index err:%s, %s, agent", err.Error(), tblname)
			} else {
				logging.Info("monitor add index:%s, agent", tblname)
			}
		}
	}
	tblname = get_adjapp_table(gameid)
	if check_table_exists(tblname) {
		add_char_column(tblname, "appname", 64)
		add_char_column(tblname, "appsimname", 64)
		add_char_column(tblname, "appid", 64)
		add_uint_column(tblname, "status")
		add_uint_column(tblname, "minute")
		add_decimal_column(tblname, "rate")
	}
	colname := []string{"day_1", "day_2", "day_3", "day_4", "day_5", "day_6", "day_7", "day_8", "day_9", "day_10", "day_11", "day_12", "day_13", "day_14", "day_30", "day_60", "day_90", "day_180"}

	tblname = get_user_daily_roi_table(gameid)
	add_decimal_column(tblname, "cost")
	add_int_column(tblname, "is_examination")
	for _, colname1 := range colname {
		add_uint_column(tblname, colname1)
	}
	tblname = get_launch_keys_table(gameid)
	if check_table_exists(tblname) {
		add_char_column(tblname, "ad_account_id", 64)
		add_char_column(tblname, "agent_id", 64)
		add_uint_column(tblname, "status")
		add_uint_column(tblname, "appid")
		add_uint_column(tblname, "currency")
		add_char_column(tblname, "token", 256)
	}
	tblname = get_app_cost_today_table(gameid)
	if check_table_exists(tblname) {
		add_char_column(tblname, "agent_id", 64)
		add_char_column(tblname, "ad_account_id", 64)
		add_decimal_column(tblname, "currency")
		add_decimal_column(tblname, "rate")

	}
	tblname = get_exchange_rate_table(gameid)
	if check_table_exists(tblname) {
		add_uint_column(tblname, "currency")
	}

}

func get_user_data_table(gameid uint32) string {
	return fmt.Sprintf("user_data_%d", gameid)
}
func get_account_table() string {
	return "monitor_account_info"
}

func check_table_exists(tblname string) bool {
	var count int
	str := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.TABLES where TABLE_SCHEMA='%s' and TABLE_NAME='%s'", config.GetConfigStr("mysql_dbname"), tblname)
	db_monitor.QueryRow(str).Scan(&count)
	if count > 0 {
		return true
	}
	return false
}

func check_column_exists(tblname, colname string) bool {
	var count int
	str := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA='%s' and TABLE_NAME='%s' and COLUMN_NAME='%s'", config.GetConfigStr("mysql_dbname"), tblname, colname)
	db_monitor.QueryRow(str).Scan(&count)
	if count > 0 {
		return true
	}
	return false
}

func check_index_exists(tblname, indexname string) bool {
	var count int
	str := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.STATISTICS where TABLE_SCHEMA='%s' and TABLE_NAME='%s' and INDEX_NAME='%s'", config.GetConfigStr("mysql_dbname"), tblname, indexname)
	db_monitor.QueryRow(str).Scan(&count)
	if count > 0 {
		return true
	}
	return false
}

func add_index(tblname, index string) error {
	if !check_index_exists(tblname, index) {
		str := fmt.Sprintf("ALTER TABLE %s ADD INDEX %s(%s);", tblname, index, index)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add index err:%s,%s,%s", err.Error(), tblname, index)
		} else {
			logging.Info("monitorserver add index:%s,%s", tblname, index)
		}
		return err
	}
	return nil
}

func add_decimal_column(tblname, column string) error {
	if !check_column_exists(tblname, column) {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s decimal(10, 2) not null DEFAULT '0' ;`, tblname, column)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add column err:%s,%s,%s", err.Error(), tblname, column)
		} else {
			logging.Info("monitorserver add column:%s,%s", tblname, column)
		}
		return err
	}
	return nil
}

func add_char_column(tblname, column string, width int) error {
	if !check_column_exists(tblname, column) {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s varchar(%d) not null default '';`, tblname, column, width)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitor add column err:%s, %s, %s", err.Error(), tblname, column)
		} else {
			logging.Info("monitor add column:%s, %s", tblname, column)
		}
		return err
	}
	return nil
}

func add_bigint_column(tblname, column string) error {
	if !check_column_exists(tblname, column) {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s bigint(20) unsigned not null default '0';`, tblname, column)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitor add column err:%s, %s, %s", err.Error(), tblname, column)
		} else {
			logging.Info("monitor add column:%s, %s", tblname, column)
		}
		return err
	}
	return nil
}

func add_int_column(tblname, column string) error {
	if !check_column_exists(tblname, column) {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s int(11) not null default '0';`, tblname, column)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitor add column err:%s, %s, %s", err.Error(), tblname, column)
		} else {
			logging.Info("monitor add column:%s, %s", tblname, column)
		}
		return err
	}
	return nil
}

func add_uint_column(tblname, column string) error {
	if !check_column_exists(tblname, column) {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s int(10) unsigned not null default '0';`, tblname, column)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitor add column err:%s, %s, %s", err.Error(), tblname, column)
		} else {
			logging.Info("monitor add column:%s, %s", tblname, column)
		}
		return err
	}
	return nil
}

func get_table_names(tblprefix string) (tblnames []string) {
	str := fmt.Sprintf("show tables like '%s'", tblprefix)
	rows, err := db_monitor.Query(str)
	if err != nil {
		logging.Warning("get_table_names error:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tblname string
		if err := rows.Scan(&tblname); err != nil {
			logging.Warning("get_table_names err:%s", err.Error())
			continue
		}
		tblnames = append(tblnames, tblname)
	}
	return
}

func update_userid_accid() {
	str := fmt.Sprintf("show tables")
	rows, err := db_monitor.Query(str)
	if err != nil {
		logging.Warning("show tables error:%s", err.Error())
		return
	}
	defer rows.Close()
	tblnames := make([]string, 0)
	for rows.Next() {
		var tblname string
		if err := rows.Scan(&tblname); err != nil {
			logging.Warning("update_userid_accid error:%s", err.Error())
			continue
		}
		tblnames = append(tblnames, tblname)
	}
	for _, tblname := range tblnames {
		if check_column_exists(tblname, "userid") {
			str = fmt.Sprintf("alter table %s change userid userid bigint(20) unsigned not null default '0' ", tblname)
			_, err = db_monitor.Exec(str)
			if err != nil {
				logging.Warning("change userid error:%s, %s", tblname, err.Error())
			}
		}
		if check_column_exists(tblname, "accid") {
			str = fmt.Sprintf("alter table %s change accid accid bigint(20) unsigned not null default '0' ", tblname)
			_, err = db_monitor.Exec(str)
			if err != nil {
				logging.Warning("change accid error:%s, %s", tblname, err.Error())
			}
		}
	}
}

func update_user_data() {
	tblnames := get_table_names("user_data_%")
	for _, tblname := range tblnames {
		str := fmt.Sprintf("alter table %s add adcode varchar(32) not null default '', add sid tinyint(2) unsigned not null default '0',add  isguid tinyint(2) unsigned not null default '0', add index adcode(adcode);", tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Warning("update_user_data error1:%s, %s", err.Error(), tblname)
		}
		str = fmt.Sprintf("alter table %s modify column isguid int(10) unsigned not null default 0", tblname)
		_, err = db_monitor.Exec(str)
		if err != nil {
			logging.Warning("update_user_data error2:%s, %s", err.Error(), tblname)
		}
	}
}

func update_user_pay() {
	tblnames := get_table_names("user_pay_%")
	for _, tblname := range tblnames {
		str := fmt.Sprintf("alter table %s add type tinyint(2) unsigned not null default '0'", tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Warning("update_user_pay error:%s, %s", err.Error(), tblname)
		}
	}
}

func update_user_detail() {
	tblnames := get_table_names("user_detail_%")
	for _, tblname := range tblnames {
		str := fmt.Sprintf("alter table %s add logoutmin int(10) unsigned not null default '0', add onlinetime int(10) unsigned not null default '0', add sceneid int(10) unsigned not null default '0', add taskid int(10) unsigned not null default '0', add level int(10) unsigned not null default '0', add sid tinyint(2) unsigned not null default '0';", tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Warning("update_user_detail error:%s, %s", err.Error(), tblname)
		}
	}
}

func check_table_empty(tblname string) bool {
	//logging.Error("check_table_empty:%s", tblname)
	str := fmt.Sprintf("select * from %s limit 1", tblname)
	rows, err := db_monitor.Query(str)
	if err != nil {
		return false
	}
	defer rows.Close()
	if err == nil && rows.Next() == false {
		return true
	}
	return false
}

func get_user_account_table(gameid uint32) string {
	return fmt.Sprintf("user_account_%d", gameid)
}

// 账号ID，账号，平台ID，vendorid（投放平台），adcode广告码，登陆IP，登陆IME，创建时间，登录时间，登出时间，在线时间，
// 当天登陆次数，连续登陆天数，首次游戏时间，首次付费时间
func create_user_account_table(gameid uint32) {
	tblname := get_user_account_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		daynum int(10) unsigned not null default '0',
		accid bigint(20) unsigned not null default '0',
		account varchar(256) not null default '',
		platid int(10) unsigned not null default '0',
		vendorid int(10) unsigned not null default '0',
		launchid varchar(32) not null default '',
		adcode varchar(32) not null default '',
		sid tinyint(2) unsigned not null default '0',
		mobilenum varchar(32) not null default '',
		mobiletime int(10) unsigned NOT NULL DEFAULT '0',
		ip int(10) unsigned NOT NULL DEFAULT '0',
		imei varchar(64) COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
		created_at int(10) unsigned NOT NULL DEFAULT '0',
		login int(10) unsigned not null default '0',
		logout int(10) unsigned not null default '0',
		onlinetime int(10) unsigned not null default '0',
		logintimes int(10) unsigned not null default '0',
		logindays int(10) unsigned not null default '0',
		firstgametime int(10) unsigned not null default '0',
		firstgamezone int(10) unsigned not null default '0',
		firstpaytime int(10) unsigned not null default '0',
		firstpayzone int(10) unsigned not null default '0',
		firstpaylevel int(10) unsigned not null default '0',
		flag int(10) unsigned not null default '0',
		isonline tinyint(2) not null default '0',
		packid int(10) unsigned not null default '0',
		PRIMARY KEY (id),
		UNIQUE KEY index_accid (accid),
		KEY index_daynum (daynum),
		KEY index_vendor (vendorid),
		KEY index_platid (platid),
		KEY index_game_zone (firstgamezone),
		KEY packid (packid),
		KEY index_pay_zone (firstpayzone)
	)ENGINE=MyISAM AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Error("db_monitor create table error:%s, %s", err.Error(), tblname)
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func LogoutAccount(gameid uint32, accid, charid uint64, packid uint32) {
	tblname := get_user_account_table(gameid)
	now := uint32(unitime.Time.Sec())
	ymd := uint32(unitime.Time.YearMonthDay())
	if accid == 0 {
		accid = charid
	}
	tbldata := get_user_data_table(gameid)
	str := fmt.Sprintf("select sum(onlinemin) from %s where accid=?", tbldata)
	row := db_monitor.QueryRow(str, accid)
	var onlinemin int64
	row.Scan(&onlinemin)

	str = fmt.Sprintf(`update %s set isonline=0,logout=?,onlinetime=?, packid=(case packid when 0 then ? else packid end) where accid=?`, tblname)
	result, err := db_monitor.Exec(str, now, onlinemin*60, packid, accid)
	var affected int64
	if err == nil {
		affected, err = result.RowsAffected()
	}
	if affected == 0 || err != nil {
		logging.Error("LogoutAccount error:%v, sql:%s, %d,%d,%d,%d", err, str, now, ymd, now, accid)
	}
}

func UpdateFirstGametime(gameid uint32, zoneid uint32, accid, charid uint64) (int64, error) {
	tblname := get_user_account_table(gameid)
	ymd := uint32(unitime.Time.YearMonthDay())
	if accid == 0 {
		accid = charid
	}
	str := fmt.Sprintf("select id from %s where accid=?", tblname)
	row := db_monitor.QueryRow(str, accid)
	var aid int64
	var err error
	err = row.Scan(&aid)
	if aid == 0 {
		return 0, err
	} else {
		str = fmt.Sprintf("update %s set firstgametime=?, firstgamezone=? where id = ? and firstgametime=0", tblname)
		_, err = db_monitor.Exec(str, ymd, zoneid, aid)
	}
	return 1, err
}

func UpdateFirstPaytime(gameid, zoneid, level uint32, accid, charid uint64) error {
	tblname := get_user_account_table(gameid)
	ymd := uint32(unitime.Time.YearMonthDay())
	if accid == 0 {
		accid = charid
	}
	str := fmt.Sprintf("update %s set firstpaytime=?, firstpayzone=?, firstpaylevel=? where accid = ? and firstpaytime=0", tblname)
	_, err := db_monitor.Exec(str, ymd, zoneid, level, accid)
	return err
}

func LoginAccount(gameid uint32, accid uint64, account string, platid int, adcode string, launchid string, ip uint32, imei string, zoneid, vendorid uint32, mobilenum string, sid uint32, update uint32) error {

	tblname := get_user_account_table(gameid)
	platid, _ = GetPlatidFromAccount(platid, account)

	now := uint32(unitime.Time.Sec())
	ymd := uint32(unitime.Time.YearMonthDay())
	ymd1 := uint32(unitime.Time.YearMonthDay(-1))
	ymd3 := uint32(unitime.Time.YearMonthDay(-3))

	var aid int64
	var mobile string
	var err error
	str := fmt.Sprintf("select id , mobilenum from %s where accid=?", tblname)
	row := db_monitor.QueryRow(str, accid)
	err = row.Scan(&aid, &mobile)

	// logging.Error("LoginAccount err:%s,", err.Error())

	if aid == 0 {

		if !check_table_exists(tblname) {
			create_user_account_table(gameid)
		}
		var firstgametime, firstgamezone uint32
		if zoneid != 0 {
			firstgametime = ymd
			firstgamezone = zoneid
		}
		if mobilenum != "" {
			str = fmt.Sprintf(`insert ignore into %s (daynum, accid, account, platid, adcode, launchid, ip, imei,created_at,login,logintimes,logindays,isonline,firstgametime,firstgamezone,vendorid , mobilenum , mobiletime)
			values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?,?,?,?,?,?)`, tblname)
			_, err = db_monitor.Exec(str, ymd, accid, account, platid, adcode, launchid, ip, imei, now, now, 1, 1, 1, firstgametime, firstgamezone, vendorid, mobilenum, now)
			if err != nil {
				logging.Error("checkInsertAccount error:%s, gameid:%d, zoneid:%d, platid:%d, accid:%d", err.Error(), gameid, zoneid, platid, accid)
			}
		} else {
			str = fmt.Sprintf(`insert ignore into %s (daynum, accid, account, platid, adcode, launchid, ip, imei,created_at,login,logintimes,logindays,isonline,firstgametime,firstgamezone,vendorid , sid)
				values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?,?,?,?,?)`, tblname)
			_, err = db_monitor.Exec(str, ymd, accid, account, platid, adcode, launchid, ip, imei, now, now, 1, 1, 1, firstgametime, firstgamezone, vendorid, sid)
			if err != nil {
				logging.Error("checkInsertAccount error:%s, gameid:%d, zoneid:%d, platid:%d, accid:%d", err.Error(), gameid, zoneid, platid, accid)
			}
		}

	} else {

		if mobilenum != "" && mobile == "" {
			str = fmt.Sprintf(`update %s set logintimes=logintimes+1,
			logindays=(case from_unixtime(login, '%%Y%%m%%d') when ? then logindays+1 when ? then logindays else 1 end),
			flag=(case when from_unixtime(login, '%%Y%%m%%d')<? then 1 else 0 end), isonline=1, login=? ,mobilenum = ? , mobiletime = ?  where id=?`, tblname)
			_, err = db_monitor.Exec(str, ymd1, ymd, ymd3, now, mobilenum, now, aid)
			if err != nil {
				logging.Error("checkInsertAccount update error:%s, gameid:%d, zoneid:%d, platid:%d, accid:%d", err.Error(), gameid, zoneid, platid, accid)
			}
		} else {
			str = fmt.Sprintf(`update %s set logintimes=logintimes+1,
			logindays=(case from_unixtime(login, '%%Y%%m%%d') when ? then logindays+1 when ? then logindays else 1 end),
			flag=(case when from_unixtime(login, '%%Y%%m%%d')<? then 1 else 0 end), isonline=1, login=?  where id=?`, tblname)
			_, err = db_monitor.Exec(str, ymd1, ymd, ymd3, now, aid)
			if err != nil {
				logging.Error("checkInsertAccount update error:%s, gameid:%d, zoneid:%d, platid:%d, accid:%d", err.Error(), gameid, zoneid, platid, accid)
			}
		}
		if update == 1 {
			str = fmt.Sprintf(`update %s set platid=?,  adcode=?, launchid=? where id=?`, tblname)
			_, err = db_monitor.Exec(str, platid, adcode, launchid, aid)
			if err != nil {
				logging.Error("checkInsertAccount platid update error:%s, gameid:%d, zoneid:%d, platid:%d, accid:%d", err.Error(), gameid, zoneid, platid, accid)
			}
		}

	}

	return err
}

func create_user_data(gameid uint32) {
	tblname := get_user_data_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		launchid varchar(64) not null default '' ,
		ad_account varchar(64) not null default '',
		username varchar(256) not null default '',
		accid bigint(20) unsigned not null default '0',
		account varchar(256) not null default '',
		platid int(10) unsigned not null default '0',
		plataccount varchar(256) not null default '',
		vendorid int(10) unsigned not null default '0',
		ip int(10) unsigned not null default '0',
		imei varchar(64) not null default '',
		userlevel int(10) unsigned not null default '1',
		viplevel int(10) unsigned not null default '0',
		money int(10) unsigned not null default '0',
		gold int(10) unsigned not null default '0',
		goldgive int(10) unsigned not null default '0',
		power int(10) unsigned not null default '0',
		firstmin int(10) unsigned not null default '0',
		lastmin int(10) unsigned not null default '0',
		onlinemin int(10) unsigned default '1',

		pay_first int(10) unsigned not null default '0',
		pay_last int(10) unsigned not null default '0',
		pay_all int(10) unsigned not null default '0',
		pay_first_day int(10) unsigned not null default '0',
		pay_last_day int(10) unsigned not null default '0',

		onlineday int(10) unsigned not null default '1',
		mobile varchar(64) not null default '',
		qq bigint(13) unsigned not null default '0',
		mail varchar(48) not null default '',
		initsubplat int(10) unsigned not null default '0',
		cursubplat int(10) unsigned not null default '0',
		inviterid int(10) unsigned not null default '0',
		isonline int(10) unsigned default '1',
		adcode varchar(32) not null default '',
		sid tinyint(2) unsigned not null default '0',
		isguid int(10) unsigned not null default '0',
		pay_first_level int(10) unsigned not null default '0',
		last_levelup int(10) unsigned not null default '0',
		last_login_time int(10) unsigned not null default '0',
		last_logout_time int(10) unsigned not null default '0',
		logintimes int(10) unsigned not null default '0',
		logindays int(10) unsigned not null default '0',
		flag int(10) unsigned not null default '0',
		packid int(10) unsigned not null default '0',
		reg_ip int(10) unsigned not null default '0',
		reg_time int(10) unsigned not null default '0',
		pay_first_time int(10) unsigned not null default '0',
		pay_last_time int(10) unsigned not null default '0',
		cash_out_all int(10) unsigned not null default '0',
		cash_out_num int(10) unsigned not null default '0',
		updated_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		primary key (id),
		unique key index_userid_platid (userid,zoneid),
		key zoneid (zoneid),
		key accid (accid),
		key userlevel (userlevel),
		key platid (platid),
		key vendorid (vendorid),
		key lastmin (lastmin),
		key firstmin (firstmin),
		key cursubplat (cursubplat),
		key initsubplat (initsubplat),
		key inviterid (inviterid),
		key pay_first_day (pay_first_day),
		key pay_last_day (pay_last_day),
		key packid (packid),
		key adcode (adcode),
		key regtime (reg_time)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_online_table(gameid, daynum uint32) string {
	return fmt.Sprintf("user_online_%d_%d", gameid, daynum)
}

func get_user_login_table(gameid, daynum uint32) string {
	return fmt.Sprintf("user_login_%d_%d", gameid, daynum)
}

func check_empty_user_online(gameid, daynum uint32) bool {
	tblname := get_user_online_table(gameid, daynum)
	if check_table_exists(tblname) == true && check_table_empty(tblname) == true {
		str := fmt.Sprintf(`
		drop table IF EXISTS %s;
		`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func check_empty_action_record(gameid, daynum uint32) bool {
	tblnames := []string{get_action_table(gameid, 1, daynum), get_action_table(gameid, 2, daynum), get_action_table(gameid, 3, daynum), get_action_table(gameid, 4, daynum)}
	for _, tblname := range tblnames {
		if check_table_exists(tblname) == true && check_table_empty(tblname) == true {
			str := fmt.Sprintf(`
			drop table IF EXISTS %s;
			`, tblname)
			_, err := db_monitor.Exec(str)
			if err != nil {
				return false
			}
			return true
		}
	}
	return false
}

func check_empty_user_economic(gameid, daynum uint32) bool {
	tblname := get_user_economic_table(gameid, daynum)
	if check_table_exists(tblname) == true && check_table_empty(tblname) == true {
		str := fmt.Sprintf(`
		drop table IF EXISTS %s;
		`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func check_empty_item_table(gameid, daynum uint32) bool {
	tblname := get_user_item_table(gameid, daynum)
	if check_table_exists(tblname) == true && check_table_empty(tblname) == true {
		str := fmt.Sprintf(`
		drop table IF EXISTS %s;
		`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func create_user_login(gameid, daynum uint32) {
	tblname := get_user_login_table(gameid, daynum)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		accountid bigint(20)  null default '0',
		accountname varchar(128) not null default '',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		primary key (id),
		key index_zoneid(zoneid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func create_user_online(gameid, daynum uint32) {
	tblname := get_user_online_table(gameid, daynum)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		onlinenum int(10) unsigned not null default '0',
		onlineinfo varchar(512) not null default '',
		timestamp_min int(10) unsigned not null default '0',
		primary key (id),
		unique key index_zoneid_timestamp_min (zoneid,timestamp_min)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_online_plat_table(gameid, daynum uint32) string {
	return fmt.Sprintf("user_online_plat_%d_%d", gameid, daynum)
}

func create_user_online_plat(gameid, daynum uint32) {
	tblname := get_user_online_plat_table(gameid, daynum)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		platid int(10) unsigned not null default '0',
		onlinenum int(10) unsigned not null default '0',
		timestamp_min int(10) unsigned not null default '0',
		primary key (id),
		unique key index_zoneid_timestamp_min (zoneid,platid,timestamp_min)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_imei_table(gameid uint32) string {
	return fmt.Sprintf("user_imei_%d", gameid)
}

func create_user_imei(gameid uint32) {
	tblname := get_user_imei_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		accountid bigint(20)  null default '0',
		accountname varchar(128) not null default '',
		imei varchar(128) not null default '',
		osname varchar(128) not null default '',
		primary key (id),
		unique key index_zoneid (zoneid,imei)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_daily_table(gameid uint32) string {
	return fmt.Sprintf("user_daily_%d", gameid)
}

func create_user_daily(gameid uint32) {
	tblname := get_user_daily_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		dau int(10) unsigned not null default '0',
		pcu int(10) unsigned not null default '0',
		arpu int(10) unsigned not null default '0',
		day_1 int(10) unsigned not null default '0',
		day_2 int(10) unsigned not null default '0',
		day_3 int(10) unsigned not null default '0',
		day_7 int(10) unsigned not null default '0',
		day_30 int(10) unsigned not null default '0',
		pay_first int(10) unsigned not null default '0',
		pay_today int(10) unsigned not null default '0',
		pay_first_num int(10) unsigned not null default '0',
		pay_today_num int(10) unsigned not null default '0',
		primary key (id),
		unique key index_zoneid (zoneid,daynum)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_daily_plat_table(gameid uint32) string {
	return fmt.Sprintf("user_daily_plat_%d", gameid)
}

func create_user_daily_plat(gameid uint32) {
	tblname := get_user_daily_plat_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		platid int(10) unsigned not null default '0',
		zoneid int(10) unsigned not null default '0',
		launchid varchar(64) not null default '' ,
		ad_account varchar(64) not null default '',
		sid tinyint(2) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		dau int(10) unsigned not null default '0',
		pcu int(10) unsigned not null default '0',
		arpu int(10) unsigned not null default '0',
		
		pay_first int(10) unsigned not null default '0',
		pay_today int(10) unsigned not null default '0',
		pay_first_num int(10) unsigned not null default '0',
		pay_today_num int(10) unsigned not null default '0',
		pay_cashout int(10) unsigned not null default '0',
		pay_cashout_num int(10) unsigned not null default '0',

		day_1 int(10) unsigned not null default '0',
		day_2 int(10) unsigned not null default '0',
		day_3 int(10) unsigned not null default '0',
		day_4 int(10) unsigned not null default '0',
		day_5 int(10) unsigned not null default '0',
		day_6 int(10) unsigned not null default '0',
		day_7 int(10) unsigned not null default '0',
		day_8 int(10) unsigned not null default '0',
		day_9 int(10) unsigned not null default '0',
		day_10 int(10) unsigned not null default '0',
		day_11 int(10) unsigned not null default '0',
		day_12 int(10) unsigned not null default '0',
		day_13 int(10) unsigned not null default '0',
		day_14 int(10) unsigned not null default '0',
		day_21 int(10) unsigned not null default '0',
		day_30 int(10) unsigned not null default '0',
		day_60 int(10) unsigned not null default '0',
		day_90 int(10) unsigned not null default '0',
		day_180 int(10) unsigned not null default '0',
		pay_1 int(10) unsigned not null default '0',
		pay_2 int(10) unsigned not null default '0',
		pay_3 int(10) unsigned not null default '0',
		pay_4 int(10) unsigned not null default '0',
		pay_5 int(10) unsigned not null default '0',
		pay_6 int(10) unsigned not null default '0',
		pay_7 int(10) unsigned not null default '0',
		pay_8 int(10) unsigned not null default '0',
		pay_9 int(10) unsigned not null default '0',
		pay_10 int(10) unsigned not null default '0',
		pay_11 int(10) unsigned not null default '0',
		pay_12 int(10) unsigned not null default '0',
		pay_13 int(10) unsigned not null default '0',
		pay_14 int(10) unsigned not null default '0',
		pay_21 int(10) unsigned not null default '0',
		pay_30 int(10) unsigned not null default '0',
		pay_45 int(10) unsigned not null default '0',
		pay_60 int(10) unsigned not null default '0',
		pay_90 int(10) unsigned not null default '0',
		pay_180 int(10) unsigned not null default '0',
		
		primary key (id),
		unique key index_zoneid_platid (platid,zoneid,daynum,launchid,ad_account),
		key platid (platid),
		key launchid (launchid),
		key daynum (daynum)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_daily_subplat_table(gameid uint32) string {
	return fmt.Sprintf("user_daily_subplat_%d", gameid)
}

func create_user_daily_subplat(gameid uint32) {
	tblname := get_user_daily_subplat_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		subplatid int(10) unsigned not null default '0',
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		dau int(10) unsigned not null default '0',
		pcu int(10) unsigned not null default '0',
		arpu int(10) unsigned not null default '0',
		day_1 int(10) unsigned not null default '0',
		day_2 int(10) unsigned not null default '0',
		day_3 int(10) unsigned not null default '0',
		day_7 int(10) unsigned not null default '0',
		day_30 int(10) unsigned not null default '0',
		pay_first int(10) unsigned not null default '0',
		pay_today int(10) unsigned not null default '0',
		pay_first_num int(10) unsigned not null default '0',
		pay_today_num int(10) unsigned not null default '0',
		primary key (id),
		unique key index_zoneid_subplatid (subplatid,zoneid,daynum),
		key subplatid (subplatid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_daily_invite_table(gameid uint32) string {
	return fmt.Sprintf("user_daily_invite_%d", gameid)
}

func create_user_daily_invite(gameid uint32) {
	tblname := get_user_daily_invite_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		inviterid int(10) unsigned not null default '0',
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		dau int(10) unsigned not null default '0',
		pcu int(10) unsigned not null default '0',
		arpu int(10) unsigned not null default '0',
		day_1 int(10) unsigned not null default '0',
		day_2 int(10) unsigned not null default '0',
		day_3 int(10) unsigned not null default '0',
		day_7 int(10) unsigned not null default '0',
		day_30 int(10) unsigned not null default '0',
		pay_first int(10) unsigned not null default '0',
		pay_today int(10) unsigned not null default '0',
		pay_first_num int(10) unsigned not null default '0',
		pay_today_num int(10) unsigned not null default '0',
		primary key (id),
		unique key index_zoneid_subplatid (inviterid,zoneid,daynum),
		key inviterid (inviterid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_subplat_record_table() string {
	return fmt.Sprintf("game_subplat_record")
}

func create_subplat_record() {
	tblname := get_subplat_record_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		subplat varchar(64) not null default '',
		gameid int(10) unsigned not null default '0',
		createtime int(10) unsigned not null default '0',
		primary key (id),
		unique key index_subplat_game (subplat,gameid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_invite_record_table() string {
	return fmt.Sprintf("game_invite_record")
}

func create_invite_record() {
	tblname := get_invite_record_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		inviter varchar(64) not null default '',
		gameid int(10) unsigned not null default '0',
		createtime int(10) unsigned not null default '0',
		primary key (id),
		unique key index_inviter_game (inviter,gameid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_detail_table(gameid, daynum uint32) string {
	return fmt.Sprintf("user_detail_%d_%d", gameid, daynum)
}

func check_empty_user_detail(gameid, daynum uint32) bool {
	tblname := get_user_detail_table(gameid, daynum)
	if check_table_exists(tblname) == true && check_table_empty(tblname) == true {
		str := fmt.Sprintf(`
		drop table IF EXISTS %s;
		`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func create_user_detail(gameid, daynum uint32) {
	tblname := get_user_detail_table(gameid, daynum)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		ip int(10) unsigned not null default '0',
		imei varchar(64) not null default '',
		min int(10) unsigned not null default '0',
		logoutmin int(10) unsigned not null default '0',
		onlinetime int(10) unsigned not null default '0',
		sceneid int(10) unsigned not null default '0',
		taskid varchar(20) not null default '',
		level int(10) unsigned not null default '0',
		sid tinyint(2) unsigned not null default '0',
		updated_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		primary key (id),
		unique key index_zoneid_userid_min (zoneid,userid,min)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_pay_table(gameid uint32) string {
	return fmt.Sprintf("user_pay_%d", gameid)
}

func create_user_pay(gameid uint32) {
	tblname := get_user_pay_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		gameorder varchar(64) not null default '',
		platorder varchar(64) not null default '',
		money int(10) unsigned not null default '0',
		goodid int(10) unsigned not null default '0',
		goodnum int(10) unsigned not null default '0',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		type tinyint(2) unsigned not null default '0',
		state tinyint unsigned not null default '0',
		curlevel int(10) unsigned not null default '0',
		isfirst tinyint(2) unsigned not null default '0',
		primary key (id),
		key daynum (daynum),
		unique key index_zoneid (zoneid,daynum,gameorder)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}
func get_user_cashout_table(gameid uint32) string {
	return fmt.Sprintf("user_cashout_%d", gameid)
}

func create_user_cashout(gameid uint32) {
	tblname := get_user_cashout_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		gameorder varchar(64) not null default '',
		platorder varchar(64) not null default '',
		money int(10) unsigned not null default '0',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		state tinyint unsigned not null default '0',
		primary key (id),
		key daynum (daynum),
		unique key index_zoneid (zoneid,daynum,gameorder)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_user_levelup_table(gameid uint32) string {
	return fmt.Sprintf("user_levelup_%d", gameid)
}

func create_user_levelup(gameid uint32) {
	tblname := get_user_levelup_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		leveltype int(10) unsigned not null default '0',
		typename varchar(64) not null default '',
		oldlevel int(10) unsigned not null default '0',
		newlevel int(10) unsigned not null default '0',
		leveltime int(10) unsigned not null default '0',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key leveltype (leveltype),
		key userid (userid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_exchange_info_table(gameid uint32) string {
	return fmt.Sprintf("user_exchange_%d", gameid)
}

func create_user_exchange(gameid uint32) {
	tblname := get_exchange_info_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		charid bigint(20) unsigned not null default '0',
		charname varchar(64) not null default '',
		recveruid bigint(20) unsigned not null default '0',
		recvername varchar(64) not null default '',
		amount int(10) unsigned not null default '0',
		agent tinyint(2) unsigned not null default '0',
		created_at int(10) unsigned not null default '0',
		recved_at int(10) unsigned not null default '0',
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key agent (agent),
		key userid (charid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}

}

func get_user_economic_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_economic_%d_%d", gameid, ymd)
}
func create_user_economic(gameid, ymd uint32) {
	tblname := get_user_economic_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		coinid int(10) unsigned not null default '0',
		coinname varchar(20) not null default '',
		coincount int(10) unsigned not null default '0',
		actionid int(10) unsigned not null default '0',
		actionname varchar(64) not null default '',
		actioncount int(10) unsigned not null default '0',
		type tinyint(2) unsigned not null default '0',
		level int(10) unsigned not null default '0',
		curcoin int(10) unsigned not null default '0',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key userid (userid),
		key actionid (actionid),
		key coinid (coinid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_user_item_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_item_%d_%d", gameid, ymd)
}
func create_user_item(gameid, ymd uint32) {
	tblname := get_user_item_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		itemtype int(10) unsigned not null default '0',
		itemid int(10) unsigned not null default '0',
		itemname varchar(30) not null default '',
		itemcount int(10) unsigned not null default '1',
		actionid int(10) unsigned not null default '0',
		actionname varchar(64) not null default '',
		type tinyint(2) unsigned not null default '0',
		level int(10) unsigned not null default '0',
		gold int(10) unsigned not null default '0',
		curnum int(10) unsigned not null default '0',
		extdata varchar(126) not null default '',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key userid (userid),
		key actionid (actionid),
		key itemtype (itemtype),
		key itemid (itemid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_user_lottery_table(gameid uint32) string {
	return fmt.Sprintf("user_lottery_%d", gameid)
}
func create_user_lottery(gameid uint32) {
	tblname := get_user_lottery_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		type int(10) unsigned not null default '0',
		name varchar(64) not null default '',
		itemtype int(10) unsigned not null default '0',
		itemid int(10) unsigned not null default '0',
		itemcount int(10) unsigned not null default '1',
		itemname varchar(64) not null default '',
		optype int(10) unsigned not null default '0',
		opid int(10) unsigned not null default '0',
		opcount int(10) unsigned not null default '1',
		opname varchar(64) not null default '',
		remaincount int(10) unsigned not null default '0',
		created_at int(10) unsigned not null default '0',
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key userid (userid),
		key itemtype_id (itemtype,itemid),
		key type (type)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

// ===
func get_user_group_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_group_%d_%d", gameid, ymd/100)
}
func create_user_group(gameid, ymd uint32) {
	tblname := get_user_group_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		groupid bigint(20) unsigned not null default '0',
		groupname varchar(64) not null default '',
		state int(10) unsigned not null default '1',
		created_at int(10) unsigned not null default '0',
		updated_at int(10) unsigned not null default '0',
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key userid (userid),
		key groupid (groupid),
		key state (state)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_user_redpack_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_redpack_%d_%d", gameid, ymd)
}
func create_user_redpack(gameid, ymd uint32) {
	tblname := get_user_redpack_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		groupid bigint(20) unsigned not null default '0',
		packid bigint(20) unsigned not null default '0',
		packtype int(10) unsigned not null default '0',
		coin float(12,3) not null default '0',
		packnum int(10) not null default '1',
		thundernum int(10) unsigned not null default '0',
		created_at int(10) unsigned not null default '0',
		primary key (id),
		key zoneid (zoneid),
		key userid (userid),
		key groupid (groupid),
		key packtype (packtype)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_user_share_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_share_%d_%d", gameid, ymd/100)
}
func create_user_share(gameid, ymd uint32) {
	tblname := get_user_share_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		created_at int(10) unsigned not null default '0',
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key userid (userid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

//===

func get_shop_transaction_table(gameid uint32) string {
	return fmt.Sprintf("shop_transaction_%d", gameid)
}
func create_shop_transaction(gameid uint32) {
	tblname := get_shop_transaction_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		type int(10) unsigned not null default '0',
		name varchar(64) not null default '',
		itemtype int(10) unsigned not null default '0',
		itemid int(10) unsigned not null default '0',
		itemcount int(10) unsigned not null default '1',
		itemname varchar(64) not null default '',
		optype int(10) unsigned not null default '0',
		opid int(10) unsigned not null default '0',
		opcount int(10) unsigned not null default '1',
		opname varchar(64) not null default '',
		remaincount int(10) unsigned not null default '0',
		created_at int(10) unsigned not null default '0',
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key userid (userid),
		key itemtype_id (itemtype,itemid),
		key type (type)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_user_transaction_table(gameid uint32) string {
	return fmt.Sprintf("user_transaction_%d", gameid)
}
func create_user_transaction(gameid uint32) {
	tblname := get_user_transaction_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		sellerid bigint(20) unsigned not null default '0',
		actid int(10) unsigned not null default '0',
		actname varchar(64) not null default '',
		coinid int(10) unsigned not null default '0',
		coincount int(10) unsigned not null default '1',
		coinname varchar(64) not null default '',
		itemtype int(10) unsigned not null default '0',
		itemid int(10) unsigned not null default '0',
		itemcount int(10) unsigned not null default '1',
		itemname varchar(64) not null default '',
		curnum int(10) unsigned not null default '0',
		sellercurnum int(10) unsigned not null default '0',
		created_at int(10) unsigned not null default '0',
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key userid (userid),
		key sellerid (sellerid),
		key coinid (coinid),
		key itemtype_id (itemtype,itemid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_mahjong_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_mahjong_%d_%d", gameid, int(ymd/100))
}

func create_user_mahjong(gameid, ymd uint32) {
	tblname := get_mahjong_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		daynum int(10) unsigned not null default '0',
		zoneid int(10) unsigned not null default '0',
		charid bigint(20) unsigned not null default '0',
		charname varchar(64) not null default '',
		platid int(10) unsigned not null default '0',
		charnum int(10) unsigned not null default '0',
		repnum int(10) unsigned not null default '0',
		roomid int(10) unsigned not null default '0',
		groomid bigint(20) unsigned not null default '0',
		optype int(10) unsigned not null default '0',
		realnum int(10) unsigned not null default '0',
		diamond decimal(10, 2) not null default '0',
		extdata varchar(256) not null default '',
		meminfo varchar(500) not null default '',
		created_at int(10) unsigned not null default '0',
		primary key (id),
		key daynum (daynum),
		key zoneid (zoneid),
		key charid (charid),
		key charnum (charnum),
		key roomid (roomid),
		key repnum (repnum)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_mjpoint_table(ymd uint32) string {
	return fmt.Sprintf("user_mjpoint_%d", int(ymd/100))
}

func create_user_mjpoint(ymd uint32) {
	tblname := get_mjpoint_table(ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		daynum int(10) unsigned not null default '0',
		gameid int(10) unsigned not null default '0',
		subgameid int(10) unsigned not null default '0',
		charid bigint(20) unsigned not null default '0',
		point int(11) not null default '0',
		rooms int(10) unsigned not null default '0',
		rounds int(10) unsigned not null default '0',
		diamond decimal(10, 2) not null default '0',
		key gameid (gameid),
		key daynum (daynum),
		unique key index_c_g_d (charid, subgameid, daynum)
	) engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_mjredpack_code_table() string {
	return "user_mjredpack_code"
}

func create_user_mjredpack_code() {
	tblname := get_mjredpack_code_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		daynum int(10) unsigned not null default '0',
		gameid int(10) unsigned not null default '0',
		subgameid int(10) unsigned not null default '0',
		charid bigint(20) unsigned not null default '0',
		code varchar(64) not null default '',
		money int(10) unsigned not null default '0',
		created int(10) unsigned not null default '0',
		key gameid (gameid, subgameid),
		key daynum (daynum),
		key charid (charid)
	) engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

// 关卡记录按年月分表
func get_checkpoint_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_checkpoint_%d_%d", gameid, uint32(ymd/100))
}

// 活动记录
func get_activity_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_activity_%d_%d", gameid, uint32(ymd/100))
}

// 任务记录
func get_task_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_task_%d_%d", gameid, uint32(ymd/100))
}

// PVP记录
func get_battle_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_battle_%d_%d", gameid, uint32(ymd/100))
}

// 根据类型获取相应的表名
func get_action_table(gameid, rtype, ymd uint32) string {
	var tblname string
	switch rtype {
	case 1:
		tblname = get_checkpoint_table(gameid, ymd)
	case 2:
		tblname = get_activity_table(gameid, ymd)
	case 3:
		tblname = get_task_table(gameid, ymd)
	case 4:
		tblname = get_battle_table(gameid, ymd)
	}
	return tblname
}

// 创建关卡、活动、任务、PVP记录表
func create_action_record(gameid, ymd uint32) {
	tblname := get_checkpoint_table(gameid, ymd)
	_create_action_record(tblname)
	tblname = get_task_table(gameid, ymd)
	_create_action_record(tblname)
	tblname = get_activity_table(gameid, ymd)
	_create_action_record(tblname)
	tblname = get_battle_table(gameid, ymd)
	_create_action_record(tblname)
}
func _create_action_record(tblname string) {
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		zoneid int(10) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		userid bigint(20) unsigned not null default '0',
		actionid int(10) unsigned not null default '0',
		actionname varchar(100) not null default '',
		acttype int(10) unsigned not null default '0',
		acttypename varchar(64) not null default '',
		starttime int(10) unsigned not null default '0',
		duration int(10) unsigned not null default '0',
		endtime int(10) unsigned not null default '0',
		state tinyint(2) unsigned not null default '0',
		power int(10) unsigned not null default '0',
		level int(10) unsigned not null default '0',
		viplevel int(10) unsigned not null default '0',
		extdata varchar(255) not null default '',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		primary key (id),
		key zoneid (zoneid),
		key daynum (daynum),
		key userid (userid),
		key actionid (actionid),
		key acttype (acttype)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_output_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_output_%d_%d", gameid, uint32(ymd/100))
}

func get_consume_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_consume_%d_%d", gameid, uint32(ymd/100))
}

func create_output_table(gameid, ymd uint32) {
	tblname := get_output_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		daynum int(10) unsigned not null default '0' comment '日期',
		subgameid int(10) unsigned not null default '0' comment '子游戏ID',
		userid bigint(20) unsigned not null default '0' comment '角色ID',
		acttype int(10) unsigned not null default '0' comment '类型',
		typename varchar(100) not null default '' comment '类型名称',
		actid int(10) unsigned not null default '0' comment '行为、道具等ID',
		actname varchar(100) not null default '' comment '行为、道具名称',
		actnum int(10) not null default '0' comment '行为、道具数量',
		coinid int(10) unsigned not null default '0' comment '产出的货币ID',
		coinnum int(10) not null default '0' comment '产出的货币数量',
		coinleft int(10) not null default '0' comment '剩余的货币数量',
		level int(10) unsigned not null default '0' comment '玩家等级',
		viplevel int(10) unsigned not null default '0' comment 'VIP等级',
		created int(10) unsigned not null default '0' comment '时间',
		primary key (id),
		key daynum (daynum),
		key userid (userid),
		key acttype (acttype),
		key actid (actid),
		key coinid (coinid)
	) engine=InnoDB auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func create_consume_table(gameid, ymd uint32) {
	tblname := get_consume_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		daynum int(10) unsigned not null default '0' comment '日期',
		subgameid int(10) unsigned not null default '0' comment '子游戏ID',
		userid bigint(20) unsigned not null default '0' comment '角色ID',
		acttype int(10) unsigned not null default '0' comment '类型',
		typename varchar(100) not null default '' comment '类型名称',
		actid int(10) unsigned not null default '0' comment '行为、道具等ID',
		actname varchar(100) not null default '' comment '行为、道具名称',
		actnum int(10) not null default '0' comment '行为、道具数量',
		coinid int(10) unsigned not null default '0' comment '消耗的货币ID',
		coinnum int(10) not null default '0' comment '消耗的货币数量',
		coinleft int(10) not null default '0' comment '剩余的货币数量',
		level int(10) unsigned not null default '0' comment '玩家等级',
		viplevel int(10) unsigned not null default '0' comment 'VIP等级',
		created int(10) unsigned not null default '0' comment '时间',
		primary key (id),
		key daynum (daynum),
		key userid (userid),
		key acttype (acttype),
		key actid (actid),
		key coinid (coinid)
	) engine=InnoDB auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_chips_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_chips_%d_%d", gameid, uint32(ymd/100))
}

func create_chips_table(gameid, ymd uint32) {
	tblname := get_chips_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		daynum int(10) unsigned not null default '0' comment '日期',
		userid bigint(20) unsigned not null default '0' comment '角色ID',
		subgameid int(10) unsigned not null default '0' comment '子游戏ID',
		roomtype int(10) unsigned not null default '0' comment '房间类型',
		roomid int(10) unsigned not null default '0' comment '房间ID',
		coinid int(10) unsigned not null default '0' comment '货币ID',
		coinbet int(10) unsigned not null default '0' comment '下注数',
		coinnum int(10) not null default '0' comment '输赢数',
		coinleft int(10) not null default '0' comment '剩余数',
		win tinyint(2) not null default '0' comment '0输，1赢',
		wincard int(11) unsigned not null default '0' comment '赢家牌型',
		owncard int(11) unsigned not null default '0' comment '玩家牌型',
		point int(10) not null default '0' comment '输赢积分',
		created int(10) unsigned not null default '0' comment '下注时间',
		primary key (id),
		key daynum (daynum),
		key userid (userid),
		key subgameid (subgameid)
	) engine=InnoDB auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_coin_table(gameid uint32) string {
	return fmt.Sprintf("user_coin_%d", gameid)
}

func create_coin_table(gameid uint32) {
	tblname := get_coin_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		time bigint(20) unsigned not null default '0' comment '角色ID',
		coinid int(10) unsigned not null default '0' comment '货币ID',
		curnum bigint(20) not null default '0' comment '货币数量',
		updated int(10) unsigned not null default '0' comment '更新时间',
		unique key uid_coinid (userid, coinid)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_exchange_rate_table(gameid uint32) string {
	return fmt.Sprintf("exchange_rate_%d", gameid)
}
func create_exchange_rate_table(gameid uint32) {
	tblname := get_exchange_rate_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		time varchar(100) not null default '' comment '日期',
		exchange_rate float(12,4) not null default '0',
		currency int(10) unsigned not null default '1' comment '货币',
		unique key time_currency (time,currency)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_launch_keys_table(gameid uint32) string {
	return fmt.Sprintf("launch_keys_%d", gameid)
}
func create_launch_keys_table(gameid uint32) {
	tblname := get_launch_keys_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		launchid varchar(64) not null default '' ,
		keywords varchar(100) not null default '' ,
		updated timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		unique key (id)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table %s err:%s", tblname, err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}
func get_money_change_table(gameid uint32) string {
	return fmt.Sprintf("money_change_%d", gameid)
}
func create_money_change_table(gameid uint32) {
	tblname := get_money_change_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		daynum int(10) unsigned not null default '0' comment '日期',
		userid bigint(20) unsigned not null default '0' comment '角色ID',
		pay_num int(10) unsigned not null default '0' comment '充值次数',
		pay_all int(10) unsigned not null default '0' comment '充值金额',
		cash_num int(10) unsigned not null default '0' comment '提现次数',
		cash_all int(10) unsigned not null default '0' comment '提现金额',
		unique key (id)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}
func get_realtime_data_table(gameid uint32, ymd uint32) string {
	return fmt.Sprintf("realtime_data_%d_%d", gameid, ymd)
}
func create_realtime_data_table(gameid uint32, ymd uint32) {
	tblname := get_realtime_data_table(gameid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		hour varchar(64) not null default '' comment '小时',
		platid int(10) unsigned not null default '0' comment '平台ID',
		launchid varchar(64) not null default '' comment '渠道',
		adcode varchar(64) not null default '' comment '广告',
		sid tinyint(2) unsigned not null default '0',
		type int(10) unsigned not null default '0' comment '查询类型',

		totaluser int(10) unsigned not null default '0' comment '总玩家',
		newuser int(10) unsigned not null default '0' comment '新增用户',
		olduser int(10) unsigned not null default '0' comment '老用户',
		dau int(10) unsigned not null default '0' comment '活跃用户',
		paynum int(10) unsigned not null default '0' comment '充值用户',
		paytotal int(10) unsigned not null default '0' comment '充值金额',
		cashnum int(10) unsigned not null default '0' comment '提现用户',
		cashtotal int(10) unsigned not null default '0' comment '提现金额',
		unique key (id)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}
func get_adjapp_table(gameid uint32) string {
	return fmt.Sprintf("adj_app_%d", gameid)
}
func create_adjapp_table(gameid uint32) {
	tblname := get_adjapp_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		content varchar(64) not null default '' comment '内容',
		unique key (id)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_change_registersrc_table(gameid uint32) string {
	return fmt.Sprintf("user_change_registersrc_%d", gameid)
}

func create_change_registersrc_table(gameid uint32) {
	tblname := get_change_registersrc_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		userid bigint(20) unsigned not null default '0',
		zoneid int(10) unsigned not null default '0',
		sid int(10) unsigned not null default '0',
		befor_platid int(10) unsigned not null default '0',
		befor_launchid varchar(64) not null default '' ,
		befor_ad_account varchar(64) not null default '',
		after_platid int(10) unsigned not null default '0',
		after_launchid varchar(64) not null default '' ,
		after_ad_account varchar(64) not null default '',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		
		unique key (id)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}

func get_user_daily_roi_table(gameid uint32) string {
	return fmt.Sprintf("user_daily_roi_%d", gameid)
}
func create_user_daily_roi(gameid uint32) {
	tblname := get_user_daily_roi_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		platid int(10) unsigned not null default '0',
		zoneid int(10) unsigned not null default '0',
		launchid varchar(64) not null default '' ,
		ad_account varchar(64) not null default '',
		sid tinyint(2) unsigned not null default '0',
		daynum int(10) unsigned not null default '0',
		new_user int(10) unsigned not null default '0',
		pay_first int(10) unsigned not null default '0',
		pay_first_num int(10) unsigned not null default '0',	
		
		pay_1 int(10) unsigned not null default '0',
		pay_2 int(10) unsigned not null default '0',
		pay_3 int(10) unsigned not null default '0',
		pay_4 int(10) unsigned not null default '0',
		pay_5 int(10) unsigned not null default '0',
		pay_6 int(10) unsigned not null default '0',
		pay_7 int(10) unsigned not null default '0',
		pay_8 int(10) unsigned not null default '0',
		pay_9 int(10) unsigned not null default '0',
		pay_10 int(10) unsigned not null default '0',
		pay_11 int(10) unsigned not null default '0',
		pay_12 int(10) unsigned not null default '0',
		pay_13 int(10) unsigned not null default '0',
		pay_14 int(10) unsigned not null default '0',
		pay_21 int(10) unsigned not null default '0',
		pay_30 int(10) unsigned not null default '0',
		pay_45 int(10) unsigned not null default '0',
		pay_60 int(10) unsigned not null default '0',
		pay_90 int(10) unsigned not null default '0',
		pay_180 int(10) unsigned not null default '0',

		cashout_1 int(10) unsigned not null default '0',
		cashout_2 int(10) unsigned not null default '0',
		cashout_3 int(10) unsigned not null default '0',
		cashout_4 int(10) unsigned not null default '0',
		cashout_5 int(10) unsigned not null default '0',
		cashout_6 int(10) unsigned not null default '0',
		cashout_7 int(10) unsigned not null default '0',
		cashout_8 int(10) unsigned not null default '0',
		cashout_9 int(10) unsigned not null default '0',
		cashout_10 int(10) unsigned not null default '0',
		cashout_11 int(10) unsigned not null default '0',
		cashout_12 int(10) unsigned not null default '0',
		cashout_13 int(10) unsigned not null default '0',
		cashout_14 int(10) unsigned not null default '0',
		cashout_21 int(10) unsigned not null default '0',
		cashout_30 int(10) unsigned not null default '0',
		cashout_45 int(10) unsigned not null default '0',
		cashout_60 int(10) unsigned not null default '0',
		cashout_90 int(10) unsigned not null default '0',
		cashout_180 int(10) unsigned not null default '0',
		
		primary key (id),
		unique key index_zoneid_platid (platid,zoneid,daynum,launchid,ad_account),
		key platid (platid),
		key launchid (launchid),
		key daynum (daynum)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}
func get_app_cost_today_table(gameid uint32) string {
	return fmt.Sprintf("app_cost_today_%d", gameid)
}

func create_app_cost_today_table(gameid uint32) {
	tblname := get_app_cost_today_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		appid int(10) unsigned not null default '0',
		launchid varchar(64) not null default '',
		daynum varchar(64) not null default '',
		cost  decimal(10, 2) not null default '0',
		unique key (id)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}
func get_app_access_table(gameid uint32) string {
	return fmt.Sprintf("app_access_%d", gameid)
}

func create_app_access_table(gameid uint32) {
	tblname := get_app_access_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		platid int(10) unsigned not null default '0',
		launchid varchar(64) not null default '',
		daynum varchar(64) not null default '',
		accessnum  int(10) unsigned not null default '0',
		unique key (id),
		key platid (platid),
		key launchid (launchid),
		key daynum (daynum)
	) engine=InnoDB default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s", err.Error())
	} else {
		logging.Info("db_monitor create table :%s", tblname)
	}
}
