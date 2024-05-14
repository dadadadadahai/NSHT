package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/unitime"
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
}

func HandleAddGameZone(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	zonename := task.R.FormValue("zonename")
	remarks := task.R.FormValue("remarks")
	if gameid != 0 && zoneid != 0 && zonename != "" {
		tblname := get_zone_table()
		str := fmt.Sprintf("insert into %s(gameid, zoneid, zonename, remarks) values(?,?,?,?)", tblname)
		result, err := db_monitor.Exec(str, gameid, zoneid, zonename, remarks)
		if err == nil {
			if lastid, err := result.LastInsertId(); err == nil && lastid > 0 {
				task.SendBinary([]byte(`{"retcode":0,"retdesc":"insert game zone success"}`))
				return
			}
		}
		if err != nil {
			task.Error("HandleAddGameZone error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":%s}`, err.Error())))
			return
		}
	}
	task.Warning("HandleAddGameZone gameid:%d, zoneid:%d, zonename:%s failed", gameid, zoneid, zonename)
	task.SendBinary([]byte(`{"retcode":1,"retdesc": "add game zone failed"}`))
}

// 暂时不启用
func HandleGameZoneListNew(task *unibase.ChanHttpTask) {
	var zonelist []map[string]interface{}
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	if gameid != 0 {
		tblname := get_zone_table()
		str := fmt.Sprintf("select zoneid, zonename, remarks from %s where gameid=?", tblname)
		row, err := db_monitor.Query(str, gameid)
		if err == nil {
			defer row.Close()
			var zoneid uint32
			var zonename, remarks string
			for row.Next() {
				if err = row.Scan(&zoneid, &zonename, &remarks); err != nil {
					task.Warning("HandleGamePlatList error:%s", err.Error())
					continue
				}
				zonelist = append(zonelist, map[string]interface{}{"zoneid": zoneid, "zonename": zonename, "remarks": remarks})
			}
		} else {
			task.Error("HandleGamePlatList gameid:%d, error:%s", gameid, err.Error())
		}
	}
	data, _ := json.Marshal(map[string][]map[string]interface{}{"data": zonelist})
	task.SendBinary(data)
	task.Debug("platlist:%v", zonelist)
}

func HandleGameZoneList(task *unibase.ChanHttpTask) {
	var zonelist []map[string]interface{}
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	if gameid != 0 {
		tblname := get_user_data_table(gameid)
		str := fmt.Sprintf("select distinct zoneid from %s ", tblname)
		row, err := db_monitor.Query(str)
		if err == nil {
			defer row.Close()
			var zoneid uint32
			for row.Next() {
				if err = row.Scan(&zoneid); err != nil {
					task.Warning("HandleGamePlatList error:%s", err.Error())
					continue
				}
				zonelist = append(zonelist, map[string]interface{}{"zoneid": zoneid, "zonename": "", "remarks": ""})
			}
		} else {
			task.Error("HandleGamePlatList gameid:%d, error:%s", gameid, err.Error())
		}
	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": zonelist})
	task.SendBinary(data)
	task.Debug("platlist:%v", zonelist)
}

func HandleDelGameZone(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	if gameid != 0 && zoneid != 0 {
		tblname := get_zone_table()
		str := fmt.Sprintf("delete from %s where gameid=? and zoneid=?", tblname)
		db_monitor.Exec(str, gameid, zoneid)
	}
	task.SendBinary([]byte(`{"retcode":0,"retdesc":"delete zone success"}`))
}

func get_game_table() string {
	return "monitor_game_info"
}
func monitor_game_default_insert() {
	tblname := get_game_table()
	if !check_table_data_exists(tblname, "id", "1") {
		return
	}
	str := fmt.Sprintf("insert into %s(id, gameid, gamename,gamekey, logolink, remarks, type,conntype,state) values(1,1,'slots','112233','/images/cq.png','admin用户所需的游戏信息',1,2,0)", tblname)
	result, err := db_monitor.Exec(str)
	if err == nil {
		if lastid, err := result.LastInsertId(); err == nil && lastid > 0 {
			logging.Debug("add admin monitor_game_info success")
			return
		}
	}
	if err != nil {
		logging.Warning("add admin monitor_game_info error")
		return
	}
}

func create_game_table() {
	tblname := get_game_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned not null default '0',
		gamename varchar(128) not null default '',
		gamekey varchar(64) not null default '',
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
}

func get_account_game_table() string {
	return "monitor_account_games"
}
func monitor_add_acc_gameinfo() {
	tblname := get_account_game_table()
	if !check_table_data_exists(tblname, "accid", "1") {
		return
	}
	str := fmt.Sprintf("insert into %s(accid, gameid) values(1,1)", tblname)
	result, err := db_monitor.Exec(str)
	if err == nil {
		if lastid, err := result.LastInsertId(); err == nil && lastid > 0 {
			logging.Debug("add admin monitor_account_games success")
			return
		}
	}
	if err != nil {
		logging.Warning("add admin monitor_account_games error")
		return
	}
}
func create_account_game_table() {
	tblname := get_account_game_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		accid bigint(20) unsigned not null default '0',
		gameid int(10) unsigned not null default '0',
		unique key accid_gameid (accid, gameid)
	) engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func check_account_games(accid interface{}) (gameids []uint32) {
	tblname := get_account_game_table()
	str := fmt.Sprintf("select gameid from %s where accid=?", tblname)
	rows, err := db_monitor.Query(str, accid)

	defer rows.Close()
	if err != nil {
		return
	}
	gameids = make([]uint32, 0)
	for rows.Next() {
		var gameid uint32
		if err := rows.Scan(&gameid); err != nil {
			continue
		}
		gameids = append(gameids, gameid)
	}
	return
}

// 处理图片上传
func ImgUpload(task *unibase.ChanHttpTask) (uint32, string) {
	uEnc := task.R.FormValue("pictures")
	filename := task.R.FormValue("pictures_name")
	ext := path.Ext(filename)
	valid := false
	for _, v := range []string{".jpg", ".png", ".gif"} {
		if ext == v {
			valid = true
			break
		}
	}
	//判断文件后缀名
	if !valid {
		return 1, "file not support"
	}
	destfile := "/images/" + unibase.Rand.RandString(12) + ext
	err := os.MkdirAll(filepath.Dir(config.GetConfigStr("static")+destfile), os.ModeDir|os.ModePerm)
	if err != nil {
		return 1, "MkdirAll  error!"
	}
	i := strings.Index(uEnc, ",")
	if i < 0 {
		logging.Error("no comma")
	}
	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(uEnc[i+1:]))
	fi, err := os.Create(config.GetConfigStr("static") + destfile) //放入指定文件夹
	if err != nil {
		return 1, "Create  error!"
	}
	defer fi.Close()
	if _, err := io.Copy(fi, dec); err != nil {
		return 1, "copy  error!"
	}
	return 0, destfile
}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	logging.Error(path)
	if err == nil {
		logging.Error("存在")
		return true, nil
	}
	if os.IsNotExist(err) {
		logging.Error("不存在")
		return false, nil
	}
	return false, err
}

// type 1端游，2手游，3页游; conntype: 1 tcp接入，2 http接入
func HandleAddGame(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	gamename := strings.TrimSpace(task.R.FormValue("gamename"))
	gamekey := strings.TrimSpace(task.R.FormValue("gamekey"))
	remarks := strings.TrimSpace(task.R.FormValue("remarks"))
	gtype := uint32(unibase.Atoi(task.R.FormValue("gtype"), 0))
	conntype := uint32(unibase.Atoi(task.R.FormValue("conntype"), 0))
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	logolink := ""
	if gameid != 0 && gamename != "" {
		tblname := get_game_table()
		str := fmt.Sprintf("insert into %s(gameid, gamename, gamekey,  remarks, type, conntype, state) values(?,?,?,?,?,?,?)", tblname)
		result, err := db_monitor.Exec(str, gameid, gamename, gamekey, remarks, gtype, conntype, state)
		if err == nil {
			//创建游戏admin自动添加管理
			tblname1 := get_account_game_table()
			str2 := fmt.Sprintf("insert into %s (accid,gameid) values(?,?)", tblname1)
			tt := 1
			_, err2 := db_monitor.Exec(str2, tt, gameid)
			if err2 != nil {
				task.Error("admin add new gameid error:%s", err2)
				return
			}
			if lastid, err := result.LastInsertId(); err == nil && lastid > 0 {
				//数据添加成功再处理图片
				//处理图片
				if i, s := ImgUpload(task); i == 1 {
					task.Error("err:%s", s)
					task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, s)))
					return
				} else {
					logolink = s
				}
				str1 := fmt.Sprintf("update %s set logolink=? where gameid=?", tblname)
				_, err1 := db_monitor.Exec(str1, logolink, gameid)
				if err1 != nil {
					task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err1.Error())))
				}
				task.SendBinary([]byte(`{"retcode":0,"retdesc":"add game success"}`))
				return
			}

		}
		if err != nil {
			task.Error("HandleAddGame error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
			return
		}

	}
	task.Warning("HandleAddGame gameid:%d, gamename:%d, logolink:%s", gameid, gamename, logolink)
	task.SendBinary([]byte(`{"retcode":1,"retdesc":"add game failed"}`))

}

func InSlice(us []uint32, e uint32) bool {
	for _, u := range us {
		if e == u {
			return true
		}
	}
	return false
}
func gamelist1(task *unibase.ChanHttpTask) {
	var gamelist []map[string]interface{}
	tblname := get_game_table()
	str := fmt.Sprintf("select gameid, gamename, gamekey	, remarks, type ,conntype ,state from %s ", tblname)
	row, err := db_monitor.Query(str)
	if err == nil {
		defer row.Close()
		var gameid, conntype, gtype, state uint32
		var gamename, gamekey, remarks string
		for row.Next() {
			if err = row.Scan(&gameid, &gamename, &gamekey, &remarks, &gtype, &conntype, &state); err != nil {
				task.Warning("HandleGameList error:%s", err.Error())
				continue
			}
			gamelist = append(gamelist, map[string]interface{}{"gameid": gameid, "gamename": gamename, "gamekey": gamekey, "remarks": remarks, "type": gtype, "conntype": conntype, "state": state})
		}
	} else {
		task.Error("HandleGameList error:%s", err.Error())
	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": gamelist})
	task.SendBinary(data)
	task.Debug("gamelist:%v", gamelist)
}
func gamelist2(task *unibase.ChanHttpTask, id uint32) {
	var gamelist []map[string]interface{}
	tblname := get_game_table()
	str := fmt.Sprintf("select gameid, gamename, gamekey	, remarks, type ,conntype ,state from %s where gameid=?", tblname)
	row, err := db_monitor.Query(str, id)
	if err == nil {
		defer row.Close()
		var gameid, conntype, gtype, state uint32
		var gamename, gamekey, remarks string
		for row.Next() {
			if err = row.Scan(&gameid, &gamename, &gamekey, &remarks, &gtype, &conntype, &state); err != nil {
				task.Warning("HandleGameList error:%s", err.Error())
				continue
			}
			gamelist = append(gamelist, map[string]interface{}{"gameid": gameid, "gamename": gamename, "gamekey": gamekey, "remarks": remarks, "type": gtype, "conntype": conntype, "state": state})
		}
	} else {
		task.Error("HandleGameList error:%s", err.Error())
	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": gamelist})
	task.SendBinary(data)
	task.Debug("gamelist:%v", gamelist)
}
func HandleGameList2(task *unibase.ChanHttpTask) {
	curtype := uint32(unibase.Atoi(task.R.FormValue("curtype"), 0))
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	//account := task.R.FormValue("account")
	if curtype == 1 {
		gamelist1(task)
	} else if curtype == 2 {
		gamelist2(task, gameid)
	}

}
func HandleGameList1(task *unibase.ChanHttpTask) {
	var gamelist []map[string]interface{}
	tblname := get_game_table()
	str := fmt.Sprintf("select gameid, gamename from %s", tblname)
	row, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleGameList error:%s", err.Error())
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"HandleGameList1 error"}`))
		return
	}
	defer row.Close()
	var gameid uint32
	var gamename string
	for row.Next() {
		if err = row.Scan(&gameid, &gamename); err != nil {
			task.Warning("HandleGameList1 error:%s", err.Error())
			continue
		}
		gamelist = append(gamelist, map[string]interface{}{"gameid": gameid, "gamename": gamename})
	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": gamelist})
	task.SendBinary(data)
	task.Debug("gamelist:%v", gamelist)
}

