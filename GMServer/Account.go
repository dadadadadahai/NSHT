package main

import (
	"fmt"
	"net"
	"strings"

	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
)

func createAccount(data *Pmd.GmUserInfo) (uint32, string) {

	if len(data.GetName()) == 0 || len(data.GetPassword()) < 6 {
		return 1, "account or password error"
	}
	tblname := get_user_account_table()
	str := fmt.Sprintf("insert into %s (username, password, bindip, gameid, zoneid, priviliege, qmaxnum, autorecv, workstate, winnum, config) values(?,md5(?),?,?,?,?,?,?,?,?,?)", tblname)
	ret, err := db_gm.Exec(str, data.GetName(), data.GetPassword(), data.GetBindip(), data.GetGameid(), data.GetZoneid(), data.GetPri(), data.GetQmaxnum(), data.GetAutorecv(), data.GetWorkstate(), data.GetWinnum(), data.GetConfig())
	if err != nil {
		logging.Error("createAccount error:%s, str:%s", err.Error(), str)
		return 1, err.Error()
	}
	tblname1 := get_gm_account_game_table()
	ins_id, _ := ret.LastInsertId()
	str1 := fmt.Sprintf("insert into %s (accid,gameid) values(?,?)", tblname1)
	_, err1 := db_gm.Exec(str1, ins_id, data.GetGameid())
	if err1 != nil {
		logging.Error("addAccGame error:%s, str:%s", err.Error(), str)
		return 1, err.Error()
	}
	return 0, ""
}

func checkAccount(username, password string, userip uint32) (retcode uint32, data *Pmd.GmUserInfo) {
	data = &Pmd.GmUserInfo{}
	tblname := get_user_account_table()
	str := fmt.Sprintf("select id, username, bindip, gameid, zoneid, priviliege, qmaxnum, autorecv, workstate, winnum, config from %s where username=? and password=md5(?)", tblname)
	row := db_gm.QueryRow(str, username, password)
	err := row.Scan(&data.Gmid, &data.Name, &data.Bindip, &data.Gameid, &data.Zoneid, &data.Pri, &data.Qmaxnum, &data.Autorecv, &data.Workstate, &data.Winnum, &data.Config)
	if err != nil {
		logging.Error("checkAccount error:%s sql:%s, username:%s, password:%s", err.Error(), str, username, password)
		return 1, data
	}
	if data.GetBindip() != 0 && uint32(data.GetBindip()) != userip {
		return 1, data
	}
	return 0, data
}

func checkAccountList() []*Pmd.GmUserInfo {
	var accountlist = make([]*Pmd.GmUserInfo, 0)
	tblname := get_user_account_table()
	str := fmt.Sprintf("select id, username, bindip, gameid, zoneid, priviliege, qmaxnum, autorecv, workstate, winnum from %s", tblname)
	rows, err := db_gm.Query(str)
	if err != nil {
		logging.Error("query gm account list error:%s", err.Error())
		return accountlist
	}
	defer rows.Close()
	for rows.Next() {
		data := &Pmd.GmUserInfo{}
		if err := rows.Scan(&data.Gmid, &data.Name, &data.Bindip, &data.Gameid, &data.Zoneid, &data.Pri, &data.Qmaxnum, &data.Autorecv, &data.Workstate, &data.Winnum); err != nil {
			logging.Error("checkAccountList error:%s", err.Error())
			continue
		}
		accountlist = append(accountlist, data)
	}
	return accountlist
}

func changePassword(gmid uint32, password, newpassword string) (uint32, string) {
	tblname := get_user_account_table()
	str := fmt.Sprintf("update %s set password=md5(?) where id=? and password=md5(?)", tblname)
	result, err := db_gm.Exec(str, newpassword, gmid, password)
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		return 1, err.Error()
	}
	return 0, ""
}
func changePassword1(gmid uint32, newpassword string) (uint32, string) {
	tblname := get_user_account_table()
	str := fmt.Sprintf("update %s set password=md5(?) where id=?", tblname)
	result, err := db_gm.Exec(str, newpassword, gmid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		return 1, err.Error()
	}
	return 0, ""
}
func deleteAccount_by_gameid(gameid uint32) (uint32, string) {
	tblname := get_user_account_table()
	str := fmt.Sprintf("delete from %s where gameid=?", tblname)
	result, err := db_gm.Exec(str, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		return 1, err.Error()
	}
	return 0, ""
}

