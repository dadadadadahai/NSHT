package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
)

func get_user_data_table(gameid uint32) string {
	return fmt.Sprintf("user_data_%d", gameid)
}
func get_app_cost_today_table(gameid uint32) string {
	return fmt.Sprintf("app_cost_today_%d", gameid)
}

func get_account_table() string {
	return "monitor_account_info"
}

func create_account_table() {
	tblname := get_account_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		account varchar(64) not null default '',
		passwd varchar(128) not null default '',
		name varchar(64) not null default '',
		per_list bigint(20) not null default '0',
		game_list varchar(512) not null default '',
		plat_list varchar(512) not null default '',
		remarks varchar(256) not null default '',
		created_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		primary key (id),
		unique key index_account (account)
	) engine=MyISAM auto_increment=1 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_monitor.Exec(str)
	if err != nil {
		logging.Warning("db_monitor create table err:%s,%s", err.Error(), str)
	} else {
		logging.Info("db_monitor create table:%s", tblname)
	}
}

func monitor_account_default_insert() { //登录用户admin创建
	tblname := get_account_table()
	if !check_table_data_exists(tblname, "id", "1") {
		return
	}
	str := fmt.Sprintf("insert into %s(id, account, name,passwd, per_list, game_list, remarks) values(1,'admin','admin',md5(123456),0,'','')", tblname)
	result, err := db_monitor.Exec(str)
	if err == nil {
		if lastid, err := result.LastInsertId(); err == nil && lastid > 0 {
			logging.Debug("add account success")
			return
		}
	}
	if err != nil {
		logging.Warning("First Creation account error")
		return
	}
}
func check_table_data_exists(tablename, field, value string) bool { //判断?表?数据是否存在
	var count int
	str := fmt.Sprintf("select 1 from %s where %s = '%s' limit 1", tablename, field, value)
	db_monitor.QueryRow(str).Scan(&count)
	if count > 0 {
		return false
	}
	return true
}
func set_account_game(id, gameid uint32) string {
	tblname := get_account_game_table()
	str := fmt.Sprintf("insert into %s(accid, gameid) values(?,?)", tblname)
	result, err := db_monitor.Exec(str, id, gameid)
	if err != nil {
		return "insert account_game error"
	}
	if lastid, err1 := result.LastInsertId(); err1 == nil && lastid > 0 {
		return ""
	}
	return ""
}
func HandleAddAccount(task *unibase.ChanHttpTask) {
	name := task.R.FormValue("username")
	account := task.R.FormValue("account")
	per_list := uint32(unibase.Atoi(task.R.FormValue("per_list"), 0)) //权限列表
	passwd := task.R.FormValue("passwd")
	game_list := task.R.FormValue("game_list") //游戏列表
	remarks := task.R.FormValue("remarks")
	if account != "" && passwd != "" && len(passwd) >= 6 {
		tblname := get_account_table()
		str := fmt.Sprintf("insert into %s(account, passwd,name, per_list, game_list, remarks) values(?,md5(?),?,?,?,?)", tblname)
		result, err := db_monitor.Exec(str, account, passwd, name, per_list, game_list, remarks)
		if err != nil {
			task.Error("HandleAddAccount error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
			return
		}
		lastid, err1 := result.LastInsertId()
		if err1 == nil && lastid > 0 {
			str2 := set_account_game(uint32(lastid), uint32(unibase.Atoi(game_list, 0))) //添加数据到account_game中
			if str2 != "" {
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, str2)))
			}
			task.SendBinary([]byte(`{"retcode":0,"retdesc":"add account success"}`))
			return
		}
	}
}