func HandleGameList3(task *unibase.ChanHttpTask) {
	var gamelist []map[string]interface{}
	accid := uint32(unibase.Atoi(task.R.FormValue("accid"), 0))
	gameids := check_account_games(accid)
	//task.Warning("gameids:%v", gameids)
	tblname := get_game_table()
	for i := range gameids {
		str := fmt.Sprintf("select gameid, gamename from %s where gameid=?", tblname)
		var gameid uint32
		var gamename string
		err := db_monitor.QueryRow(str, gameids[i]).Scan(&gameid, &gamename)
		if err == nil {
			gamelist = append(gamelist, map[string]interface{}{"gameid": gameid, "gamename": gamename})
		} else {
			task.Error("HandleGameList3 error:%s", err.Error())
		}
	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": gamelist})
	task.SendBinary(data)
	task.Debug("gamelist:%v", gamelist)
}

func HandleGameList(task *unibase.ChanHttpTask) {
	var gamelist []map[string]interface{}
	accid, _ := GetSecureCookieEx(task.R, config.GetConfigStr("secret_key"), "sid")
	gameids := check_account_games(accid)
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	//task.Error("state:%d",state)
	tblname := get_game_table()
	str := fmt.Sprintf("select gameid, gamename, logolink, remarks, type from %s where state=?", tblname)
	row, err := db_monitor.Query(str, state)
	if err == nil {
		defer row.Close()
		var gameid, gtype uint32
		var gamename, logolink, remarks string
		for row.Next() {
			if err = row.Scan(&gameid, &gamename, &logolink, &remarks, &gtype); err != nil {
				task.Warning("HandleGameList error:%s", err.Error())
				continue
			}
			if gameids != nil && len(gameids) > 0 && InSlice(gameids, gameid) == false {
				continue
			}
			gamelist = append(gamelist, map[string]interface{}{"gameid": gameid, "gamename": gamename, "logolink": logolink, "remarks": remarks, "type": gtype})
		}
	} else {
		task.Error("HandleGameList error:%s", err.Error())
	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": gamelist})
	task.SendBinary(data)
	task.Debug("gamelist:%v", gamelist)
}

func HandleGameUpdate(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	modify_gamename := task.R.FormValue("modify_gamename")
	modify_gamekey := task.R.FormValue("modify_gamekey")
	modify_remarks := task.R.FormValue("modify_remarks")
	modify_type := uint32(unibase.Atoi(task.R.FormValue("modify_type"), 0))
	modify_conntype := uint32(unibase.Atoi(task.R.FormValue("modify_conntype"), 0))
	modify_state := uint32(unibase.Atoi(task.R.FormValue("modify_state"), 0))
	pictures := task.R.FormValue("pictures")
	//pictures_name := task.R.FormValue("pictures_name")
	tblname := get_game_table()
	message := "update game "
	if modify_gamename != "" {
		str := fmt.Sprintf("update %s set gamename=? where gameid=?", tblname)
		_, err := db_monitor.Exec(str, modify_gamename, gameid)
		if err != nil {
			task.Error("HandleGameUpdate gamename error:%s", err.Error())
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"update game gamename failed"}`))
		} else {
			message += "gamename "
		}
	}
	if modify_gamekey != "" {
		str := fmt.Sprintf("update %s set gamekey=? where gameid=?", tblname)
		_, err := db_monitor.Exec(str, modify_gamekey, gameid)
		if err != nil {
			task.Error("HandleGameUpdate gamekey error:%s", err.Error())
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"update game gamekey failed"}`))
		} else {
			message += "gamekey "
		}
	}
	if modify_remarks != "" {
		str := fmt.Sprintf("update %s set remarks=? where gameid=?", tblname)
		_, err := db_monitor.Exec(str, modify_remarks, gameid)
		if err != nil {
			task.Error("HandleGameUpdate remarks error:%s", err.Error())
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"update game remarks failed"}`))
		} else {
			message += "remarks "
		}
	}
	if modify_type != 0 {
		str := fmt.Sprintf("update %s set type=? where gameid=?", tblname)
		_, err := db_monitor.Exec(str, modify_type, gameid)
		if err != nil {
			task.Error("HandleGameUpdate type error:%s", err.Error())
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"update game type failed"}`))
		} else {
			message += "type "
		}
	}
	if modify_conntype != 0 {
		str := fmt.Sprintf("update %s set conntype=? where gameid=?", tblname)
		_, err := db_monitor.Exec(str, modify_conntype, gameid)
		if err != nil {
			task.Error("HandleGameUpdate conntype error:%s", err.Error())
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"update game conntype failed"}`))
		} else {
			message += "conntype "
		}
	}
	if modify_state != 3 {
		str := fmt.Sprintf("update %s set state=? where gameid=?", tblname)
		_, err := db_monitor.Exec(str, modify_state, gameid)
		if err != nil {
			task.Error("HandleGameUpdate state error:%s", err.Error())
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"update game state failed"}`))
		} else {
			message += "state "
		}
	}
	if pictures != "" { //修改游戏图片
		logolink1 := ""
		i, s := ImgUpload(task)
		if i == 1 {
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"upload game pictures error"}`))
			return
		} else {
			logolink1 = s
		}
		//找到修改前的图片删除
		logolink := ""
		str1 := fmt.Sprintf("select logolink from %s where gameid=?", tblname)
		err1 := db_monitor.QueryRow(str1, gameid).Scan(&logolink)
		if err1 != nil {
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err1.Error())))
			return
		}
		//sxtr,_ := os.Getwd()
		//logging.Error("str := %s",sxtr)
		logolink = "./monitor_www/static" + logolink
		err2 := os.Remove(logolink) //删除图片出现问题
		if err2 != nil {
			logging.Error("delete %s err", logolink)
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"delete picture error"}`))
			return
		} else {
			logging.Debug("delete %s success", logolink)
		}
		//再修改新图片地址
		str := fmt.Sprintf("update %s set logolink=? where gameid=?", tblname)
		_, err := db_monitor.Exec(str, logolink1, gameid)
		if err != nil {
			task.Error("HandleGameUpdate logolink error:%s", err.Error())
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"update game logolink failed"}`))
		} else {
			message += "logolink "
		}
	}
	message += "success!"
	task.SendBinary([]byte(fmt.Sprintf(`{"retcode":0,"retdesc":"%s"}`, message)))
}

func HandleDelGame(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	logolink := ""
	if gameid != 0 {
		tblname := get_game_table()
		//删除游戏前  先删除服务器上的图片
		str1 := fmt.Sprintf("select logolink from %s where gameid=?", tblname)
		err1 := db_monitor.QueryRow(str1, gameid).Scan(&logolink)
		if err1 != nil {
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err1.Error())))
			return
		}
		sxtr, _ := os.Getwd()
		logging.Error("str := %s", sxtr)
		logolink = "./monitor_www/static" + logolink
		err2 := os.Remove(logolink) //删除图片出现问题
		if err2 != nil {
			logging.Error("delete %s err", logolink)
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"delete picture error"}`))
			return
		} else {
			logging.Error("delete %s success", logolink)
		}

		str := fmt.Sprintf("delete from %s where gameid=? ", tblname)
		result, err := db_monitor.Exec(str, gameid)
		if err == nil {
			if rows, err := result.RowsAffected(); err == nil && rows > 0 {
				tblname = get_zone_table()
				str = fmt.Sprintf("delete from %s where gameid=? ", tblname)
				db_monitor.Exec(str, gameid)
			}
		}
		if err != nil {
			task.Error("HandleDelGame error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
			return
		}

		//先查找对应的account表数据 并删除 admin除外
		tblname4 := get_account_game_table()
		str4 := fmt.Sprintf("select accid from %s where gameid=?", tblname4)
		row4, err4 := db_monitor.Query(str4, gameid)

		if err4 == nil {
			defer row4.Close()
			var accid []uint32
			for row4.Next() {
				var del_accid uint32
				if err4 = row4.Scan(&del_accid); err4 != nil {
					task.Warning("select delect accid error:%s", err4.Error())
					continue
				}
				if del_accid != 1 {
					accid = append(accid, del_accid)
				}

			}
			//删除account表中此游戏下的用户数据
			tblname5 := get_account_table()
			str5 := fmt.Sprintf("delete from %s where id=?", tblname5)
			for _, i := range accid {
				_, err5 := db_monitor.Exec(str5, &i)
				if err5 != nil {
					task.Warning("delete accid error:%s,id:%d", err5.Error(), i)
					continue
				}
			}
		}
		//删除account_games表数据
		tblname3 := get_account_game_table()
		str3 := fmt.Sprintf("delete from %s where gameid=? ", tblname3)
		_, err3 := db_monitor.Exec(str3, &gameid)
		if err3 != nil {
			task.Error("delete account_games error:%s", err3.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err2.Error())))
			return
		}
	}
	task.SendBinary([]byte(`{"retcode":0,"retdesc":"delete game success"}`))
}

