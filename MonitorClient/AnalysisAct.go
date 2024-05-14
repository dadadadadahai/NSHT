package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

// 玩法分析
func HandleUserActAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("starttime")), 10, 64)
	etime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("endtime")), 10, 64)
	actionid := uint32(unibase.Atoi(task.R.FormValue("actionid"), 0))
	actiontype := uint32(unibase.Atoi(task.R.FormValue("actiontype"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	if actionid != 0 {
		where += fmt.Sprintf(" AND actionid=%d ", actionid)
	}
	if zoneid != 0 {
		where += fmt.Sprintf(" AND a.zoneid=%d ", zoneid)
	}
	if plats != "" {
		platlist := strings.Split(plats, ",")
		if len(platlist) == 1 {
			where += fmt.Sprintf(" AND platid=%s ", platlist[0])
		} else {

			//如果不是1，需要根据权限列获得相应的渠道id
			if len(platlist) != 1 {
				ptlist := getAccountPlatList(task)
				if len(ptlist) > 0 {
					platlist = ptlist
				}
			}
			where += fmt.Sprintf(" AND platid in (%s) ", strings.Join(platlist, ","))
		}
	} else {
		//如果为空，也要检测下是否有权限设置 ,没有设置就是最大权限
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			where += fmt.Sprintf(" AND platid in (%s) ", strings.Join(ptlist, ","))
		}
	}
	ymd := unitime.Time.YearMonthDay()
	etime += 86400
	edate := UnixToYmd(etime)
	if ymd-edate > 10000 {
		edate = ymd
		etime = uint64(unitime.Time.Sec())
	}
	if etime-stime > (300 * 24 * 3600) {
		stime = etime - 15*24*3600
	}
	where = fmt.Sprintf(" AND starttime>=%d AND endtime<%d ", stime, etime) + where
	var acttable string
	switch actiontype {
	case 1:
		acttable = "user_checkpoint_"
	case 2:
		acttable = "user_activity_"
	case 3:
		acttable = "user_task_"
	case 4:
		acttable = "user_battle_"
	default:
		return
	}
	retm := make(map[string]map[string]interface{})
	tbmap := make(map[string]uint32)
	for i, j := time.Unix(int64(stime), 0), 0; TimeToYmd(&i, j) < edate; j++ {
		tbname := fmt.Sprintf("%s%d_%d", acttable, gameid, TimeToYmd(&i, j)/100)
		if tbmap[tbname] > 0 || !check_table_exists(tbname) {
			continue
		}
		//新增完成最快时间、完成平均时间、完成最低战力、完成平均战力、失败平均时间、失败平均能力,完成平均vip
		tbmap[tbname] = 1
		str := fmt.Sprintf(`select daynum, actionid, min(duration)/60, avg(duration)/60, min(a.power), avg(a.power), avg(a.viplevel),
			count(*) from %s as a, %s as b where state=1 and a.userid=b.userid %s group by daynum, actionid`, tbname, tbuser, where)
		rows, err := db_monitor.Query(str)
		if err != nil {
			task.Error("HandleUserActAnalysis error:%s, sql:%s", err.Error(), str)
			continue
		}
		for rows.Next() {
			var daynum, actionid, minpower, passnum uint32
			var mintime, avgtime, avgvip, avgpower float32
			if err = rows.Scan(&daynum, &actionid, &mintime, &avgtime, &minpower, &avgpower, &avgvip, &passnum); err != nil {
				task.Error("HandleUserActAnalysis error:%s", err.Error())
				continue
			}
			data := map[string]interface{}{
				"daynum":       daynum,
				"actionid":     actionid,
				"actionname":   "",
				"acttypename":  "",
				"usernum":      0,
				"actionnum":    passnum,
				"passnum":      passnum,
				"lossnum":      0,
				"avgnum":       "1.00",
				"duration":     fmt.Sprintf("%.2f", avgtime),
				"percent":      "100.00",
				"fastmin":      fmt.Sprintf("%.2f", mintime),
				"minpower":     minpower,
				"avgpower":     fmt.Sprintf("%.2f", avgpower),
				"lossavgmin":   "0.00",
				"lossavgpower": "0.00",
				"viplevel":     fmt.Sprintf("%.2f", avgvip),
			}
			retm[fmt.Sprintf("%d%d", daynum, actionid)] = data
		}
		rows.Close()

		str = fmt.Sprintf(`select daynum, actionid, avg(duration)/60, avg(a.power), count(*) from %s as a,
			%s as b where state>=2 and a.userid=b.userid %s group by daynum, actionid`, tbname, tbuser, where)
		rows, err = db_monitor.Query(str)
		if err != nil {
			task.Error("HandleUserActAnalysis error:%s, sql:%s", err.Error(), str)
			continue
		}
		for rows.Next() {
			var daynum, actionid, lossnum uint32
			var avgtime, avgpower float32
			if err = rows.Scan(&daynum, &actionid, &avgtime, &avgpower, &lossnum); err != nil {
				task.Error("HandleUserActAnalysis error:%s", err.Error())
				continue
			}
			k := fmt.Sprintf("%d%d", daynum, actionid)
			if _, ok := retm[k]; !ok {
				retm[k] = map[string]interface{}{
					"daynum":       daynum,
					"actionid":     actionid,
					"actionname":   "",
					"acttypename":  "",
					"usernum":      0,
					"actionnum":    lossnum,
					"passnum":      0,
					"lossnum":      lossnum,
					"avgnum":       1,
					"duration":     0,
					"percent":      "100.00",
					"fastmin":      0,
					"minpower":     0,
					"avgpower":     0,
					"lossavgmin":   fmt.Sprintf("%.2f", avgtime),
					"lossavgpower": fmt.Sprintf("%.2f", avgpower),
					"viplevel":     0,
				}
			} else {
				retm[k]["lossnum"] = lossnum
				retm[k]["lossavgmin"] = fmt.Sprintf("%.2f", avgtime)
				retm[k]["lossavgpower"] = fmt.Sprintf("%.2f", avgpower)
				retm[k]["actionnum"] = lossnum + retm[k]["passnum"].(uint32)
				retm[k]["percent"] = fmt.Sprintf("%.2f", 100*float32(retm[k]["passnum"].(uint32))/float32(retm[k]["actionnum"].(uint32)))
			}
		}
		rows.Close()

		str = fmt.Sprintf(`select daynum, actionid, actionname, acttypename, count(distinct a.userid) from %s as a,
			%s as b where a.userid=b.userid %s group by daynum, actionid`, tbname, tbuser, where)
		rows, err = db_monitor.Query(str)
		if err != nil {
			task.Error("HandleUserActAnalysis error:%s, sql:%s", err.Error(), str)
			continue
		}
		for rows.Next() {
			var daynum, actionid, usernum uint32
			var actionname, acttypename string
			if err = rows.Scan(&daynum, &actionid, &actionname, &acttypename, &usernum); err != nil {
				task.Error("HandleUserActAnalysis error:%s", err.Error())
				continue
			}
			k := fmt.Sprintf("%d%d", daynum, actionid)
			retm[k]["actionname"] = actionname
			retm[k]["acttypename"] = acttypename
			retm[k]["usernum"] = usernum
			retm[k]["avgnum"] = fmt.Sprintf("%.2f", float32(retm[k]["actionnum"].(uint32))/float32(usernum))
		}
		rows.Close()
	}
	retl := make([]map[string]interface{}, 0)
	for _, data := range retm {
		retl = append(retl, data)
	}
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retl, retm, tbmap = nil, nil, nil
	return
}

// 升级统计
func HandleUserLevelupAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	btype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	tblname := get_user_levelup_table(gameid)
	if zoneid != 0 {
		where += fmt.Sprintf(" %s.zoneid=%d AND ", tblname, zoneid)
	}
	if btype != 0 {
		where += fmt.Sprintf(" leveltype=%d AND ", btype)
	}
	if plats != "" {
		platlist := strings.Split(plats, ",")
		if len(platlist) == 1 {
			where += fmt.Sprintf(" platid=%s AND ", platlist[0])
		} else if len(platlist) > 1 {

			//如果不是1，需要根据权限列获得相应的渠道id
			if len(platlist) != 1 {
				ptlist := getAccountPlatList(task)
				if len(ptlist) > 0 {
					platlist = ptlist
				}
			}
			where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(platlist, ","))
		}
	} else {

		//如果为空，也要检测下是否有权限设置 ,没有设置就是最大权限
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(ptlist, ","))
		}
	}
	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)
	str := fmt.Sprintf(`select typename, newlevel, count(*), avg(leveltime), min(leveltime),
		max(leveltime) from %s,%s where %s.userid=%s.userid and %s group by leveltype, newlevel`,
		tblname, tbuser, tblname, tbuser, where)
	task.Debug("HandleLevelupAnalysis sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleLevelupAnalysis error:%s, sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var typename string
		var avgtime float32
		var level, usernum, mintime, maxtime uint32
		if err := rows.Scan(&typename, &level, &usernum, &avgtime, &mintime, &maxtime); err != nil {
			task.Error("HandleLevelupAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"typename": typename,
			"level":    level,
			"usernum":  usernum,
			"avgtime":  fmt.Sprintf("%.2f", avgtime/60),
			"mintime":  fmt.Sprintf("%.2f", float32(mintime)/float32(60)),
			"maxtime":  fmt.Sprintf("%.2f", float32(maxtime)/float32(60)),
		}
		retl = append(retl, data)
	}
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retl = nil
	return
}