func HandleUpdateAccount(task *unibase.ChanHttpTask) {
	accid := uint32(unibase.Atoi(task.R.FormValue("accid"), 0))
	per_list := task.R.FormValue("per_list")   //权限列表
	game_list := task.R.FormValue("game_list") //游戏列表
	remarks := task.R.FormValue("remarks")
	if accid != 0 {
		tblname := get_account_table()
		str := fmt.Sprintf("update %s set per_list=?, game_list=?, remarks=? where id=?", tblname)
		result, err := db_monitor.Exec(str, per_list, game_list, remarks, accid)
		if err == nil {
			if rows, err := result.RowsAffected(); err == nil && rows > 0 {
				task.SendBinary([]byte(`{"retcode":0,"retdesc":"add account success"}`))
				return
			}
		}
		if err != nil {
			task.Error("HandleUpdateAccount error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":%s}`, err.Error())))
			return
		}
	}
	//task.Warning("HandleUpdateAccount accid:%d", accid)
	task.SendBinary([]byte(`{"retcode":1,"retdesc":"add account failed"}`))
}
func modify_account(task *unibase.ChanHttpTask) {
	id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	modify_name := task.R.FormValue("modify_name")
	modify_password := task.R.FormValue("modify_password")
	modify_per := uint32(unibase.Atoi(task.R.FormValue("modify_per"), 0))
	per_bool := task.R.FormValue("per_bool")
	add_game_list := task.R.FormValue("add_game_list")
	delete_game_list := task.R.FormValue("delete_game_list")

	data := "modify"
	if id != 0 {
		tblname := get_account_table()
		if modify_name != "" { //修改昵称
			str := fmt.Sprintf("update %s set  name=? where id=?", tblname)
			result, err := db_monitor.Exec(str, modify_name, id)
			if err != nil {
				task.Error("modify_name error:%s", err.Error())
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
				return
			}
			if rows, err := result.RowsAffected(); err == nil && rows > 0 {
				data += " name "
			}
		}
		if modify_password != "" { //修改密码
			str := fmt.Sprintf("update %s set  passwd=md5(?) where id=?", tblname)
			result, err := db_monitor.Exec(str, modify_password, id)
			if err != nil {
				task.Error("modify_passwd error:%s", err.Error())
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
				return
			}
			if rows, err := result.RowsAffected(); err == nil && rows > 0 {
				data += " passwd "
			}
		}
		if per_bool != "" { //修改权限
			str := fmt.Sprintf("update %s set  per_list=? where id=?", tblname)
			result, err := db_monitor.Exec(str, modify_per, id)
			if err != nil {
				task.Error("modify_per_list error:%s", err.Error())
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
				return
			}
			if rows, err := result.RowsAffected(); err == nil && rows > 0 {
				data += " per_list "
			}
		}
		var temp []string
		temp = json_to_arr(add_game_list, 1)
		if temp != nil { //添加用户信息的游戏

			//task.Warning("add_game_list:%V",temp)
			//task.Warning("add_game_list:%s",temp)

			tblname2 := get_account_game_table()
			str := fmt.Sprintf("insert into %s(accid, gameid) values(?,?)", tblname2)
			for i := range temp { //添加入Account_game表中
				_, err := db_monitor.Exec(str, id, temp[i])
				if err != nil {
					task.Warning("account_game add_game_list error", err.Error())
					task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
					return
				}
			}
			//添加用户表中的游戏
			tblname1 := get_account_table()
			str1 := fmt.Sprintf("select game_list from %s where id=?", tblname1)
			var game_list string
			err1 := db_monitor.QueryRow(str1, id).Scan(&game_list)
			if err1 != nil {
				task.Error("add_account_table:%s", err1.Error())
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err1.Error())))
			}
			for i := 0; i < len(temp); i++ {
				game_list += " " + temp[i]
			}
			//修改account_info表
			str2 := fmt.Sprintf("update %s set game_list=? where id=?", tblname1)
			_, err2 := db_monitor.Exec(str2, game_list, id)
			if err2 != nil {
				task.Error("update account game_list error:%s", err2.Error())
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err2.Error())))
				return
			}
			data += " add game_list "
		}
		var temp1 []string
		temp1 = json_to_arr(delete_game_list, 1)
		if temp1 != nil { //删除用户信息的游戏
			task.Warning("temp1:%s", temp1)
			tblname = get_account_game_table()
			str := fmt.Sprintf("delete from %s where accid=? and gameid=?", tblname)
			for i := range temp1 {
				_, err := db_monitor.Exec(str, id, temp1[i])
				if err != nil {
					task.Error("delete account_game error:%s", err.Error())
					task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
					return
				}
			}

			tblname1 := get_account_table()
			str1 := fmt.Sprintf("select game_list from %s where id=?", tblname1)
			var game_list string
			err1 := db_monitor.QueryRow(str1, id).Scan(&game_list)
			if err1 != nil {
				task.Error("delete_account_table:%s", err1.Error())
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err1.Error())))
			}
			var games []string
			games = json_to_arr(game_list, 2)
			task.Warning("1 games:%s", games)
			for i := range games { //删除游戏
				for j := 0; j < len(temp1); j++ {
					if games[i] == temp1[j] {
						games[i] = ""
					}
				}
			}
			var arr string
			for i := range games { //删除游戏
				if games[i] != "" {
					if i == len(games)-1 {
						arr += games[i]
					} else {
						arr += games[i] + " "
					}
				}
			}
			task.Warning("arr:%s", arr)

			//修改account_info表
			str2 := fmt.Sprintf("update %s set game_list=? where id=?", tblname1)
			_, err2 := db_monitor.Exec(str2, arr, id)
			if err2 != nil {
				task.Error("update account game_list error:%s", err2.Error())
				task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err2.Error())))
				return
			}
			data += " delete game_list "
		}
	}
	data += " success !"
	//task.Warning("modify_account accid:%d", id)
	send, _ := json.Marshal(map[string]interface{}{"retcode": 0, "retdesc": data})
	task.SendBinary(send)
}
func json_to_arr(s string, stype uint32) (arr []string) {

	var tt = []byte(s)
	if tt[0] == '[' && tt[1] == ']' {
		return nil
	}
	ii := 0
	if tt[ii] == '[' {
		tt = append(tt[:ii], tt[ii+1:]...)
	}
	if ii = len(tt) - 1; tt[ii] == ']' {
		tt = append(tt[:ii], tt[ii+1:]...)
	}
	if stype == 1 {
		var mm string = string(tt)
		arr = strings.Split(mm, ",")
	} else if stype == 2 {
		var mm string = string(tt)
		arr = strings.Split(mm, " ")
	}

	return arr
}

func HandleSettingAccount(task *unibase.ChanHttpTask) {
	curtype := uint32(unibase.Atoi(task.R.FormValue("curtype"), 0))
	if curtype == 1 { //设置用户信息
		modify_account(task)
	} else if curtype == 2 { //删除用户信息
		delete_account(task)
	}

}
func delete_account(task *unibase.ChanHttpTask) {
	accid := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	if accid != 0 {
		tblname := get_account_table()
		str := fmt.Sprintf("delete from %s where id=? ", tblname)
		_, err := db_monitor.Exec(str, accid)
		if err != nil {
			task.Error("delete account error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
			return
		}
		//继续删除account_gamede表中的数据
		tblname = get_account_game_table()
		str = fmt.Sprintf("delete from %s where accid=? ", tblname)
		_, err = db_monitor.Exec(str, accid)
		if err != nil {
			task.Error("delete account_game error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":"%s"}`, err.Error())))
			return
		}
	}
	task.SendBinary([]byte(`{"retcode":0,"retdesc":"delete game success"}`))
}

