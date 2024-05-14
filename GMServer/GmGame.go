package main

import (
	"errors"
	"fmt"
	"strconv"

	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type Game struct {
	Id         uint64
	GameId     uint32 //游戏ID
	GameName   string //游戏名称
	GameLink   string //游戏主页
	GmLink     string //游戏GM接口地址
	GameType   uint32 //游戏类型,1手游，2端游，3页游
	Status     uint32 //状态，1有效，0无效
	Createtime uint32
}

func NewGame() *Game {
	return &Game{}
}

func (self *Game) IsValid() bool {
	return self.Id != uint64(0) && self.Status == uint32(1)
}

func get_gm_game_table() string {
	return "gm_games"
}

func get_gm_account_game_table() string {
	return "gm_account_games"
}

func create_gm_game() {
	tblname := get_gm_game_table()
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		gameid int(10) unsigned NOT NULL default '0',
		gamename varchar(64) NOT NULL default '',
		gamelink varchar(64) NOT NULL default '',
		gmlink varchar(64) NOT NULL default '',
		gamekey varchar(64) NOT NULL default '',
		gametype int(10) unsigned NOT NULL default '1',
		status tinyint(2) unsigned NOT NULL default '1',
		createtime int(10) unsigned NOT NULL default '0',
		primary key (id),
		unique key index_gameid_gamename (gameid,gamename)
	)engine=MyISAM auto_increment=1001 default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
	tblname = get_gm_account_game_table()
	str = fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		accid int(10) unsigned NOT NULL default '0',
		gameid int(10) unsigned NOT NULL default '0',
		primary key (id),
		unique key index_gameid_accid (accid, gameid)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err = db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

// 创建游戏
func create_game(gameid uint32, name, gamelink, gmlink string, gametype uint32) error {
	tblname := get_gm_game_table()
	str := fmt.Sprintf("replace into %s(gameid, gamename, gamelink, gmlink, gametype, createtime) values (?,?,?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, gameid, name, gamelink, gmlink, gametype, unitime.Time.Sec())
	if err == nil {
		_, err = result.LastInsertId()
	}
	return err
}

// 添加游戏
func add_game(gamename, gamelink, gmlink, gamekey, zonename string, gametype, zoneid, gameid uint32) (uint32, string) {
	//先判断添加的数据是否重复
	if !check_table_data_exists_2("gm_zones", "gameid", "zoneid", gameid, zoneid) {
		return 1, "check gm_zones err:区服id已重复!!"
	}
	if !check_table_data_exists1("gm_games", "gameid", gameid) {
		return 1, "check gm_games err:游戏id已重复!!"
	}
	//添加入game表
	tblname := get_gm_game_table()
	str := fmt.Sprintf("insert into %s(gameid, gamename, gamelink, gmlink,gamekey, gametype, createtime) values (?,?,?,?,?,?,?)", tblname)
	_, err := db_gm.Exec(str, gameid, gamename, gamelink, gmlink, gamekey, gametype, unitime.Time.Sec())
	if err != nil {
		logging.Error("add_gm_games error:%s, str:%s", err.Error(), str)
		return 1, err.Error()
	}
	//添加入大区表
	err1 := add_game_zone(gameid, zoneid, zonename, gmlink)
	if err1 != nil {
		logging.Error("add_game_zone err:%s", err1)
		return 1, err1.Error()
	}
	return 0, ""
}

// 添加修改类型
func (self *GmTask) ParseAddModifyTypes_CS(rev *Pmd.AddModifyTypes_CS) bool {
	if !self.VerifyOk {
		return false
	}
	tblname := get_gm_modify_types_table()
	str := fmt.Sprintf("insert into %s(gameid, typeid, typename) values (?,?,?)", tblname)
	_, err := db_gm.Exec(str, rev.GetGameid(), rev.GetTypeid(), rev.GetTypename())
	if err != nil {
		logging.Error("add_gm_games error:%s, str:%s", err.Error(), str)
		rev.Retcode = proto.Uint32(1)
		rev.Retdesc = proto.String(err.Error())
		self.SendCmd(rev)
		return false
	}
	content := "添加信息类型：" + rev.GetTypename()
	add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)

	rev.Retcode = proto.Uint32(0)
	self.SendCmd(rev)
	return true
}

// 删除  修改信息类型
func (self *GmTask) ParseDeleteModifyTypesMessage_CS(rev *Pmd.DeleteModifyTypesMessage_CS) bool {
	if !self.VerifyOk {
		return false
	}
	tblname := get_gm_modify_types_table()
	str := fmt.Sprintf("delete from %s where gameid=? and typeid=?", tblname)
	result, err := db_gm.Exec(str, rev.GetGameid(), rev.GetTypeid())
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		logging.Error("delete modify_types_message error,gameid:%d,typeid:%d", rev.GetGameid(), rev.GetTypeid())
		rev.Retcode = proto.Uint32(1)
		rev.Retdesc = proto.String("删除失败！")
		self.SendCmd(rev)
		return false
	}
	content := fmt.Sprintf("删除信息类型%d", rev.GetTypeid())
	add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	rev.Retcode = proto.Uint32(0)
	self.SendCmd(rev)
	return true
}

