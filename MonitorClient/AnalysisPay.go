package main

//充值分析
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/unitime"
	"github.com/golang/protobuf/proto"
	"github.com/tidwall/gjson"
)

func get_realtime_data_table(gameid uint32, ymd uint32) string {
	return fmt.Sprintf("realtime_data_%d_%d", gameid, ymd)
}
func get_user_pay_table(gameid uint32) string {
	return fmt.Sprintf("user_pay_%d", gameid)
}
func get_adjapp_table(gameid uint32) string {
	return fmt.Sprintf("adj_app_%d", gameid)
}
func get_user_cashout_table(gameid uint32) string {
	return fmt.Sprintf("user_cashout_%d", gameid)
}
func get_money_change_table(gameid uint32) string {
	return fmt.Sprintf("money_change_%d", gameid)
}

func get_user_lottery_table(gameid uint32) string {
	return fmt.Sprintf("user_lottery_%d", gameid)
}

func get_user_levelup_table(gameid uint32) string {
	return fmt.Sprintf("user_levelup_%d", gameid)
}

func get_shop_transaction_table(gameid uint32) string {
	return fmt.Sprintf("shop_transaction_%d", gameid)
}
func get_change_registersrc_table(gameid uint32) string {
	return fmt.Sprintf("user_change_registersrc_%d", gameid)
}

