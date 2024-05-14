package main

import (
	"errors"
	"fmt"

	"git.code4.in/mobilegameserver/logging"
	sjson "github.com/bitly/go-simplejson"
)

func get_device_table() string {
	return "monitor_device"
}

// mac地址，device设备标识，sid操作系统(1、IOS, 2、Android, 3、WindowsPhone),渠道ID，广告码，激活日期，激活时间戳
func create_device_table() {
	tblname := get_device_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		mac varchar(32) not null default '',
		device varchar(32) not null default '',
		sid tinyint(2) unsigned not null default '0',
		platid int(10) unsigned not null default '0',
		adcode varchar(32) not null default '',
		ip varchar(32) not null default '',
		cdate int(10) unsigned not null default '0',
		ctime int(10) unsigned not null default '0',
		primary key (id),
		key index_date (cdate),
		key index_platid (platid),
		unique key index_mac (mac,sid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func get_game_device_table(gameid uint32) string {
	return fmt.Sprintf("monitor_device_%d", gameid)
}

func create_game_device_table(gameid uint32) {
	tblname := get_game_device_table(gameid)
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		vendorid int(10) unsigned NOT NULL default '0',
		deviceid bigint(20) NOT NULL default '0',
		platid int(10) unsigned not null default '0',
		ip varchar(32) not null default '',
		adcode varchar(32) not null default '',
		cdate int(10) unsigned not null default '0',
		ctime int(10) unsigned not null default '0',
		primary key (id),
		key index_vendor (vendorid),
		key index_date (cdate),
		key index_platid (platid),
		unique key index_deviceid (deviceid)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
	if !check_column_exists(tblname, "vendorid") {
		str := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN vendorid int(10) unsigned not null DEFAULT 0;`, tblname)
		_, err := db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add column err:%s,%s,vendorid", err.Error(), tblname)
		} else {
			logging.Info("monitorserver add column:%s,vendorid", tblname)
		}
		str = fmt.Sprintf(`ALTER TABLE %s ADD index index_vendor (vendorid);`, tblname)
		_, err = db_monitor.Exec(str)
		if err != nil {
			logging.Error("monitorserver add index err:%s,%s,vendorid", err.Error(), tblname)
		} else {
			logging.Info("monitorserver add index:%s,vendorid", tblname)
		}
	}
}

// 设备统计
func HandleDeviceCount(data *sjson.Json) error {
	gameid, mac, sid, device := uint32(data.Get("gameid").MustInt()), data.Get("mac").MustString(), data.Get("sid").MustInt(), data.Get("device").MustString()
	if gameid <= 0 || mac == "" || sid <= 0 {
		return errors.New("参数错误，游戏ID或mac为空")
	}
	tblname := get_device_table()
	str := fmt.Sprintf("select id from %s where mac=? and sid=?", tblname)
	row := db_monitor.QueryRow(str, mac, sid)
	var deviceid int64
	err := row.Scan(&deviceid)
	logging.Debug("deviceid:%d, err:%v", deviceid, err)
	if deviceid == 0 {
		str = fmt.Sprintf("insert into %s(mac, sid, device, platid, ip, adcode, cdate, ctime) values(?,?,?,?,?,?,?,?)", tblname)
		result, err := db_monitor.Exec(str, mac, sid, device, data.Get("platid").MustInt(), data.Get("ip").MustString(), data.Get("adcode").MustString(), data.Get("curdate").MustInt(), data.Get("curtime").MustInt())
		if err == nil {
			deviceid, err = result.LastInsertId()
		}
	}
	if err != nil && deviceid == 0 {
		logging.Error("sql:%s, deviceid:%d, err:%v", str, deviceid, err)
		return err
	}
	tblname = get_game_device_table(gameid)
	str = fmt.Sprintf("select id from %s where deviceid = ?", tblname)
	row = db_monitor.QueryRow(str, deviceid)
	var tmpid int64
	err = row.Scan(&tmpid)
	if err != nil {
		logging.Error("sql:%s, err:%s, tmpid:%d, deviceid:%d", str, err.Error(), tmpid, deviceid)
	}
	if tmpid == 0 {
		str = fmt.Sprintf("insert into %s(deviceid, platid, ip, adcode, cdate, ctime,vendorid) values(?,?,?,?,?,?,?)", tblname)
		result, err := db_monitor.Exec(str, deviceid, data.Get("platid").MustInt(), data.Get("ip").MustString(), data.Get("adcode").MustString(), data.Get("curdate").MustInt(), data.Get("curtime").MustInt(), data.Get("vendorid").MustInt())
		if err == nil {
			_, err = result.LastInsertId()
		}
	}
	return err
}