// 修改  修改信息类型
func (self *GmTask) ParseModifyModifyTypesMessage_CS(rev *Pmd.ModifyModifyTypesMessage_CS) bool {
	if !self.VerifyOk {
		return false
	}
	tblname := get_gm_modify_types_table()
	str := fmt.Sprintf("update %s set typename=? where gameid=? and typeid=?", tblname)
	result, err := db_gm.Exec(str, rev.GetTypename(), rev.GetGameid(), rev.GetTypeid())
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		logging.Error("update modify_types_message error,typename:%s,gameid:%d,typeid:%d", rev.GetTypename(), rev.GetGameid(), rev.GetTypeid())
		rev.Retcode = proto.Uint32(1)
		rev.Retdesc = proto.String("修改失败！")
		self.SendCmd(rev)
		return false
	}
	content := fmt.Sprintf("修改信息类型%d ， %s", rev.GetTypeid(), rev.GetTypename())
	add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	rev.Retcode = proto.Uint32(0)
	self.SendCmd(rev)
	return true
}

// 获取修改信息列表
func (self *GmTask) ParseGetModifyTypesList_CS(rev *Pmd.GetModifyTypesList_CS) bool {
	if !self.VerifyOk {
		return false
	}
	//无条件获取列表
	if rev.GetCurtype() == 1 {
		tblname := get_gm_modify_types_table()
		str := fmt.Sprintf("select gameid,typeid,typename from %s ", tblname)
		rows, err := db_gm.Query(str)
		if err != nil {
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.ModifyTypesList, 0)
		for rows.Next() {
			data := &Pmd.ModifyTypesList{}
			if err := rows.Scan(&data.Gameid, &data.Typeid, &data.Typename); err != nil {
				continue
			}
			tblname1 := get_gm_game_table()
			str1 := fmt.Sprintf("select gamename from %s where gameid=?", tblname1)
			row2 := db_gm.QueryRow(str1, data.Gameid)
			err1 := row2.Scan(&data.Gamename)
			if err1 != nil {
				logging.Error("select gamename error:%s, str:%s", err1.Error(), str1)
				continue
			}
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(modify_types_list) == 0")
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("len(modify_types_list) == 0")
			self.SendCmd(rev)
			return false
		}
		//logging.Error("rev.data = ",rev.Data)
		rows.Close()
	} else if rev.GetCurtype() == 2 {
		tblname := get_gm_modify_types_table()
		str := fmt.Sprintf("select typeid,typename from %s where gameid=?", tblname)
		rows, err := db_gm.Query(str, rev.GetGameid())
		if err != nil {
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.ModifyTypesList, 0)
		for rows.Next() {
			data := &Pmd.ModifyTypesList{}
			if err := rows.Scan(&data.Typeid, &data.Typename); err != nil {
				continue
			}
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(modify_types_list1) == 0")
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("len(modify_types_list1) == 0")
			self.SendCmd(rev)
			return false
		}
		rows.Close()
	}
	self.SendCmd(rev)
	return true
}

// 获取列表
func (self *GmTask) ParseGmGameOrZonelistGet_CS(rev *Pmd.GmGameOrZonelistGet_CS) bool {
	if !self.VerifyOk {
		return false
	}
	//logging.Error("进入函数")
	if rev.GetCurtype() == 1 { //1、查询gm_games表返回游戏名列表
		// tblname := get_gm_game_table()
		str := fmt.Sprintf("select id,name from apps ")
		rows, err := unibase.DBZ.Query(str)
		if err != nil {
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.GmListInfo, 0)
		for rows.Next() {
			data := &Pmd.GmListInfo{}
			if err := rows.Scan(&data.Gameid, &data.Gamename); err != nil {
				continue
			}
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(user_list) == 0")
			ty := -1
			*rev.Curtype = uint32(ty)
			self.SendCmd(rev)
			return false
		}
		//logging.Error("rev.data = ",rev.Data)
		rows.Close()
	} else if rev.GetCurtype() == 2 { //2、查询gm_zones表返回游戏大区名列表
		tblname := get_gm_zone_table()
		str := fmt.Sprintf("select zoneid,zonename from %s where gameid=?", tblname)
		rows, err := db_gm.Query(str, rev.GetGameid())
		if err != nil {
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.GmListInfo, 0)
		for rows.Next() {
			data := &Pmd.GmListInfo{}
			if err := rows.Scan(&data.Zoneid, &data.Zonename); err != nil {
				continue
			}
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(user_list) == 0")
			ty := -1
			*rev.Curtype = uint32(ty)
			self.SendCmd(rev)
			return false
		}
		//logging.Error("rev.data = ",rev.Data)
		rows.Close()
	}
	self.SendCmd(rev)
	return true
}