func HandleAccountList(task *unibase.ChanHttpTask) {
	curtype := uint32(unibase.Atoi(task.R.FormValue("curtype"), 0))
	account := task.R.FormValue("account")
	if curtype == 1 {
		accountlist1(task)
	} else if curtype == 2 {
		accountlist2(task, account)
	}
}
func accountlist2(task *unibase.ChanHttpTask, account1 string) {
	var accountlist []map[string]interface{}
	tblname := get_account_table()
	str := fmt.Sprintf("select id, account,name, per_list, plat_list, remarks from %s where account=?", tblname)
	row, err := db_monitor.Query(str, account1)
	if err == nil {
		defer row.Close()
		var accid uint64
		var account, name, per_list, plat_list, remarks string
		for row.Next() {
			if err = row.Scan(&accid, &account, &name, &per_list, &plat_list, &remarks); err != nil {
				task.Error("HandleAccountList error:%s", err.Error())
				continue
			}
			//通过accid获取gameids  拼接gamelist
			tblname1 := get_game_table()
			gameids := check_account_games(accid)
			var temp = ""
			for i := 0; i < len(gameids); i++ {
				str1 := fmt.Sprintf("select gamename from %s where gameid=?", tblname1)
				var gamename string
				err1 := db_monitor.QueryRow(str1, gameids[i]).Scan(&gamename)
				if err1 == nil {
					if i == len(gameids)-1 {
						temp += gamename
					} else {
						temp += gamename + " "
					}
				} else {
					task.Error("HandleGameList2 error:%s", err1.Error())
				}
			}
			accountlist = append(accountlist, map[string]interface{}{"accid": accid, "account": account, "name": name, "per_list": per_list, "game_list": temp, "plat_list": plat_list, "remarks": remarks})
		}
	} else {
		task.Error("HandleAccountList error:%s", err.Error())
	}
	data, _ := json.Marshal(map[string][]map[string]interface{}{"data": accountlist})
	task.SendBinary(data)
	task.Debug("accountlist:%v", accountlist)
}
func accountlist1(task *unibase.ChanHttpTask) {
	var accountlist []map[string]interface{}
	tblname := get_account_table()
	str := fmt.Sprintf("select id, account,name, per_list,plat_list, remarks from %s ", tblname)
	row, err := db_monitor.Query(str)
	if err == nil {
		defer row.Close()
		var accid uint64
		var account, name, per_list, plat_list, remarks string
		for row.Next() {
			if err = row.Scan(&accid, &account, &name, &per_list, &plat_list, &remarks); err != nil {
				task.Warning("HandleAccountList error:%s", err.Error())
				continue
			}
			//通过accid获取gameids  拼接gamelist
			tblname1 := get_game_table()
			gameids := check_account_games(accid)
			var temp = ""
			for i := 0; i < len(gameids); i++ {
				str := fmt.Sprintf("select gamename from %s where gameid=?", tblname1)
				var gamename string
				err := db_monitor.QueryRow(str, gameids[i]).Scan(&gamename)
				if err == nil {
					if i == len(gameids)-1 {
						temp += gamename
					} else {
						temp += gamename + " "
					}
				} else {
					task.Error("HandleGameList3 error:%s", err.Error())
				}
			}

			accountlist = append(accountlist, map[string]interface{}{"accid": accid, "account": account, "name": name, "per_list": per_list, "game_list": temp, "plat_list": plat_list, "remarks": remarks})
		}
	} else {
		task.Error("HandleAccountList error:%s", err.Error())
	}
	data, _ := json.Marshal(map[string][]map[string]interface{}{"data": accountlist})
	task.SendBinary(data)
	task.Debug("accountlist:%v", accountlist)
}

