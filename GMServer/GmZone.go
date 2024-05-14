package main

import (
	"fmt"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type Zone struct {
	Id         uint64
	GameId     uint32
	ZoneId     uint32
	ZoneName   string
	GmLink     string
	Status     uint32
	Createtime uint32
}

func (self *Zone) IsValid() bool {
	return self.Id != uint64(0) && self.Status == uint32(1)
}

func get_gm_zone_table() string {
	return "gm_zones"
}

func create_gm_zone() {
	tblname := get_gm_zone_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		zoneid int(10) unsigned NOT NULL default '0',
		zonename varchar(64) NOT NULL default '',
		gmlink varchar(64) NOT NULL default '',
		status tinyint(2) unsigned NOT NULL default '1',
		createtime int(10) unsigned NOT NULL default '0',
		primary key (id),
		unique key index_gameid_zoneid (gameid, zoneid)
	)engine=MyISAM auto_increment=1001 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func create_zone(gameid, zoneid uint32, name, gmlink string) error {
	tblname := get_gm_zone_table()
	str := fmt.Sprintf("insert into %s(gameid, zoneid, zonename, gmlink, createtime) values (?,?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, gameid, zoneid, name, gmlink, unitime.Time.Sec())
	if err == nil {
		_, err = result.LastInsertId()
	}
	return err
}

func update_zone_name(gameid, zoneid uint32, name string) error {
	tblname := get_gm_zone_table()
	str := fmt.Sprintf("update %s set zonename=? where gameid=? and zoneid=?", tblname)
	result, err := db_gm.Exec(str, name, gameid, zoneid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func update_zone_gmlink(gameid, zoneid uint32, gmlink string) error {
	tblname := get_gm_zone_table()
	str := fmt.Sprintf("update %s set gmlink=? where gameid=? and zoneid=?", tblname)
	result, err := db_gm.Exec(str, gmlink, gameid, zoneid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func update_zone_status(gameid, zoneid uint32) error {
	tblname := get_gm_zone_table()
	str := fmt.Sprintf("update %s set status=0 where gameid=? and zoneid=?", tblname)
	result, err := db_gm.Exec(str, gameid, zoneid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func get_zone_by_id(gameid, zoneid uint32) *Zone {
	m := &Zone{}
	tblname := get_gm_zone_table()
	str := fmt.Sprintf("select id, gameid, zoneid, zonename, gmlink, status, createtime from %s where gameid=? and zoneid=?", tblname)
	row := db_gm.QueryRow(str, gameid, zoneid)
	err := row.Scan(&m.Id, &m.GameId, &m.ZoneId, &m.ZoneName, &m.GmLink, &m.Status, &m.Createtime)
	if err != nil {
		return nil
	}
	return m
}

func delete_zone(gameid, zoneid uint32, force bool) error {
	str, tblname := "", get_gm_zone_table()
	if force {
		str = fmt.Sprintf("delete from %s where gameid=? and zoneid=?", tblname)
	} else {
		str = fmt.Sprint("update %s set status=0 where gameid=? and zoneid=?", tblname)
	}
	result, err := db_gm.Exec(str, gameid, zoneid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func delete_zone_by_gameid(gameid uint32, force bool) error {
	str, tblname := "", get_gm_zone_table()
	if force {
		str = fmt.Sprintf("delete from %s where gameid=?", tblname)
	} else {
		str = fmt.Sprint("update %s set status=0 where gameid=?", tblname)
	}
	result, err := db_gm.Exec(str, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func get_zone_by_gameid(gameid uint32) []*Zone {
	ml := make([]*Zone, 0)
	tblname := get_gm_zone_table()
	str := fmt.Sprintf("select id, gameid, zoneid, zonename, gmlink, status, createtime from %s where gameid=? and status=1", tblname)
	rows, err := db_gm.Query(str, gameid)
	if err != nil {
		logging.Error("get_zone_by_gameid error:%s", err.Error())
		return ml
	}
	defer rows.Close()
	for rows.Next() {
		m := &Zone{}
		if err := rows.Scan(&m.Id, &m.GameId, &m.ZoneId, &m.ZoneName, &m.GmLink, &m.Status, &m.Createtime); err != nil {
			logging.Error("get_zone_by_gameid error:%s", err.Error())
			continue
		}
		ml = append(ml, m)
	}
	return ml
}

func get_valid_zone() []*Zone {
	ml := make([]*Zone, 0)
	tblname := get_gm_zone_table()
	str := fmt.Sprintf("select id, gameid, zoneid, zonename, gmlink, status, createtime from %s where status=1", tblname)
	rows, err := db_gm.Query(str)
	if err != nil {
		logging.Error("get_valid_zone error:%s", err.Error())
		return ml
	}
	defer rows.Close()
	for rows.Next() {
		m := &Zone{}
		if err := rows.Scan(&m.Id, &m.GameId, &m.ZoneId, &m.ZoneName, &m.GmLink, &m.Status, &m.Createtime); err != nil {
			logging.Error("get_valid_zone error:%s", err.Error())
			continue
		}
		ml = append(ml, m)
	}
	return ml
}

func get_all_Zone() []*Zone {
	ml := make([]*Zone, 0)
	tblname := get_gm_zone_table()
	str := fmt.Sprintf("select id, gameid, zoneid, zonename, gmlink, status, createtime from %s where status=1", tblname)
	rows, err := db_gm.Query(str)
	if err != nil {
		logging.Error("get_all_Zone error:%s", err.Error())
		return ml
	}
	defer rows.Close()
	for rows.Next() {
		m := &Zone{}
		if err := rows.Scan(&m.Id, &m.GameId, &m.ZoneId, &m.ZoneName, &m.GmLink, &m.Status, &m.Createtime); err != nil {
			logging.Error("get_all_Zone error:%s", err.Error())
			continue
		}
		ml = append(ml, m)
	}
	return ml
}