// 大区操作
func (self *GmTask) ParseRequestZoneOperation_CS(rev *Pmd.RequestZoneOperation_CS) bool {
	if !self.VerifyOk {
		return false
	}
	if rev.GetCurtype() == 1 { //获取大区列表
		tblname := get_gm_zone_table()
		str := fmt.Sprintf("select count(*) from %s where gameid=?", tblname)
		row := db_gm.QueryRow(str, rev.GetGameid())
		var count uint32
		if err := row.Scan(&count); err != nil {
			logging.Error("checkFeedbackList error:%s", err.Error())
			return false
		}
		if rev.GetPerpage() == 0 {
			*rev.Perpage, *rev.Curpage = 15, 1
		}
		*rev.Maxpage = count / rev.GetPerpage()
		if count > 0 && *rev.Maxpage == 0 {
			*rev.Maxpage = 1
		} else if count%rev.GetPerpage() != 0 {
			*rev.Maxpage += 1
		}
		if rev.GetCurpage() > rev.GetMaxpage() {
			logging.Debug("checkFeedbackList sql:%s ", str)
			return false
		}
		var str1 string
		if rev.GetGameid() == 1001 {
			str1 = fmt.Sprintf("select id,gameid,zoneid,zonename,gmlink,status from %s order by id asc limit %d, %d", tblname, (rev.GetCurpage()-1)*rev.GetPerpage(), rev.GetPerpage())
		} else {
			str1 = fmt.Sprintf("select id,gameid,zoneid,zonename,gmlink,status from %s where gameid=%d order by id asc limit %d, %d", tblname, rev.GetGameid(), (rev.GetCurpage()-1)*rev.GetPerpage(), rev.GetPerpage())
		}
		rows, err := db_gm.Query(str1)
		if err != nil {
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.ZoneList, 0)
		for rows.Next() {
			data := &Pmd.ZoneList{}
			if err := rows.Scan(&data.Id, &data.Gameid, &data.Zoneid, &data.Zonename, &data.Gmlink, &data.Status); err != nil {
				continue
			}
			tblname1 := get_gm_game_table()
			str2 := fmt.Sprintf("select gamename from %s where gameid=?", tblname1)
			row2 := db_gm.QueryRow(str2, data.GetGameid())
			err2 := row2.Scan(&data.Gamename)
			if err2 != nil {
				logging.Error("select gamename error ")
			}
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(zone_list) == 0")
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("查询失败！")
			return false
		}
		rows.Close()
	} else if rev.GetCurtype() == 2 { //修改大区信息
		tblname := get_gm_zone_table()
		str := fmt.Sprintf("update %s set zonename=?,gmlink=?,status=? where gameid=? and zoneid=?", tblname)
		result, err := db_gm.Exec(str, rev.GetModifygzonename(), rev.GetModifygmlink(), rev.GetModifystatus(), rev.GetGameid(), rev.GetZoneid())
		if err == nil {
			_, err = result.RowsAffected()
		}
		//logging.Error("sql:%s",str)
		if err != nil {
			logging.Error("update zonemsg error,name:%s,gmlink:%s,status:%d", rev.GetModifygzonename(), rev.GetModifygmlink(), rev.GetModifystatus())
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("未查询到此数据！")
			self.SendCmd(rev)
			return false
		}
		content := fmt.Sprintf("修改大区%s", rev.GetZoneid())
		add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	} else if rev.GetCurtype() == 3 { //通过大区id获取大区信息
		tblname := get_gm_zone_table()
		str := fmt.Sprintf("select id,gameid,zoneid,zonename,gmlink,status from %s where gameid=? and zoneid=?", tblname)
		rows, err := db_gm.Query(str, rev.GetGameid(), rev.GetZoneid())
		if err != nil {
			self.SendCmd(rev)
			return false
		}
		//rev.Data = make([]*Pmd.ZoneList, 0)
		for rows.Next() {
			data := &Pmd.ZoneList{}
			if err := rows.Scan(&data.Id, &data.Gameid, &data.Zoneid, &data.Zonename, &data.Gmlink, &data.Status); err != nil {
				continue
			}
			tblname1 := get_gm_game_table()
			str2 := fmt.Sprintf("select gamename from %s where gameid=?", tblname1)
			row2 := db_gm.QueryRow(str2, data.GetGameid())
			err2 := row2.Scan(&data.Gamename)
			if err2 != nil {
				logging.Error("select gamename error ")
			}
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(zone_list) == 0,rev.GetCurtype():%d", rev.GetCurtype())
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("未查询到此数据！")
			self.SendCmd(rev)
			return false
		}
		rows.Close()
	} else if rev.GetCurtype() == 4 { //删除大区
		tblname := get_gm_zone_table()
		str := fmt.Sprintf("delete from %s where gameid=? and zoneid=?", tblname)
		result, err := db_gm.Exec(str, rev.GetGameid(), rev.GetZoneid())
		if err == nil {
			_, err = result.RowsAffected()
			content := fmt.Sprintf("删除大区%s", rev.GetZoneid())
			add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
		}
		if err != nil {
			logging.Error("delete zonemsg error:%s,", err)
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("删除失败！")
			self.SendCmd(rev)
		}
	}
	self.SendCmd(rev)
	return true
}