func deleteAccount(gmid uint32) (uint32, string) {
	tblname := get_user_account_table()
	str := fmt.Sprintf("delete from %s where id=?", tblname)
	result, err := db_gm.Exec(str, gmid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		return 1, err.Error()
	}
	return 0, ""
}
func deleteAccount_Games_by_accid(gmid uint32) (uint32, string) {
	tblname := get_gm_account_game_table()
	str := fmt.Sprintf("delete from %s where accid=?", tblname)
	result, err := db_gm.Exec(str, gmid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		return 1, err.Error()
	}
	return 0, ""
}
func deleteAccount_Games_by_gameid(gameid uint32) (uint32, string) {
	tblname := get_gm_account_game_table()
	str := fmt.Sprintf("delete from %s where gameid=?", tblname)
	result, err := db_gm.Exec(str, gameid)
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		return 1, err.Error()
	}
	return 0, ""
}

func modifyPri(id uint64, pri uint64) error {
	tblname := get_user_account_table()
	str := fmt.Sprintf("update %s set priviliege=? where id=?", tblname)
	result, err := db_gm.Exec(str, pri, id)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}
func update_account_pasword(id uint64, password string) error {
	tblname := get_user_account_table()
	str := fmt.Sprintf("update %s set password=md5(?) where id=?", tblname)
	result, err := db_gm.Exec(str, password, id)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}
func update_account_bindip(id uint64, bindip uint32) error {
	tblname := get_user_account_table()
	str := fmt.Sprintf("update %s set bindip=? where id=?", tblname)
	result, err := db_gm.Exec(str, bindip, id)
	if err == nil {
		_, err = result.RowsAffected()
	}
	return err
}

func (self *GmTask) ParseSetPasswordGmUserPmd_CS(rev *Pmd.SetPasswordGmUserPmd_CS) bool {
	if !self.VerifyOk {
		return false
	}
	//self.Info("gm:%s ParseSetPasswordGmUserPmd_CS data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	//logging.Error("id= %d",uint32(self.GetId()))
	//retcode, retdesc := changePassword(uint32(self.GetId()), rev.GetOldpassword(), rev.GetNewpassword())
	retcode, retdesc := changePassword1(uint32(self.GetId()), rev.GetNewpassword())
	rev.Retcode = proto.Uint32(retcode)
	rev.Retdesc = proto.String(retdesc)
	if retcode == 0 {
		content := "修改密码"
		add_manager_action_record(uint32(self.GetId()), self.Data.GetGameid(), content)
	}
	self.SendCmd(rev)
	return true
}

// 解析添加GM账号
func (self *GmTask) ParseAddNewGmUserPmd_CS(rev *Pmd.AddNewGmUserPmd_CS) bool {
	if !self.VerifyOk {
		return false
	}
	retcode, retdesc := createAccount(rev.GetData())
	send := rev
	send.Data = nil
	send.Retcode = proto.Uint32(retcode)
	send.Retdesc = proto.String(retdesc)
	if retcode == 0 {
		content := "添加GM账号"
		add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	}
	self.SendCmd(send)
	return true
}

// 解析删除GM账号
func (self *GmTask) ParseRequestDelGmUserPmd_CS(rev *Pmd.RequestDelGmUserPmd_CS) bool {
	if !self.VerifyOk {
		return false
	}
	retcode, retdesc := deleteAccount(rev.GetGmid())
	_, retdesc1 := deleteAccount_Games_by_accid(rev.GetGmid())
	retdesc2 := retdesc + retdesc1
	send := rev
	send.Retcode = proto.Uint32(retcode)
	send.Retdesc = proto.String(retdesc2)
	if retcode == 0 {
		content := fmt.Sprintf("删除GM账号%d", rev.GetGmid())
		add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	}
	self.SendCmd(send)
	return true
}

func (self *GmTask) ParseGmAccountListGmUserPmd_CS(rev *Pmd.GmAccountListGmUserPmd_CS) bool {
	if !self.VerifyOk {
		return false
	}
	rev.Data = checkAccountList()
	self.SendCmd(rev)
	return true
}

func (self *GmTask) ParseRequestModifyPriGmUserPmd_C(rev *Pmd.RequestModifyPriGmUserPmd_CS) bool {
	if !self.VerifyOk {
		return false
	}
	//retcode, retdesc := modifyPri(rev.GetGmid(), rev.GetPri())
	//send := rev
	//send.Retcode = proto.Uint32(retcode)
	//send.Retdesc = proto.String(retdesc)
	//self.SendCmd(send)
	return true
}