// 处理上传游戏图片
func HandlerImgUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		f, h, err := r.FormFile("gameimage")
		r.ParseForm()
		logging.Error("filename:%s", h.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		filename := h.Filename
		ext := path.Ext(filename)
		valid := false
		for _, v := range []string{".jpg", ".png", ".gif"} {
			if ext == v {
				valid = true
				break
			}
		}
		if !valid {
			w.Write([]byte(`{"retcode": 1, "retdesc":"file not support"}`))
			return
		}
		destfile := "/images/" + unibase.Rand.RandString(12) + ext
		os.MkdirAll(filepath.Dir(config.GetConfigStr("static")+destfile), os.ModeDir|os.ModePerm)
		t, err := os.Create(config.GetConfigStr("static") + destfile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer t.Close()
		if _, err := io.Copy(t, f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		str := fmt.Sprintf(`{"retcode": 0, "gameimage":"%s"}`, destfile)
		w.Write([]byte(str))
		return
	}
}

func HandleGamePlatList(task *unibase.ChanHttpTask) {
	var platlist []uint32
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	if gameid != 0 {
		tblname := get_user_data_table(gameid)
		ptlist := getAccountPlatList(task)
		var str string
		if len(ptlist) > 0 {
			str = fmt.Sprintf("select distinct platid from %s where platid in (%s)", tblname, strings.Join(ptlist, ","))
		} else {
			str = fmt.Sprintf("select distinct platid from %s", tblname)
		}
		row, err := db_monitor.Query(str)
		if err == nil {
			defer row.Close()
			var platid uint32
			for row.Next() {
				if err = row.Scan(&platid); err != nil {
					task.Warning("HandleGamePlatList error:%s", err.Error())
					continue
				}
				platlist = append(platlist, platid)
			}
		} else {
			task.Error("HandleGamePlatList gameid:%d, error:%s", gameid, err.Error())
		}
	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": platlist})
	task.SendBinary(data)
	task.Debug("platlist:%v", platlist)
}

// 汇率查询
func HandleExchangeRateList(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))

	var ratelist []map[string]interface{}
	if gameid != 0 {
		tblname := get_exchange_rate_table(gameid)
		str := fmt.Sprintf("select time , exchange_rate,currency from %s", tblname)

		row, err := db_monitor.Query(str)
		if err == nil {
			defer row.Close()

			for row.Next() {
				var time, exchange_rate, currency string
				if err = row.Scan(&time, &exchange_rate, &currency); err != nil {
					task.Warning("HandleExchangeRateList error:%s", err.Error())
					continue
				}
				ratelist = append(ratelist, map[string]interface{}{"times": time, "exchange_rate": exchange_rate, "currency": currency})
			}
		} else {
			task.Error("HandleExchangeRateList gameid:%d, error:%s", gameid, err.Error())
		}

	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": ratelist})
	task.SendBinary(data)
	task.Debug("platlist:%v", ratelist)

}
func HandleExchangeRateUpdate(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	acttype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	time := strings.TrimSpace(task.R.FormValue("time"))
	currency := uint32(unibase.Atoi(task.R.FormValue("currency"), 1))
	exchange_rate := strings.TrimSpace(task.R.FormValue("exchange_rate"))

	if gameid != 0 {
		tblname := get_exchange_rate_table(gameid)
		var str string
		if acttype == 1 {
			str = fmt.Sprintf("insert into %s set exchange_rate = ? ,currency=?, time= ? ", tblname)

		} else {
			str = fmt.Sprintf("update %s set exchange_rate = ? where  currency= ? and  time=?", tblname)
		}
		_, err := db_monitor.Exec(str, exchange_rate, currency, time)
		if err != nil {
			task.Error("HandleExchangeRateUpdate error:%s", err.Error())

			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
			return
		} else {
			if currency != 1 {
				tblcost := get_app_cost_today_table(gameid)
				tbllaunch := get_launch_keys_table(gameid)
				timek := strings.Replace(time, "-", "", -1)

				str = fmt.Sprintf("update %s set cost = currency * %s , rate = %s  where daynum = %s and ad_account_id in (SELECT ad_account_id from %s where currency = %d) ", tblcost, exchange_rate, exchange_rate, timek, tbllaunch, currency)

				_, err = db_monitor.Exec(str)
			}

			task.SendBinary([]byte(`{"retcode":0,"retdesc":"action success"}`))
			return
		}

	}
	task.Warning("HandleExchangeRateUpdate gameid:%d, time:%d, exchange_rate:%s", gameid, time, exchange_rate)
	task.SendBinary([]byte(`{"retcode":1,"retdesc":"action failed"}`))
}
func HandleLaunchChannelList(task *unibase.ChanHttpTask) {
	task.Info("unitime.Time.Now():", unitime.Time.Now())
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))

	var launchlist []map[string]interface{}
	if gameid != 0 {
		tblname := get_launch_keys_table(gameid)
		str := fmt.Sprintf("select id , launchid,keywords,ad_account_id,agent_id,status,token,appid,currency, updated from %s", tblname)

		row, err := db_monitor.Query(str)
		if err == nil {
			defer row.Close()

			for row.Next() {
				var id, status, appid int
				var launchid, keywords, updated, ad_account_id, agent_id, token, currency string
				if err = row.Scan(&id, &launchid, &keywords, &ad_account_id, &agent_id, &status, &token, &appid, &currency, &updated); err != nil {
					task.Warning("HandleExchangeRateList error:%s", err.Error())
					continue
				}
				launchlist = append(launchlist, map[string]interface{}{"id": id, "launchid": launchid, "keywords": keywords, "updated": updated, "ad_account_id": ad_account_id, "agent_id": agent_id, "status": status, "token": token, "appid": appid, "currency": currency})
			}
		} else {
			task.Error("HandleExchangeRateList gameid:%d, error:%s", gameid, err.Error())
		}

	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": launchlist})
	task.SendBinary(data)
	task.Debug("platlist:%v", launchlist)
}
func HandleLaunchKeywordsAction(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	acttype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	channel := strings.TrimSpace(task.R.FormValue("channel"))
	id := strings.TrimSpace(task.R.FormValue("id"))
	token := strings.TrimSpace(task.R.FormValue("token"))
	status := uint32(unibase.Atoi(task.R.FormValue("status"), 0))
	ad_account_id := strings.TrimSpace(task.R.FormValue("ad_account_id"))
	agent_id := strings.TrimSpace(task.R.FormValue("agent_id"))
	appid := uint32(unibase.Atoi(task.R.FormValue("appid"), 0))
	currency := uint32(unibase.Atoi(task.R.FormValue("currency"), 1))

	if gameid != 0 {
		tblname := get_launch_keys_table(gameid)
		if acttype == 1 {
			strs := fmt.Sprintf("select count(*) from %s where launchid = ? and ad_account_id= ? ", tblname)

			couns := db_monitor.QueryRow(strs, channel, ad_account_id)

			var count uint32

			couns.Scan(&count)

			if count > 0 {
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"该渠道已存在该账号，不可重复添加"}`)))
				return
			}

			str := fmt.Sprintf("insert into %s set launchid = ? , token= ? , ad_account_id= ? , agent_id = ? , updated = ? , status=? , appid=?,currency=?", tblname)
			_, err := db_monitor.Exec(str, channel, token, ad_account_id, agent_id, unitime.Time.Now(), status, appid, currency)
			if err != nil {
				task.Error("HandleLaunchKeywordsAction error:%s", err.Error())

				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
				return
			} else {
				task.SendBinary([]byte(`{"retcode":0,"retdesc":"action success"}`))
				return
			}

		} else if acttype == 2 {
			str := fmt.Sprintf("update %s set launchid = ? , token = ? ,ad_account_id= ? , agent_id = ? , updated = ? , status = ? , appid=? , currency=? where id = ? ", tblname)
			_, err := db_monitor.Exec(str, channel, token, ad_account_id, agent_id, unitime.Time.Now(), status, appid, currency, id)
			if err != nil {
				task.Error("HandleLaunchKeywordsAction error:%s", err.Error())

				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
				return
			} else {
				task.SendBinary([]byte(`{"retcode":0,"retdesc":"action success"}`))
				return
			}
		} else if acttype == 4 {
			str := fmt.Sprintf("update %s set ", tblname)
			if agent_id != "" {
				str += fmt.Sprintf(" agent_id = %s and  ", agent_id)
			}
			if token != "" {
				str += fmt.Sprintf(" token = %s and  ", token)
			}
			str += fmt.Sprintf(" status = ? where ad_account_id in (%s)  ", id)

			fmt.Println("str:", str)

			_, err := db_monitor.Exec(str, status)
			if err != nil {
				task.Error("HandleLaunchKeywordsAction error:%s", err.Error())

				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
				return
			} else {
				task.SendBinary([]byte(`{"retcode":0,"retdesc":"action success"}`))
				return
			}
		} else {
			str := fmt.Sprintf("delete from %s  where id = ? ", tblname)
			_, err := db_monitor.Exec(str, id)
			if err != nil {
				task.Error("HandleLaunchKeywordsAction error:%s", err.Error())

				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
				return
			} else {
				task.SendBinary([]byte(`{"retcode":0,"retdesc":"action success"}`))
				return
			}
		}
	}
	task.Warning("HandleLaunchKeywordsAction gameid:%d, channel:%d, ad_account_id:%s", gameid, channel, ad_account_id)
	task.SendBinary([]byte(`{"retcode":1,"retdesc":"action failed"}`))

}
func HandleAppNumberList(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	actype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))

	var applist []map[string]interface{}
	if gameid != 0 {
		tblname := get_adjapp_table(gameid)

		var str string
		if actype == 0 {
			str = fmt.Sprintf("select id , appname , appid , status , minute , rate from %s where id !=1000 and  appid > 0 ", tblname)
		} else {
			str = fmt.Sprintf("select id , appname , appid , status , minute , rate from %s where id !=1000 ", tblname)
		}

		row, err := db_monitor.Query(str)
		if err == nil {
			defer row.Close()

			for row.Next() {
				var id, status, minute int
				var appname, appid string
				var rate float32
				if err = row.Scan(&id, &appname, &appid, &status, &minute, &rate); err != nil {
					task.Warning("HandleExchangeRateList error:%s", err.Error())
					continue
				}
				applist = append(applist, map[string]interface{}{"id": id, "appname": appname, "appid": appid, "status": status, "minute": minute, "rate": rate})
			}
		} else {
			task.Error("HandleExchangeRateList gameid:%d, error:%s", gameid, err.Error())
		}

	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": applist})
	task.SendBinary(data)
	task.Debug("applist:%v", applist)

}
func HandleUpdateStatusApp(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	status := uint32(unibase.Atoi(task.R.FormValue("status"), 0))
	minute := uint32(unibase.Atoi(task.R.FormValue("minute"), 0))
	rate := strings.TrimSpace(task.R.FormValue("rate"))

	if gameid != 0 {
		tblname := get_adjapp_table(gameid)

		str := fmt.Sprintf("update %s set status = ? , minute = ? , rate = ? where id = ? ", tblname)

		_, err := db_monitor.Exec(str, status, minute, rate, id)

		if err != nil {
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"action failed"}`))
		} else {
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"action success"}`))
		}

	} else {
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"gameid error"}`))
	}

}
func HandleAppNumberAction(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	appid := strings.TrimSpace(task.R.FormValue("appid"))

	if gameid != 0 {
		tblname := get_adjapp_table(gameid)

		str := fmt.Sprintf("update %s set appid = ? where id = ? ", tblname)
		_, err := db_monitor.Exec(str, appid, id)
		if err != nil {
			task.Error("HandleAppNumberAction error:%s", err.Error())

			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
			return
		} else {
			task.SendBinary([]byte(`{"retcode":0,"retdesc":"action success"}`))
			return
		}

	}
	task.Warning("HandleAppNumberAction gameid:%d, ", gameid)
	task.SendBinary([]byte(`{"retcode":1,"retdesc":"action failed"}`))

}
func HandleDownloadexport(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	stime := uint32(unibase.Atoi(task.R.FormValue("stime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("etime"), 0))
	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	tbldata := get_user_data_table(gameid)

	where := " zoneid != 0 "

	appid := "0"

	if id > 0 {
		where += fmt.Sprintf(" and platid = %d", id)

		tblname := get_adjapp_table(gameid)

		str := fmt.Sprintf("select appid from %s where id = ? ", tblname)

		row := db_monitor.QueryRow(str, id)

		row.Scan(&appid)

	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" AND curr_launchid='%s' ", gameSystem)

	}

	where += fmt.Sprintf(" AND reg_time between %d and %d ", stime, etime)

	file, err := os.OpenFile("monitor_www/static/json/gps_adid.txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)

	if err != nil {
		task.Error("HandleDownloadFile error:%s", err.Error())
		task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
	}

	con := ""

	_, _ = io.WriteString(file, con)

	str := fmt.Sprintf(`select imei from %s where %s and imei !="" `, tbldata, where)

	rows, err := db_monitor.Query(str)

	defer rows.Close()

	if err != nil {
		task.Error("HandleDownloadFile error:%s", err.Error())

		task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
		return
	} else {
		for rows.Next() {

			var imei, text string

			if err = rows.Scan(&imei); err != nil {

				task.Error("HandleDownloadFile error:%s", err.Error())
				continue
			}

			if appid == "0" {

				text = fmt.Sprintf("%s\n", imei)
			} else {
				text = fmt.Sprintf("%s,%s\n", imei, appid)

			}
			fmt.Println("text:", text)
			_, err = file.WriteString(text)

		}
	}

	task.SendBinary([]byte(fmt.Sprintf(`{"retcode":0,"retdesc":"操作成功"}`)))

}