// 获取管理员操作记录
// 获取管理员操作记录
func (self *GmTask) ParseGetManagerActionRecordList_CS(rev *Pmd.GetManagerActionRecordList_CS) bool {
	if !self.VerifyOk {
		return false
	}

	var perpage uint32
	perpage = 100

	tblname := get_manager_action_record_table()
	tblgame := get_user_account_table()

	str := ""
	str2 := ""
	var count uint32

	if rev.GetManagerid() != 0 {

		str = fmt.Sprintf("select a.id , a.gameid , a.gmid ,b.username, a.content , FROM_UNIXTIME(a.created_at) from %s as a , %s as b where a.gmid = b.id and a.gameid = ? and a.created_at between ? and  ? and a.gmid = ? order by a.created_at desc limit ? , ?", tblname, tblgame)
		str2 = fmt.Sprintf("select count(*) from %s where gameid = ? and created_at between ? and  ?  and gmid = ?", tblname)
		row := db_gm.QueryRow(str2, rev.GetGameid(), rev.GetStarttime(), rev.GetEndtime(), rev.GetManagerid())
		row.Scan(&count)
		rows, err := db_gm.Query(str, rev.GetGameid(), rev.GetStarttime(), rev.GetEndtime(), rev.GetManagerid(), (rev.GetCurpage()-1)*perpage, perpage)
		if err != nil {

			self.SendCmd(rev)
			return false
		}
		defer rows.Close()

		for rows.Next() {

			data := &Pmd.ManagerActionRecord{}
			err = rows.Scan(&data.Id, &data.Gameid, &data.Gmid, &data.Gamename, &data.Content, &data.CreatedAt)

			if err != nil {
				continue
			}
			// var game string
			// str = fmt.Sprintf("select  gamename from %s where gameid = ? ", get_gm_game_table())
			// result := db_gm.QueryRow(str, data.GetGameid())

			// result.Scan(&game)
			// data.Gamename = &game

			rev.Data = append(rev.Data, data)

		}
	} else {
		str = fmt.Sprintf("select a.id , a.gameid , a.gmid ,b.username, a.content , FROM_UNIXTIME(a.created_at) from %s as a , %s as b where a.gmid = b.id and a.gameid = ? and a.created_at between ? and  ?  order by a.created_at desc limit ? , ?", tblname, tblgame)

		str2 = fmt.Sprintf("select count(*) from %s where gameid = ? and created_at between ? and  ?", tblname)
		row := db_gm.QueryRow(str2, rev.GetGameid(), rev.GetStarttime(), rev.GetEndtime(), tblname)
		row.Scan(&count)
		rows, err := db_gm.Query(str, rev.GetGameid(), rev.GetStarttime(), rev.GetEndtime(), (rev.GetCurpage()-1)*perpage, perpage)

		if err != nil {

			self.SendCmd(rev)
			return false
		}
		defer rows.Close()

		for rows.Next() {

			data := &Pmd.ManagerActionRecord{}
			err = rows.Scan(&data.Id, &data.Gameid, &data.Gmid, &data.Gamename, &data.Content, &data.CreatedAt)

			if err != nil {
				continue
			}
			// var game string
			// str = fmt.Sprintf("select  gamename from %s where gameid = ? ", get_gm_game_table())
			// result := db_gm.QueryRow(str, data.GetGameid())

			// result.Scan(&game)
			// data.Gamename = &game

			rev.Data = append(rev.Data, data)

		}
	}

	*rev.Maxpage = count / perpage
	if count > 0 && *rev.Maxpage == 0 {
		*rev.Maxpage = 1
	} else if count%perpage != 0 {
		*rev.Maxpage += 1
	}
	// if rev.GetCurpage() > rev.GetMaxpage() {
	// 	logging.Debug("checkmanageractionrecord sql:%s ", str)
	// 	return false
	// }

	self.SendCmd(rev)
	return true
}
func add_manager_action_record(gameid uint32, gmid uint32, content string) bool {

	tblname := get_manager_action_record_table()
	str := fmt.Sprintf("insert into %s(gameid, gmid, content, created_at) values (?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, gameid, gmid, content, unitime.Time.Sec())
	if err == nil {
		_, err = result.LastInsertId()
	}
	return true
}

// 操作限制ip/机器码
func (self *GmTask) ParseModifyLimitIporcode_CS(rev *Pmd.ModifyLimitIporcode_CS) bool {
	if !self.VerifyOk {
		return false
	}

	typeid := rev.GetTypeid()
	var content string
	var canrecord = true
	if typeid == 1 {
		//add
		content = "添加限制ip/机器码，账号为" + rev.GetCode()
		tblname := get_limit_iporcode_table()
		str := fmt.Sprintf("insert into %s(gameid,zoneid, optype, limittype, code, content, starttime,endtime,created_at) values (?,?,?,?,?,?,?,?,?)", tblname)
		result, err := db_gm.Exec(str, rev.GetGameid(), rev.GetZoneid(), rev.GetOptype(), rev.GetLimittype(), rev.GetCode(), rev.GetContent(), rev.GetStarttime(), rev.GetEndtime(), unitime.Time.Sec())
		if err == nil {
			_, err = result.LastInsertId()
		}
		if err != nil {
			logging.Error("delete zonemsg error:%s,", err)
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("添加失败！")
			canrecord = false
		}

	} else {
		//delete
		content = "删除限制ip/机器码，账号为" + rev.GetCode()
		tblname := get_limit_iporcode_table()
		str := fmt.Sprintf("delete from %s where id=?", tblname)
		result, err := db_gm.Exec(str, rev.GetId())
		if err == nil {
			_, err = result.LastInsertId()
		}
		if err != nil {
			logging.Error("delete zonemsg error:%s,", err)
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("删除失败！")
			canrecord = false

		}
	}

	if canrecord {

		add_manager_action_record(rev.GetGameid(), rev.GetGmid(), content)
	}
	self.SendCmd(rev)

	return true
}

func (self *GmTask) ParseGetLimitIporcodeList_CS(rev *Pmd.GetLimitIporcodeList_CS) bool {

	tblname := get_limit_iporcode_table()

	where := ""

	if rev.GetOptype() > 0 {
		if rev.GetOptype() == 1 {
			where += " and optype = 1"
		} else {
			where += " and optype = 2"
		}

	}
	if rev.GetCode() != "" {
		where += " and code = " + rev.GetCode()
	}
	if rev.GetLimittype() > 0 {
		if rev.GetLimittype() == 1 {
			where += " and limittype = 1"
		} else {
			where += " and limittype = 2"
		}
	}

	var perpage uint32
	perpage = 100
	var count uint32

	str := fmt.Sprintf("select id , gameid , zoneid , optype , limittype, code, content, FROM_UNIXTIME(starttime), FROM_UNIXTIME(endtime) from %s where gameid = ? and zoneid= ? and created_at between ? and ?  %s limit ? , ?", tblname, where)

	rows, err := db_gm.Query(str, rev.GetGameid(), rev.GetZoneid(), rev.GetStarttime(), rev.GetEndtime(), (rev.GetCurpage()-1)*perpage, perpage)
	str1 := fmt.Sprintf("select count(*) from %s where gameid = ? and zoneid = ?  and created_at between ? and ?  %s ", tblname, where)

	row := db_gm.QueryRow(str1, rev.GetGameid(), rev.GetZoneid(), rev.GetStarttime(), rev.GetEndtime())

	row.Scan(&count)

	*rev.Maxpage = count / perpage
	if count > 0 && *rev.Maxpage == 0 {
		*rev.Maxpage = 1
	} else if count%perpage != 0 {
		*rev.Maxpage += 1
	}

	if err != nil {
		self.SendCmd(rev)
		return false
	}
	defer rows.Close()
	for rows.Next() {

		data := &Pmd.LimitIporcode{}
		err = rows.Scan(&data.Id, &data.Gameid, &data.Zoneid, &data.Optype, &data.Limittype, &data.Code, &data.Content, &data.Starttime, &data.Endtime)

		if err != nil {
			continue
		}

		rev.Data = append(rev.Data, data)

	}
	self.SendCmd(rev)
	return true
}

// 获取列表
func (self *GmTask) ParseModifyGameretInfolist_CS(rev *Pmd.ModifyGameretInfolist_CS) bool {
	if !self.VerifyOk {
		return false
	}
	if rev.GetCurtype() == 1 { //1、查询gm_games表返回游戏名列表
		tblname := get_gm_game_table()

		str := fmt.Sprintf("select count(*) from %s", tblname)
		row := db_gm.QueryRow(str)
		var count uint32
		if err := row.Scan(&count); err != nil {
			logging.Error("checkFeedbackList error:%s", err.Error())
			return false
		}
		if rev.GetPerpage() == 0 {
			*rev.Perpage, *rev.Curpage = 15, 1
		}
		*rev.Maxpage = count / rev.GetPerpage()
		if count > 0 && *rev.Maxpage == 0 {
			*rev.Maxpage = 1
		} else if count%rev.GetPerpage() != 0 {
			*rev.Maxpage += 1
		}
		if rev.GetCurpage() > rev.GetMaxpage() {
			logging.Debug("checkFeedbackList sql:%s ", str)
			return false
		}
		str1 := fmt.Sprintf("select gameid,gamename,gamelink,gmlink,gamekey,gametype,status from %s order by id asc limit %d, %d", tblname, (rev.GetCurpage()-1)*rev.GetPerpage(), rev.GetPerpage())

		rows, err := db_gm.Query(str1)
		if err != nil {
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.GameListInfo, 0)
		for rows.Next() {
			data := &Pmd.GameListInfo{}
			if err := rows.Scan(&data.Gameid, &data.Gamename, &data.Gamelink, &data.Gmlink, &data.Gamekey, &data.Gametype, &data.Status); err != nil {
				continue
			}
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(user_list) == 0")
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("查询失败！")
			return false
		}
		rows.Close()
	} else if rev.GetCurtype() == 2 { //2、修改游戏配置
		var retcode uint32 = 0
		retdesc := ""
		if rev.GetModifygamename() != "" {
			err := update_game_name(rev.GetGameid(), rev.GetModifygamename())
			if err != nil {
				retcode = 1
				retdesc += "update gamename error"
			}
		}
		if rev.GetModifygamelink() != "" {
			err := update_game_gamelink(rev.GetGameid(), rev.GetModifygamelink())
			if err != nil {
				retcode = 1
				retdesc += "update gamelink error"
			}
		}
		if rev.GetModifygmlink() != "" {
			err := update_game_gmlink(rev.GetGameid(), rev.GetModifygmlink())
			if err != nil {
				retcode = 1
				retdesc += "update gmlink error"
			}
		}
		if rev.GetModifygametype() != 0 {
			err := update_game_gametype(rev.GetGameid(), rev.GetModifygametype())
			if err != nil {
				retcode = 1
				retdesc += "update gametype error"
			}
		}
		if rev.GetModifygamekey() != "" {
			err := update_game_gamekey(rev.GetGameid(), rev.GetModifygamekey())
			if err != nil {
				retcode = 1
				retdesc += "update gamekey error"
			}
		}
		if rev.GetModifystatus() != 2 { //如若不修改status 赋值为2
			err := update_game_status(rev.GetGameid(), rev.GetModifystatus())
			if err != nil {
				retcode = 1
				retdesc += "update status error"
			}
		}
		rev.Retcode = proto.Uint32(retcode)
		rev.Retdesc = proto.String(retdesc)

		if retcode == 0 {
			content := fmt.Sprintf("修改游戏%d", rev.GetGameid())
			add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
		}

	} else if rev.GetCurtype() == 3 { //通过gameid查询游戏数据
		tblname := get_gm_game_table()
		str := fmt.Sprintf("select gameid,gamename,gamelink,gmlink,gamekey,gametype,status from %s where gameid=?", tblname)
		rows, err := db_gm.Query(str, rev.GetGameid())
		if err != nil {
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.GameListInfo, 0)
		for rows.Next() {
			data := &Pmd.GameListInfo{}
			if err := rows.Scan(&data.Gameid, &data.Gamename, &data.Gamelink, &data.Gmlink, &data.Gamekey, &data.Gametype, &data.Status); err != nil {
				continue
			}
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(game_list) == 0,rev.GetCurtype():%d", rev.GetCurtype())
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String("未查询到此数据！")
			self.SendCmd(rev)
			return false
		}
		rows.Close()
	}
	self.SendCmd(rev)
	return true
}
func get_gameids_by_accid(accid uint32) []uint32 {
	retl := make([]uint32, 0)
	if accid == 0 {
		return retl
	}
	tblname := get_gm_account_game_table()

	var str string
	//admin
	if accid == 1001 {
		str := fmt.Sprintf("select id from apps ")
		rows, err := unibase.DBZ.Query(str)
		if err != nil {
			return retl
		}
		for rows.Next() {
			var gameid uint32
			if err := rows.Scan(&gameid); err != nil {
				continue
			}
			retl = append(retl, gameid)
		}
		rows.Close()
	} else {
		str = fmt.Sprintf("select gameid from %s where accid=%d", tblname, accid)
		rows, err := db_gm.Query(str)
		if err != nil {
			return retl
		}
		for rows.Next() {
			var gameid uint32
			if err := rows.Scan(&gameid); err != nil {
				continue
			}
			retl = append(retl, gameid)
		}
		rows.Close()
	}

	return retl
}

func get_gamezone_by_accid(accid uint32) (retl []*Pmd.GameZoneInfo, retids []uint32, err error) {
	retl, retids = make([]*Pmd.GameZoneInfo, 0), make([]uint32, 0)
	gameids := get_gameids_by_accid(accid)
	if len(gameids) == 0 {
		return
	}
	tbl1 := get_gm_game_table()
	tbl2 := get_gm_zone_table()
	where := ""
	for _, gameid := range gameids {
		where += strconv.Itoa(int(gameid)) + ","
	}
	str := fmt.Sprintf("select a.gameid, a.gamename, b.zoneid, b.zonename from %s as a left join %s as b on a.gameid=b.gameid where a.gameid in (%s)", tbl1, tbl2, where[:len(where)-1])
	rows, err := db_gm.Query(str)
	if err != nil {
		return
	}
	for rows.Next() {
		gz := &Pmd.GameZoneInfo{}
		if err = rows.Scan(&gz.Gameid, &gz.Gamename, &gz.Zoneid, &gz.Zonename); err != nil {
			continue
		}
		retl = append(retl, gz)
	}
	rows.Close()
	for _, gid := range gameids {
		exists := false
		for _, gz := range retl {
			if gz.GetGameid() == gid {
				exists = true
				break
			}
		}
		if !exists {
			retids = append(retids, gid)
		}
	}
	return
}

func get_game_gmlink_and_key(gameid, zoneid uint32) (gamekey string, gmlinks []string) {
	tbl1 := get_gm_game_table()
	tbl2 := get_gm_zone_table()
	str := fmt.Sprintf("select zoneid, g.gmlink, gamekey, z.gmlink from %s as g left join %s as z on g.gameid=z.gameid where g.gameid=? ", tbl1, tbl2)
	rows, err := db_gm.Query(str, gameid)

	if err != nil {
		logging.Error("get gmlink error, gameid: %d", gameid)
		return
	}
	defer rows.Close()
	gmlinks = make([]string, 0)
	for rows.Next() {
		var zonelink, gmlink string
		var tzoneid uint32
		if err = rows.Scan(&tzoneid, &gmlink, &gamekey, &zonelink); err != nil {
			logging.Error("get gmlink error, gameid: %d", gameid)
			continue
		}
		if zoneid != 0 && tzoneid != zoneid {
			continue
		}
		if zonelink != "" {
			gmlinks = append(gmlinks, zonelink)
		} else if gmlink != "" {
			gmlinks = append(gmlinks, gmlink)
		}
	}
	return
}

func update_game_name(gameid uint32, name string) error {
	tblname := get_gm_game_table()
	str := fmt.Sprintf("update %s set gamename=? where gameid=?", tblname)
	result, err := db_gm.Exec(str, name, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func update_game_gmlink(gameid uint32, gmlink string) error {
	tblname := get_gm_game_table()
	str := fmt.Sprintf("update %s set gmlink=? where gameid=?", tblname)
	result, err := db_gm.Exec(str, gmlink, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func update_game_gamelink(gameid uint32, gamelink string) error {
	tblname := get_gm_game_table()
	str := fmt.Sprintf("update %s set gamelink=? where gameid=?", tblname)
	result, err := db_gm.Exec(str, gamelink, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func update_game_gametype(gameid, gametype uint32) error {
	tblname := get_gm_game_table()
	str := fmt.Sprintf("update %s set gametype=? where gameid=?", tblname)
	result, err := db_gm.Exec(str, gametype, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func update_game_status(gameid, status uint32) error {
	tblname := get_gm_game_table()
	str := fmt.Sprintf("update %s set status=? where gameid=?", tblname)
	result, err := db_gm.Exec(str, status, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func update_game_gamekey(gameid uint32, gamekey string) error {
	tblname := get_gm_game_table()
	str := fmt.Sprintf("update %s set gamekey=? where gameid=?", tblname)
	result, err := db_gm.Exec(str, gamekey, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func get_game_by_id(gameid uint32) *Game {
	m := &Game{}
	tblname := get_gm_game_table()
	str := fmt.Sprintf("select id, gameid, gamename, gamelink, gmlink, gametype, status, createtime from %s where gameid=?", tblname)
	row := db_gm.QueryRow(str, gameid)
	err := row.Scan(&m.Id, &m.GameId, &m.GameName, &m.GameLink, &m.GmLink, &m.GameType, &m.Status, &m.Createtime)
	if err != nil {
		return nil
	}
	return m
}

func add_game_zone(gameid, zoneid uint32, zonename, gmlink string) error {
	m := get_game_by_id(gameid)
	if m != nil && m.Id != 0 {
		if gmlink == "" {
			gmlink = m.GmLink
		}
		return create_zone(gameid, zoneid, zonename, gmlink)
	}
	return errors.New("gameid error")
}

// 解析添加新游戏  这是最高权限GM用户的功能
func (self *GmTask) ParseAddNewGameGmUserPmd_CS(rev *Pmd.AddNewGameGmUserPmd_CS) bool {
	if !self.VerifyOk {
		return false
	}
	retcode, retdesc := add_game(rev.GetCreategamename(), rev.GetGamelink(), rev.GetGmlink(), rev.GetGamekey(), rev.GetCreatezonename(), rev.GetGametype(), rev.GetCreatezoneid(), rev.GetCreategameid()) //添加游戏
	send := rev
	send.Retcode = proto.Uint32(retcode)
	send.Retdesc = proto.String(retdesc)
	if retcode == 0 {
		content := "添加新游戏" + rev.GetCreategamename()
		add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	}
	self.SendCmd(send)
	return true
}

// 解析删除游戏
func (self *GmTask) ParseRequestDelGamePmd_CS(rev *Pmd.RequestDelGamePmd_CS) bool {
	if !self.VerifyOk {
		return false
	}
	err := delete_game(rev.GetGameid(), true) //删除game表数据
	if err != nil {
		retcode := 1
		retdesc := fmt.Sprintf("delete game error:%s", err)
		send := rev
		send.Retcode = proto.Uint32(uint32(retcode))
		send.Retdesc = proto.String(retdesc)
		self.SendCmd(send)
		return false
	}
	//再删除游戏下的大区  用户
	retcode, retdesc := deleteAccount_by_gameid(rev.GetGameid())
	_, retdesc1 := deleteAccount_Games_by_gameid(rev.GetGameid())

	send := rev
	send.Retcode = proto.Uint32(retcode)
	send.Retdesc = proto.String(retdesc + retdesc1)
	if retcode == 0 {
		content := fmt.Sprintf("删除游戏%d", rev.GetGameid())
		add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	}
	self.SendCmd(send)
	return true
}

func delete_game(gameid uint32, force bool) error {
	str, tblname := "", get_gm_game_table()
	if force {
		str = fmt.Sprintf("delete from %s where gameid=?", tblname)
	} else {
		str = fmt.Sprintf("update %s set status=0 where gameid=?", tblname)
	}
	result, err := db_gm.Exec(str, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	if force && err == nil {
		delete_zone_by_gameid(gameid, force)
	}
	return err
}

func get_valid_game() []*Game {
	ml := make([]*Game, 0)
	tblname := get_gm_game_table()
	str := fmt.Sprintf("select id, gameid, gamename, gamelink, gmlink, gametype, status, createtime from %s where status=1", tblname)
	rows, err := db_gm.Query(str)
	if err != nil {
		logging.Error("get_valid_game error:%s", err.Error())
		return ml
	}
	defer rows.Close()
	for rows.Next() {
		m := &Game{}
		if err := rows.Scan(&m.Id, &m.GameId, &m.GameName, &m.GameLink, &m.GmLink, &m.GameType, &m.Status, &m.Createtime); err != nil {
			logging.Error("get_valid_game error:%s", err.Error())
			continue
		}
		ml = append(ml, m)
	}
	return ml
}

func get_all_Game() []*Game {
	ml := make([]*Game, 0)
	tblname := get_gm_game_table()
	str := fmt.Sprintf("select id, gameid, gamename, gamelink, gmlink, gametype, status, createtime from %s", tblname)
	rows, err := db_gm.Query(str)
	if err != nil {
		logging.Error("get_valid_game error:%s", err.Error())
		return ml
	}
	defer rows.Close()
	for rows.Next() {
		m := &Game{}
		if err := rows.Scan(&m.Id, &m.GameId, &m.GameName, &m.GameLink, &m.GmLink, &m.GameType, &m.Status, &m.Createtime); err != nil {
			logging.Error("get_valid_game error:%s", err.Error())
			continue
		}
		ml = append(ml, m)
	}
	return ml
}
