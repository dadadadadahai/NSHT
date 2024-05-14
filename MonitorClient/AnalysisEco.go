package main

//经济消费分析
import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

// 每日货币
func HandleUserDailyCoin(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("starttime")), 10, 64)
	etime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("endtime")), 10, 64)

	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))

	var where string
	tbuser := get_user_data_table(gameid)

	if zoneid != 0 {
		where += fmt.Sprintf(" AND %s.zoneid=%d ", tbuser, zoneid)
	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" AND %s.launchid='%s' ", tbuser, gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf(" AND %s.ad_account='%s' ", tbuser, launchAccount)
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
	sdate := UnixToYmd(stime)
	coindata := make(map[uint32]map[uint32]map[string]interface{})
	for i, j := time.Unix(int64(etime), 0), 0; TimeToYmd(&i, j) > sdate; j-- {

		tbname := fmt.Sprintf("user_economic_%d_%d", gameid, TimeToYmd(&i, j))
		if !check_table_exists(tbname) {
			continue
		}
		str := fmt.Sprintf("select daynum, count(distinct %s.userid),sum(coincount), type from %s, %s where %s.userid=%s.userid %s group by type,daynum", tbname, tbname, tbuser, tbname, tbuser, where)
		rows, err := db_monitor.Query(str)
		task.Debug("sql:%s", str)
		if err != nil {
			task.Error("HandleUseriDailyCoin error:%s sql:%s", err.Error(), str)
			continue
		}
		for rows.Next() {
			var daynum, num, cointype uint32
			var coincount float32
			coinid := uint32(1)
			if err = rows.Scan(&daynum, &num, &coincount, &cointype); err != nil {
				task.Error("HandleUseriDailyCoin error:%s", err.Error())
				continue
			}
			var data map[string]interface{}
			var ok bool
			if _, ok = coindata[daynum]; !ok {
				coindata[daynum] = make(map[uint32]map[string]interface{})
			}
			if data, ok = coindata[daynum][coinid]; !ok {
				data = map[string]interface{}{
					"coinid":     0.00,
					"daynum":     daynum,
					"output":     0, //产出
					"outputuser": 0, //产出账号数
					"outputavg":  0.00,
					"input":      0, //消耗
					"inputuser":  0, //消耗账号数
					"inputavg":   0.00,
				}
				coindata[daynum][coinid] = data
			}

			if cointype == 1 {
				data["output"] = fmt.Sprintf("%.2f", coincount/100)
				data["outputuser"] = num
				data["outputavg"] = fmt.Sprintf("%.2f", coincount/100/float32(num))
			} else {
				data["input"] = fmt.Sprintf("%.2f", coincount/100)
				data["inputuser"] = num
				data["inputavg"] = fmt.Sprintf("%.2f", coincount/100/float32(num))
			}

		}
		rows.Close()
	}
	retl := make([]map[string]interface{}, 0)
	for _, datamap := range coindata {
		for _, data := range datamap {
			retl = append(retl, data)
		}
		datamap = nil
	}
	coindata = nil
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

// 产出、消耗分布
func HandleUserCoinDistribution(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("starttime")), 10, 64)
	etime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("endtime")), 10, 64)
	coinid := uint32(unibase.Atoi(task.R.FormValue("coinid"), 0))
	actionid := uint32(unibase.Atoi(task.R.FormValue("actionid"), 0))
	actiontype := uint32(unibase.Atoi(task.R.FormValue("actiontype"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	if coinid != 0 {
		where += fmt.Sprintf(" AND coinid=%d AND type=%d ", coinid, actiontype)
	} else {
		where += fmt.Sprintf(" AND type=%d ", actiontype)
	}

	if actionid != 0 {
		where += fmt.Sprintf(" AND actionid=%d ", actionid)
	}
	if zoneid != 0 {
		where += fmt.Sprintf(" AND %s.zoneid=%d ", tbuser, zoneid)
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
	edate := UnixToYmd(etime)
	if ymd-edate > 10000 {
		edate = ymd
		etime = uint64(unitime.Time.Sec())
	}
	if etime-stime > (300 * 24 * 3600) {
		stime = etime - 15*24*3600
	}
	coindata := make(map[uint32]map[string]interface{})
	var total float32
	for i, j := time.Unix(int64(stime), 0), 0; TimeToYmd(&i, j) <= edate; j++ {
		tbname := fmt.Sprintf("user_economic_%d_%d", gameid, TimeToYmd(&i, j))
		if !check_table_exists(tbname) {
			continue
		}
		str := fmt.Sprintf("select actionid, actionname, count(distinct %s.userid), sum(actioncount), sum(coincount) from %s, %s where %s.userid=%s.userid %s group by actionid", tbname, tbname, tbuser, tbname, tbuser, where)
		rows, err := db_monitor.Query(str)
		task.Debug("sql:%s", str)
		if err != nil {
			task.Error("HandleUserCoinDistribution error:%s, sql:%s", err.Error(), str)
			continue
		}
		for rows.Next() {
			var actionid, num, actionnum uint32
			var coincount float32
			var actionname string
			if err = rows.Scan(&actionid, &actionname, &num, &actionnum, &coincount); err != nil {
				task.Error("HandleUserCoinDistribution error:%s", err.Error())
				continue
			}
			var data map[string]interface{}
			var ok bool
			if data, ok = coindata[actionid]; !ok {
				data = map[string]interface{}{
					"actionid":  actionid,
					"num":       uint32(0),
					"actionnum": uint32(0),
					"coincount": float32(0),
					"coinavg":   float32(0),
				}
				coindata[actionid] = data
				if actionname != "" {
					data["actionid"] = actionname
				}
			}
			data["num"] = data["num"].(uint32) + num
			data["actionnum"] = data["actionnum"].(uint32) + actionnum
			data["coincount"] = data["coincount"].(float32) + coincount
			total += coincount
		}
		rows.Close()
	}
	retl := make([]map[string]interface{}, 0)
	for _, data := range coindata {
		data["coinavg"] = fmt.Sprintf("%.2f", data["coincount"].(float32)/float32(data["num"].(uint32)))
		data["percent"] = fmt.Sprintf("%.2f", 100*data["coincount"].(float32)/total)
		data["coincount"] = fmt.Sprintf("%.2f", data["coincount"].(float32))
		retl = append(retl, data)
	}
	coindata = nil
	//task.Debug("HandleUserCoinDistribution retl:%v", retl)
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

// 等级产出、消耗分布
func HandleUserCoinLevelDistribution(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("starttime")), 10, 64)
	etime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("endtime")), 10, 64)
	coinid := uint32(unibase.Atoi(task.R.FormValue("coinid"), 0))
	actionid := uint32(unibase.Atoi(task.R.FormValue("actionid"), 0))
	actiontype := uint32(unibase.Atoi(task.R.FormValue("actiontype"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	if coinid != 0 {
		where += fmt.Sprintf(" AND coinid=%d AND type=%d ", coinid, actiontype)
	} else {
		where += fmt.Sprintf(" AND type=%d ", actiontype)
	}
	if actionid != 0 {
		where += fmt.Sprintf(" AND actionid=%d ", actionid)
	}
	if zoneid != 0 {
		where += fmt.Sprintf(" AND %s.zoneid=%d ", tbuser, zoneid)
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
	edate := UnixToYmd(etime)
	if ymd-edate > 10000 {
		edate = ymd
		etime = uint64(unitime.Time.Sec())
	}
	if etime-stime > (300 * 24 * 3600) {
		stime = etime - 15*24*3600
	}
	coindata := make(map[uint32]map[string]interface{})
	var total float32
	for i, j := time.Unix(int64(stime), 0), 0; TimeToYmd(&i, j) <= edate; j++ {
		tbname := fmt.Sprintf("user_economic_%d_%d", gameid, TimeToYmd(&i, j))
		if !check_table_exists(tbname) {
			continue
		}
		str := fmt.Sprintf("select actionid, actionname, cast((level/5) as unsigned) as userlevel, sum(coincount) from %s, %s where %s.userid=%s.userid %s group by actionid, userlevel", tbname, tbuser, tbname, tbuser, where)
		rows, err := db_monitor.Query(str)
		task.Debug("sql:%s", str)
		if err != nil {
			task.Error("HandleUserCoinDistribution error:%s, sql:%s", err.Error(), str)
			continue
		}
		for rows.Next() {
			var actionid, level uint32
			var coincount float32
			var actionname string
			if err = rows.Scan(&actionid, &actionname, &level, &coincount); err != nil {
				task.Error("HandleUserCoinDistribution error:%s", err.Error())
				continue
			}
			var data map[string]interface{}
			var ok bool
			if data, ok = coindata[actionid]; !ok {
				data = map[string]interface{}{
					"actionid":    actionid,
					"coincount":   float32(0),
					"coincount0":  float32(0),
					"coincount1":  float32(0),
					"coincount2":  float32(0),
					"coincount3":  float32(0),
					"coincount4":  float32(0),
					"coincount5":  float32(0),
					"coincount6":  float32(0),
					"coincount7":  float32(0),
					"coincount8":  float32(0),
					"coincount9":  float32(0),
					"coincount10": float32(0),
					"coincount11": float32(0),
					"coincount12": float32(0),
					"coincount13": float32(0),
					"coincount14": float32(0),
					"coincount15": float32(0),
					"coincount16": float32(0),
					"coincount17": float32(0),
					"coincount18": float32(0),
					"coincount19": float32(0),
					"coincount20": float32(0),
				}
				coindata[actionid] = data
				if actionname != "" {
					data["actionid"] = actionname
				}
			}
			data["coincount"] = data["coincount"].(float32) + coincount
			if level >= 20 {
				level = 20
			}
			dk := fmt.Sprintf("coincount%d", level)
			data[dk] = data[dk].(float32) + coincount
			total += coincount
		}
		rows.Close()
	}
	retl := make([]map[string]interface{}, 0)
	for _, data := range coindata {
		for i := 0; i < 21; i++ {
			dk := fmt.Sprintf("coincount%d", i)
			data[dk] = fmt.Sprintf("%.2f", data[dk])
		}
		data["coincount"] = fmt.Sprintf("%.2f", data["coincount"])
		retl = append(retl, data)
	}
	coindata = nil
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
