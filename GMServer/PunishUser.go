/*
*惩罚玩家相关的函数
 */

package main

import (
	"fmt"

	"code.google.com/p/goprotobuf/proto"

	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

func savePunish(data *Pmd.PunishUserInfo, gmid uint32) uint32 {
	tblname := "gm_punish"
	now := uint32(unitime.Time.Sec())
	var id uint32
	// if data.GetCharid() > 0 {
	str := fmt.Sprintf("insert into %s (id , gameid, zoneid, charid, gmid, type, reason, starttime, endtime,created_at,pointnum,multiple) values (?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, data.GetTaskid(), data.GetGameid(), data.GetZoneid(), data.GetCharid(), gmid, data.GetPtype(), data.GetReason(), data.GetStarttime(), data.GetEndtime(), now, data.GetPunishvalue(), data.GetMultiple())
	if err != nil {
		logging.Error("savePunish error:%s, sql:%s", err.Error(), str)
		return 0
	}
	tmpid, _ := result.LastInsertId()
	id = uint32(tmpid)
	data.Taskid = proto.Uint32(id)
	data.Gmid = proto.Uint32(gmid)
	// } else if data.GetPid() != "" {
	// 	str := fmt.Sprintf("insert into %s (gameid, zoneid, pid, gmid, type, reason, starttime, endtime, created_at) values (?,?,?,?,?,?,?,?,?)", tblname)
	// 	result, err := db_gm.Exec(str, data.GetGameid(), data.GetZoneid(), data.GetPid(), gmid, data.GetPtype(), data.GetReason(), data.GetStarttime(), data.GetEndtime(), now)
	// 	if err != nil {
	// 		logging.Error("savePunish error:%s, sql:%s", err.Error(), str)
	// 		return 0
	// 	}
	// 	tmpid, _ := result.LastInsertId()
	// 	id = uint32(tmpid)
	// 	// data.Taskid = proto.Uint32(id)
	// 	data.Gmid = proto.Uint32(gmid)
	// }

	return uint32(id)
}

func updatePunishState(taskid, state uint32) {
	tblname := "gm_punish"
	now := uint32(unitime.Time.Sec())
	str := fmt.Sprintf("update %s set state=?,updated_at=? where id=?", tblname)
	_, err := db_gm.Exec(str, state, now, taskid)
	if err != nil {
		logging.Error("updatePunishState error:%s", err.Error())
	}
}
func updatePunishStateByUserid(charid, state uint32) {
	tblname := "gm_punish"
	now := uint32(unitime.Time.Sec())
	str := fmt.Sprintf("update %s set state=?,updated_at=? where charid=?", tblname)
	_, err := db_gm.Exec(str, state, now, charid)
	if err != nil {
		logging.Error("updatePunishState error:%s", err.Error())
	}
}
func deletePunish(taskid uint32) {
	tblname := "gm_punish"
	str := fmt.Sprintf("delete %s where id=?", tblname)
	_, err := db_gm.Exec(str, taskid)
	if err != nil {
		logging.Error("updatePunishState error:%s", err.Error())
	}
}
func getMaxid() uint32 {
	tblname := "gm_punish"
	str := fmt.Sprintf("select IFNULL(Max(id),0) from %s", tblname)
	row := db_gm.QueryRow(str)

	var count uint32
	row.Scan(&count)
	return count + 1
}

func checkPunishList(gameid, zoneid uint32, charid uint64, pid string, curpage, perpage, ptype, state uint32, starttime, endtime uint64, punishvalue string) (maxpage uint32, retl []*Pmd.PunishUserInfo) {
	tblname := "gm_punish"
	// if charid != 0 {
	now := uint32(unitime.Time.Sec())
	retl = make([]*Pmd.PunishUserInfo, 0)
	where := fmt.Sprintf(" (created_at>%d and created_at<%d) and state!=3  ", starttime, endtime)
	if charid > 0 {
		where += fmt.Sprintf(" and charid=%d", charid)
	}
	if ptype > 0 {
		where += fmt.Sprintf(" and type=%d", ptype)
	}
	if ptype > 6 {
		where += fmt.Sprintf(" and type < 6")
	}
	if state == 1 {
		where += fmt.Sprintf(" and endtime<%d", now)
	} else if state == 2 {
		where += fmt.Sprintf(" and endtime>%d", now)
	}
	if punishvalue != "" || len(punishvalue) > 0 {
		where += fmt.Sprintf(" and pointnum = %s", punishvalue)
	}
	//查找总数
	str := fmt.Sprintf("select count(*) from %s where %s", tblname, where)

	row := db_gm.QueryRow(str)
	var count uint32
	if err := row.Scan(&count); err != nil {
		logging.Error("checkPunishList error:%s", err.Error())
		return
	}
	maxpage = count / perpage
	if count > 0 && maxpage == 0 {
		maxpage = 1
	}
	if curpage > maxpage {
		return
	}
	//查找详细数据
	str = fmt.Sprintf("select id, gameid, zoneid, gmid, charid, charname, ip, type, reason, starttime, endtime, created_at, (case when endtime > unix_timestamp(now()) or ( type = 4 AND (created_at + endtime) > unix_timestamp(now()) ) then 2 else 1 end), pointnum from %s where %s order by id desc limit ?, ?", tblname, where)
	rows, err := db_gm.Query(str, (curpage-1)*perpage, perpage)
	if err != nil {
		logging.Error("checkBroadcastiList error:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		data := &Pmd.PunishUserInfo{}
		if err := rows.Scan(&data.Taskid, &data.Gameid, &data.Zoneid, &data.Gmid, &data.Charid, &data.Charname, &data.Ip, &data.Ptype, &data.Reason, &data.Starttime, &data.Endtime, &data.Punishtime, &data.State, &data.Punishvalue); err != nil {
			logging.Error("checkBroadcastiList error:%s", err.Error())
			continue
		}
		retl = append(retl, data)
	}
	// }
	// if pid != "" {
	// 	retl = make([]*Pmd.PunishUserInfo, 0)
	// 	where := fmt.Sprintf("gameid=%d and zoneid=%d and state!=3", gameid, zoneid)
	// 	where += fmt.Sprintf(" and pid='%s'", pid)
	// 	if ptype > 0 {
	// 		where += fmt.Sprintf(" and type=%d", ptype)
	// 	}
	// 	//查找总数
	// 	str := fmt.Sprintf("select count(*) from %s where %s", tblname, where)
	// 	row := db_gm.QueryRow(str)
	// 	var count uint32
	// 	if err := row.Scan(&count); err != nil {
	// 		logging.Error("checkPunishList error:%s", err.Error())
	// 		return
	// 	}
	// 	maxpage = count / perpage
	// 	if count > 0 && maxpage == 0 {
	// 		maxpage = 1
	// 	}
	// 	if curpage > maxpage {
	// 		return
	// 	}
	// 	//查找详细数据
	// 	str = fmt.Sprintf("select id, gameid, zoneid, gmid, charid, pid, charname, ip, type, reason, starttime, endtime, created_at,state from %s where %s order by id desc limit ?, ?", tblname, where)
	// 	rows, err := db_gm.Query(str, (curpage-1)*perpage, perpage)
	// 	if err != nil {
	// 		logging.Error("checkBroadcastiList1 error:%s", err.Error())
	// 		return
	// 	}
	// 	defer rows.Close()
	// 	for rows.Next() {
	// 		data := &Pmd.PunishUserInfo{}
	// 		if err := rows.Scan(&data.Taskid, &data.Gameid, &data.Zoneid, &data.Gmid, &data.Charid, &data.Pid, &data.Charname, &data.Ip, &data.Ptype, &data.Reason, &data.Starttime, &data.Endtime, &data.Punishtime, &data.State); err != nil {
	// 			logging.Error("checkBroadcastiList2 error:%s", err.Error())
	// 			continue
	// 		}
	// 		retl = append(retl, data)
	// 	}
	// }

	return
}
func checkrechargerewardlog(gameid, zoneid, curpage, perpage uint32) (maxpage uint32, retl []*Pmd.RechargeRewardLogGmUserPmd_CS_DataInfo) {
	tblname := get_exclusive_rewards_table()
	// if charid != 0 {

	retl = make([]*Pmd.RechargeRewardLogGmUserPmd_CS_DataInfo, 0)
	where := fmt.Sprintf(" gameid=%d and zoneid=%d ", gameid, zoneid)

	//查找总数
	str := fmt.Sprintf("select count(*) from %s where %s", tblname, where)

	row := db_gm.QueryRow(str)
	var count uint32
	if err := row.Scan(&count); err != nil {
		logging.Error("checkrechargerewardlog error:%s", err.Error())
		return
	}
	maxpage = count / perpage
	if count > 0 && maxpage == 0 {
		maxpage = 1
	}
	if curpage > maxpage {
		return
	}
	//查找详细数据
	str = fmt.Sprintf("select id, content,FROM_UNIXTIME(created_at) from %s where %s order by id desc limit ?, ?", tblname, where)
	rows, err := db_gm.Query(str, (curpage-1)*perpage, perpage)
	if err != nil {
		logging.Error("checkBroadcastiList error:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		data := &Pmd.RechargeRewardLogGmUserPmd_CS_DataInfo{}
		if err := rows.Scan(&data.Id, &data.Interval, &data.Time); err != nil {
			logging.Error("checkBroadcastiList error:%s", err.Error())
			continue
		}
		retl = append(retl, data)
	}
	return
}