// 解析
func (self *GmTask) ParseSetPriGameGmUserPmd_CS(rev *Pmd.SetPriGameGmUserPmd_CS) bool {
	if !self.VerifyOk {
		return false
	}
	tblname := get_user_account_table()
	var Retcode uint32
	var Retdesc string
	if rev.GetCurtype() == 1 { //1、默认查询所有用户
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
		//logging.Warning("curpage:%d ... perpage:%d ... maxpage:%d",rev.GetCurpage(),rev.GetPerpage(),rev.GetMaxpage())//
		if rev.GetCurpage() > rev.GetMaxpage() {
			logging.Debug("checkFeedbackList sql:%s ", str)
			return false
		}
		str1 := fmt.Sprintf("select id,username,bindip,gameid,zoneid,priviliege,qmaxnum,autorecv,workstate,winnum,config from %s order by id asc limit ?, ?", tblname)

		rows, err := db_gm.Query(str1, (rev.GetCurpage()-1)*rev.GetPerpage(), rev.GetPerpage())

		if err != nil {
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String(fmt.Sprintf("select userinfo error:%s", err.Error()))
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.GmAccountInfo, 0)
		for rows.Next() {
			data := &Pmd.GmAccountInfo{}
			var bindip uint32

			err := rows.Scan(&data.Id, &data.Username, &bindip, &data.Gameid, &data.Zoneid, &data.Priviliege, &data.Qmaxnum, &data.Autorecv, &data.Workstate, &data.Winnum, &data.Config)

			if err != nil {
				continue
			}
			var ip [4]byte
			ip[0] = byte(bindip & 0xFF)
			ip[1] = byte((bindip >> 8) & 0xFF)
			ip[2] = byte((bindip >> 16) & 0xFF)
			ip[3] = byte((bindip >> 24) & 0xFF)
			temp := new(string)
			*temp = net.IPv4(ip[3], ip[2], ip[1], ip[0]).String()
			data.Bindip = temp

			tblname1 := get_gm_account_game_table()

			str2 := fmt.Sprintf("select gameid from %s where accid = ? ", tblname1)

			rows1, _ := db_gm.Query(str2, data.GetId())
			defer rows1.Close()

			var gamename = ""
			for rows1.Next() {
				var gameid uint32
				rows1.Scan(&gameid)

				gamename += search_field_value("name", "apps", "id", gameid) //查找游戏名称  查找大区名称

				fmt.Println("gamename:", gamename)
				gamename += " , "

			}

			data.Gamename = proto.String(gamename)

			// data.Zonename = search_field_value("zonename", "gm_zones", "zoneid", data.GetZoneid())
			rev.Data = append(rev.Data, data)
		}
		if len(rev.Data) == 0 { //一般不会出现这种情况
			logging.Error("len(user_list) == 0")
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String(fmt.Sprintf("用户列表为空！"))
			self.SendCmd(rev)
			return false
		}
		rows.Close()
	} else if rev.GetCurtype() == 2 { //2、使用用户名模糊查找
		rev.Data = make([]*Pmd.GmAccountInfo, 0)
		str := fmt.Sprintf("select id,username,bindip,gameid,zoneid,priviliege,qmaxnum,autorecv,workstate,winnum,config from %s where username like ?", tblname)
		rows, err := db_gm.Query(str, "%"+rev.GetUsername()+"%")
		if err != nil {
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String(fmt.Sprintf("select userinfo error:%s", err.Error()))
			self.SendCmd(rev)
			return false
		}
		rev.Data = make([]*Pmd.GmAccountInfo, 0)
		defer rows.Close()
		for rows.Next() {
			data := &Pmd.GmAccountInfo{}
			var bindip uint32
			if err := rows.Scan(&data.Id, &data.Username, &bindip, &data.Gameid, &data.Zoneid, &data.Priviliege, &data.Qmaxnum, &data.Autorecv, &data.Workstate, &data.Winnum, &data.Config); err != nil {
				continue
			}
			var ip [4]byte
			ip[0] = byte(bindip & 0xFF)
			ip[1] = byte((bindip >> 8) & 0xFF)
			ip[2] = byte((bindip >> 16) & 0xFF)
			ip[3] = byte((bindip >> 24) & 0xFF)
			temp := new(string)
			*temp = net.IPv4(ip[3], ip[2], ip[1], ip[0]).String()
			data.Bindip = temp
			tblname1 := get_gm_account_game_table()

			str2 := fmt.Sprintf("select gameid from %s where accid = ? ", tblname1)

			rows1, _ := db_gm.Query(str2, data.GetId())
			defer rows1.Close()

			var gamename = ""
			for rows1.Next() {
				var gameid uint32
				rows1.Scan(&gameid)
				gamename += search_field_value("name", "apps", "id", gameid) //查找游戏名称  查找大区名称
				gamename += " , "

			}

			data.Gamename = proto.String(gamename)
			// data.Gamename = search_field_value("gamename", "gm_games", "gameid", data.GetGameid()) //查找游戏名称  查找大区名称
			// data.Zonename = search_field_value("zonename", "gm_zones", "zoneid", data.GetZoneid())
			rev.Data = append(rev.Data, data)
		}
	} else if rev.GetCurtype() == 3 { //通过id 查找返回用户信息
		str := fmt.Sprintf("select priviliege,bindip from %s where id=?", tblname)
		rev.Data = make([]*Pmd.GmAccountInfo, 0)
		data := &Pmd.GmAccountInfo{}
		var bindip uint32
		err := db_gm.QueryRow(str, rev.GetId()).Scan(&data.Priviliege, &bindip)
		var ip [4]byte
		ip[0] = byte(bindip & 0xFF)
		ip[1] = byte((bindip >> 8) & 0xFF)
		ip[2] = byte((bindip >> 16) & 0xFF)
		ip[3] = byte((bindip >> 24) & 0xFF)
		temp := new(string)

		*temp = net.IPv4(ip[3], ip[2], ip[1], ip[0]).String()
		data.Bindip = temp
		if err != nil { //未找到当前用户名信息 用1表示
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String(fmt.Sprintf("未查询到此用户！"))
			self.SendCmd(rev)
			return false
		}
		rev.Data = append(rev.Data, data)
	} else if rev.GetCurtype() == 4 { //修改用户信息
		Retdesc = "修改 "
		if rev.GetPassword() != "" {
			err := update_account_pasword(rev.GetId(), rev.GetPassword())
			if err != nil {
				rev.Retcode = proto.Uint32(1)
				rev.Retdesc = proto.String(fmt.Sprintf("修改密码失败！%s", err.Error()))
				self.SendCmd(rev)
				return false
			} else {
				Retdesc += "密码 "
			}
		}
		if rev.GetPbool() != "" { //修改权限
			err := modifyPri(rev.GetId(), rev.GetPriviliege())
			if err != nil {
				rev.Retcode = proto.Uint32(1)
				rev.Retdesc = proto.String(fmt.Sprintf("修改权限失败！%s", err.Error()))
				self.SendCmd(rev)
				return false
			} else {

				Retdesc += "权限 "
			}
		}
		if rev.GetBbool() != "" { //修改ip
			err := update_account_bindip(rev.GetId(), rev.GetBindip())
			if err != nil {
				rev.Retcode = proto.Uint32(1)
				rev.Retdesc = proto.String(fmt.Sprintf("修改IP失败！%s", err.Error()))
				self.SendCmd(rev)
				return false
			} else {
				Retdesc += "IP "
			}
		}
		fmt.Println("修改对应游戏:", rev.GetGamestr())
		if rev.GetGamestr() != "" { //修改对应游戏
			tblname1 := get_gm_account_game_table()
			str := fmt.Sprintf("delete from %s where accid = %d ", tblname1, rev.GetId())
			_, err := db_gm.Exec(str)
			if err != nil {
				rev.Retcode = proto.Uint32(1)
				rev.Retdesc = proto.String(fmt.Sprintf("设置权限平台失败！%s", err.Error()))
				self.SendCmd(rev)
				return false
			}
			var gameids = strings.Split(rev.GetGamestr(), ",")

			for _, v := range gameids {
				str2 := fmt.Sprintf("insert into %s (accid,gameid) values (?,?); ", tblname1)
				_, err := db_gm.Exec(str2, rev.GetId(), v)
				if err != nil {

				}
			}

		}
		content := fmt.Sprintf("修改管理员%d信息", rev.GetId())
		add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
		Retcode = 0
		Retdesc += "成功！"
		rev.Retcode = proto.Uint32(Retcode)
		rev.Retdesc = proto.String(Retdesc)
	}
	self.SendCmd(rev)
	return true
}
func search_field_value(field1, tablename, field2 string, value uint32) string {
	var data string
	str := fmt.Sprintf("select %s from %s where %s = %d ", field1, tablename, field2, value)
	err := unibase.DBZ.QueryRow(str).Scan(&data)

	if err != nil {
		fmt.Printf("scan failed, err:%vn", err)
		data = ""
		return data
	}

	return data
}