func HandleDelAccount(task *unibase.ChanHttpTask) {
	accid := uint32(unibase.Atoi(task.R.FormValue("accid"), 0))
	if accid != 0 {
		tblname := get_account_table()
		str := fmt.Sprintf("delete from %s where id=? ", tblname)
		_, err := db_monitor.Exec(str, accid)
		if err != nil {
			task.Error("HandleDelGame error:%s", err.Error())
			task.SendBinary([]byte(fmt.Sprintf(`{"retcode":1,"retdesc":%s}`, err.Error())))
			return
		}
	}
	task.SendBinary([]byte(`{"retcode":0,"retdesc":"delete game success"}`))
}

func HandleLoginAccount(task *unibase.ChanHttpTask) {
	account := task.R.FormValue("account")
	passwd := task.R.FormValue("passwd")

	if account != "" && passwd != "" && CheckXSRF(task.R, config.GetConfigStr("secret_key")) {
		tblname := get_account_table()
		str := fmt.Sprintf("select id, name from %s where account=? and passwd=md5(?) ", tblname)
		row := db_monitor.QueryRow(str, account, passwd)
		var accid uint64
		var name string
		if err := row.Scan(&accid, &name); err == nil {
			sid := strconv.FormatUint(accid, 10)
			session := sm.SessionInit(sid)
			session.Set("Sid", sid)
			session.Set("LRT", time.Now().Unix())
			SetCookie(task.W, "name", name)
			SetCookie(task.W, "account", account)

			SetSecureCookieEx(task.W, config.GetConfigStr("secret_key"), "sid", sid, session.createtime)
			task.SendBinary([]byte(`{"retcode":0,"retdesc":"login success"}`))
			return
		}
	}
	task.SendBinary([]byte(`{"retcode": 1, "retdesc":"login failed"}`))
	task.Debug("HandleLoginAccount account or password is none :%s", account)
}

// 获得玩家对应的渠道
func getAccountPlatList(task *unibase.ChanHttpTask) []string {
	sid, _ := GetSecureCookieEx(task.R, config.GetConfigStr("secret_key"), "sid")
	tblname := get_account_table()
	str := fmt.Sprintf("select id, account, plat_list from %s where id=%s", tblname, sid)
	var ret []string
	row, err := db_monitor.Query(str)
	if err == nil {
		defer row.Close()
		var accid uint64
		var account, plat_list string
		for row.Next() {
			if err = row.Scan(&accid, &account, &plat_list); err != nil {
				task.Error("getAccountPlatList error:%s", err.Error())
				continue
			}
			if plat_list != "" {
				ret = strings.Split(plat_list, ",")
			}
		}
	} else {
		task.Error("getAccountPlatList error:%s", err.Error())
	}
	//task.Warning("ret= %s  sid=%s", ret,sid)
	return ret
}
func HandleGetPri(task *unibase.ChanHttpTask) {
	account_name := GetCookie(task.R, "account")

	var per uint32
	tblname := get_account_table()
	str := fmt.Sprintf("select per_list from %s where account=?", tblname)

	err := db_monitor.QueryRow(str, account_name).Scan(&per)

	if err != nil { //未找到当前用户名信息
		task.SendBinary([]byte(`{"retcode": 1, "retdesc":"get user_priviliege failed"}`))
	}
	data, _ := json.Marshal(map[string]interface{}{"retcode": 0, "data": per})
	task.SendBinary(data)
}
func get_launch_keys_table(gameid uint32) string {
	return fmt.Sprintf("launch_keys_%d", gameid)
}
func get_exchange_rate_table(gameid uint32) string {
	return fmt.Sprintf("exchange_rate_%d", gameid)
}
