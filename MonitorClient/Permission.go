package main

import (
	"encoding/json"
	"fmt"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
)

//总的条目表
func get_permission_item_table() string {
	return "monitor_permission_item"
}

func create_permission_item_table() {
	tblname := get_permission_item_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL default '0',
		name varchar(128) not null default '',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		unique key index_perid_name (id,name)
	) engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func HandleAddPerItem(task *unibase.ChanHttpTask) {
	itemid := uint32(unibase.Atoi(task.R.FormValue("itemid"), 0))
	itemname := task.R.FormValue("itemname")
	if itemid != 0 && itemname != "" {
		tblname := get_permission_item_table()
		str := fmt.Sprintf("insert into %s(id, name) values(?,?)", tblname)
		result, err := db_monitor.Exec(str, itemid, itemname)
		if err == nil {
			if lastid, err := result.LastInsertId(); err == nil && lastid > 0 {
				task.SendBinary([]byte(`{"retcode":0,"retdesc":"add item success"}`))
				return
			}
		}
		if err != nil {
			task.Error("HandleAddPerItem error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":%s}`, err.Error())))
			return
		}
	}
	task.Warning("HandleAddPerItem itemid:%d, itemname:%s", itemid, itemname)
	task.SendBinary([]byte(`{"retcode":1,"retdesc":"add item failed"}`))
}

func HandleItemList(task *unibase.ChanHttpTask) {
	var itemlist []map[string]interface{}
	tblname := get_permission_item_table()
	str := fmt.Sprintf("select id, name from %s ", tblname)
	row, err := db_monitor.Query(str)
	if err == nil {
		defer row.Close()
		var itemid uint64
		var itemname string
		for row.Next() {
			if err = row.Scan(&itemid, &itemname); err != nil {
				task.Warning("HandleItemList error:%s", err.Error())
				continue
			}
			itemlist = append(itemlist, map[string]interface{}{"itemid": itemid, "itemname": itemname})
		}
	} else {
		task.Error("HandleItemList error:%s", err.Error())
	}
	data, _ := json.Marshal(map[string][]map[string]interface{}{"data": itemlist})
	task.SendBinary(data)
	task.Debug("itemlist:%v", itemlist)
}

func HandleDelItem(task *unibase.ChanHttpTask) {
	itemid := uint32(unibase.Atoi(task.R.FormValue("itemid"), 0))
	if itemid != 0 {
		tblname := get_permission_item_table()
		str := fmt.Sprintf("delete from %s where id=? ", tblname)
		_, err := db_monitor.Exec(str, itemid)
		if err != nil {
			task.Error("HandleDelItem error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":%s}`, err.Error())))
			return
		}
	}
	task.SendBinary([]byte(`{"retcode":0,"retdesc":"delete game success"}`))
}
