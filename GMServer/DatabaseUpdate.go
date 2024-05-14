package main

import (
	"fmt"
	"strconv"
	"strings"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

func check_db_init() {
	create_user_account()
	create_punish_table()
	create_broadcast_table()
	create_package_code()
	create_packcode_record()
	create_packcode_type()
	create_feedback_table()
	create_user_mail()
	create_user_mail_ext()
	create_blackwhitelist_table()
	create_gm_zone()
	create_gm_game()
	create_cpay_table()
	create_user_primenu_table()
	create_act_control()
	create_gm_modify_types_table()
	//默认用户创建
	creat_defaultuser()
	create_manager_action_record_code()
	create_limit_iporcode_table()
	create_exclusive_rewards_table()
}

func creat_defaultuser() {
	insert_defaultuser()
	insert_defaultgame()
	insert_defaultzone()
	insert_defaultconnt()
}

// 创建数据表时，自动添加一个默认账户 包含
// gm_user gm_games gm_zones gm_account_games
// 插入对应的默认数据
// gamename admin
// zonename admin
// username admin
// password 123456
// priviliege 0 最高的权限等级
func check_table_data_exists_2(tablename, field1, field2 string, value1, value2 uint32) bool { //判断?表?数据是否存在
	var count int
	str := fmt.Sprintf("select 1 from %s where %s = %d and %s = %d limit 1", tablename, field1, value1, field2, value2)
	db_gm.QueryRow(str).Scan(&count)
	if count > 0 {
		return false
	}
	return true
}

func check_table_data_exists1(tablename, field string, value uint32) bool { //判断?表?数据是否存在
	var count int
	str := fmt.Sprintf("select 1 from %s where %s = %d limit 1", tablename, field, value)
	db_gm.QueryRow(str).Scan(&count)
	if count > 0 {
		return false
	}
	return true
}
func check_table_data_exists(tablename, field, value string) bool { //判断?表?数据是否存在
	var count int
	str := fmt.Sprintf("select 1 from %s where %s = '%s' limit 1", tablename, field, value)
	db_gm.QueryRow(str).Scan(&count)
	if count > 0 {
		return false
	}
	return true
}
func insert_defaultuser() { //gm_user表插入默认用户
	tblname := get_user_account_table()
	if !check_table_data_exists(tblname, "id", "1001") {
		return
	}
	str := fmt.Sprintf("insert into %s (id,username,password,bindip,gameid,zoneid,priviliege,qmaxnum,autorecv,workstate,winnum,config) values(1001,'admin',md5('123456'),0,1001,0,0,0,0,0,0,0)", tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm insert to gm_user err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm insert default value to table:%s", tblname)
	}
}
func insert_defaultgame() { //gm_games表插入默认游戏
	tblname := get_gm_game_table()
	if !check_table_data_exists(tblname, "id", "1001") {
		return
	}
	str := fmt.Sprintf("insert into %s (id,gameid,gamename,gamelink,gmlink,gamekey,gametype,status,createtime) values(1001,1001,'admin','','',0,1,1,0)", tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm insert to gm_games err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm insert default value to table:%s", tblname)
	}
}

func insert_defaultzone() { //gm_zones表插入默认大区
	tblname := get_gm_zone_table()
	if !check_table_data_exists(tblname, "id", "1001") {
		return
	}
	str := fmt.Sprintf("insert into %s (id,gameid,zoneid,zonename,gmlink,status,createtime) values(1001,1001,1001,'admin','',1,0)", tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm insert to gm_zones err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm insert default value to table:%s", tblname)
	}
}
func insert_defaultconnt() { //gm_account_games表插入默认用户数据与gameid连接
	tblname := get_gm_account_game_table()
	if !check_table_data_exists(tblname, "id", "1001") {
		return
	}
	str := fmt.Sprintf("insert into %s (id,accid,gameid) values(1001,1001,1001)", tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm insert to gm_account_games err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm insert default value to table:%s", tblname)
	}
}

func check_db_update() {

}

func check_table_exists(tblname string) bool {
	var count int
	str := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.TABLES where TABLE_SCHEMA='%s' and TABLE_NAME='%s'", config.GetConfigStr("mysql_dbname"), tblname)
	db_gm.QueryRow(str).Scan(&count)
	if count > 0 {
		return true
	}
	return false
}

func check_column_exists(tblname, colname string) bool {
	var count int
	str := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA='%s' and TABLE_NAME='%s' and COLUMN_NAME='%s'", config.GetConfigStr("mysql_dbname"), tblname, colname)
	db_gm.QueryRow(str).Scan(&count)
	if count > 0 {
		return true
	}
	return false
}

func check_index_exists(tblname, indexname string) bool {
	var count int
	str := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.STATISTICS where TABLE_SCHEMA='%s' and TABLE_NAME='%s' and INDEX_NAME='%s'", config.GetConfigStr("mysql_dbname"), tblname, indexname)
	db_gm.QueryRow(str).Scan(&count)
	if count > 0 {
		return true
	}
	return false
}

func add_index(tblname, index string) error {
	if !check_index_exists(tblname, index) {
		str := fmt.Sprintf("ALTER TABLE %s ADD INDEX %s(%s);", tblname, index, index)
		_, err := db_gm.Exec(str)
		if err != nil {
			logging.Error("db_gm add index err:%s,%s,%s", err.Error(), tblname, index)
		} else {
			logging.Info("db_gm add index:%s,%s", tblname, index)
		}
		return err
	}
	return nil
}

func add_char_column(tblname, column string, width int) error {
	if !check_column_exists(tblname, column) {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s varchar(%d) not null default '';`, tblname, column, width)
		_, err := db_gm.Exec(str)
		if err != nil {
			logging.Error("db_gm add column err:%s, %s, %s", err.Error(), tblname, column)
		} else {
			logging.Info("db_gm add column:%s, %s", tblname, column)
		}
		return err
	}
	return nil
}

func add_int_column(tblname, column string) error {
	if !check_column_exists(tblname, column) {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s int(11) not null default '0';`, tblname, column)
		_, err := db_gm.Exec(str)
		if err != nil {
			logging.Error("db_gm add column err:%s, %s, %s", err.Error(), tblname, column)
		} else {
			logging.Info("db_gm add column:%s, %s", tblname, column)
		}
		return err
	}
	return nil
}
func get_gm_modify_types_table() string {
	return "gm_modify_types"
}
func create_gm_modify_types_table() {
	tblname := get_gm_modify_types_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		typeid int(10) unsigned NOT NULL default '0',
		typename varchar(64) NOT NULL default '',
		primary key (id),
		unique key index_gameid_typeid (gameid, typeid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_user_primenu_table() string {
	return "gm_primenu"
}

func create_user_primenu_table() {
	tblname := get_user_primenu_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		accid bigint(20) unsigned NOT NULL default '0',
		menuid int(10) unsigned NOT NULL default '0',
		unique key index_accid_menuid (accid, menuid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_menuids(accid uint64) (retl []uint32) {
	retl = make([]uint32, 0)
	tblname := get_user_primenu_table()
	str := fmt.Sprintf("select menuid from %s where accid=?", tblname)
	rows, err := db_gm.Query(str, accid)
	if err != nil {
		return retl
	}
	for rows.Next() {
		var menuid uint32
		if err = rows.Scan(&menuid); err != nil {
			continue
		}
		retl = append(retl, menuid)
	}
	rows.Close()
	return retl
}

func get_act_control_table() string {
	return "gm_act_control"
}

func create_act_control() {
	tblname := get_act_control_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		packid int(10) unsigned NOT NULL default '0',
		platid int(10) unsigned NOT NULL default '0',
		actid bigint(20) unsigned NOT NULL default '0',
		actname varchar(64) NOT NULL default '',
		state int(10) unsigned NOT NULL default '0',
		stime int(10) unsigned NOT NULL default '0',
		etime int(10) unsigned NOT NULL default '0',
		gmid int(10) unsigned NOT NULL default '0',
		created int(10) unsigned NOT NULL default '0',
		primary key (id),
		key gameid (gameid),
		key packid_platid (packid, platid),
		key stime (stime),
		key etime (etime)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_user_chat_table(gameid, zoneid, ymd uint32) string {
	return fmt.Sprintf("gm_chat_%d_%d_%d", gameid, zoneid, ymd)
}

func create_user_chat(gameid, zoneid, ymd uint32) {
	tblname := get_user_chat_table(gameid, zoneid, ymd)
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		cpid bigint(20) unsigned NOT NULL default '0',
		platid int(10) unsigned NOT NULL default '0',
		accid bigint(20) unsigned NOT NULL default '0',
		charid bigint(20) unsigned NOT NULL default '0',
		charname varchar(64) NOT NULL default '',
		type int(10) unsigned NOT NULL default '0',
		otherid bigint(10) unsigned NOT NULL default '0',
		othername varchar(64) NOT NULL default '',
		content varchar(512) NOT NULL default '',
		created_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key index_cpid (cpid),
		key index_platid (platid),
		key index_accid (accid),
		key index_charid (charid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func drop_user_chat(gameid, zoneid, ymd uint32) {
	tblname := get_user_chat_table(gameid, zoneid, ymd)
	str := fmt.Sprintf("drop table if exists %s", tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm drop table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm drop table:%s", tblname)
	}
}

func get_user_account_table() string {
	return "gm_user"
}

func create_user_account() {
	tblname := get_user_account_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		username varchar(64) NOT NULL default '',
		password varchar(64) NOT NULL default '',
		bindip int(10) unsigned NOT NULL default '0',
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		priviliege bigint(20) NOT NULL default '0',
		qmaxnum int(10) unsigned NOT NULL DEFAULT '0',
		autorecv int(10) unsigned NOT NULL DEFAULT '0',
		workstate int(10) unsigned NOT NULL DEFAULT '0',
		winnum int(10) unsigned NOT NULL DEFAULT '0',
		config varchar(1024) NOT NULL default '',
		primary key (id),
		unique key index_account (username)
	)engine=MyISAM auto_increment=1001 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_user_mail_table() string {
	return "gm_mail"
}

// type 0个人邮件，1全区邮件, 2条件式多人邮件
func create_user_mail() {
	tblname := get_user_mail_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		type int(10) unsigned NOT NULL default '0',
		gmid int(10) unsigned NOT NULL default '0',
		charid bigint(20) unsigned NOT NULL default '0',
		subject varchar(128) NOT NULL default '',
		content varchar(512) NOT NULL default '',
		attachment varchar(300) NOT NULL default '',
		gold int(10) unsigned NOT NULL default '0',
		goldbind int(10) unsigned NOT NULL default '0',
		money int(10) unsigned NOT NULL default '0',
		state int(10) unsigned NOT NULL default '0',
		recordtime int(10) unsigned NOT NULL default '0',
		ext text,
		primary key (id),
		key game_zone_charid (gameid, zoneid, charid),
		key gmid (gmid),
		key recordtime (recordtime)
	)engine=MyISAM auto_increment=1001 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

// func checkoutmails(gameid, zoneid, gmid uint32, charid, stime, etime uint64) (retl []*Pmd.MailRecordData) {
// 	tblname := get_user_mail_table()
// 	where := fmt.Sprintf("gameid=%d and (zoneid=0 or zoneid=%d) ", gameid, zoneid)
// 	if charid != 0 {
// 		where += fmt.Sprintf(" AND charid=%d ", charid)
// 	}
// 	if gmid != 0 {
// 		where += fmt.Sprintf(" AND gmid=%d", gmid)
// 	}
// 	where += fmt.Sprintf(" AND (recordtime between %d and %d)", stime, etime)
// 	str := fmt.Sprintf("select id, gmid, charid, subject, content, attachment,recordtime,gold,money from %s where %s", tblname, where)
// 	logging.Debug("sql:%s", str)
// 	rows, err := db_gm.Query(str)
// 	if err != nil {
// 		logging.Error("checkoutmails error:%s", err.Error())
// 		return
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		data := &Pmd.MailRecordData{}
// 		if err = rows.Scan(&data.Id, &data.Charid, &data.Recvid, &data.Subject, &data.Content, &data.Attachment, &data.Ts, &data.Money, &data.Gold); err != nil {
// 			logging.Error("checkoutmails error:%s", err.Error())
// 			return
// 		}
// 		retl = append(retl, data)
// 	}
// 	return
// }

func get_user_mail_ext_table() string {
	return "gm_mail_ext"
}

func create_user_mail_ext() {
	tblname := get_user_mail_ext_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		mailid bigint(10) unsigned NOT NULL default '0',
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		charid bigint(20) unsigned NOT NULL default '0',
		state int(10) unsigned NOT NULL default '0',
		primary key (id),
		key index_mailid (mailid),
		key game_zone_charid (gameid, zoneid, charid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func saveMailExtCharids(mailid uint64, gameid, zoneid uint32, charids []uint64) {
	if charids == nil || len(charids) == 0 {
		return
	}
	tblname := get_user_mail_ext_table()
	str := fmt.Sprintf("insert into %s(mailid, gameid, zoneid, charid) values ", tblname)
	values := make([]interface{}, 0)
	args := ""
	for _, charid := range charids {
		values = append(values, mailid, gameid, zoneid, charid)
		args += "(?,?,?,?),"
	}
	_, err := db_gm.Exec(str+args[:(len(args)-1)], values...)
	if err != nil {
		logging.Error("saveMailExtCharids error:%s, values:%v", err.Error(), values)
	}
}

func saveMailExtZoneids(mailid uint64, gameid uint32, zoneids []uint32) {
	if zoneids == nil || len(zoneids) == 0 {
		return
	}
	tblname := get_user_mail_ext_table()
	str := fmt.Sprintf("insert into %s(mailid, gameid, zoneid) values ", tblname)
	values := make([]interface{}, 0)
	args := ""
	for _, zoneid := range zoneids {
		values = append(values, mailid, gameid, zoneid)
		args += "(?,?,?),"
	}
	_, err := db_gm.Exec(str+args[:(len(args)-1)], values...)
	if err != nil {
		logging.Error("saveMailExtZoneids error:%s, values:%v", err.Error(), values)
	}
}

func updateMailExtState(mailid uint64, gameid, zoneid uint32, charids []uint64) {
	tblname := get_user_mail_ext_table()
	if charids == nil || len(charids) == 0 {
		str := fmt.Sprintf("update %s set state=1 where mailid=? and gameid=? and zoneid=?", tblname)
		_, err := db_gm.Exec(str, mailid, gameid, zoneid)
		if err != nil {
			logging.Error("updateMailExtState error:%s, mailid:%d, gameid:%d, zoneid:%d", err.Error(), mailid, gameid, zoneid)
		}
	} else {
		args := strings.Repeat("?,", len(charids))
		values := make([]interface{}, 0)
		values = append(values, mailid, gameid, zoneid)
		for _, charid := range charids {
			values = append(values, charid)
		}
		str := fmt.Sprintf("update %s set state=1 where mailid=? and gameid=? and zoneid=? and charid in (%s)", tblname, args[:len(args)-1])
		_, err := db_gm.Exec(str, values...)
		if err != nil {
			logging.Error("updateMailExtState error:%s, mailid:%d, gameid:%d, zoneid:%d, charids:%v", err.Error(), mailid, gameid, zoneid, charids)
		}
	}
}

func saveGmMail(data *Pmd.GmMailInfo, gmid uint32) uint64 {
	atts := data.GetAttachment()
	atts_str := ""
	for _, att := range atts {
		atts_str += fmt.Sprintf("%d*%d*%d,", att.GetItemtype(), att.GetItemid(), att.GetItemnum())
	}
	if atts_str != "" {
		atts_str = atts_str[:len(atts_str)-1]
	}
	tblname := get_user_mail_table()
	now := unitime.Time.Sec()
	str := fmt.Sprintf("insert into %s(gameid,zoneid,type,gmid,charid,subject,content,attachment,recordtime,gold,goldbind,money) values(?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, data.GetGameid(), data.GetZoneid(), data.GetType(), gmid, data.GetCharid(), data.GetSubject(), data.GetContent(), atts_str, now, data.GetGold(), data.GetGoldbind(), data.GetMoney())
	if err != nil {
		logging.Error("saveGmMail error:%s", err.Error())
		return 0
	}
	lastid, _ := result.LastInsertId()
	return uint64(lastid)
}

func saveGmMailEx(data *Pmd.GmMailInfoEx, gmid uint32, ext string) uint64 {
	atts := data.GetAttachment()
	atts_str := ""
	for _, att := range atts {
		atts_str += fmt.Sprintf("%d*%d*%d,", att.GetItemtype(), att.GetItemid(), att.GetItemnum())
	}
	if atts_str != "" {
		atts_str = atts_str[:len(atts_str)-1]
	}
	tblname := get_user_mail_table()
	now := unitime.Time.Sec()
	str := fmt.Sprintf("insert into %s(gameid,zoneid,type,gmid,charid,subject,content,attachment,recordtime,gold,goldbind,money,ext) values(?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, data.GetGameid(), data.GetZoneid(), 2, gmid, 0, data.GetSubject(), data.GetContent(), atts_str, now, data.GetGold(), data.GetGoldbind(), data.GetMoney(), ext)
	if err != nil {
		logging.Error("saveGmMailEx error:%s", err.Error())
		return 0
	}
	lastid, _ := result.LastInsertId()
	return uint64(lastid)
}

func get_blackwhitelist_table() string {
	return "gm_blackwhitelist"
}

func create_blackwhitelist_table() {
	tblname := get_blackwhitelist_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		subgameid int(10) unsigned NOT NULL default '0',
		charid bigint(20) unsigned NOT NULL default '0',
		charname varchar(32) NOT NULL default '',
		gmid int(10) unsigned NOT NULL default '0',
		type int(10) unsigned NOT NULL default '0',
		state int(10) unsigned NOT NULL default '0',
		winrate int(10) unsigned NOT NULL default '0',
		setchips int(10) unsigned NOT NULL default '0',
		curchips int(10) unsigned NOT NULL default '0',
		settimes int(10) unsigned NOT NULL default '0',
		curtimes int(10) unsigned NOT NULL default '0',
		itimes int(10) unsigned NOT NULL default '0',
		recordtime int(10) unsigned NOT NULL default '0',
		updated_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key game_zone_charid (gameid, zoneid, charid),
		key subgameid (subgameid)
	)engine=MyISAM auto_increment=1001 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func saveBlackWihteList(data *Pmd.BlackWhitelistInfo, gmid uint32) uint64 {
	tblname := get_blackwhitelist_table()
	now := unitime.Time.Sec()
	str := fmt.Sprintf("insert into %s(gameid,zoneid,subgameid,charid,charname,gmid,type,state,winrate,setchips,curchips,settimes,curtimes,itimes,recordtime) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, data.GetGameid(), data.GetZoneid(), data.GetSubgameid(), data.GetCharid(), data.GetCharname(), gmid, data.GetType(), data.GetState(), data.GetWinrate(), data.GetSetchips(),
		data.GetCurchips(), data.GetSettimes(), data.GetCurtimes(), data.GetIntervaltimes(), now)
	if err != nil {
		logging.Error("saveBlackWihteList error:%s", err.Error())
		return 0
	}
	lastid, _ := result.LastInsertId()
	return uint64(lastid)
}

func updateBlackWhiteList(bid uint32, gmid uint32, data *Pmd.BlackWhitelistInfo) {
	tblname := get_blackwhitelist_table()
	now := unitime.Time.Sec()
	if data == nil {
		str := fmt.Sprintf("update %s set state=1, updated_at=%d, gmid=%d where id=%d", tblname, bid, now, gmid)
		db_gm.Exec(str)
	} else {
		str := fmt.Sprintf("update %s set subgameid=?,setchips=?,curchips=?,winrate=?,state=?,type=?,settimes=?,curtimes=?,itimes=?,updated_at=?,gmid=? where id=%d", tblname, bid)
		db_gm.Exec(str, data.GetSubgameid(), data.GetSetchips(), data.GetCurchips(), data.GetWinrate(), data.GetState(), data.GetType(), data.GetSettimes(), data.GetCurtimes(), data.GetIntervaltimes(), now, gmid)
	}
}

func delBlackWhiteList(gmid uint32, ids []uint32) {
	tblname := get_blackwhitelist_table()
	now := unitime.Time.Sec()
	where := ""
	if len(ids) == 1 {
		where += fmt.Sprintf("id=%d", ids[0])
	} else if len(ids) > 1 {
		ids_str := ""
		for _, id := range ids {
			ids_str += fmt.Sprintf("%d,", id)
		}
		where += fmt.Sprintf("id in (%s)", ids_str[:len(ids_str)-1])
	} else {
		return
	}
	str := fmt.Sprintf("update %s set state=1, updated_at=%d, gmid=%d where %s", tblname, now, gmid, where)
	db_gm.Exec(str)
}

func get_feedback_table() string {
	return "gm_feedback"
}

func create_feedback_table() {
	tblname := get_feedback_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		platid int(10) unsigned NOT NULL default '0',
		charid bigint(20) unsigned NOT NULL default '0',
		charname varchar(64) NOT NULL default '',
		userlevel int(10) unsigned NOT NULL default '0',
		viplevel int(10) unsigned NOT NULL default '0',
		feedbackid int(10) unsigned NOT NULL default '0',
		subject varchar(256) NOT NULL default '',
		content varchar(256) NOT NULL default '',
		star int(10) unsigned NOT NULL default '0',
		recordtime int(10) unsigned NOT NULL default '0',
		type int(10) unsigned NOT NULL default '0',
		action int(10) unsigned NOT NULL default '1',
		reply varchar(256) NOT NULL default '',
		phonenum varchar(12) NOT NULL default '',
		primary key (id),
		key gamezone (gameid, zoneid, platid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func saveFeedback(ftype uint32, data []*Pmd.FeedbackData) error {
	if len(data) > 0 {
		tblname := get_feedback_table()
		args := strings.Repeat("(?,?,?,?,?,?,?,?,?,?,?,?,?,?),", len(data))
		vals := make([]interface{}, 0)
		for _, tmp := range data {
			vals = append(vals, tmp.GetGameid(), tmp.GetZoneid(), tmp.GetPlatid(), tmp.GetCharid(), tmp.GetCharname(), tmp.GetUserlevel(), tmp.GetViplevel(), tmp.GetFeedbackid(), tmp.GetSubject(), tmp.GetContent(), tmp.GetStar(), tmp.GetRecordtime(), ftype, tmp.GetPhonenum())
		}
		str := fmt.Sprintf("insert into %s(gameid,zoneid,platid,charid,charname,userlevel,viplevel,feedbackid,subject,content,star,recordtime, type, phonenum) values %s", tblname, args[:len(args)-1])
		_, err := db_gm.Exec(str, vals...)
		if err != nil {
			logging.Error("ParseFeedbackGmUserPmd_CS error:%s", err.Error())
			return err
		}
	}
	return nil
}

func checkFeedback(recordid uint32) (charid uint64, ftype uint32, content string) {
	tblname := get_feedback_table()
	str := fmt.Sprintf("select charid, type, content from %s where id=?", tblname)
	row := db_gm.QueryRow(str, recordid)
	if err := row.Scan(&charid, &ftype, &content); err != nil {
		logging.Error("checkFeedback error:%s", err.Error())
	}
	return
}

func checkFeedbackList(gameid, zoneid, platid, ftype, action uint32, charid uint64, charname string, starttime, endtime uint64, curpage, perpage uint32) (maxpage uint32, retl []*Pmd.FeedbackData) {
	retl = make([]*Pmd.FeedbackData, 0)
	var tblname string = get_feedback_table()
	var where string
	where = fmt.Sprintf(" gameid=%d AND zoneid=%d AND (recordtime>=%d AND recordtime<%d) ", gameid, zoneid, starttime, endtime)
	if zoneid == 0 {
		where = fmt.Sprintf(" gameid=%d AND (recordtime>=%d AND recordtime<%d) ", gameid, starttime, endtime)
	}

	if ftype > 100 {
		where += fmt.Sprintf(" AND type != 0")
	} else {
		where += fmt.Sprintf(" AND type = %d ", ftype)
	}
	if action != 0 {
		where += fmt.Sprintf(" AND action=%d ", action)
	}
	if platid != 0 {
		where += fmt.Sprintf(" AND platid=%d ", platid)
	}
	var vals = make([]interface{}, 0)
	if charname != "" {
		where += " AND charname=? "
		vals = append(vals, charname)
	}
	str := fmt.Sprintf("select count(*) from %s where %s", tblname, where)
	row := db_gm.QueryRow(str, vals...)
	var count uint32
	if err := row.Scan(&count); err != nil {
		logging.Error("checkFeedbackList error:%s", err.Error())
		return
	}
	if perpage == 0 {
		perpage, curpage = 15, 1
	}
	maxpage = count / perpage
	if count > 0 && maxpage == 0 {
		maxpage = 1
	}
	if curpage > maxpage {
		logging.Debug("checkFeedbackList sql:%s vals:%v", str, vals)
		return
	}

	where += fmt.Sprintf(" order by id desc limit %d, %d", (curpage-1)*perpage, perpage)
	str = fmt.Sprintf("select id,gameid,zoneid,platid,charid,charname,userlevel,viplevel,feedbackid,subject,content,star,recordtime, action, reply, phonenum from %s where %s", tblname, where)
	logging.Debug("checkFeedbackList sql:%s, vals:%v", str, vals)
	rows, err := db_gm.Query(str, vals...)
	if err != nil {
		logging.Error("checkFeedbackList error:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		data := &Pmd.FeedbackData{}
		if err = rows.Scan(&data.Recordid, &data.Gameid, &data.Zoneid, &data.Platid, &data.Charid, &data.Charname, &data.Userlevel, &data.Viplevel, &data.Feedbackid, &data.Subject,
			&data.Content, &data.Star, &data.Recordtime, &data.State, &data.Reply, &data.Phonenum); err != nil {
			logging.Error("checkFeedbackList error:%s", err.Error())
			return
		}
		retl = append(retl, data)
	}
	return
}

func updateFeedbackState(recordid, state uint32, content string) {
	tblname := get_feedback_table()
	str := fmt.Sprintf("update %s set action=%d, reply=? where id=?", tblname, state)
	_, err := db_gm.Exec(str, content, recordid)
	if err != nil {
		logging.Error("update_feedback_state error:%s", err.Error())
	}
}

func create_broadcast_table() {
	tblname := "gm_broadcast"
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		gmid int(10) unsigned NOT NULL default '0',
		countryid int(10) unsigned NOT NULL default '0',
		sceneid int(10) unsigned NOT NULL default '0',
		starttime int(10) unsigned NOT NULL default '0',
		nexttime int(10) unsigned NOT NULL default '0',
		endtime int(10) unsigned NOT NULL default '0',
		intervaltime int(10) unsigned NOT NULL default '0',
		type int(10) unsigned NOT NULL default '0',
		title varchar(128) NOT NULL default '',
		content varchar(512) NOT NULL default '',
		state tinyint(2) NOT NULL default '0',
		updated_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key index_game (gameid),
		key index_zone (zoneid),
		key index_starttime (starttime),
		key index_nexttime (nexttime),
		key index_state (state),
		key index_endtime (endtime)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func create_punish_table() {
	tblname := "gm_punish"
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		charid int(20) unsigned NOT NULL default '0',
		charname varchar(128) NOT NULL default '',
		ip int(10) unsigned NOT NULL default '0',
		gmid int(10) unsigned NOT NULL default '0',
		type int(10) unsigned NOT NULL default '0',
		reason varchar(512) NOT NULL default '',
		starttime int(10) unsigned NOT NULL default '0',
		endtime int(10) unsigned NOT NULL default '0',
		pointnum varchar(64) NOT NULL default '',
		multiple int(10) unsigned NOT NULL default '0',
		state tinyint(2) NOT NULL default '0',
		created_at int(10) unsigned NOT NULL default '0',
		updated_at int(10) unsigned NOT NULL default '0',
		pid varchar(128) NOT NULL default '',
		primary key (id)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_cpay_table() string {
	return "gm_cpay"
}

func create_cpay_table() {
	tblname := get_cpay_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		gmid bigint(20) unsigned NOT NULL default '0',
		packid int(20) unsigned NOT NULL default '0',
		platid int(10) unsigned NOT NULL default '0',
		payplatid int(10) unsigned NOT NULL default '0',
		minlevel int(10) unsigned NOT NULL default '0',
		maxlevel int(10) unsigned NOT NULL default '0',
		minmoney int(10) unsigned NOT NULL default '0',
		maxmoney int(10) unsigned NOT NULL default '0',
		stime int(10) unsigned NOT NULL default '0',
		etime int(10) unsigned NOT NULL default '0',
		state int(10) unsigned NOT NULL default '0',
		created_at int(10) unsigned NOT NULL default '0',
		updated_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key gameid (gameid),
		key packid (packid),
		key platid (platid),
		key state (state)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_packcode_type_table() string {
	return "gm_packcode_type"
}

func create_packcode_type() {
	tblname := get_packcode_type_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		content varchar(384) NOT NULL default '',
		gmid bigint(20) unsigned NOT NULL default '0',
		created_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key gameid (gameid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_packcode_record_table() string {
	return "gm_packcode_record"
}

func create_packcode_record() {
	tblname := get_packcode_record_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneidmax int(10) unsigned NOT NULL default '0',
		zoneidmin int(10) unsigned NOT NULL default '0',
		ctype bigint(20) unsigned NOT NULL default '0',
		cnum int(10) unsigned NOT NULL default '0',
		packid int(20) unsigned NOT NULL default '0',
		platid int(10) unsigned NOT NULL default '0',
		cdesc varchar(128) NOT NULL default '',
		filename varchar(128) NOT NULL default '',
		stime int(10) unsigned NOT NULL default '0',
		etime int(10) unsigned NOT NULL default '0',
		climit int(10) unsigned NOT NULL default '0',
		gmid bigint(20) unsigned NOT NULL default '0',
		state int(10) unsigned NOT NULL default '0',
		created_at int(10) unsigned NOT NULL default '0',
		updated_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key gameid (gameid),
		key packid (packid),
		key platid (platid),
		key ctype (ctype),
		key state (state)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func create_package_code() {
	tmpname := "gm_packcode_"
	tblnums := 10
	for i := 0; i < tblnums; i++ {
		tblname := tmpname + strconv.Itoa(i)
		if check_table_exists(tblname) == true {
			continue
		}
		str := fmt.Sprintf(`
		create table IF NOT EXISTS %s (
			id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
			recordid int(10) unsigned NOT NULL default '0',
			gameid int(10) unsigned not null default '0',
			code varchar(128) not null default '',
			uzoneid int(10) unsigned not null default '0',
			uuid bigint(20) unsigned not null default '0',
			utime int(10) unsigned not null default '0',
			state tinyint(2) unsigned not null default '0',
			created_at int(10) unsigned not null default '0',
			primary key(id),
			key recordid (recordid),
			key uuid (uuid),
			unique key index_code (gameid,code)
		) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
		`, tblname)
		_, err := db_gm.Exec(str)
		if err != nil {
			logging.Warning("db_gm create table error:%s", err.Error())
		} else {
			logging.Info("db_gm create table %s success", tblname)
		}
	}
}

func create_manager_action_record_code() {
	tblname := get_manager_action_record_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		gmid bigint(20) unsigned NOT NULL default '0',
		content varchar(255) NOT NULL default '',
		created_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key gameid (gameid),
		key gmid (gmid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}
func get_manager_action_record_table() string {
	return "manager_action_record"
}

func create_limit_iporcode_table() {
	tblname := get_limit_iporcode_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		optype int(10) unsigned NOT NULL default '1',
		limittype int(10) unsigned NOT NULL default '1',
		code varchar(255) NOT NULL default '',
		content varchar(255) NOT NULL default '',
		starttime int(10) unsigned NOT NULL default '0',
		endtime int(10) unsigned NOT NULL default '0',
		created_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key gameid (gameid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_limit_iporcode_table() string {
	return "limit_iporcode"
}
func get_exclusive_rewards_table() string {
	return "exclusive_rewards"
}
func create_exclusive_rewards_table() {
	tblname := get_exclusive_rewards_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		content varchar(255) NOT NULL default '',
		created_at int(10) unsigned NOT NULL default '0',
		primary key (id),
		key gameid (gameid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

// 拼接sql条件语句，如果涉及到键值顺序，请自行拼接，此函数不带指定顺序功能
func ConstructWhereSQL(wm map[string]interface{}) (where string, values []interface{}) {
	if wm == nil || len(wm) == 0 {
		return " 1 = 1", nil
	}
	fields := make([]string, 0, len(wm))
	for k, v := range wm {
		fields = append(fields, fmt.Sprintf("%s=?", k))
		values = append(values, v)
	}
	where = strings.Join(fields, " AND ")
	return
}

// 拼接update语句,um更新的map值，wm条件map值
func ConstructUpdateSQL(tblname string, um, wm map[string]interface{}) (upsql string, values []interface{}) {
	if um == nil || len(um) == 0 {
		return "", nil
	}
	fields := make([]string, 0, len(um))
	for k, v := range um {
		fields = append(fields, fmt.Sprintf("%s=?", k))
		values = append(values, v)
	}
	where, tmp := ConstructWhereSQL(wm)
	values = append(values, tmp...)
	upsql = fmt.Sprintf("UPDATE %s SET %s WHERE %s", tblname, strings.Join(fields, ", "), where)
	return
}

// 拼接插入操作语句
func ConstructInsertSQL(tblname string, vm map[string]interface{}) (insql string, values []interface{}) {
	if vm == nil || len(vm) == 0 {
		return "", nil
	}
	fields := make([]string, 0, len(vm))
	for k, v := range vm {
		fields = append(fields, k)
		values = append(values, v)
	}
	tmp := strings.Repeat("?, ", len(fields))
	tmp = tmp[:len(tmp)-2]
	insql = fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", tblname, strings.Join(fields, ", "), tmp)
	return
}

// 拼接查询语句
func ConstructSelectSQL(tblname string, fields []string, wm map[string]interface{}) (ssql string, values []interface{}) {
	fieldStr := ""
	if fields == nil || len(fields) == 0 {
		fieldStr = "*"
	} else {
		fieldStr = strings.Join(fields, ",")
	}
	where, values := ConstructWhereSQL(wm)
	ssql = fmt.Sprintf("SELECT %s FROM %s WHERE %s", fieldStr, tblname, where)
	return
}
