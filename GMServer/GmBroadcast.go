/*
*操作公告、广播相关的函数
 */

package main

import (
	"fmt"
	"math"

	"code.google.com/p/goprotobuf/proto"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

func saveBroadcast(data *Pmd.BroadcastInfo, gmid uint32) uint32 {
	tblname := "gm_broadcast"
	now := uint32(unitime.Time.Sec())
	str := fmt.Sprintf("insert into %s (gameid, zoneid, gmid, countryid, sceneid, starttime, nexttime, endtime, intervaltime, type, title, content, updated_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, data.GetGameid(), data.GetZoneid(), gmid, data.GetCountryid(), data.GetSceneid(), data.GetStarttime(), data.GetStarttime()+data.GetIntervaltime(), data.GetEndtime(), data.GetIntervaltime(), data.GetBtype(), data.GetTitle(), data.GetContent(), now)
	if err != nil {
		logging.Error("saveBroadcast error:%s, sql:%s", err.Error(), str)
		return 0
	}
	tmpid, _ := result.LastInsertId()
	id := uint32(tmpid)
	data.Taskid = proto.Uint32(id)
	data.Gmid = proto.Uint32(gmid)
	btype := data.GetBtype()
	if btype == 6 || btype == 4 || btype == 5 {
		str = fmt.Sprintf("update %s set state=3 where id!=%d and gameid=%d and zoneid=%d and state=0 and endtime>%d and type=%d", tblname, tmpid, data.GetGameid(), data.GetZoneid(), now, btype)
		logging.Debug("saveBroadcast sql:%s", str)
		db_gm.Exec(str)
	}
	return id
}

func saveShutdownBroadcast(gameid, zoneid, gmid, servertime, lefttime uint32, desc string) uint32 {
	tblname := "gm_broadcast"
	str := fmt.Sprintf("insert into %s (gameid, zoneid, gmid, endtime, type, content, updated_at) values (?,?,?,?,?,?,?)", tblname)
	result, err := db_gm.Exec(str, gameid, zoneid, gmid, servertime+lefttime, 3, desc, servertime)
	if err != nil {
		logging.Error("saveShutdownBroadcast error:%s, sql:%s", err.Error(), str)
		return 0
	}
	tmpid, _ := result.LastInsertId()
	return uint32(tmpid)
}

func checkNextBroadcast(gameid, zoneid uint32) (retl []*Pmd.BroadcastInfo) {
	tblname := "gm_broadcast"
	now := uint32(unitime.Time.Sec())
	where := fmt.Sprintf("gameid=%d and (zoneid=%d or zoneid=0) and state=0 and starttime<%d and %d<endtime and (nexttime between 0 and %d)", gameid, zoneid, now, now, now)
	str := fmt.Sprintf("select id, gameid, zoneid, gmid, countryid, sceneid, starttime, endtime, intervaltime, type,  content from %s where %s", tblname, where)
	retl = make([]*Pmd.BroadcastInfo, 0)
	rows, err := db_gm.Query(str)
	if err != nil {
		logging.Error("checkNextBroadcast error:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		data := &Pmd.BroadcastInfo{}
		if err := rows.Scan(&data.Taskid, &data.Gameid, &data.Zoneid, &data.Gmid, &data.Countryid, &data.Sceneid, &data.Starttime, &data.Endtime, &data.Intervaltime, &data.Btype,  &data.Content); err != nil {
			logging.Error("checkNextBroadcast error:%s", err.Error())
			continue
		}
		retl = append(retl, data)
	}
	str = fmt.Sprintf("update %s set nexttime=nexttime+intervaltime where %s", tblname, where)
	_, err = db_gm.Exec(str)
	if err != nil {
		logging.Error("checkNextBroadcast update nexttime error:%s", err.Error())
	}
	return
}

func updateBroadcastState(taskid, state uint32) {
	tblname := "gm_broadcast"
	now := uint32(unitime.Time.Sec())
	str := fmt.Sprintf("update %s set state=?,updated_at=? where id=?", tblname)
	_, err := db_gm.Exec(str, state, now, taskid)
	if err != nil {
		logging.Error("updateBroadcastState error:%s", err.Error())
	}
}

func checkBroadcastList(gameid, zoneid, countryid, sceneid, btype, endtime, curpage, perpage uint32) (maxpage uint32, retl []*Pmd.BroadcastInfo) {
	tblname := "gm_broadcast"
	retl = make([]*Pmd.BroadcastInfo, 0)
	where := fmt.Sprintf(" gameid=%d AND (zoneid=%d or zoneid=0) AND endtime>%d ", gameid, zoneid, endtime)
	if btype != 0 {
		where += fmt.Sprintf(" AND type=%d ", btype)
	}
	if countryid != 0 {
		where += fmt.Sprintf(" AND countryid=%d ", countryid)
	}
	if sceneid != 0 {
		where += fmt.Sprintf(" AND sceneid=%d ", sceneid)
	}
	where += " AND state = 0 "
	//查询总页数
	str := fmt.Sprintf("select count(*) from %s where %s", tblname, where)
	var count uint32
	row := db_gm.QueryRow(str)
	if err := row.Scan(&count); err != nil {
		logging.Error("checkBroadcastiList error:%s", err.Error())
		return
	}
	if curpage == 0 {
		curpage = 1
	}
	if perpage == 0 {
		perpage = 200
	}
	maxpage = uint32(math.Ceil(float64(count) / float64(perpage)))
	if curpage > maxpage {
		return
	}
	str = fmt.Sprintf("select id, gameid, zoneid, gmid, countryid, sceneid, starttime, endtime, intervaltime, type, title, content, updated_at from %s where %s order by id desc limit %d, %d", tblname, where, (curpage-1)*perpage, perpage)
	rows, err := db_gm.Query(str)
	if err != nil {
		logging.Error("checkBroadcastiList error:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		data := &Pmd.BroadcastInfo{}
		if err := rows.Scan(&data.Taskid, &data.Gameid, &data.Zoneid, &data.Gmid, &data.Countryid, &data.Sceneid, &data.Starttime, &data.Endtime, &data.Intervaltime, &data.Btype, &data.Title, &data.Content, &data.Recordtime); err != nil {
			logging.Error("checkBroadcastiList error:%s", err.Error())
			continue
		}
		retl = append(retl, data)
	}
	return
}