func HandleUserPayAnalysis_new(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))

	var where string

	tbuser := get_user_data_table(gameid)
	tbmoney := get_money_change_table(gameid)

	if zoneid != 0 {
		where += fmt.Sprintf(" a.zoneid=%d AND ", zoneid)
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
	if gameSystem != "all" {
		where += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
	}

	if launchAccount != "" {
		where += fmt.Sprintf(" ad_account = '%s' AND ", launchAccount)
	}
	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d) and type=0 limit 500", stime, etime)

	str := fmt.Sprintf(`select platid,  daynum ,  a.userid , a.pay_num , a.pay_all , a.cash_num , a.cash_all from %s as a inner join %s as b on (a.userid=b.userid) where %s`, tbmoney, tbuser, where)

	rows, err := db_monitor.Query(str)

	if err != nil {
		task.Error("HandleUserPayAnalysis_new error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var userid uint64
		var platid, daynum, pay_num, pay_all, cash_num, cash_all uint32

		if err = rows.Scan(&platid, &daynum, &userid, &pay_num, &pay_all, &cash_num, &cash_all); err != nil {
			task.Error("HandleUserPayAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"Platid":   platid,
			"Daynum":   daynum,
			"Charid":   userid,
			"Pay_num":  pay_num,
			"Pay_all":  pay_all,
			"Cash_num": cash_num,
			"Cash_all": cash_all,
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

// 充值明细
func HandleUserPayAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	charid, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("charid")), 10, 64)
	var where string
	tbuser := get_user_data_table(gameid)
	tbpay := get_user_pay_table(gameid)
	if zoneid != 0 {
		where += fmt.Sprintf(" a.zoneid=%d AND ", zoneid)
	}
	if charid != 0 {
		where += fmt.Sprintf("a.userid=%d AND ", charid)
	} else if plats != "" {
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
	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d) and type=0 limit 500", stime, etime)
	str := fmt.Sprintf(`select platid, b.zoneid, a.money, a.created_at, platorder, gameorder, a.userid, username,
		curlevel, isfirst from %s as a inner join %s as b on (a.userid=b.userid and a.zoneid=b.zoneid) where %s`, tbpay, tbuser, where)
	task.Debug("HandleUserPayAnalysis sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserPayAnalysis error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var userid uint64
		var platid, zoneid, money, curlevel, isfirst uint32
		var platorder, gameorder, username, createtime string
		if err = rows.Scan(&platid, &zoneid, &money, &createtime, &platorder, &gameorder, &userid, &username, &curlevel, &isfirst); err != nil {
			task.Error("HandleUserPayAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"Platid":     platid,
			"Zoneid":     zoneid,
			"Createtime": createtime,
			"Platorder":  platorder,
			"Gameorder":  gameorder,
			"Money":      money,
			"Charid":     userid,
			"Charname":   username,
			"Userlevel":  curlevel,
			"Isfirst":    isfirst,
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

// 充值分布的金额描述
var PayLevelDescMap = map[uint64]string{
	1:  "0-6",
	2:  "6-18",
	3:  "18-30",
	4:  "30-100",
	5:  "100-200",
	6:  "200-500",
	7:  "500-1000",
	8:  "1000-2000",
	9:  "2000-5000",
	10: "5000-10000",
	11: "10000-30000",
	12: ">30000",
}

// 充值额度分布
func HandleUserPayDistribution(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	tblname := get_user_pay_table(gameid)
	if zoneid != 0 {
		where += fmt.Sprintf(" %s.zoneid=%d AND ", tblname, zoneid)
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
	where += fmt.Sprintf(" (daynum>=%d AND daynum<=%d) AND type=0 group by %s.userid", stime, etime, tblname)
	str := fmt.Sprintf(`select (case when amount<6 then 1 when amount<18 then 2 when amount<30 then 3 when
		 amount<100 then 4 when amount<200 then 5 when amount<500 then 6 when
		 amount<1000 then 7 when amount<2000 then 8 when amount<5000 then 9 when
		 amount<10000 then 10 when amount<30000 then 11 else 12 end) as level, count(*),
		 sum(amount), sum(times) from (select accid, sum(rolemoney) as amount, sum(roletimes) as times from
		 ( select %s.accid, cast(sum(%s.money)/100 as unsigned) as rolemoney, count(*) as roletimes from %s left
		 join %s on %s.userid=%s.userid where %s ) as tmpdata group by accid) as tmpdata2 group by level`,
		tbuser, tblname, tblname, tbuser, tblname, tbuser, where)
	rows, err := db_monitor.Query(str)
	task.Debug("HandleUserPayDistribution sql:%s", str)
	if err != nil {
		task.Error("HandleUserPayDistribution error:%s, sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	var totaluser, totalamount, totaltimes uint64
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var level, usernum, amount, paytimes uint64
		if err = rows.Scan(&level, &usernum, &amount, &paytimes); err != nil {
			task.Error("HandleUserPayDistribution error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"level":     level,
			"leveldesc": PayLevelDescMap[level],
			"usernum":   usernum,
			"amount":    amount,
			"paytimes":  paytimes,
		}
		retl = append(retl, data)
		totaluser += usernum
		totalamount += amount
		totaltimes += paytimes
	}
	for _, data := range retl {
		data["userpercent"] = fmt.Sprintf("%.2f", 100*float32(data["usernum"].(uint64))/float32(totaluser))
		data["amountpercent"] = fmt.Sprintf("%.2f", 100*float32(data["amount"].(uint64))/float32(totalamount))
		data["timespercent"] = fmt.Sprintf("%.2f", 100*float32(data["paytimes"].(uint64))/float32(totaltimes))
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

// 首冲等级-所选时间段内首冲玩家的首冲等级
func HandleUserPayLevel(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	tbpay := get_user_pay_table(gameid)
	if zoneid != 0 {
		where += fmt.Sprintf(" %s.zoneid=%d AND ", tbpay, zoneid)
	}
	if plats != "" {
		platlist := strings.Split(plats, ",")
		if len(platlist) == 1 {
			where += fmt.Sprintf(" platid=%s AND ", platlist[0])
		} else {
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

	where += fmt.Sprintf(" (daynum>=%d AND daynum<=%d) AND type=0 AND isfirst=1 ", stime, etime)
	str := fmt.Sprintf(`select curlevel, count(*) from (select min(curlevel) as curlevel, accid from %s left
	join %s on %s.userid=%s.userid where %s group by accid) as a group by curlevel`, tbpay, tbuser, tbpay, tbuser, where)

	task.Debug("HandleUserPayLevel sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserPayLevel error:%s, sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	var total uint32
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var level, num uint32
		if err = rows.Scan(&level, &num); err != nil {
			task.Error("HandleUserPayLevel error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"level": level,
			"num":   num,
		}
		total += num
		retl = append(retl, data)
	}
	for _, data := range retl {
		data["percent"] = fmt.Sprintf("%.2f", 100*float32(data["num"].(uint32))/float32(total))
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

// 充值排名
func HandleUserPayRank(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	tblname := get_user_pay_table(gameid)
	if zoneid != 0 {
		where += fmt.Sprintf(" a.zoneid=%d AND ", zoneid)
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
	where += fmt.Sprintf(" (daynum>=%d AND daynum<=%d) AND type=0 group by a.userid order by amount desc limit 200", stime, etime)
	str := fmt.Sprintf(`select platid, a.userid, account, username, userlevel, cast(sum(a.money)/100 as unsigned) as amount, count(*) as times,
		cast(max(a.money)/100 as unsigned), cast(min(a.money)/100 as unsigned), from_unixtime(last_login_time, "%%Y-%%m-%%d %%h:%%m:%%s"),
		max(created_at) from %s as a inner join %s as b on (a.userid=b.userid and a.zoneid=b.zoneid) where %s `, tblname, tbuser, where)
	task.Debug("HandleUserPayRank sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserPayRank error:%s, sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	var rank uint32 = 0
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var platid, charid, amount, times, maxamount, minamount, userlevel uint64
		var account, charname, lastpay, lastlogin string
		if err := rows.Scan(&platid, &charid, &account, &charname, &userlevel, &amount, &times, &maxamount, &minamount, &lastlogin, &lastpay); err != nil {
			task.Error("HandleUserPayRank error:%s", err.Error())
			continue
		}
		rank += 1
		data := map[string]interface{}{
			"rank":      rank,
			"platid":    platid,
			"charid":    charid,
			"account":   account,
			"charname":  charname,
			"charlevel": userlevel,
			"amount":    amount,
			"times":     times,
			"maxamount": maxamount,
			"minamount": minamount,
			"lastlogin": lastlogin,
			"lastpay":   lastpay,
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

// 积分兑换明细
func HandleUserPointDetail(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	charid, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("charid")), 10, 64)
	var where string
	tbuser := get_user_data_table(gameid)
	tbpay := get_user_pay_table(gameid)
	if zoneid != 0 {
		where += fmt.Sprintf(" %s.zoneid=%d AND ", tbuser, zoneid)
	}
	if charid != 0 {
		where += fmt.Sprintf("%s.userid=%d ", tbpay, charid)
	} else if plats != "" {
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
	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d) and (type=2 or type=3) ", stime, etime)
	str := fmt.Sprintf(`select platid, %s.zoneid, %s.created_at, platorder, gameorder, %s.userid, username,
		curlevel, (case type when 2 then %s.money else -%s.money end) from %s, %s where %s.userid=%s.userid and %s`,
		tbpay, tbpay, tbpay, tbpay, tbpay, tbpay, tbuser, tbpay, tbuser, where)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserPointDetail error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var userid, point int64
		var platid, zoneid, curlevel uint32
		var platorder, gameorder, username, createtime string
		if err = rows.Scan(&platid, &zoneid, &createtime, &platorder, &gameorder, &userid, &username, &curlevel, &point); err != nil {
			task.Error("HandleUserPointDetail error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"Platid":     platid,
			"Zoneid":     zoneid,
			"Createtime": createtime,
			"Platorder":  platorder,
			"Gameorder":  gameorder,
			"Charid":     userid,
			"Charname":   username,
			"Userlevel":  curlevel,
			"Point":      point,
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

// 抽奖分析
func HandleUserLotteryAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	btype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	tblname := get_user_lottery_table(gameid)
	if zoneid != 0 {
		where += fmt.Sprintf(" %s.zoneid=%d AND ", tblname, zoneid)
	}
	if btype != 0 {
		where += fmt.Sprintf(" type=%d AND ", btype)
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

	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d) and type=0 ", stime, etime)
	str := fmt.Sprintf(`select daynum, name, itemname, count(distinct %s.userid), sum(opcount),
		count(*) from %s, %s where %s.userid=%s.userid and %s group by daynum, type, itemtype, itemid`,
		tblname, tblname, tbuser, tblname, tbuser, where)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserLotteryAnalysis error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var daynum, usernum, amount, times uint32
		var typename, itemname string
		if err = rows.Scan(&daynum, &typename, &itemname, &usernum, &amount, &times); err != nil {
			task.Error("HandleUserLotteryAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"daynum":   daynum,
			"typename": typename,
			"itemname": itemname,
			"usernum":  usernum,
			"amount":   amount,
			"times":    times,
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

// 商店交易记录
func HandleShopTransactionAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	btype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	var where string
	tbuser := get_user_data_table(gameid)
	tblname := get_shop_transaction_table(gameid)
	if zoneid != 0 {
		where += fmt.Sprintf(" %s.zoneid=%d AND ", tblname, zoneid)
	}
	if btype != 0 {
		where += fmt.Sprintf(" type=%d AND ", btype)
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
	str := fmt.Sprintf(`select name, opname, itemname, count(distinct accid), count(*), sum(itemcount),
		sum(opcount), avg(opcount/itemcount) from %s, %s where %s.userid=%s.userid and %s group by type,
		itemtype, itemid`, tblname, tbuser, tblname, tbuser, where)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserTransactionAnalysis error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var avgprice float32
		var usernum, times, itemcount, amount uint32
		var typename, opname, itemname string
		if err = rows.Scan(&typename, &opname, &itemname, &usernum, &times, &itemcount, &amount, &avgprice); err != nil {
			task.Error("HandleUserTransactionAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"typename":  typename,
			"opname":    opname,
			"itemname":  itemname,
			"usernum":   usernum,
			"amount":    amount,
			"times":     times,
			"itemcount": itemcount,
			"avgprice":  fmt.Sprintf("%.2f", avgprice),
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

// LTV
func HandleUserLTVAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	tblname := get_user_daily_plat_table(gameid)
	tbldataname := get_user_data_table(gameid)
	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))

	where := fmt.Sprintf(" zoneid != 0 AND ")
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
	if gameSystem != "all" {
		where += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf(" ad_account = '%s' AND ", launchAccount)
	}

	where2 := where
	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)

	str := fmt.Sprintf(`select daynum, sum(day_1),sum(pay_today),sum(pay_cashout),sum(pay_1), sum(pay_2),sum(pay_3),sum(pay_4),sum(pay_5),sum(pay_6),sum(pay_7), sum(pay_8), sum(pay_9), sum(pay_10),sum(pay_11),sum(pay_12),sum(pay_13),sum(pay_14),sum(pay_21),sum(pay_30),sum(pay_45),sum(pay_60),sum(pay_90),sum(pay_180) from %s where %s group by daynum`, tblname, where)
	rows, err := db_monitor.Query(str)
	task.Debug("HandleUserLTVAnalysis sql:%s", str)
	if err != nil {
		task.Error("HandleUserLTVAnalysis error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var daynum, newuser, paytoday, cashouttoday, pay_1, pay_2, pay_3, pay_4, pay_5, pay_6, pay_7, pay_8, pay_9, pay_10, pay_11, pay_12, pay_13, pay_14, pay_21, pay_30, pay_45, pay_60, pay_90, pay_180 uint32
		if err = rows.Scan(&daynum, &newuser, &paytoday, &cashouttoday, &pay_1, &pay_2, &pay_3, &pay_4, &pay_5, &pay_6, &pay_7, &pay_8, &pay_9, &pay_10, &pay_11, &pay_12, &pay_13, &pay_14, &pay_21, &pay_30, &pay_45, &pay_60, &pay_90, &pay_180); err != nil {
			task.Error("HandleUserLTVAnalysis error:%s", err.Error())
			continue
		}
		var zero = newuser == 0
		if zero {
			newuser = 1
		}
		data := map[string]interface{}{
			"daynum":       daynum,
			"newuser":      newuser,
			"paytoday":     pay_1,
			"paynewall":    0.00,
			"cashouttoday": cashouttoday,
			"newpay":       fmt.Sprintf("%.2f", float32(pay_1/100)),
			"pay_1":        fmt.Sprintf("%.2f", float32(pay_1/100)/float32(newuser)),
			"pay_2":        fmt.Sprintf("%.2f", float32(pay_2/100)/float32(newuser)),
			"pay_3":        fmt.Sprintf("%.2f", float32(pay_3/100)/float32(newuser)),
			"pay_4":        fmt.Sprintf("%.2f", float32(pay_4/100)/float32(newuser)),
			"pay_5":        fmt.Sprintf("%.2f", float32(pay_5/100)/float32(newuser)),
			"pay_6":        fmt.Sprintf("%.2f", float32(pay_6/100)/float32(newuser)),
			"pay_7":        fmt.Sprintf("%.2f", float32(pay_7/100)/float32(newuser)),
			"pay_8":        fmt.Sprintf("%.2f", float32(pay_8/100)/float32(newuser)),
			"pay_9":        fmt.Sprintf("%.2f", float32(pay_9/100)/float32(newuser)),
			"pay_10":       fmt.Sprintf("%.2f", float32(pay_10/100)/float32(newuser)),
			"pay_11":       fmt.Sprintf("%.2f", float32(pay_11/100)/float32(newuser)),
			"pay_12":       fmt.Sprintf("%.2f", float32(pay_12/100)/float32(newuser)),
			"pay_13":       fmt.Sprintf("%.2f", float32(pay_13/100)/float32(newuser)),
			"pay_14":       fmt.Sprintf("%.2f", float32(pay_14/100)/float32(newuser)),
			"pay_21":       fmt.Sprintf("%.2f", float32(pay_21/100)/float32(newuser)),
			"pay_30":       fmt.Sprintf("%.2f", float32(pay_30/100)/float32(newuser)),
			"pay_45":       fmt.Sprintf("%.2f", float32(pay_45/100)/float32(newuser)),
			"pay_60":       fmt.Sprintf("%.2f", float32(pay_60/100)/float32(newuser)),
			"pay_90":       fmt.Sprintf("%.2f", float32(pay_90/100)/float32(newuser)),
			"pay_180":      fmt.Sprintf("%.2f", float32(pay_180/100)/float32(newuser)),
		}
		if zero {
			data["newuser"] = 0
		}
		//获取新增玩家累计充值

		strall := fmt.Sprintf(`select sum(pay_all) from %s where %s reg_time >= UNIX_TIMESTAMP(?) and  reg_time <= UNIX_TIMESTAMP(?) `, tbldataname, where2)

		timestamp1, timestamp2 := fmt.Sprintf("%d000000", daynum), fmt.Sprintf("%d235959", daynum)
		all := db_monitor.QueryRow(strall, timestamp1, timestamp2)

		var allnewpay uint32

		all.Scan(&allnewpay)

		data["paynewall"] = fmt.Sprintf("%.2f", float32(allnewpay/100))

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

//ltv去提现

func HandleUserLTVCashAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	tblname := get_user_daily_roi_table(gameid)
	tbldataname := get_user_data_table(gameid)
	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))
	tbltodaycost := get_app_cost_today_table(gameid)

	//获取汇率
	ratetable := get_exchange_rate_table(gameid)
	strrate := fmt.Sprintf(`select time , exchange_rate from %s where currency = 1`, ratetable)

	rowrate, err := db_monitor.Query(strrate)

	ratearr := make(map[string]float32)

	defer rowrate.Close()

	if err == nil {

		for rowrate.Next() {
			var time string
			var rate float32
			rowrate.Scan(&time, &rate)
			ratekey := strings.Replace(time, "-", "", -1)
			ratearr[string(ratekey)] = float32(rate)

		}

	}

	where, whereapp := fmt.Sprintf(" zoneid != 0 AND "), ""

	if plats != "" {
		platlist := strings.Split(plats, ",")
		if len(platlist) == 1 {
			where += fmt.Sprintf(" platid=%s AND ", platlist[0])

			whereapp += fmt.Sprintf(" appid=%s AND ", platlist[0])
		} else if len(platlist) > 1 {
			//如果不是1，需要根据权限列获得相应的渠道id
			if len(platlist) != 1 {
				ptlist := getAccountPlatList(task)
				if len(ptlist) > 0 {
					platlist = ptlist
				}
			}
			where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(platlist, ","))
			whereapp += fmt.Sprintf(" appid in (%s) AND ", strings.Join(platlist, ","))
		}
	} else {

		//如果为空，也要检测下是否有权限设置 ,没有设置就是最大权限
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(ptlist, ","))
			whereapp += fmt.Sprintf(" appid in (%s) AND ", strings.Join(ptlist, ","))
		}
	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
		whereapp += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf(" ad_account = '%s' AND ", launchAccount)
	}

	where2 := where

	// if stime < 20230312 {
	// 	stime = 20230312
	// }
	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)
	whereapp += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)
	costarr := make(map[uint32]float32)

	//获取消耗
	str1 := fmt.Sprintf(`select daynum,sum(cost) from %s where %s group by daynum`, tbltodaycost, whereapp)
	rowsapp, err := db_monitor.Query(str1)

	if err != nil {
		task.Error("HandleRecoveryData error:%s sql:%s", err.Error(), str1)

	} else {
		for rowsapp.Next() {
			var daynumapp uint32
			var costapp float32

			rowsapp.Scan(&daynumapp, &costapp)
			costarr[daynumapp] = costapp

		}
	}
	defer rowsapp.Close()

	str := fmt.Sprintf(`select daynum, sum(new_user),sum(pay_1), sum(pay_2),sum(pay_3),sum(pay_4),sum(pay_5),sum(pay_6),sum(pay_7), sum(pay_8), sum(pay_9), sum(pay_10),sum(pay_11),sum(pay_12),sum(pay_13),sum(pay_14),sum(pay_21),sum(pay_30),sum(pay_45),sum(pay_60),sum(pay_90),sum(pay_180),sum(cashout_1), sum(cashout_2),sum(cashout_3),sum(cashout_4),sum(cashout_5),sum(cashout_6),sum(cashout_7), sum(cashout_8), sum(cashout_9), sum(cashout_10),sum(cashout_11),sum(cashout_12),sum(cashout_13),sum(cashout_14),sum(cashout_21),sum(cashout_30),sum(cashout_45),sum(cashout_60),sum(cashout_90),sum(cashout_180),sum(cost) from %s where %s group by daynum`, tblname, where)
	rows, err := db_monitor.Query(str)
	task.Debug("HandleUserLTVAnalysis sql:%s", str)
	if err != nil {
		task.Error("HandleUserLTVAnalysis error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var daynum, newuser, pay_1, pay_2, pay_3, pay_4, pay_5, pay_6, pay_7, pay_8, pay_9, pay_10, pay_11, pay_12, pay_13, pay_14, pay_21, pay_30, pay_45, pay_60, pay_90, pay_180, cashout_1, cashout_2, cashout_3, cashout_4, cashout_5, cashout_6, cashout_7, cashout_8, cashout_9, cashout_10, cashout_11, cashout_12, cashout_13, cashout_14, cashout_21, cashout_30, cashout_45, cashout_60, cashout_90, cashout_180 uint32

		var cost float32
		if err = rows.Scan(&daynum, &newuser, &pay_1, &pay_2, &pay_3, &pay_4, &pay_5, &pay_6, &pay_7, &pay_8, &pay_9, &pay_10, &pay_11, &pay_12, &pay_13, &pay_14, &pay_21, &pay_30, &pay_45, &pay_60, &pay_90, &pay_180, &cashout_1, &cashout_2, &cashout_3, &cashout_4, &cashout_5, &cashout_6, &cashout_7, &cashout_8, &cashout_9, &cashout_10, &cashout_11, &cashout_12, &cashout_13, &cashout_14, &cashout_21, &cashout_30, &cashout_45, &cashout_60, &cashout_90, &cashout_180, &cost); err != nil {
			task.Error("HandleUserLTVAnalysis error:%s", err.Error())
			continue
		}
		var zero = newuser == 0
		if zero {
			newuser = 1
		}
		_, ok := costarr[daynum]
		if ok {
			cost = costarr[daynum]
		}
		data := map[string]interface{}{
			"daynum":        daynum,
			"newuser":       newuser,
			"cost":          cost,
			"rate":          0.00,
			"paytoday":      pay_1,
			"paynewall":     0.00,
			"cashoutnewall": 0.00,
			"cashouttoday":  fmt.Sprintf("%.2f", float32(cashout_1/100)),
			"newpay":        fmt.Sprintf("%.2f", float32(pay_1/100)),
			"pay_1":         fmt.Sprintf("%.2f", (float32(pay_1/100)-float32(cashout_1/100))/float32(newuser)),
			"pay_2":         fmt.Sprintf("%.2f", (float32(pay_2/100)-float32(cashout_2/100))/float32(newuser)),
			"pay_3":         fmt.Sprintf("%.2f", (float32(pay_3/100)-float32(cashout_3/100))/float32(newuser)),
			"pay_4":         fmt.Sprintf("%.2f", (float32(pay_4/100)-float32(cashout_4/100))/float32(newuser)),
			"pay_5":         fmt.Sprintf("%.2f", (float32(pay_5/100)-float32(cashout_5/100))/float32(newuser)),
			"pay_6":         fmt.Sprintf("%.2f", (float32(pay_6/100)-float32(cashout_6/100))/float32(newuser)),
			"pay_7":         fmt.Sprintf("%.2f", (float32(pay_7/100)-float32(cashout_7/100))/float32(newuser)),
			"pay_8":         fmt.Sprintf("%.2f", (float32(pay_8/100)-float32(cashout_8/100))/float32(newuser)),
			"pay_9":         fmt.Sprintf("%.2f", (float32(pay_9/100)-float32(cashout_9/100))/float32(newuser)),
			"pay_10":        fmt.Sprintf("%.2f", (float32(pay_10/100)-float32(cashout_10/100))/float32(newuser)),
			"pay_11":        fmt.Sprintf("%.2f", (float32(pay_11/100)-float32(cashout_11/100))/float32(newuser)),
			"pay_12":        fmt.Sprintf("%.2f", (float32(pay_12/100)-float32(cashout_12/100))/float32(newuser)),
			"pay_13":        fmt.Sprintf("%.2f", (float32(pay_13/100)-float32(cashout_13/100))/float32(newuser)),
			"pay_14":        fmt.Sprintf("%.2f", (float32(pay_14/100)-float32(cashout_14/100))/float32(newuser)),
			"pay_21":        fmt.Sprintf("%.2f", (float32(pay_21/100)-float32(cashout_21/100))/float32(newuser)),
			"pay_30":        fmt.Sprintf("%.2f", (float32(pay_30/100)-float32(cashout_30/100))/float32(newuser)),
			"pay_45":        fmt.Sprintf("%.2f", (float32(pay_45/100)-float32(cashout_45/100))/float32(newuser)),
			"pay_60":        fmt.Sprintf("%.2f", (float32(pay_60/100)-float32(cashout_60/100))/float32(newuser)),
			"pay_90":        fmt.Sprintf("%.2f", (float32(pay_90/100)-float32(cashout_90/100))/float32(newuser)),
			"pay_180":       fmt.Sprintf("%.2f", (float32(pay_180/100)-float32(cashout_180/100))/float32(newuser)),
		}
		strdaynum := fmt.Sprintf("%d", daynum)
		if _, ok := ratearr[strdaynum]; ok {

			data["rate"] = ratearr[strdaynum]
		} else {
			data["rate"] = 4.9
		}
		if zero {
			data["newuser"] = 0
		}
		//获取新增玩家累计充值

		strall := fmt.Sprintf(`select sum(pay_all) , sum(cash_out_all) from %s where %s pay_first_day = ? `, tbldataname, where2)

		fmt.Printf("strall:%s", strall)

		// timestamp1, timestamp2 := fmt.Sprintf("%d000000", daynum), fmt.Sprintf("%d235959", daynum)
		all := db_monitor.QueryRow(strall, daynum)

		var allnewpay, allnewcashout uint32

		all.Scan(&allnewpay, &allnewcashout)

		data["paynewall"] = fmt.Sprintf("%.2f", float32(allnewpay/100))
		data["cashoutnewall"] = fmt.Sprintf("%.2f", float32(allnewcashout/100))

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

//arppu

func HandleUserARPPUAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	tblname := get_user_daily_roi_table(gameid)
	tbldataname := get_user_data_table(gameid)
	tbltodaycost := get_app_cost_today_table(gameid)
	// tblpay := get_user_pay_table(gameid)
	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))
	//获取汇率
	ratetable := get_exchange_rate_table(gameid)
	strrate := fmt.Sprintf(`select time , exchange_rate from %s where currency = 1`, ratetable)

	rowrate, err := db_monitor.Query(strrate)

	ratearr := make(map[string]float32)

	defer rowrate.Close()

	if err == nil {

		for rowrate.Next() {
			var time string
			var rate float32
			rowrate.Scan(&time, &rate)
			ratekey := strings.Replace(time, "-", "", -1)
			ratearr[string(ratekey)] = float32(rate)

		}

	}
	where := fmt.Sprintf(" zoneid != 0 AND ")
	where3 := fmt.Sprintf(" b.zoneid != 0 AND ")
	whereapp := ""
	if plats != "" {
		platlist := strings.Split(plats, ",")
		if len(platlist) == 1 {
			where += fmt.Sprintf(" platid=%s AND ", platlist[0])
			where3 += fmt.Sprintf(" b.platid=%s AND ", platlist[0])
			whereapp += fmt.Sprintf(" appid=%s AND ", platlist[0])
		} else if len(platlist) > 1 {
			//如果不是1，需要根据权限列获得相应的渠道id
			if len(platlist) != 1 {
				ptlist := getAccountPlatList(task)
				if len(ptlist) > 0 {
					platlist = ptlist
				}
			}
			where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(platlist, ","))
			where3 += fmt.Sprintf(" b.platid in (%s) AND ", strings.Join(platlist, ","))
			whereapp += fmt.Sprintf(" appid in (%s) AND ", strings.Join(platlist, ","))
		}
	} else {

		//如果为空，也要检测下是否有权限设置 ,没有设置就是最大权限
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(ptlist, ","))
			where3 += fmt.Sprintf(" b.platid in (%s) AND ", strings.Join(ptlist, ","))
			whereapp += fmt.Sprintf(" appid in (%s) AND ", strings.Join(ptlist, ","))
		}
	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
		where3 += fmt.Sprintf(" b.launchid = '%s' AND ", gameSystem)
		whereapp += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf(" ad_account = '%s' AND ", launchAccount)
		where3 += fmt.Sprintf(" b.ad_account = '%s' AND ", launchAccount)
	}

	where2 := where
	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)
	whereapp += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)
	costarr := make(map[uint32]float32)

	//获取消耗
	str1 := fmt.Sprintf(`select daynum,sum(cost) from %s where %s group by daynum`, tbltodaycost, whereapp)
	rowsapp, err := db_monitor.Query(str1)

	if err != nil {
		task.Error("HandleRecoveryData error:%s sql:%s", err.Error(), str1)

	} else {
		for rowsapp.Next() {
			var daynumapp uint32
			var costapp float32

			rowsapp.Scan(&daynumapp, &costapp)
			costarr[daynumapp] = costapp

		}
	}
	defer rowsapp.Close()

	str := fmt.Sprintf(`select daynum,sum(pay_first_num), sum(pay_first),sum(pay_1), sum(pay_2),sum(pay_3),sum(pay_4),sum(pay_5),sum(pay_6),sum(pay_7), sum(pay_8), sum(pay_9), sum(pay_10),sum(pay_11),sum(pay_12),sum(pay_13),sum(pay_14),sum(pay_21),sum(pay_30),sum(pay_45),sum(pay_60),sum(pay_90),sum(pay_180),sum(cost),sum(cashout_1), sum(cashout_2),sum(cashout_3),sum(cashout_4),sum(cashout_5),sum(cashout_6),sum(cashout_7), sum(cashout_8), sum(cashout_9), sum(cashout_10),sum(cashout_11),sum(cashout_12),sum(cashout_13),sum(cashout_14),sum(cashout_21),sum(cashout_30),sum(cashout_45),sum(cashout_60),sum(cashout_90),sum(cashout_180) from %s where %s group by daynum`, tblname, where)
	rows, err := db_monitor.Query(str)
	task.Debug("HandleUserLTVAnalysis sql:%s", str)
	if err != nil {
		task.Error("HandleUserLTVAnalysis error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var daynum, paytoday, payuser, pay_1, pay_2, pay_3, pay_4, pay_5, pay_6, pay_7, pay_8, pay_9, pay_10, pay_11, pay_12, pay_13, pay_14, pay_21, pay_30, pay_45, pay_60, pay_90, pay_180, cashout_1, cashout_2, cashout_3, cashout_4, cashout_5, cashout_6, cashout_7, cashout_8, cashout_9, cashout_10, cashout_11, cashout_12, cashout_13, cashout_14, cashout_21, cashout_30, cashout_45, cashout_60, cashout_90, cashout_180 uint32
		var cost float32
		if err = rows.Scan(&daynum, &payuser, &paytoday, &pay_1, &pay_2, &pay_3, &pay_4, &pay_5, &pay_6, &pay_7, &pay_8, &pay_9, &pay_10, &pay_11, &pay_12, &pay_13, &pay_14, &pay_21, &pay_30, &pay_45, &pay_60, &pay_90, &pay_180, &cost, &cashout_1, &cashout_2, &cashout_3, &cashout_4, &cashout_5, &cashout_6, &cashout_7, &cashout_8, &cashout_9, &cashout_10, &cashout_11, &cashout_12, &cashout_13, &cashout_14, &cashout_21, &cashout_30, &cashout_45, &cashout_60, &cashout_90, &cashout_180); err != nil {
			task.Error("HandleUserLTVAnalysis error:%s", err.Error())
			continue
		}
		var zero = payuser == 0
		if zero {
			payuser = 1
		}
		_, ok := costarr[daynum]
		if ok {
			cost = costarr[daynum]
		}
		data := map[string]interface{}{
			"daynum":     daynum,
			"payuser":    payuser,
			"paytoday":   fmt.Sprintf("%.2f", float32(paytoday/100)),
			"paynewall":  0.00,
			"cashnewall": 0.00,
			"avgpaynum":  0.00,
			"cost":       cost,
			"rate":       0.00,
			"newpay":     fmt.Sprintf("%.2f", float32(pay_1/100)),
			"pay_1":      fmt.Sprintf("%.2f", (float32(pay_1/100)-float32(cashout_1/100))/float32(payuser)),
			"pay_2":      fmt.Sprintf("%.2f", (float32(pay_2/100)-float32(cashout_2/100))/float32(payuser)),
			"pay_3":      fmt.Sprintf("%.2f", (float32(pay_3/100)-float32(cashout_3/100))/float32(payuser)),
			"pay_4":      fmt.Sprintf("%.2f", (float32(pay_4/100)-float32(cashout_4/100))/float32(payuser)),
			"pay_5":      fmt.Sprintf("%.2f", (float32(pay_5/100)-float32(cashout_5/100))/float32(payuser)),
			"pay_6":      fmt.Sprintf("%.2f", (float32(pay_6/100)-float32(cashout_6/100))/float32(payuser)),
			"pay_7":      fmt.Sprintf("%.2f", (float32(pay_7/100)-float32(cashout_7/100))/float32(payuser)),
			"pay_8":      fmt.Sprintf("%.2f", (float32(pay_8/100)-float32(cashout_8/100))/float32(payuser)),
			"pay_9":      fmt.Sprintf("%.2f", (float32(pay_9/100)-float32(cashout_9/100))/float32(payuser)),
			"pay_10":     fmt.Sprintf("%.2f", (float32(pay_10/100)-float32(cashout_10/100))/float32(payuser)),
			"pay_11":     fmt.Sprintf("%.2f", (float32(pay_11/100)-float32(cashout_11/100))/float32(payuser)),
			"pay_12":     fmt.Sprintf("%.2f", (float32(pay_12/100)-float32(cashout_12/100))/float32(payuser)),
			"pay_13":     fmt.Sprintf("%.2f", (float32(pay_13/100)-float32(cashout_13/100))/float32(payuser)),
			"pay_14":     fmt.Sprintf("%.2f", (float32(pay_14/100)-float32(cashout_14/100))/float32(payuser)),
			"pay_21":     fmt.Sprintf("%.2f", (float32(pay_21/100)-float32(cashout_21/100))/float32(payuser)),
			"pay_30":     fmt.Sprintf("%.2f", (float32(pay_30/100)-float32(cashout_30/100))/float32(payuser)),
			"pay_45":     fmt.Sprintf("%.2f", (float32(pay_45/100)-float32(cashout_45/100))/float32(payuser)),
			"pay_60":     fmt.Sprintf("%.2f", (float32(pay_60/100)-float32(cashout_60/100))/float32(payuser)),
			"pay_90":     fmt.Sprintf("%.2f", (float32(pay_90/100)-float32(cashout_90/100))/float32(payuser)),
			"pay_180":    fmt.Sprintf("%.2f", (float32(pay_180/100)-float32(cashout_180/100))/float32(payuser)),
		}
		strdaynum := fmt.Sprintf("%d", daynum)
		if _, ok := ratearr[strdaynum]; ok {

			data["rate"] = ratearr[strdaynum]
		} else {
			data["rate"] = 4.9
		}
		if zero {
			data["payuser"] = 0
		}

		//获取新增玩家累计付费次数,累计充值
		// strpaynum := fmt.Sprintf(`SELECT count(*) , sum(a.money) from %s as a LEFT JOIN %s as b on a.userid = b.userid where %s b.pay_first_day =? `, tblpay, tbldataname, where3)

		// allpaynum := db_monitor.QueryRow(strpaynum, daynum)

		var allpaynumtoday, allnewpay, allcashout uint32

		// allpaynum.Scan(&allpaynumtoday, &allnewpay)

		//获取累计提现
		strnewcashmoney := fmt.Sprintf(`SELECT sum(pay_all_num), sum(pay_all) , sum(cash_out_all) from %s where %s pay_first_day =? `, tbldataname, where2)

		strcashmoney := db_monitor.QueryRow(strnewcashmoney, daynum)

		strcashmoney.Scan(&allpaynumtoday, &allnewpay, &allcashout)

		data["paynewall"] = fmt.Sprintf("%.2f", float32(allnewpay/100))

		if payuser != 0 {
			data["avgpaynum"] = fmt.Sprintf("%.2f", float32(allpaynumtoday)/float32(payuser))
		}

		data["cashnewall"] = fmt.Sprintf("%.2f", float32(allcashout/100))

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
func HandleRecoveryData(task *unibase.ChanHttpTask) {
	go func(task *unibase.ChanHttpTask) {
		plats := strings.TrimSpace(task.R.FormValue("platlist"))
		// sdate := strings.TrimSpace(task.R.FormValue("sdate"))

		// edate := strings.TrimSpace(task.R.FormValue("edate"))
		gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
		// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

		stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
		etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
		stretime := fmt.Sprintf("%d", etime)
		tblname := get_user_daily_plat_table(gameid)
		tbldata := get_user_data_table(gameid)
		tbltodaycost := get_app_cost_today_table(gameid)
		gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
		launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))
		adaccountid := strings.TrimSpace(task.R.FormValue("ad_account_id"))
		agent_id := strings.TrimSpace(task.R.FormValue("agent_id"))
		fillcost := float32(unibase.Atoi(task.R.FormValue("cost"), 0))

		//获取汇率
		ratetable := get_exchange_rate_table(gameid)
		strrate := fmt.Sprintf(`select time , exchange_rate from %s where currency = 1`, ratetable)

		rowrate, err := db_monitor.Query(strrate)

		ratearr := make(map[string]float32)
		if err != nil {
			task.Error("HandleRecoveryData error:%s sql:%s", err.Error(), strrate)

		} else {

			for rowrate.Next() {
				var time string
				var rate float32
				rowrate.Scan(&time, &rate)
				ratekey := strings.Replace(time, "-", "", -1)
				ratearr[string(ratekey)] = float32(rate)

			}
			rowrate.Close()

		}

		retl := make([]map[string]interface{}, 0)

		allnewuser, allpay24num, allfirstnum, allpaynum, allcashoutnum, allisband := uint32(0), uint32(0), uint32(0), uint32(0), uint32(0), uint32(0)

		allcost, allpay, allcash, terncost, allfirstpay := float32(0.00), float32(0.00), float32(0.00), float32(0.00), float32(0.00)

		var where, wherekey, whereapp = "", "", ""

		if plats != "" {
			platlist := strings.Split(plats, ",")
			if len(platlist) == 1 {
				where += fmt.Sprintf(" platid=%s AND ", platlist[0])
				whereapp += fmt.Sprintf(" appid=%s AND ", platlist[0])
			} else if len(platlist) > 1 {
				where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(platlist, ","))
				whereapp += fmt.Sprintf(" appid in (%s) AND ", strings.Join(platlist, ","))
			}
		}
		if gameSystem != "all" {
			where += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
			whereapp += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
		}

		strsaa := ""
		if adaccountid != "" || agent_id != "" || launchAccount != "" {

			if launchAccount != "" {
				wherekey += fmt.Sprintf(" keywords = '%s' AND ", launchAccount)
			}

			if adaccountid != "" {
				wherekey += fmt.Sprintf(" ad_account_id='%s' AND ", adaccountid)
				whereapp += fmt.Sprintf(" ad_account_id='%s' AND ", adaccountid)
			}

			if agent_id != "" {
				wherekey += fmt.Sprintf(" agent_id='%s' AND ", agent_id)
				whereapp += fmt.Sprintf(" agent_id = '%s' AND ", agent_id)
			}
			wherekey += " id > 0 "
			tbl := get_launch_keys_table(gameid)

			str := fmt.Sprintf("select keywords from %s  where %s ", tbl, wherekey)
			res, err := db_monitor.Query(str)

			var keys []string
			var keys1 []string

			if err == nil {
				for res.Next() {
					var keywords string
					res.Scan(&keywords)
					keys = append(keys, `"`+keywords+`"`)
					keys1 = append(keys1, keywords)
				}

				if len(keys) > 0 {
					str1 := strings.Join(keys, ",")

					strsaa = strings.Join(keys1, ",")

					where += fmt.Sprintf(" ad_account in (%s) AND ", str1)
				} else {
					if launchAccount != "" {
						where += fmt.Sprintf(" ad_account = '%s' AND ", launchAccount)
					} else {
						where += fmt.Sprintf(" ad_account = '00000' AND ")
					}
				}
			}
			res.Close()

		}
		if strsaa != "" {
			launchAccount = strsaa
		}
		whereapp += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)
		costarr := make(map[string]float32)
		if fillcost != 0 {
			costarr[stretime] = fillcost
		} else {
			//获取消耗
			str := fmt.Sprintf(`select daynum,sum(cost) from %s where %s group by daynum`, tbltodaycost, whereapp)
			rowsapp, err := db_monitor.Query(str)

			if err != nil {
				task.Error("HandleRecoveryData error:%s sql:%s", err.Error(), str)

			} else {
				if rowsapp != nil {

					for rowsapp.Next() {
						var daynumapp string
						var costapp float32

						rowsapp.Scan(&daynumapp, &costapp)
						costarr[daynumapp] = costapp

					}

				}
				rowsapp.Close()
			}

		}

		if etime > uint32(unitime.Time.YearMonthDay()) {
			etime = uint32(unitime.Time.YearMonthDay())
		}

		if etime == uint32(unitime.Time.YearMonthDay()) {

			// tblpay := get_user_pay_table(gameid)
			// tblcash := get_user_cashout_table(gameid)
			//获取今日新增玩家
			// strnew := fmt.Sprintf(`select count(accid) from %s where %s reg_time >= UNIX_TIMESTAMP(?)`, tbldata, where)

			// resnew := db_monitor.QueryRow(strnew, etime)

			// var tdnewuser uint32

			// resnew.Scan(&tdnewuser)

			//获取今日充值信息
			var moneypay, numpay, moneycash, numcash, isband, tdnewuser uint32

			strdate := fmt.Sprintf("select sum(pay_all) , sum(cash_out_all) , sum(case when pay_all >0 then 1 else 0 end) as paynum , sum(case when cash_out_all >0 then 1 else 0 end) as cashoutnum ,  sum(case when plataccount != '' then 1 else 0 end) as isband , count(accid) as newuser from %s where %s reg_time >= UNIX_TIMESTAMP(?) ", tbldata, where)

			strdatearr := db_monitor.QueryRow(strdate, etime)

			strdatearr.Scan(&moneypay, &moneycash, &numpay, &numcash, &isband, &tdnewuser)

			data := map[string]interface{}{
				"daynum":       etime,
				"newuser":      tdnewuser,
				"paytoday":     fmt.Sprintf("%.2f", float32(moneypay)/100),
				"cashouttoday": fmt.Sprintf("%.2f", float32(moneycash)/100),
				"firstpay":     fmt.Sprintf("%.2f", float32(moneypay)/100),
				"firstnum":     numpay,
				"firstrate":    0.00,
				"pay24num":     numpay,
				"paynum":       numpay,
				"cashoutnum":   numcash,
				"pay_1":        0.00,
				"pay_2":        0.00,
				"pay_3":        0.00,
				"pay_4":        0.00,
				"pay_5":        0.00,
				"pay_6":        0.00,
				"pay_7":        0.00,
				"pay_8":        0.00,
				"pay_9":        0.00,
				"pay_10":       0.00,
				"pay_11":       0.00,
				"pay_12":       0.00,
				"pay_13":       0.00,
				"pay_14":       0.00,
				"pay_21":       0.00,
				"pay_30":       0.00,
				"pay_45":       0.00,
				"pay_60":       0.00,
				"pay_90":       0.00,
				"pay_180":      0.00,
				"pay_today":    0.00,
				"cost":         0.00,
				"single":       0.00,
				"rate":         0.00,
				"cash_today":   0.00,
				"is_band":      isband,
			}
			if tdnewuser != 0 {

				data["firstrate"] = fmt.Sprintf("%.2f%%", (float32(numpay)/float32(tdnewuser))*100)
			}

			allnewuser += tdnewuser
			allpay24num += numpay
			allfirstnum += numpay
			allpaynum += numpay
			allcashoutnum += numcash
			allisband += isband
			firstpaya, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(moneypay)/100), 32)

			pay, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(moneypay)/100), 32)
			cash, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(moneycash)/100), 32)

			allpay += *proto.Float32(float32(pay))
			allcash += *proto.Float32(float32(cash))
			allfirstpay += *proto.Float32(float32(firstpaya))

			if _, ok := ratearr[stretime]; ok {

				data["rate"] = ratearr[stretime]
			}
			// date := time.Now().Format("2006-01-02")
			// costarr := make(map[string]float32)

			costnum := fillcost
			if fillcost == 0 {
				//获取消耗
				_, ok := costarr[stretime]
				if ok {
					costnum = costarr[stretime]
				}

			}
			if costnum != 0 {
				allcost += float32(costnum)

				data["cost"] = costnum

				if tdnewuser != 0 {
					data["single"] = fmt.Sprintf("%.2f", (float32(costnum) / float32(tdnewuser)))
				}
				_, ok := ratearr[stretime]
				if ok && ratearr[stretime] > float32(0) {

					// data["pay_today"] = fmt.Sprintf("%.2f", (float32((moneypay-moneycash)/100)/float32(costnum*ratearr[stretime]))*100)
					data["pay_today"] = fmt.Sprintf("%.2f%%", ((float32(moneypay) - float32(moneycash)) / 100 / float32(costnum*ratearr[stretime]) * 100))

					// data["pay_today"] = fmt.Sprintf("%.2f", (float32(moneypay/100)/float32(fillcost*ratearr[stretime]))*100)

					terncost += float32(costnum * ratearr[stretime])
				} else {

					// data["pay_today"] = fmt.Sprintf("%.2f", (float32(moneypay/100)/float32(fillcost))*100)
					// data["pay_today"] = fmt.Sprintf("%.2f", (float32((moneypay-moneycash)/100)/float32(costnum))*100)
					data["pay_today"] = fmt.Sprintf("%.2f%%", ((float32(moneypay) - float32(moneycash)) / 100 / float32(costnum) * 100))

					terncost += float32(costnum)

				}
			}
			retl = append(retl, data)
		}
		where2 := where
		where += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)

		str := fmt.Sprintf(`select daynum,sum(day_1),sum(pay_today),sum(pay_cashout),  sum(pay_1), sum(pay_2),sum(pay_3),sum(pay_4),sum(pay_5),sum(pay_6),sum(pay_7), sum(pay_8), sum(pay_9), sum(pay_10),sum(pay_11),sum(pay_12),sum(pay_13),sum(pay_14),sum(pay_21),sum(pay_30),sum(pay_45),sum(pay_60),sum(pay_90),sum(pay_180) , sum(pay_first) , sum(pay_first_num) , sum(cost) ,  sum(cashout_1), sum(cashout_2),sum(cashout_3),sum(cashout_4),sum(cashout_5),sum(cashout_6),sum(cashout_7), sum(cashout_8), sum(cashout_9), sum(cashout_10),sum(cashout_11),sum(cashout_12),sum(cashout_13),sum(cashout_14),sum(cashout_21),sum(cashout_30),sum(cashout_45),sum(cashout_60),sum(cashout_90),sum(cashout_180) from %s where %s group by daynum`, tblname, where)
		rows, err := db_monitor.Query(str)

		if err != nil {
			task.Error("HandleRecoveryData error:%s sql:%s", err.Error(), str)
			return
		} else {
			defer rows.Close()
		}

		for rows.Next() {
			var newuser, paytoday_all, paytoday, cashouttoday, cashouttodayall, pay_1, pay_2, pay_3, pay_4, pay_5, pay_6, pay_7, pay_8, pay_9, pay_10, pay_11, pay_12, pay_13, pay_14, pay_21, pay_30, pay_45, pay_60, pay_90, pay_180, firstpay, firstnum, pay24num, cashout_1, cashout_2, cashout_3, cashout_4, cashout_5, cashout_6, cashout_7, cashout_8, cashout_9, cashout_10, cashout_11, cashout_12, cashout_13, cashout_14, cashout_21, cashout_30, cashout_45, cashout_60, cashout_90, cashout_180, paynum, cashoutnum, isband uint32

			var cost float32

			var daynum string
			if err = rows.Scan(&daynum, &newuser, &paytoday_all, &cashouttodayall, &pay_1, &pay_2, &pay_3, &pay_4, &pay_5, &pay_6, &pay_7, &pay_8, &pay_9, &pay_10, &pay_11, &pay_12, &pay_13, &pay_14, &pay_21, &pay_30, &pay_45, &pay_60, &pay_90, &pay_180, &firstpay, &firstnum, &cost, &cashout_1, &cashout_2, &cashout_3, &cashout_4, &cashout_5, &cashout_6, &cashout_7, &cashout_8, &cashout_9, &cashout_10, &cashout_11, &cashout_12, &cashout_13, &cashout_14, &cashout_21, &cashout_30,
				&cashout_45, &cashout_60, &cashout_90, &cashout_180); err != nil {
				task.Error("HandleRecoveryData error:%s", err.Error())
				continue
			}
			var zero = newuser == 0
			// if zero {
			// 	newuser = 1
			// }

			data := map[string]interface{}{
				"daynum":       daynum,
				"newuser":      0,
				"firstpay":     fmt.Sprintf("%.2f", float32(pay_1)/100),
				"firstnum":     firstnum,
				"firstrate":    0.00,
				"paytoday":     0.00,
				"cashouttoday": 0.00,
				"pay24num":     0,
				"paynum":       0,
				"cashoutnum":   0,
				"pay_1":        0.00,
				"pay_2":        0.00,
				"pay_3":        0.00,
				"pay_4":        0.00,
				"pay_5":        0.00,
				"pay_6":        0.00,
				"pay_7":        0.00,
				"pay_8":        0.00,
				"pay_9":        0.00,
				"pay_10":       0.00,
				"pay_11":       0.00,
				"pay_12":       0.00,
				"pay_13":       0.00,
				"pay_14":       0.00,
				"pay_21":       0.00,
				"pay_30":       0.00,
				"pay_45":       0.00,
				"pay_60":       0.00,
				"pay_90":       0.00,
				"pay_180":      0.00,
				"pay_today":    0.00,
				"cost":         0.00,
				"single":       0.00,
				"rate":         0.00,
				"is_band":      0,
				// "cash_today":   0.00,
			}
			//获取新增玩家累计充值,累计提现,24小时内充值人数

			strall := fmt.Sprintf(`select sum(pay_all) , sum(cash_out_all) , sum(case when pay_first_time >0 and (pay_first_time - reg_time ) < 24*60*60 then 1 else 0 end) ,sum(case when pay_all > 0 then 1 else 0 end) ,sum(case when cash_out_all > 0 then 1 else 0 end) , sum(case when plataccount != '' then 1 else 0 end) as isband,count(accid) as newuser  from %s where %s reg_time >= UNIX_TIMESTAMP(?) and  reg_time <= UNIX_TIMESTAMP(?) `, tbldata, where2)

			timestamp1, timestamp2 := fmt.Sprintf("%s000000", daynum), fmt.Sprintf("%s235959", daynum)
			all := db_monitor.QueryRow(strall, timestamp1, timestamp2)

			all.Scan(&paytoday, &cashouttoday, &pay24num, &paynum, &cashoutnum, &isband, &newuser)

			data["paytoday"] = fmt.Sprintf("%.2f", float32(paytoday)/100)
			data["cashouttoday"] = fmt.Sprintf("%.2f", float32(cashouttoday)/100)
			data["pay24num"] = pay24num
			data["paynum"] = paynum
			data["cashoutnum"] = cashoutnum
			data["is_band"] = isband
			data["newuser"] = newuser

			allnewuser += newuser

			allpay24num += pay24num
			allfirstnum += firstnum
			allpaynum += paynum
			allcashoutnum += cashoutnum
			allisband += isband

			firstpaya, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(pay_1)/100), 32)

			pay, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(paytoday)/100), 32)
			cash, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(cashouttoday)/100), 32)

			allpay += *proto.Float32(float32(pay))
			allcash += *proto.Float32(float32(cash))
			allfirstpay += *proto.Float32(float32(firstpaya))

			// allpay += float32(paytoday / 100)
			// allcash += float32(cashouttoday / 100)

			strdaynum := fmt.Sprintf("%s", daynum)

			if zero {
				data["newuser"] = 0
			}
			if fillcost == 0 {
				//获取消耗
				_, ok := costarr[strdaynum]
				if ok {
					cost = costarr[strdaynum]
				}

			} else {
				cost = fillcost
			}
			if newuser != 0 {
				data["single"] = fmt.Sprintf("%.2f", (float32(cost) / float32(newuser)))
				data["firstrate"] = fmt.Sprintf("%.2f%%", (float32(firstnum)/float32(newuser))*100)
			}
			if _, ok := ratearr[strdaynum]; ok {

				data["rate"] = ratearr[strdaynum]
			}
			data["cost"] = cost
			if cost > 0 {
				allcost += float32(cost)

				_, ok := ratearr[strdaynum]

				if ok && float32(ratearr[strdaynum]) > 0 {

					data["rate"] = ratearr[strdaynum]

					data["pay_1"] = fmt.Sprintf("%.2f%%", ((float32(pay_1)-float32(cashout_1))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_2"] = fmt.Sprintf("%.2f%%", ((float32(pay_2)-float32(cashout_2))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_3"] = fmt.Sprintf("%.2f%%", ((float32(pay_3)-float32(cashout_3))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_4"] = fmt.Sprintf("%.2f%%", ((float32(pay_4)-float32(cashout_4))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_5"] = fmt.Sprintf("%.2f%%", ((float32(pay_5)-float32(cashout_5))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_6"] = fmt.Sprintf("%.2f%%", ((float32(pay_6)-float32(cashout_6))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_7"] = fmt.Sprintf("%.2f%%", ((float32(pay_7)-float32(cashout_7))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_8"] = fmt.Sprintf("%.2f%%", ((float32(pay_8)-float32(cashout_8))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_9"] = fmt.Sprintf("%.2f%%", ((float32(pay_9)-float32(cashout_9))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_10"] = fmt.Sprintf("%.2f%%", ((float32(pay_10)-float32(cashout_10))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_11"] = fmt.Sprintf("%.2f%%", ((float32(pay_11)-float32(cashout_11))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_12"] = fmt.Sprintf("%.2f%%", ((float32(pay_12)-float32(cashout_12))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_13"] = fmt.Sprintf("%.2f%%", ((float32(pay_13)-float32(cashout_13))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_14"] = fmt.Sprintf("%.2f%%", ((float32(pay_14)-float32(cashout_14))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_21"] = fmt.Sprintf("%.2f%%", ((float32(pay_21)-float32(cashout_21))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_30"] = fmt.Sprintf("%.2f%%", ((float32(pay_30)-float32(cashout_30))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_45"] = fmt.Sprintf("%.2f%%", ((float32(pay_45)-float32(cashout_45))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_60"] = fmt.Sprintf("%.2f%%", ((float32(pay_60)-float32(cashout_60))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_90"] = fmt.Sprintf("%.2f%%", ((float32(pay_90)-float32(cashout_90))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_180"] = fmt.Sprintf("%.2f%%", ((float32(pay_180)-float32(cashout_180))/100/float32(cost*ratearr[strdaynum]))*100)
					data["pay_today"] = fmt.Sprintf("%.2f%%", ((float32(paytoday)-float32(cashouttoday))/100/float32(cost*ratearr[strdaynum]))*100)
					// data["cash_today"] = fmt.Sprintf("%.2f%%", ((float32(paytoday)-float32(cashouttoday))/100/float32(cost*ratearr[strdaynum]))*100)

					terncost += float32(cost * ratearr[strdaynum])

				} else {
					data["pay_1"] = fmt.Sprintf("%.2f%%", ((float32(pay_1)-float32(cashout_1))/100/float32(cost))*100)
					data["pay_2"] = fmt.Sprintf("%.2f%%", ((float32(pay_2)-float32(cashout_2))/100/float32(cost))*100)
					data["pay_3"] = fmt.Sprintf("%.2f%%", ((float32(pay_3)-float32(cashout_3))/100/float32(cost))*100)
					data["pay_4"] = fmt.Sprintf("%.2f%%", ((float32(pay_4)-float32(cashout_4))/100/float32(cost))*100)
					data["pay_5"] = fmt.Sprintf("%.2f%%", ((float32(pay_5)-float32(cashout_5))/100/float32(cost))*100)
					data["pay_6"] = fmt.Sprintf("%.2f%%", ((float32(pay_6)-float32(cashout_6))/100/float32(cost))*100)
					data["pay_7"] = fmt.Sprintf("%.2f%%", ((float32(pay_7)-float32(cashout_7))/100/float32(cost))*100)
					data["pay_8"] = fmt.Sprintf("%.2f%%", ((float32(pay_8)-float32(cashout_8))/100/float32(cost))*100)
					data["pay_9"] = fmt.Sprintf("%.2f%%", ((float32(pay_9)-float32(cashout_9))/100/float32(cost))*100)
					data["pay_10"] = fmt.Sprintf("%.2f%%", ((float32(pay_10)-float32(cashout_10))/100/float32(cost))*100)
					data["pay_11"] = fmt.Sprintf("%.2f%%", ((float32(pay_11)-float32(cashout_11))/100/float32(cost))*100)
					data["pay_12"] = fmt.Sprintf("%.2f%%", ((float32(pay_12)-float32(cashout_12))/100/float32(cost))*100)
					data["pay_13"] = fmt.Sprintf("%.2f%%", ((float32(pay_13)-float32(cashout_13))/100/float32(cost))*100)
					data["pay_14"] = fmt.Sprintf("%.2f%%", ((float32(pay_14)-float32(cashout_14))/100/float32(cost))*100)
					data["pay_21"] = fmt.Sprintf("%.2f%%", ((float32(pay_21)-float32(cashout_21))/100/float32(cost))*100)
					data["pay_30"] = fmt.Sprintf("%.2f%%", ((float32(pay_30)-float32(cashout_30))/100/float32(cost))*100)
					data["pay_45"] = fmt.Sprintf("%.2f%%", ((float32(pay_45)-float32(cashout_45))/100/float32(cost))*100)
					data["pay_60"] = fmt.Sprintf("%.2f%%", ((float32(pay_60)-float32(cashout_60))/100/float32(cost))*100)
					data["pay_90"] = fmt.Sprintf("%.2f%%", ((float32(pay_90)-float32(cashout_90))/100/float32(cost))*100)
					data["pay_180"] = fmt.Sprintf("%.2f%%", ((float32(pay_180)-float32(cashout_180))/100/float32(cost))*100)
					data["pay_today"] = fmt.Sprintf("%.2f%%", ((float32(paytoday)-float32(cashouttoday))/100/float32(cost))*100)

					// data["cash_today"] = fmt.Sprintf("%.2f%%", ((float32(paytoday)-float32(cashouttoday))/100/float32(cost))*100)

					terncost += float32(cost)
				}

			}

			retl = append(retl, data)
		}

		alldata := map[string]interface{}{
			"daynum":       "汇总",
			"newuser":      allnewuser,
			"paytoday":     allpay,
			"cashouttoday": allcash,
			"firstpay":     allfirstpay,
			"firstnum":     allfirstnum,
			"paynum":       allpaynum,
			"cashoutnum":   allcashoutnum,
			"firstrate":    "",
			"pay_today":    0.00,
			"pay24num":     allpay24num,
			"cost":         allcost,
			"pay_1":        "",
			"pay_2":        "",
			"pay_3":        "",
			"pay_4":        "",
			"pay_5":        "",
			"pay_6":        "",
			"pay_7":        "",
			"pay_8":        "",
			"pay_9":        "",
			"pay_10":       "",
			"pay_11":       "",
			"pay_12":       "",
			"pay_13":       "",
			"pay_14":       "",
			"pay_21":       "",
			"pay_30":       "",
			"pay_45":       "",
			"pay_60":       "",
			"pay_90":       "",
			"pay_180":      "",
			// "cash_today": 0.00,

			"single":  0.00,
			"rate":    0.00,
			"is_band": allisband,
		}
		if terncost != 0 {
			alldata["pay_today"] = fmt.Sprintf("%.2f%%", ((float32(allpay) - float32(allcash)) / float32(terncost) * 100))
			// alldata["cash_today"] = fmt.Sprintf("%.2f%%", (float32(allpay-allcash) / float32(terncost) * 100))
		}

		if allnewuser != 0 {
			alldata["single"] = fmt.Sprintf("%.2f", (float32(allcost) / float32(allnewuser)))
			alldata["firstrate"] = fmt.Sprintf("%.2f%%", (float32(allfirstnum)/float32(allnewuser))*100)
		}

		retl = append(retl, alldata)

		var data []byte
		if len(retl) == 0 {
			data = []byte("[]")
		} else {
			data, _ = json.Marshal(retl)
		}
		task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
		retl = nil
	}(task)
}

func HandleRecoveryDataCash(task *unibase.ChanHttpTask) {
	go func(task *unibase.ChanHttpTask) {
		plats := strings.TrimSpace(task.R.FormValue("platlist"))
		gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
		stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
		etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
		stretime := fmt.Sprintf("%d", etime)
		tblname := get_user_daily_roi_table(gameid)
		tbldata := get_user_data_table(gameid)
		tbltodaycost := get_app_cost_today_table(gameid)
		gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
		launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))
		adaccountid := strings.TrimSpace(task.R.FormValue("ad_account_id"))
		agent_id := strings.TrimSpace(task.R.FormValue("agent_id"))
		fillcost := float32(unibase.Atoi(task.R.FormValue("cost"), 0))

		//获取汇率
		ratetable := get_exchange_rate_table(gameid)
		strrate := fmt.Sprintf(`select time , exchange_rate from %s where currency = 1`, ratetable)

		rowrate, err := db_monitor.Query(strrate)

		ratearr := make(map[string]float32)

		defer rowrate.Close()

		if err == nil {

			for rowrate.Next() {
				var time string
				var rate float32
				rowrate.Scan(&time, &rate)
				ratekey := strings.Replace(time, "-", "", -1)
				ratearr[string(ratekey)] = float32(rate)

			}

		}
		retl := make([]map[string]interface{}, 0)

		allnewuser, allfirstnum := uint32(0), uint32(0)

		allcost, allpay, allcash, terncost, allfirstpay, allfirstcash := float32(0.00), float32(0.00), float32(0.00), float32(0.00), float32(0.00), float32(0.00)

		var where, wherekey, whereapp = "", "", ""

		if plats != "" {
			platlist := strings.Split(plats, ",")
			if len(platlist) == 1 {
				where += fmt.Sprintf(" platid=%s AND ", platlist[0])
				whereapp += fmt.Sprintf(" appid=%s AND ", platlist[0])
			} else if len(platlist) > 1 {
				where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(platlist, ","))
				whereapp += fmt.Sprintf(" appid in (%s) AND ", strings.Join(platlist, ","))
			}
		}
		if gameSystem != "all" {
			where += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
			whereapp += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
		}

		strsaa := ""
		if adaccountid != "" || agent_id != "" || launchAccount != "" {

			if launchAccount != "" {
				wherekey += fmt.Sprintf(" keywords = '%s' AND ", launchAccount)
			}

			if adaccountid != "" {
				wherekey += fmt.Sprintf(" ad_account_id='%s' AND ", adaccountid)
				whereapp += fmt.Sprintf(" ad_account_id='%s' AND ", adaccountid)
			}

			if agent_id != "" {
				wherekey += fmt.Sprintf(" agent_id='%s' AND ", agent_id)
				whereapp += fmt.Sprintf(" agent_id='%s' AND ", agent_id)
			}

			wherekey += " id > 0 "

			tbl := get_launch_keys_table(gameid)

			str := fmt.Sprintf("select keywords from %s where %s ", tbl, wherekey)

			res, err := db_monitor.Query(str)

			defer res.Close()

			var keys []string
			var keys1 []string

			if err == nil {
				for res.Next() {
					var keywords string
					res.Scan(&keywords)
					keys = append(keys, `"`+keywords+`"`)
					keys1 = append(keys1, keywords)
				}

				if len(keys) > 0 {
					str1 := strings.Join(keys, ",")
					strsaa = strings.Join(keys1, ",")
					where += fmt.Sprintf(" ad_account in (%s) AND ", str1)
				} else {
					if launchAccount != "" {
						where += fmt.Sprintf(" ad_account = '%s' AND ", launchAccount)
					} else {
						where += fmt.Sprintf(" ad_account = '00000' AND ")
					}
				}
			}

		}
		if strsaa != "" {
			launchAccount = strsaa
		}
		if etime > uint32(unitime.Time.YearMonthDay()) {
			etime = uint32(unitime.Time.YearMonthDay())
		}
		whereapp += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)
		costarr := make(map[string]float32)
		if fillcost != 0 {
			costarr[stretime] = fillcost
		} else {
			//获取消耗
			str := fmt.Sprintf(`select daynum,sum(cost) from %s where %s group by daynum`, tbltodaycost, whereapp)
			rowsapp, err := db_monitor.Query(str)

			if err != nil {
				task.Error("HandleRecoveryData error:%s sql:%s", err.Error(), str)

			} else {
				for rowsapp.Next() {
					var daynumapp string
					var costapp float32

					rowsapp.Scan(&daynumapp, &costapp)
					costarr[daynumapp] = costapp

				}
				rowsapp.Close()
			}

		}
		if etime == uint32(unitime.Time.YearMonthDay()) {

			// tblpay := get_user_pay_table(gameid)
			// tblcash := get_user_cashout_table(gameid)
			//获取今日新增玩家
			strnew := fmt.Sprintf(`select count(accid) from %s where %s reg_time >= UNIX_TIMESTAMP(?)`, tbldata, where)

			resnew := db_monitor.QueryRow(strnew, etime)

			var tdnewuser uint32

			resnew.Scan(&tdnewuser)

			//获取今日充值信息
			// etime1, etime2 := fmt.Sprintf("%d000000", etime), fmt.Sprintf("%d235959", etime)

			// strpay := fmt.Sprintf("select sum(b.money) , count(distinct c.userid) from %s as b inner join %s as c on (b.userid=c.userid and b.zoneid=c.zoneid) where %s b.daynum=? and b.type=0 and c.pay_first_day = ? ", tblpay, tbldata, where)

			// respay := db_monitor.QueryRow(strpay, etime, etime)

			// var moneypay, numpay uint32

			// respay.Scan(&moneypay, &numpay)

			// //获取今日提现信息

			// strcash := fmt.Sprintf("select sum(b.money) , count(distinct c.userid) from %s as b inner join %s as c on (b.userid=c.userid and b.zoneid=c.zoneid) where %s  c.pay_first_day = ?", tblcash, tbldata, where)

			// rescash := db_monitor.QueryRow(strcash, etime, etime)

			// var moneycash, numcash uint32

			// rescash.Scan(&moneycash, &numcash)

			//今日首冲

			strfirst := fmt.Sprintf("select  sum(pay_all),sum(cash_out_all), count(distinct userid) from %s where %s pay_first_day = ? ", tbldata, where)

			resfirst := db_monitor.QueryRow(strfirst, etime)

			var firstpay, firstnum, firstcashout uint32

			resfirst.Scan(&firstpay, &firstcashout, &firstnum)

			data := map[string]interface{}{
				"daynum":       etime,
				"newuser":      tdnewuser,
				"allpay":       fmt.Sprintf("%.2f", float32(firstpay)/100),
				"allcash":      fmt.Sprintf("%.2f", float32(firstcashout)/100),
				"firstpay":     fmt.Sprintf("%.2f", float32(firstpay)/100),
				"firstpaycash": fmt.Sprintf("%.2f", ((float32(firstpay) - float32(firstcashout)) / 100)),
				"firstnum":     firstnum,
				"firstrate":    0.00,

				"roi_1":        0.00,
				"roi_2":        0.00,
				"roi_3":        0.00,
				"roi_4":        0.00,
				"roi_5":        0.00,
				"roi_6":        0.00,
				"roi_7":        0.00,
				"roi_8":        0.00,
				"roi_9":        0.00,
				"roi_10":       0.00,
				"roi_11":       0.00,
				"roi_12":       0.00,
				"roi_13":       0.00,
				"roi_14":       0.00,
				"roi_21":       0.00,
				"roi_30":       0.00,
				"roi_45":       0.00,
				"roi_60":       0.00,
				"roi_90":       0.00,
				"roi_180":      0.00,
				"roi_realtime": 0.00,
				"cost":         0.00,
				"single":       0.00,
				"rate":         0.00,
			}
			if tdnewuser != 0 {

				data["firstrate"] = fmt.Sprintf("%.2f%%", (float32(firstnum)/float32(tdnewuser))*100)
			}

			allnewuser += tdnewuser

			allfirstnum += firstnum
			firstpaya, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(firstpay)/100), 32)
			firstcasha, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(firstcashout)/100), 32)

			pay, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(firstpay)/100), 32)
			cash, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(firstcashout)/100), 32)

			allpay += *proto.Float32(float32(pay))
			allcash += *proto.Float32(float32(cash))
			allfirstpay += *proto.Float32(float32(firstpaya))
			allfirstcash += *proto.Float32(float32(firstcasha))

			if _, ok := ratearr[stretime]; ok {

				data["rate"] = ratearr[stretime]
			}
			// date := time.Now().Format("2006-01-02")

			costnum := fillcost
			if fillcost == 0 {
				//获取消耗
				_, ok := costarr[stretime]
				if ok {
					costnum = costarr[stretime]
				}

			}

			if costnum > 0 {
				if tdnewuser != 0 {
					data["single"] = fmt.Sprintf("%.2f", (float32(costnum) / float32(tdnewuser)))
				}

				data["cost"] = costnum
				allcost += float32(costnum)

				_, ok := ratearr[stretime]
				if ok && ratearr[stretime] > float32(0) {

					data["roi_realtime"] = fmt.Sprintf("%.2f", ((float32(firstpay)-float32(firstcashout))/100/float32(costnum*ratearr[stretime]))*100)
					terncost += float32(costnum * ratearr[stretime])
				} else {

					data["roi_realtime"] = fmt.Sprintf("%.2f", ((float32(firstpay)-float32(firstcashout))/100/float32(costnum))*100)
					terncost += float32(costnum)

				}

			}
			retl = append(retl, data)
		}
		where2 := where

		where += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)

		str := fmt.Sprintf(`select daynum,sum(new_user),sum(pay_1), sum(pay_2),sum(pay_3),sum(pay_4),sum(pay_5),sum(pay_6),sum(pay_7), sum(pay_8), sum(pay_9), sum(pay_10),sum(pay_11),sum(pay_12),sum(pay_13),sum(pay_14),sum(pay_21),sum(pay_30),sum(pay_45),sum(pay_60),sum(pay_90),sum(pay_180) , sum(pay_first) , sum(pay_first_num),sum(cashout_1), sum(cashout_2),sum(cashout_3),sum(cashout_4),sum(cashout_5),sum(cashout_6),sum(cashout_7), sum(cashout_8), sum(cashout_9), sum(cashout_10),sum(cashout_11),sum(cashout_12),sum(cashout_13),sum(cashout_14),sum(cashout_21),sum(cashout_30),sum(cashout_45),sum(cashout_60),sum(cashout_90),sum(cashout_180) , sum(cost) from %s where %s group by daynum`, tblname, where)
		rows, err := db_monitor.Query(str)
		//task.Debug("HandleRecoveryData sql:%s", str)
		if err != nil {
			task.Error("HandleRecoveryData error:%s sql:%s", err.Error(), str)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var newuser, allpayto, allcashto, pay_1, pay_2, pay_3, pay_4, pay_5, pay_6, pay_7, pay_8, pay_9, pay_10, pay_11, pay_12, pay_13, pay_14, pay_21, pay_30, pay_45, pay_60, pay_90, pay_180, cashout_1, cashout_2, cashout_3, cashout_4, cashout_5, cashout_6, cashout_7, cashout_8, cashout_9, cashout_10, cashout_11, cashout_12, cashout_13, cashout_14, cashout_21, cashout_30, cashout_45, cashout_60, cashout_90, cashout_180, firstpay, firstnum uint32
			var cost float32

			var daynum string
			if err = rows.Scan(&daynum, &newuser, &pay_1, &pay_2, &pay_3, &pay_4, &pay_5, &pay_6, &pay_7, &pay_8, &pay_9, &pay_10, &pay_11, &pay_12, &pay_13, &pay_14, &pay_21, &pay_30, &pay_45, &pay_60, &pay_90, &pay_180, &firstpay, &firstnum, &cashout_1, &cashout_2, &cashout_3, &cashout_4, &cashout_5, &cashout_6, &cashout_7, &cashout_8, &cashout_9, &cashout_10, &cashout_11, &cashout_12, &cashout_13, &cashout_14, &cashout_21, &cashout_30, &cashout_45, &cashout_60, &cashout_90, &cashout_180, &cost); err != nil {
				task.Error("HandleRecoveryData error:%s", err.Error())
				continue
			}

			var zero = newuser == 0
			// if zero {
			// 	newuser = 1
			// }
			data := map[string]interface{}{
				"daynum":       daynum,
				"newuser":      newuser,
				"firstpay":     fmt.Sprintf("%.2f", float32(pay_1)/100),
				"firstpaycash": fmt.Sprintf("%.2f", ((float32(pay_1) - float32(cashout_1)) / 100)),
				"firstnum":     firstnum,
				"firstrate":    0.00,
				"allpay":       0.00,
				"allcash":      0.00,
				"roi_1":        0.00,
				"roi_2":        0.00,
				"roi_3":        0.00,
				"roi_4":        0.00,
				"roi_5":        0.00,
				"roi_6":        0.00,
				"roi_7":        0.00,
				"roi_8":        0.00,
				"roi_9":        0.00,
				"roi_10":       0.00,
				"roi_11":       0.00,
				"roi_12":       0.00,
				"roi_13":       0.00,
				"roi_14":       0.00,
				"roi_21":       0.00,
				"roi_30":       0.00,
				"roi_45":       0.00,
				"roi_60":       0.00,
				"roi_90":       0.00,
				"roi_180":      0.00,
				"roi_realtime": 0.00,
				"cost":         0.00,
				"single":       0.00,
				"rate":         0.00,
			}
			//获取当日首冲新增玩家累计充值,累计提现

			strall := fmt.Sprintf(`select sum(pay_all) , sum(cash_out_all) from %s where %s pay_first_day = ? `, tbldata, where2)

			all := db_monitor.QueryRow(strall, daynum)

			all.Scan(&allpayto, &allcashto)

			data["allpay"] = fmt.Sprintf("%.2f", float32(allpayto)/100)
			data["allcash"] = fmt.Sprintf("%.2f", float32(allcashto)/100)

			allfirstnum += firstnum
			allnewuser += newuser
			firstpaya, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(pay_1)/100), 32)
			firstcasha, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(cashout_1)/100), 32)

			pay, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(allpayto)/100), 32)
			cash, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(allcashto)/100), 32)

			allpay += *proto.Float32(float32(pay))
			allcash += *proto.Float32(float32(cash))
			allfirstpay += *proto.Float32(float32(firstpaya))
			allfirstcash += *proto.Float32(float32(firstcasha))

			strdaynum := fmt.Sprintf("%s", daynum)

			if zero {
				data["newuser"] = 0

			}
			if fillcost > 0 {
				cost = fillcost
			} else {
				_, ok := costarr[strdaynum]
				if ok {
					cost = costarr[strdaynum]
				}
			}
			if newuser != 0 {
				data["single"] = fmt.Sprintf("%.2f", (float32(cost) / float32(newuser)))
				data["firstrate"] = fmt.Sprintf("%.2f%%", (float32(firstnum)/float32(newuser))*100)
			}
			if _, ok := ratearr[strdaynum]; ok {

				data["rate"] = ratearr[strdaynum]
			}
			data["cost"] = cost

			if cost > 0 {
				allcost += float32(cost)

				_, ok := ratearr[strdaynum]

				if ok && float32(ratearr[strdaynum]) > 0 {

					data["rate"] = ratearr[strdaynum]

					data["roi_1"] = fmt.Sprintf("%.2f%%", ((float32(pay_1)-float32(cashout_1))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_2"] = fmt.Sprintf("%.2f%%", ((float32(pay_2)-float32(cashout_2))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_3"] = fmt.Sprintf("%.2f%%", ((float32(pay_3)-float32(cashout_3))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_4"] = fmt.Sprintf("%.2f%%", ((float32(pay_4)-float32(cashout_4))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_5"] = fmt.Sprintf("%.2f%%", ((float32(pay_5)-float32(cashout_5))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_6"] = fmt.Sprintf("%.2f%%", ((float32(pay_6)-float32(cashout_6))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_7"] = fmt.Sprintf("%.2f%%", ((float32(pay_7)-float32(cashout_7))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_8"] = fmt.Sprintf("%.2f%%", ((float32(pay_8)-float32(cashout_8))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_9"] = fmt.Sprintf("%.2f%%", ((float32(pay_9)-float32(cashout_9))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_10"] = fmt.Sprintf("%.2f%%", ((float32(pay_10)-float32(cashout_10))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_11"] = fmt.Sprintf("%.2f%%", ((float32(pay_11)-float32(cashout_11))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_12"] = fmt.Sprintf("%.2f%%", ((float32(pay_12)-float32(cashout_12))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_13"] = fmt.Sprintf("%.2f%%", ((float32(pay_13)-float32(cashout_13))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_14"] = fmt.Sprintf("%.2f%%", ((float32(pay_14)-float32(cashout_14))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_21"] = fmt.Sprintf("%.2f%%", ((float32(pay_21)-float32(cashout_21))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_30"] = fmt.Sprintf("%.2f%%", ((float32(pay_30)-float32(cashout_30))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_45"] = fmt.Sprintf("%.2f%%", ((float32(pay_45)-float32(cashout_45))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_60"] = fmt.Sprintf("%.2f%%", ((float32(pay_60)-float32(cashout_60))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_90"] = fmt.Sprintf("%.2f%%", ((float32(pay_90)-float32(cashout_90))/100/float32(cost*ratearr[strdaynum]))*100)
					data["roi_180"] = fmt.Sprintf("%.2f%%", ((float32(pay_180)-float32(cashout_180))/100/float32(cost*ratearr[strdaynum]))*100)

					data["roi_realtime"] = fmt.Sprintf("%.2f%%", ((float32(allpayto)-float32(allcashto))/100/float32(cost*ratearr[strdaynum]))*100)

					terncost += float32(cost * ratearr[strdaynum])

				} else {
					data["roi_1"] = fmt.Sprintf("%.2f%%", ((float32(pay_1)-float32(cashout_1))/100/float32(cost))*100)
					data["roi_2"] = fmt.Sprintf("%.2f%%", ((float32(pay_2)-float32(cashout_2))/100/float32(cost))*100)
					data["roi_3"] = fmt.Sprintf("%.2f%%", ((float32(pay_3)-float32(cashout_3))/100/float32(cost))*100)
					data["roi_4"] = fmt.Sprintf("%.2f%%", ((float32(pay_4)-float32(cashout_4))/100/float32(cost))*100)
					data["roi_5"] = fmt.Sprintf("%.2f%%", ((float32(pay_5)-float32(cashout_5))/100/float32(cost))*100)
					data["roi_6"] = fmt.Sprintf("%.2f%%", ((float32(pay_6)-float32(cashout_6))/100/float32(cost))*100)
					data["roi_7"] = fmt.Sprintf("%.2f%%", ((float32(pay_7)-float32(cashout_7))/100/float32(cost))*100)
					data["roi_8"] = fmt.Sprintf("%.2f%%", ((float32(pay_8)-float32(cashout_8))/100/float32(cost))*100)
					data["roi_9"] = fmt.Sprintf("%.2f%%", ((float32(pay_9)-float32(cashout_9))/100/float32(cost))*100)
					data["roi_10"] = fmt.Sprintf("%.2f%%", ((float32(pay_10)-float32(cashout_10))/100/float32(cost))*100)
					data["roi_11"] = fmt.Sprintf("%.2f%%", ((float32(pay_11)-float32(cashout_11))/100/float32(cost))*100)
					data["roi_12"] = fmt.Sprintf("%.2f%%", ((float32(pay_12)-float32(cashout_12))/100/float32(cost))*100)
					data["roi_13"] = fmt.Sprintf("%.2f%%", ((float32(pay_13)-float32(cashout_13))/100/float32(cost))*100)
					data["roi_14"] = fmt.Sprintf("%.2f%%", ((float32(pay_14)-float32(cashout_14))/100/float32(cost))*100)
					data["roi_21"] = fmt.Sprintf("%.2f%%", ((float32(pay_21)-float32(cashout_21))/100/float32(cost))*100)
					data["roi_30"] = fmt.Sprintf("%.2f%%", ((float32(pay_30)-float32(cashout_30))/100/float32(cost))*100)
					data["roi_45"] = fmt.Sprintf("%.2f%%", ((float32(pay_45)-float32(cashout_45))/100/float32(cost))*100)
					data["roi_60"] = fmt.Sprintf("%.2f%%", ((float32(pay_60)-float32(cashout_60))/100/float32(cost))*100)
					data["roi_90"] = fmt.Sprintf("%.2f%%", ((float32(pay_90)-float32(cashout_90))/100/float32(cost))*100)
					data["roi_180"] = fmt.Sprintf("%.2f%%", ((float32(pay_180)-float32(cashout_180))/100/float32(cost))*100)

					data["roi_realtime"] = fmt.Sprintf("%.2f%%", ((float32(allpayto)-float32(allcashto))/100/float32(cost))*100)

					terncost += float32(cost)
				}

			}

			retl = append(retl, data)
		}

		alldata := map[string]interface{}{
			"daynum":       "汇总",
			"newuser":      allnewuser,
			"allpay":       allpay,
			"allcash":      allcash,
			"firstpay":     allfirstpay,
			"firstpaycash": fmt.Sprintf("%.2f", (float32(allfirstpay) - float32(allfirstcash))),
			"firstnum":     allfirstnum,
			"firstrate":    "",
			"pay_today":    0.00,
			"cost":         allcost,
			"roi_1":        "",
			"roi_2":        "",
			"roi_3":        "",
			"roi_4":        "",
			"roi_5":        "",
			"roi_6":        "",
			"roi_7":        "",
			"roi_8":        "",
			"roi_9":        "",
			"roi_10":       "",
			"roi_11":       "",
			"roi_12":       "",
			"roi_13":       "",
			"roi_14":       "",
			"roi_21":       "",
			"roi_30":       "",
			"roi_45":       "",
			"roi_60":       "",
			"roi_90":       "",
			"roi_180":      "",
			"roi_realtime": 0.00,

			"single": 0.00,
			"rate":   0.00,
		}
		if terncost != 0 {
			alldata["roi_realtime"] = fmt.Sprintf("%.2f%%", (float32(allpay-allcash) / float32(terncost) * 100))
		}

		if allnewuser != 0 {
			alldata["single"] = fmt.Sprintf("%.2f", (float32(allcost) / float32(allnewuser)))
			alldata["firstrate"] = fmt.Sprintf("%.2f%%", (float32(allfirstnum)/float32(allnewuser))*100)
		}

		retl = append(retl, alldata)

		var data []byte
		if len(retl) == 0 {
			data = []byte("[]")
		} else {
			data, _ = json.Marshal(retl)
		}
		task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
		retl = nil
	}(task)
}

func getadjustreport(plats, sdate, edate, gameSystem, launchAccount string, gameid uint32) map[string]float32 {

	client := &http.Client{
		Timeout: time.Second * 60,
	}

	url := "https://dash.adjust.com/control-center/reports-service/report?utc_offset=-03:00&dimensions=day&metrics=cost,installs&date_period=" + sdate + ":" + edate + "&ad_spend_mode=network"

	var app_token__in string

	tbladj := get_adjapp_table(gameid)

	if plats != "" {

		str := fmt.Sprintf("select content from %s where id = %s", tbladj, plats)

		res := db_monitor.QueryRow(str)

		res.Scan(&app_token__in)

	}
	if app_token__in == "" {

		str := fmt.Sprintf("select content from %s where id = 1000", tbladj)

		res := db_monitor.QueryRow(str)

		res.Scan(&app_token__in)
	}
	url += "&app_token__in=" + app_token__in

	if gameSystem != "all" {
		if gameSystem == "unattributed" {
			url += "&partner__in=facebook"
		} else if gameSystem == "Google Ads ACI" {
			url += "&partner__in=adwords"
		} else if gameSystem == "Kwai for Business" {
			url += "&partner__in=kuaishou_global"
		}

	}

	if launchAccount != "" {
		url += "&campaign_network__in=" + launchAccount
	}

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", "Bearer Mrzx6vznfuPSSgce_9Ru")

	// unibase.HttpClient.HttpRequestGet()

	cost := make(map[string]float32)

	resp, err := client.Do(req)

	if err != nil {

		return cost
	}

	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err == nil {

		result1 := gjson.Get(string(body), "rows.#.cost")

		result := gjson.Get(string(body), "rows.#.day")
		for key, day := range result.Array() {

			costkey := strings.Replace(day.String(), "-", "", -1)

			result1_arr := result1.Array()

			cost[string(costkey)] = float32(result1_arr[key].Float())
		}

	}

	return cost
}

// 金钱排名
func HandleUserMoneyRank(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("starttime")), 10, 64)
	etime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("endtime")), 10, 64)
	coinid := uint32(unibase.Atoi(task.R.FormValue("coinid"), 0))
	var where string
	//如果输入的货币id为0，默认为元宝
	if coinid == 0 {
		coinid = 8
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

	if zoneid != 0 {
		where += fmt.Sprintf(" AND zoneid=%d ", zoneid)
	}

	if coinid != 0 {
		where += fmt.Sprintf(" AND coinid=%d ", coinid)
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
	retl := make([]map[string]interface{}, 0)
	for i, j := time.Unix(int64(stime), 0), 0; TimeToYmd(&i, j) < edate; j++ {
		tbname := fmt.Sprintf("user_economic_%d_%d", gameid, TimeToYmd(&i, j))
		if !check_table_exists(tbname) {
			continue
		}
		//str := fmt.Sprintf("select userid,curcoin,coinid,min(unix_timestamp(now())-UNIX_TIMESTAMP(created_at)) from %s, %s where %s.userid=%s.userid %s group by userid", tbname, tbname, tbuser, tbname, tbuser, where)
		//str := fmt.Sprintf("select enconmic.userid,enconmic.curcoin,enconmic.coinid,%s.zoneid,%s.username from %s enconmic, %s  where unix_timestamp(created_at) = (select max(unix_timestamp(created_at)) from %s where %s.userid = userid) %s  order by curcoin desc limit 200", tbuser, tbuser, tbname, tbuser, tbname, tbuser, where)
		//str := fmt.Sprintf("select economic.userid, economic.curcoin, economic.coinid, economic.zoneid, economic.zoneid from %s economic where economic.coinid= %d and unix_timestamp(economic.created_at) = (select max(unix_timestamp(created_at)) from %s where economic.userid=userid)  %s limit 200",tbname,coinid,tbname,where)
		str := fmt.Sprintf("select userid,max(unix_timestamp(created_at)) from %s where curcoin>1000 %s group by userid order by curcoin desc limit 50", tbname, where)
		rows, err := db_monitor.Query(str)
		task.Debug("sql:%s", str)
		if err != nil {
			task.Error("HandleUserMoneyRank error:%s sql:%s", err.Error(), str)
			continue
		}

		var rank uint32 = 0

		for rows.Next() {
			var userid, time_stamp uint32
			if err = rows.Scan(&userid, &time_stamp); err != nil {
				task.Error("HandleUserMoneyRank error:%s", err.Error())
				continue
			}

			//根据筛选出来的值再次查询
			str := fmt.Sprintf("select userid,curcoin,coinid,zoneid from %s where userid=%d and unix_timestamp(created_at)=%d %s", tbname, userid, time_stamp, where)

			rows2, err := db_monitor.Query(str)
			task.Debug("sql:%s", str)
			if err != nil {
				task.Error("HandleUserMoneyRank error:%s sql:%s", err.Error(), str)
				continue
			}

			for rows2.Next() {
				var userid, coinid, curcoin, zoneid uint32
				var username string

				tbname := get_user_data_table(gameid)

				if err = rows2.Scan(&userid, &curcoin, &coinid, &zoneid); err != nil {
					task.Error("HandleUserMoneyRank error:%s", err.Error())
					continue
				}
				str := fmt.Sprintf("select username from %s where userid = %d ", tbname, userid)
				rows3, err := db_monitor.Query(str)
				if err != nil {
					task.Error("HandleUserMoneyRank error:%s sql:%s", err.Error(), str)
					continue
				}
				for rows3.Next() {
					if err = rows3.Scan(&username); err != nil {
						task.Error("HandleUserMoneyRank error:%s", err.Error())
						continue
					}
				}
				rows3.Close()

				rank += 1
				data := map[string]interface{}{
					"rank":     rank,
					"zoneid":   zoneid,
					"userid":   userid,
					"username": username,
					"coinid":   coinid,
					"curcoin":  curcoin,
				}
				// task.Debug("new userid=%d,curcoin=%d",userid,curcoin)
				//要查找一下是否有重复的
				var bFind bool = false
				del_list := []int{}
				for k, v := range retl {
					if v["userid"] == userid {
						task.Debug("重复删除:%d", userid)
						del_list = append(del_list, k)
						retl[k] = data
						bFind = true
						break
					}
				}
				if bFind == false {
					retl = append(retl, data)
				}

			}
			rows2.Close()
		}
		rows.Close()

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
func HandleUserChangeInfo(task *unibase.ChanHttpTask) {
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := strings.TrimSpace(task.R.FormValue("starttime"))

	etime := strings.TrimSpace(task.R.FormValue("endtime"))

	sdate := fmt.Sprintf("%s 00:00:00", stime)
	edate := fmt.Sprintf("%s 23:59:59", etime)
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	userid := uint32(unibase.Atoi(task.R.FormValue("userid"), 0))

	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))

	tblname := get_change_registersrc_table(gameid)
	tbldata := get_user_data_table(gameid)

	where := fmt.Sprintf(" a.zoneid != 0 AND ")

	if plats != "" {
		platlist := strings.Split(plats, ",")
		if len(platlist) == 1 {
			where += fmt.Sprintf(" a.after_platid=%s AND ", platlist[0])

		} else if len(platlist) > 1 {
			//如果不是1，需要根据权限列获得相应的渠道id
			if len(platlist) != 1 {
				ptlist := getAccountPlatList(task)
				if len(ptlist) > 0 {
					platlist = ptlist
				}
			}
			where += fmt.Sprintf(" a.after_platid in (%s) AND ", strings.Join(platlist, ","))
		}
	} else {

		//如果为空，也要检测下是否有权限设置 ,没有设置就是最大权限
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			where += fmt.Sprintf(" a.after_platid in (%s) AND ", strings.Join(ptlist, ","))
		}
	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" a.after_launchid = '%s' AND ", gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf(" a.after_ad_account = '%s' AND ", launchAccount)
	}
	if userid != 0 {
		where += fmt.Sprintf(" a.userid = %d AND ", userid)
	}

	where += fmt.Sprintf(" a.created_at >= '%s' And a.created_at <= '%s'  ", sdate, edate)

	str := fmt.Sprintf(`select a.userid , a.befor_platid ,a.befor_launchid, a.befor_ad_account , a.after_platid, a.after_launchid , a.after_ad_account , a.created_at , FROM_UNIXTIME(b.reg_time) from %s as a left join %s as b on a.userid = b.userid where %s  `, tblname, tbldata, where)
	rows, err := db_monitor.Query(str)
	task.Debug("HandleUserChangeInfo sql:%s", str)
	if err != nil {
		task.Error("HandleUserChangeInfo error:%s sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var userid, befor_platid, after_platid uint32
		var befor_launchid, befor_ad_account, after_launchid, after_ad_account, created_at, reg_time string

		if err = rows.Scan(&userid, &befor_platid, &befor_launchid, &befor_ad_account, &after_platid, &after_launchid, &after_ad_account, &created_at, &reg_time); err != nil {
			task.Error("HandleUserChangeInfo error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"userid":           userid,
			"befor_platid":     befor_platid,
			"befor_launchid":   befor_launchid,
			"befor_ad_account": befor_ad_account,
			"after_launchid":   after_launchid,
			"after_platid":     after_platid,
			"after_ad_account": after_ad_account,
			"created_at":       created_at,
			"reg_time":         reg_time,
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
func HandleRecoveryExaminationList(task *unibase.ChanHttpTask) {
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))

	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))

	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))

	ad_account_id := strings.TrimSpace(task.R.FormValue("ad_account_id"))
	agent_id := strings.TrimSpace(task.R.FormValue("agent_id"))

	tblname := get_user_daily_plat_table(gameid)
	tblkeys := get_launch_keys_table(gameid)

	where := fmt.Sprintf(` a.platid < 100 AND `)

	if plats != "" {
		platlist := strings.Split(plats, ",")
		if len(platlist) == 1 {
			where += fmt.Sprintf(" a.platid=%s AND ", platlist[0])
		} else if len(platlist) > 1 {
			//如果不是1，需要根据权限列获得相应的渠道id
			if len(platlist) != 1 {
				ptlist := getAccountPlatList(task)
				if len(ptlist) > 0 {
					platlist = ptlist
				}
			}
			where += fmt.Sprintf(" a.platid in (%s) AND ", strings.Join(platlist, ","))
		}
	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" a.launchid = '%s' AND ", gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf(" a.ad_account = '%s' AND ", launchAccount)
	}
	if ad_account_id != "" {
		where += fmt.Sprintf(" b.ad_account_id = '%s' AND ", ad_account_id)
	}
	if agent_id != "" {
		where += fmt.Sprintf(" b.agent_id = '%s' AND ", agent_id)
	}
	where += fmt.Sprintf(" (a.daynum>=%d and a.daynum<=%d)", stime, etime)

	str := fmt.Sprintf(`select a.daynum, a.launchid , a.ad_account , a.cost , IFNULL(b.ad_account_id ,''), IFNULL(b.agent_id,'') from %s as a left join %s as b on a.ad_account = b.keywords where %s`, tblname, tblkeys, where)

	rows, err := db_monitor.Query(str)
	//task.Debug("HandleRecoveryExaminationList sql:%s", str)
	if err != nil {
		task.Error("HandleRecoveryExaminationList error:%s sql:%s", err.Error(), str)
		return
	}
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {

		var daynum, launchid, ad_account, ad_account_id, agent_id string
		var cost float32

		if err = rows.Scan(&daynum, &launchid, &ad_account, &cost, &ad_account_id, &agent_id); err != nil {
			task.Error("HandleRecoveryExaminationList error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"daynum":        daynum,
			"launchid":      launchid,
			"ad_account":    ad_account,
			"cost":          cost,
			"ad_account_id": ad_account_id,
			"agent_id":      agent_id,
		}

		retl = append(retl, data)

	}
	defer rows.Close()
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retl = nil
}

func HandleRecoveryExaminationListApp(task *unibase.ChanHttpTask) {
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))

	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))

	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))

	agent_id := strings.TrimSpace(task.R.FormValue("agent_id"))
	ad_account_id := strings.TrimSpace(task.R.FormValue("ad_account_id"))

	tblname := get_app_cost_today_table(gameid)

	where := ""

	if plats != "" {
		platlist := strings.Split(plats, ",")
		if len(platlist) == 1 {
			where += fmt.Sprintf(" appid=%s AND ", platlist[0])
		} else if len(platlist) > 1 {
			//如果不是1，需要根据权限列获得相应的渠道id
			if len(platlist) != 1 {
				ptlist := getAccountPlatList(task)
				if len(ptlist) > 0 {
					platlist = ptlist
				}
			}
			where += fmt.Sprintf(" appid in (%s) AND ", strings.Join(platlist, ","))
		}
	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" launchid = '%s' AND ", gameSystem)
	}
	if ad_account_id != "" {
		where += fmt.Sprintf(" ad_account_id = '%s' AND ", ad_account_id)
	}
	if agent_id != "" {
		where += fmt.Sprintf(" agent_id = '%s' AND ", agent_id)
	}
	where += fmt.Sprintf(" (daynum>=%d and daynum<=%d)", stime, etime)

	str := fmt.Sprintf(`select id , appid , launchid , agent_id , daynum , ad_account_id ,cost ,currency ,rate from %s where %s`, tblname, where)

	rows, err := db_monitor.Query(str)
	//task.Debug("HandleRecoveryExaminationList sql:%s", str)
	if err != nil {
		task.Error("HandleRecoveryExaminationList error:%s sql:%s", err.Error(), str)
		return
	}
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {

		var launchid, agent_id, ad_account_id string
		var id, appid, daynum int32
		var cost, currency, rate float32

		if err = rows.Scan(&id, &appid, &launchid, &agent_id, &daynum, &ad_account_id, &cost, &currency, &rate); err != nil {
			task.Error("HandleRecoveryExaminationList error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"id":            id,
			"appid":         appid,
			"launchid":      launchid,
			"daynum":        daynum,
			"agent_id":      agent_id,
			"cost":          cost,
			"currency":      currency,
			"rate":          rate,
			"ad_account_id": ad_account_id,
		}

		retl = append(retl, data)

	}
	defer rows.Close()
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retl = nil
}
func HandleRecoveryExamination(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	keywords := strings.TrimSpace(task.R.FormValue("keywords"))
	daynum := uint32(unibase.Atoi(task.R.FormValue("daynum"), 0))
	cost := strings.TrimSpace(task.R.FormValue("cost"))

	tblnamep := get_user_daily_plat_table(gameid)
	tblnamer := get_user_daily_roi_table(gameid)

	str1 := fmt.Sprintf(`update %s set cost= %s , is_examination = 1 where ad_account = "%s" and  daynum = %d `, tblnamep, cost, keywords, daynum)
	str2 := fmt.Sprintf(`update %s set cost= %s , is_examination = 1 where ad_account = "%s" and  daynum = %d `, tblnamer, cost, keywords, daynum)

	_, err := db_monitor.Exec(str1)
	fmt.Println("str1:", str1)
	if err != nil {
		task.Error("HandleRecoveryExamination error:%s sql:%s", err.Error(), str1)
	}
	_, err = db_monitor.Exec(str2)
	if err != nil {
		task.Error("HandleRecoveryExamination error:%s sql:%s", err.Error(), str2)
	}

	task.SendBinary([]byte(`{"retcode":0,"retdesc":"action success"}`))
}

func HandleRecoveryExaminationApp(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	actiontype := uint32(unibase.Atoi(task.R.FormValue("actiontype"), 0))
	cost := strings.TrimSpace(task.R.FormValue("cost"))
	agent := strings.TrimSpace(task.R.FormValue("agent"))
	tblname := get_app_cost_today_table(gameid)

	var code = 1

	if actiontype == 1 {
		id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
		str1 := fmt.Sprintf(`update %s set cost= %s * rate , currency = %s ,  agent_id = '%s' where id = %d  `, tblname, cost, cost, agent, id)
		_, err := db_monitor.Exec(str1)
		if err != nil {
			code = 2
			task.Error("HandleRecoveryExamination error:%s sql:%s", err.Error(), str1)
		}
	} else if actiontype == 2 {
		daynum := uint32(unibase.Atoi(task.R.FormValue("daynum"), 0))
		platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
		launch := strings.TrimSpace(task.R.FormValue("launch"))
		account := strings.TrimSpace(task.R.FormValue("account"))
		currency := uint32(unibase.Atoi(task.R.FormValue("currency"), 1))

		var rate = float64(1)

		a, _ := strconv.ParseFloat(cost, 64)
		var costsource = a
		fmt.Println("currency:", currency)
		if currency != 1 {
			fmt.Println("currency计算")
			tblrate := get_exchange_rate_table(gameid)

			strrate := fmt.Sprintf(`select exchange_rate from %s where currency = ? and DATE_FORMAT(time, "%%Y%%m%%d") = ? `, tblrate)

			res := db_monitor.QueryRow(strrate, currency, daynum)

			var rater float64

			res.Scan(&rater)

			fmt.Println("rater:", rater)

			if rater != 0 {
				rate = rater

				costsource = a * float64(rater)

			}

		}

		str := fmt.Sprintf("insert into %s (daynum, appid, launchid, agent_id, cost,ad_account_id,currency,rate) values(?,?,?,?,?,?,?,?)", tblname)
		if _, err := db_monitor.Exec(str, daynum, platid, launch, agent, costsource, account, cost, rate); err != nil {
			code = 2
			task.Error("HandleRecoveryExamination error:%s sql:%s", err.Error(), str)
		}

	}
	if code == 2 {
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"action failed"}`))
	} else {
		task.SendBinary([]byte(`{"retcode":0,"retdesc":"action success"}`))
	}
}
func getadjustreportall(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	client := &http.Client{
		Timeout: time.Second * 600,
	}
	url := fmt.Sprintf("https://dash.adjust.com/control-center/reports-service/report?utc_offset=-03:00&dimensions=day,app,campaign_network&metrics=cost,installs&date_period=2023-03-21:2023-05-07&ad_spend_mode=network")

	var app_token__in string

	tbladj := get_adjapp_table(gameid)

	str := fmt.Sprintf("select content from %s where id = 1000", tbladj)

	res := db_monitor.QueryRow(str)

	res.Scan(&app_token__in)

	// url += "&app_token__in=" + app_token__in

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", "Bearer Mrzx6vznfuPSSgce_9Ru")

	resp, err := client.Do(req)

	if err != nil {
		task.Error("getadjustreport %s err:", err.Error())
		return
	}
	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err == nil {
		tblnamep := get_user_daily_plat_table(gameid)
		tblnamer := get_user_daily_roi_table(gameid)
		result := gjson.Get(string(body), "rows")

		result.ForEach(func(key, value gjson.Result) bool {
			campaign_network := gjson.Get(value.String(), "campaign_network")
			cost := gjson.Get(value.String(), "cost")
			day := gjson.Get(value.String(), "day")

			sdate1 := strings.Replace(day.String(), "-", "", -1)

			if campaign_network.String() != "unknown" && cost.Float() > 0 {

				str1 := fmt.Sprintf(`update %s  set cost= ? where ad_account = "%s" and  daynum = ? and is_examination != 1`, tblnamep, campaign_network.String())
				str2 := fmt.Sprintf(`update %s  set cost= ? where ad_account = "%s" and  daynum = ? and is_examination != 1`, tblnamer, campaign_network.String())

				_, err = db_monitor.Exec(str1, cost.Float(), sdate1)
				if err != nil {
					task.Error("getadjustreport %s err:%s,%s", campaign_network.String(), str, err.Error())
				}
				_, err = db_monitor.Exec(str2, cost.Float(), sdate1)
				if err != nil {
					task.Error("getadjustreport %s err:%s,%s", campaign_network.String(), str, err.Error())
				}
			}

			return true
		})

	}
}
