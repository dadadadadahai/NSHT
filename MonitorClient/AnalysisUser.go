package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

func get_user_daily_table(gameid uint32) string {
	return fmt.Sprintf("user_daily_%d", gameid)
}

func get_user_daily_plat_table(gameid uint32) string {
	return fmt.Sprintf("user_daily_plat_%d", gameid)
}
func get_user_daily_roi_table(gameid uint32) string {
	return fmt.Sprintf("user_daily_roi_%d", gameid)
}

func get_user_account_table(gameid uint32) string {
	return fmt.Sprintf("user_account_%d", gameid)
}

func get_mahjong_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_mahjong_%d_%d", gameid, int(ymd/100))
}

// 每日数据
type UserDailyData struct {
	Gameid        uint32
	Zoneid        uint32
	Platid        uint32
	Daynum        uint32  //日期
	Newuser       uint32  //新增玩家
	Dau           uint32  //活跃玩家数
	Pcu           uint32  //最高在线
	Paynum        uint32  //充值玩家数
	Paytoday      uint32  //充值金额
	Arppu         float32 //付费金额/付费玩家数
	Arpu          float32 //付费金额/活跃玩家数
	Payrate       float32 //付费玩家数/活跃玩家数
	Retainedday2  float32 //次日流存数
	Retainedday3  float32 //三日流存数
	Retainedday7  float32 //七日流存数
	Retainedday30 float32 //30日
	Recordpoint   int64   //兑入兑出积分
	Recordtime    uint32  //当前分钟
	Login3days    uint32  //连续登陆3天的玩家数
	Avglogintimes float32 //平均登陆次数
	Avgonlinemin  float32 //平均在线分钟数
	Refluxnum     uint32  //回流玩家数(3天一周期)
}

func CheckUserDailyData(gameid, zoneid, stime, etime uint32, platlist []string, gameSystem string, launchAccount string) (retl []map[string]interface{}) {
	where := fmt.Sprintf(" zoneid !=0 AND ")
	if stime == etime {
		day15 := uint32(unitime.Time.YearMonthDay(-15))
		where += fmt.Sprintf(" daynum>=%d ", day15)
	} else if stime == 0 && etime != 0 {
		where += fmt.Sprintf(" daynum<=%d ", etime)
	} else if etime == 0 && stime != 0 {
		where += fmt.Sprintf(" daynum>=%d ", stime)
	} else {
		where += fmt.Sprintf(" (daynum>=%d AND daynum <=%d) ", stime, etime)
	}
	if len(platlist) == 1 {
		where += fmt.Sprintf(" AND platid=%s ", platlist[0])
	} else if len(platlist) > 1 {
		where += fmt.Sprintf(" AND platid in (%s) ", strings.Join(platlist, ","))
	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" AND launchid='%s' ", gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf(" AND ad_account='%s' ", launchAccount)
	}
	tblname := get_user_daily_plat_table(gameid)
	str := fmt.Sprintf(`SELECT daynum, sum(day_1), sum(dau), MAX(pcu),sum(pay_today_num), sum(pay_today)/100,sum(day_2),sum(day_3),sum(day_4),sum(day_5),sum(day_6),sum(day_7),sum(day_8),sum(day_9),sum(day_10),sum(day_11),sum(day_12),sum(day_13),sum(day_14),sum(day_30),sum(day_60),sum(day_90),sum(day_180),sum(ol3d), avg(avglogintimes), avg(avgonlinemin), sum(refluxnum), sum(pay_first_num),sum(pay_cashout)/100 ,sum(pay_cashout_num) from %s WHERE %s group by daynum  order by daynum desc`, tblname, where)
	rows, err := db_monitor.Query(str)
	logging.Debug("CheckUserDailyData sql:%s", str)
	if err != nil {
		logging.Error("CheckUserDailyData error:%s, sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var daynum, day_1, dau, pcu, paynum, day_2, day_3, day_4, day_5, day_6, day_7, day_8, day_9, day_10, day_11, day_12, day_13, day_14, day_30, day_60, day_90, day_180, ol3d, refluxnum, pay_first, cashoutnum uint32
		var paytoday, cashouttoady float32
		var avglogin, avgonline float32
		if err := rows.Scan(&daynum, &day_1, &dau, &pcu, &paynum, &paytoday, &day_2, &day_3, &day_4, &day_5, &day_6, &day_7, &day_8, &day_9, &day_10, &day_11, &day_12, &day_13, &day_14, &day_30, &day_60, &day_90, &day_180, &ol3d, &avglogin, &avgonline, &refluxnum, &pay_first, &cashouttoady, &cashoutnum); err != nil {
			logging.Error("CheckUserDailyData err:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"Gameid":         gameid,
			"Zoneid":         zoneid,
			"Daynum":         daynum,
			"Newuser":        day_1,
			"Dau":            dau,
			"Pcu":            pcu,
			"Paytoday":       paytoday,
			"Cashouttoady":   cashouttoady,
			"Cashoutnum":     cashoutnum,
			"Paynum":         paynum,
			"Retainedday2":   "0.00",
			"Retainedday3":   "0.00",
			"Retainedday4":   "0.00",
			"Retainedday5":   "0.00",
			"Retainedday6":   "0.00",
			"Retainedday7":   "0.00",
			"Retainedday8":   "0.00",
			"Retainedday9":   "0.00",
			"Retainedday10":  "0.00",
			"Retainedday11":  "0.00",
			"Retainedday12":  "0.00",
			"Retainedday13":  "0.00",
			"Retainedday14":  "0.00",
			"Retainedday30":  "0.00",
			"Retainedday60":  "0.00",
			"Retainedday90":  "0.00",
			"Retainedday180": "0.00",
			"Arpu":           "0.00",
			"Arppu":          "0.00",
			"Payrate":        "0.00",
			"Login3days":     ol3d,
			"Avglogintimes":  fmt.Sprintf("%.2f", avglogin),
			"Avgonlinemin":   fmt.Sprintf("%.2f", avgonline),
			"Refluxnum":      refluxnum,
			"Payfirst":       pay_first,
			"Cashout":        "0.00",
			"CashoutRatio":   "0.00",
			"CashoutScale":   "0.00",
		}
		if day_1 != 0 {
			data["Retainedday2"] = fmt.Sprintf("%.2f", float32(100*day_2)/float32(day_1))
			data["Retainedday3"] = fmt.Sprintf("%.2f", float32(100*day_3)/float32(day_1))
			data["Retainedday4"] = fmt.Sprintf("%.2f", float32(100*day_4)/float32(day_1))
			data["Retainedday5"] = fmt.Sprintf("%.2f", float32(100*day_5)/float32(day_1))
			data["Retainedday6"] = fmt.Sprintf("%.2f", float32(100*day_6)/float32(day_1))
			data["Retainedday7"] = fmt.Sprintf("%.2f", float32(100*day_7)/float32(day_1))
			data["Retainedday8"] = fmt.Sprintf("%.2f", float32(100*day_8)/float32(day_1))
			data["Retainedday9"] = fmt.Sprintf("%.2f", float32(100*day_9)/float32(day_1))
			data["Retainedday10"] = fmt.Sprintf("%.2f", float32(100*day_10)/float32(day_1))
			data["Retainedday11"] = fmt.Sprintf("%.2f", float32(100*day_11)/float32(day_1))
			data["Retainedday12"] = fmt.Sprintf("%.2f", float32(100*day_12)/float32(day_1))
			data["Retainedday13"] = fmt.Sprintf("%.2f", float32(100*day_13)/float32(day_1))
			data["Retainedday14"] = fmt.Sprintf("%.2f", float32(100*day_14)/float32(day_1))
			data["Retainedday30"] = fmt.Sprintf("%.2f", float32(100*day_30)/float32(day_1))
			data["Retainedday60"] = fmt.Sprintf("%.2f", float32(100*day_60)/float32(day_1))
			data["Retainedday90"] = fmt.Sprintf("%.2f", float32(100*day_90)/float32(day_1))
			data["Retainedday180"] = fmt.Sprintf("%.2f", float32(100*day_180)/float32(day_1))
		}
		if dau != 0 {
			data["Arpu"] = fmt.Sprintf("%.2f", float32(paytoday)/float32(dau))
			data["Payrate"] = fmt.Sprintf("%.2f", float32(100*paynum)/float32(dau))
			data["CashoutRatio"] = fmt.Sprintf("%.2f", float32(100*cashoutnum)/float32(dau))
		}
		if paynum != 0 {
			data["Arppu"] = fmt.Sprintf("%.2f", float32(paytoday)/float32(paynum))
		}
		if paytoday != 0 {
			data["CashoutScale"] = fmt.Sprintf("%.2f", float32(cashouttoady)/float32(paytoday))
		}
		retl = append(retl, data)
	}
	return
}
func CheckUserDailyfirstData(gameid, zoneid, stime, etime uint32, platlist []string, gameSystem string, launchAccount string) (retl []map[string]interface{}) {
	where := fmt.Sprintf(" zoneid !=0 AND ")
	if stime == etime {
		day15 := uint32(unitime.Time.YearMonthDay(-15))
		where += fmt.Sprintf(" daynum>=%d ", day15)
	} else if stime == 0 && etime != 0 {
		where += fmt.Sprintf(" daynum<=%d ", etime)
	} else if etime == 0 && stime != 0 {
		where += fmt.Sprintf(" daynum>=%d ", stime)
	} else {
		where += fmt.Sprintf(" (daynum>=%d AND daynum <=%d) ", stime, etime)
	}
	if len(platlist) == 1 {
		where += fmt.Sprintf(" AND platid=%s ", platlist[0])
	} else if len(platlist) > 1 {
		where += fmt.Sprintf(" AND platid in (%s) ", strings.Join(platlist, ","))
	}
	if gameSystem != "all" {
		where += fmt.Sprintf(" AND launchid='%s' ", gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf(" AND ad_account='%s' ", launchAccount)
	}
	tblname := get_user_daily_roi_table(gameid)
	str := fmt.Sprintf(`SELECT daynum, sum(pay_first_num), sum(day_2),sum(day_3),sum(day_4),sum(day_5),sum(day_6),sum(day_7),sum(day_8),sum(day_9),sum(day_10),sum(day_11),sum(day_12),sum(day_13),sum(day_14),sum(day_30),sum(day_60),sum(day_90),sum(day_180) from %s WHERE %s group by daynum  order by daynum desc`, tblname, where)
	rows, err := db_monitor.Query(str)
	logging.Debug("CheckUserDailyData sql:%s", str)
	if err != nil {
		logging.Error("CheckUserDailyData error:%s, sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var daynum, day_1, day_2, day_3, day_4, day_5, day_6, day_7, day_8, day_9, day_10, day_11, day_12, day_13, day_14, day_30, day_60, day_90, day_180 uint32

		if err := rows.Scan(&daynum, &day_1, &day_2, &day_3, &day_4, &day_5, &day_6, &day_7, &day_8, &day_9, &day_10, &day_11, &day_12, &day_13, &day_14, &day_30, &day_60, &day_90, &day_180); err != nil {
			logging.Error("CheckUserDailyData err:%s", err.Error())
			continue
		}
		data := map[string]interface{}{

			"Daynum":         daynum,
			"Newuser":        day_1,
			"Retainedday2":   "0.00",
			"Retainedday3":   "0.00",
			"Retainedday4":   "0.00",
			"Retainedday5":   "0.00",
			"Retainedday6":   "0.00",
			"Retainedday7":   "0.00",
			"Retainedday8":   "0.00",
			"Retainedday9":   "0.00",
			"Retainedday10":  "0.00",
			"Retainedday11":  "0.00",
			"Retainedday12":  "0.00",
			"Retainedday13":  "0.00",
			"Retainedday14":  "0.00",
			"Retainedday30":  "0.00",
			"Retainedday60":  "0.00",
			"Retainedday90":  "0.00",
			"Retainedday180": "0.00",
		}
		if day_1 != 0 {
			data["Retainedday2"] = fmt.Sprintf("%.2f", float32(100*day_2)/float32(day_1))
			data["Retainedday3"] = fmt.Sprintf("%.2f", float32(100*day_3)/float32(day_1))
			data["Retainedday4"] = fmt.Sprintf("%.2f", float32(100*day_4)/float32(day_1))
			data["Retainedday5"] = fmt.Sprintf("%.2f", float32(100*day_5)/float32(day_1))
			data["Retainedday6"] = fmt.Sprintf("%.2f", float32(100*day_6)/float32(day_1))
			data["Retainedday7"] = fmt.Sprintf("%.2f", float32(100*day_7)/float32(day_1))
			data["Retainedday8"] = fmt.Sprintf("%.2f", float32(100*day_8)/float32(day_1))
			data["Retainedday9"] = fmt.Sprintf("%.2f", float32(100*day_9)/float32(day_1))
			data["Retainedday10"] = fmt.Sprintf("%.2f", float32(100*day_10)/float32(day_1))
			data["Retainedday11"] = fmt.Sprintf("%.2f", float32(100*day_11)/float32(day_1))
			data["Retainedday12"] = fmt.Sprintf("%.2f", float32(100*day_12)/float32(day_1))
			data["Retainedday13"] = fmt.Sprintf("%.2f", float32(100*day_13)/float32(day_1))
			data["Retainedday14"] = fmt.Sprintf("%.2f", float32(100*day_14)/float32(day_1))
			data["Retainedday30"] = fmt.Sprintf("%.2f", float32(100*day_30)/float32(day_1))
			data["Retainedday60"] = fmt.Sprintf("%.2f", float32(100*day_60)/float32(day_1))
			data["Retainedday90"] = fmt.Sprintf("%.2f", float32(100*day_90)/float32(day_1))
			data["Retainedday180"] = fmt.Sprintf("%.2f", float32(100*day_180)/float32(day_1))
		}

		retl = append(retl, data)
	}
	return
}
func CheckUserMonthData(gameid, zoneid, stime, etime uint32, platlist []string) (retl []map[string]interface{}) {
	where := fmt.Sprintf(" zoneid=%d AND ", zoneid)
	where += fmt.Sprintf(" (daynum>=%d AND daynum <%d) ", stime, etime)
	if len(platlist) == 1 {
		where += fmt.Sprintf(" AND platid=%s ", platlist[0])
	} else if len(platlist) > 1 {
		where += fmt.Sprintf(" AND platid in (%s) ", strings.Join(platlist, ","))
	}
	tblname := get_user_daily_plat_table(gameid)
	str := fmt.Sprintf("SELECT platid, sum(day_1), sum(dau), sum(pay_today_num), cast(sum(pay_today/100) as unsigned), sum(day_2), sum(day_3), sum(day_7), sum(day_30) from %s WHERE %s group by platid order by platid desc", tblname, where)
	rows, err := db_monitor.Query(str)
	logging.Debug("CheckUserMonthData sql:%s", str)
	if err != nil {
		logging.Error("CheckUserMonthData error:%s", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var platid, day_1, dau, paynum, paytoday, day_2, day_3, day_7, day_30 uint32
		if err := rows.Scan(&platid, &day_1, &dau, &paynum, &paytoday, &day_2, &day_3, &day_7, &day_30); err != nil {
			logging.Error("CheckUserMonthData err:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"Gameid":        gameid,
			"Zoneid":        zoneid,
			"Platid":        platid,
			"Newuser":       day_1,
			"Dau":           dau,
			"Paytoday":      paytoday,
			"Paynum":        paynum,
			"Retainedday2":  "0.00",
			"Retainedday3":  "0.00",
			"Retainedday7":  "0.00",
			"Retainedday30": "0.00",
			"Arpu":          "0.00",
			"Arppu":         "0.00",
			"Payrate":       "0.00",
		}
		if day_1 != 0 {
			data["Retainedday2"] = fmt.Sprintf("%.2f", float32(100*day_2)/float32(day_1))
			data["Retainedday3"] = fmt.Sprintf("%.2f", float32(100*day_3)/float32(day_1))
			data["Retainedday7"] = fmt.Sprintf("%.2f", float32(100*day_7)/float32(day_1))
			data["Retainedday30"] = fmt.Sprintf("%.2f", float32(100*day_30)/float32(day_1))
		}
		if dau != 0 {
			data["Arpu"] = fmt.Sprintf("%.2f", float32(paytoday)/float32(dau))
			data["Payrate"] = fmt.Sprintf("%.2f", float32(100*paynum)/float32(dau))
		}
		if paynum != 0 {
			data["Arppu"] = fmt.Sprintf("%.2f", float32(paytoday)/float32(paynum))
		}
		retl = append(retl, data)
	}
	return
}

func CheckUserFormData(gameid, zoneid, stime, etime uint32, platlist []string) (retl []*UserDailyData) {
	tblname := get_user_daily_plat_table(gameid)
	where := fmt.Sprintf(" zoneid=%d AND ", zoneid)
	if len(platlist) == 1 {
		where += fmt.Sprintf(" platid=%s AND ", platlist[0])
	} else if len(platlist) > 1 {
		where += fmt.Sprintf(" platid in (%s) AND ", strings.Join(platlist, ","))
	}
	where += fmt.Sprintf(" (daynum>=%d AND daynum <=%d) ", stime, etime)
	str := fmt.Sprintf("SELECT daynum, platid, zoneid, day_1, dau, pay_today_num, cast(pay_today/100 as unsigned), day_2, day_3, day_7, day_30, ol3d, avglogintimes, avgonlinemin, refluxnum from %s WHERE %s order by daynum desc", tblname, where)
	logging.Debug("CheckUserFormData sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		logging.Error("CheckUserDailyPlatData error:%s", err.Error())
		return
	}
	defer rows.Close()
	retm := make(map[uint32]map[uint32]map[uint32]map[string]interface{})
	for rows.Next() {
		var tmpzoneid, tmpplatid, daynum, day_1, dau, paynum, paytoday, day_2, day_3, day_7, day_30, ol3d, refluxnum uint32
		var avglogin, avgonline float32
		if err := rows.Scan(&daynum, &tmpplatid, &tmpzoneid, &day_1, &dau, &paynum, &paytoday, &day_2, &day_3, &day_7, &day_30, &ol3d, &avglogin, &avgonline, &refluxnum); err != nil {
			logging.Error("CheckUserDailyPlatData err:%s", err.Error())
			continue
		}
		zonedaydata := map[string]interface{}{
			"zoneid":    tmpzoneid,
			"platid":    0,
			"daynum":    daynum,
			"day_1":     day_1,
			"dau":       dau,
			"paytoday":  paytoday,
			"paynum":    paynum,
			"day_2":     day_2,
			"day_3":     day_3,
			"day_7":     day_7,
			"day_30":    day_30,
			"ol3d":      ol3d,
			"avglogin":  avglogin,
			"avgonline": avgonline,
			"refluxnum": refluxnum,
		}
		logging.Debug("zonedaydata:%v", zonedaydata)
		if _, ok := retm[daynum]; !ok {
			retm[daynum] = make(map[uint32]map[uint32]map[string]interface{})
		}
		if _, ok := retm[daynum][tmpplatid]; !ok {
			retm[daynum][tmpplatid] = make(map[uint32]map[string]interface{})
		}
		retm[daynum][tmpplatid][tmpzoneid] = zonedaydata
	}
	if zoneid == 0 {
		for daynum, platdatamap := range retm {
			for tmpplatid, zonedatamap := range platdatamap {
				data := &UserDailyData{}
				data.Gameid = gameid
				data.Zoneid = 0
				data.Platid = tmpplatid
				data.Daynum = daynum
				var day_2, day_3, day_7, day_30 uint32
				var avglogin, avgonline float32
				for _, zonedata := range zonedatamap {
					data.Newuser += zonedata["day_1"].(uint32)
					data.Dau += zonedata["dau"].(uint32)
					data.Paynum += zonedata["paynum"].(uint32)
					data.Paytoday += zonedata["paytoday"].(uint32)
					data.Login3days += zonedata["ol3d"].(uint32)
					data.Refluxnum += zonedata["refluxnum"].(uint32)
					day_2 += zonedata["day_2"].(uint32)
					day_3 += zonedata["day_3"].(uint32)
					day_7 += zonedata["day_7"].(uint32)
					day_30 += zonedata["day_30"].(uint32)
					avglogin += zonedata["avglogin"].(float32)
					avgonline += zonedata["avgonline"].(float32)
				}
				data.Avglogintimes = avglogin / float32(len(zonedatamap))
				data.Avgonlinemin = avgonline / float32(len(zonedatamap))
				if data.Newuser != 0 {
					data.Retainedday2 = float32(100*day_2) / float32(data.Newuser)
					data.Retainedday3 = float32(100*day_3) / float32(data.Newuser)
					data.Retainedday7 = float32(100*day_7) / float32(data.Newuser)
					data.Retainedday30 = float32(100*day_30) / float32(data.Newuser)
				}
				if data.Paynum != 0 {
					data.Arppu = float32(data.Paytoday) / float32(data.Paynum)
				}
				if data.Dau != 0 {
					data.Arpu = float32(data.Paytoday) / float32(data.Dau)
					data.Payrate = float32(100*data.Paynum) / float32(data.Dau)
				}
				retl = append(retl, data)

			}
		}
	} else {
		for daynum, platdatamap := range retm {
			for tmpplatid, zonedatamap := range platdatamap {
				for zoneid, zonedata := range zonedatamap {
					data := &UserDailyData{}
					data.Gameid = gameid
					data.Zoneid = zoneid
					data.Platid = tmpplatid
					data.Daynum = daynum
					data.Newuser = zonedata["day_1"].(uint32)
					data.Dau = zonedata["dau"].(uint32)
					data.Paynum = zonedata["paynum"].(uint32)
					data.Paytoday = zonedata["paytoday"].(uint32)
					data.Login3days = zonedata["ol3d"].(uint32)
					data.Refluxnum = zonedata["refluxnum"].(uint32)
					data.Avgonlinemin = zonedata["avglogin"].(float32)
					data.Avglogintimes = zonedata["avgonline"].(float32)
					if data.Newuser != 0 {
						data.Retainedday2 = float32(100*zonedata["day_2"].(uint32)) / float32(data.Newuser)
						data.Retainedday3 = float32(100*zonedata["day_3"].(uint32)) / float32(data.Newuser)
						data.Retainedday7 = float32(100*zonedata["day_7"].(uint32)) / float32(data.Newuser)
						data.Retainedday30 = float32(100*zonedata["day_30"].(uint32)) / float32(data.Newuser)
					}
					if data.Paynum != 0 {
						data.Arppu = float32(data.Paytoday) / float32(data.Paynum)
					}
					if data.Dau != 0 {
						data.Arpu = float32(data.Paytoday) / float32(data.Dau)
						data.Payrate = float32(100*data.Paynum) / float32(data.Dau)
					}
					retl = append(retl, data)
				}
			}
		}
	}
	return
}

func HandleDailyData(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))

	var platlist []string
	if plats != "" {
		platlist = strings.Split(plats, ",")
	}
	//如果不是1，需要根据权限列获得相应的渠道id
	if len(platlist) != 1 {
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			platlist = ptlist
		}
	}
	datalist := CheckUserDailyData(gameid, zoneid, stime, etime, platlist, gameSystem, launchAccount)
	var data []byte
	if len(datalist) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(datalist)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	datalist = nil
}
func HandleDailyFirstData(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))

	var platlist []string
	if plats != "" {
		platlist = strings.Split(plats, ",")
	}
	//如果不是1，需要根据权限列获得相应的渠道id
	if len(platlist) != 1 {
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			platlist = ptlist
		}
	}
	datalist := CheckUserDailyfirstData(gameid, zoneid, stime, etime, platlist, gameSystem, launchAccount)
	var data []byte
	if len(datalist) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(datalist)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	datalist = nil
}

func HandleMonthData(task *unibase.ChanHttpTask) {

	fmt.Println(time.Now().Format("2006-01-02"))

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var platlist []string
	if plats != "" {
		platlist = strings.Split(plats, ",")
	}
	//如果不是1，需要根据权限列获得相应的渠道id
	if len(platlist) != 1 {
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			platlist = ptlist
		}
	}
	datalist := CheckUserMonthData(gameid, zoneid, stime, etime, platlist)
	var data []byte
	if len(datalist) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(datalist)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	datalist = nil
	//task.Debug("datalist:%v", datalist)
}

// platid -- [时间，总玩家数、在线人数、新角色数、DAU、付费人数、付费总额、新增用户付费、昨日新增、次日留存,提现人数 ，提现金额 ]
func CheckUserRealPlatData(gameid, zoneid uint32, platlist []string, gameSystem string, launchAccount string, time, usertype uint32) (retl []map[string]interface{}) {

	tblrealtime := get_realtime_data_table(gameid, time)

	where := fmt.Sprintf(" where  platid != 0 and type = %d ", usertype)

	if len(platlist) == 1 {
		where += fmt.Sprintf("  AND platid=%s ", platlist[0])
	} else if len(platlist) > 1 {
		where += fmt.Sprintf(" AND platid in (%s) ", strings.Join(platlist, ","))
	}

	if gameSystem != "all" {
		where += fmt.Sprintf(" AND launchid = '%s' ", gameSystem)
	}
	if launchAccount != "" {
		where += fmt.Sprintf("  AND adcode ='%s' ", launchAccount)
	}

	str := fmt.Sprintf("select hour,sum(totaluser),sum(newuser),sum(olduser),sum(dau),sum(paynum),sum(paytotal),sum(cashnum),sum(cashtotal)  from %s %s group by hour", tblrealtime, where)

	rows, err := db_monitor.Query(str)

	if err != nil {

		return
	}
	defer rows.Close()

	for rows.Next() {
		var hour string
		var totaluser, newuser, olduser, dau, paynum, paytotal, cashnum, cashtotal uint32

		rows.Scan(&hour, &totaluser, &newuser, &olduser, &dau, &paynum, &paytotal, &cashnum, &cashtotal)

		tmpdata := map[string]interface{}{
			"Recordtime":   hour + ":00:00 - " + hour + ":59:59",
			"Totaluser":    totaluser,
			"Onlinenum":    0,
			"Newuser":      newuser,
			"Olduser":      olduser,
			"Dau":          dau,
			"Paynum":       paynum,
			"Paytoday":     float32(paytotal / 100),
			"Cashoutnum":   cashnum,
			"Cashouttoday": float32(cashtotal / 100),
			"Arppu":        "0.00",
			"Arpu":         "0.00",
			"Payrate":      "0.00",
			"CashoutRatio": "0.00",
			"CashoutScale": "0.00",
		}

		if paynum != 0 {
			tmpdata["Arppu"] = fmt.Sprintf("%.2f", float32(paytotal)/float32(100*paynum))
		}
		if dau != 0 {
			tmpdata["Arpu"] = fmt.Sprintf("%.2f", float32(paytotal)/float32(100*dau))
			tmpdata["Payrate"] = fmt.Sprintf("%.2f", 100*float32(paynum)/float32(dau))
			tmpdata["CashoutRatio"] = fmt.Sprintf("%.2f", 100*float32(cashnum)/float32(dau))
		}
		if paytotal != 0 {
			tmpdata["CashoutScale"] = fmt.Sprintf("%.2f", 100*float32(cashtotal)/float32(paytotal))
		}
		retl = append(retl, tmpdata)

	}
	return
}

func HandleRealData(task *unibase.ChanHttpTask) {

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))

	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))
	times := uint32(unibase.Atoi(task.R.FormValue("time"), unitime.Time.YearMonthDay(0)))
	usertype := uint32(unibase.Atoi(task.R.FormValue("usertype"), 0))

	var platlist []string
	if plats != "" {
		platlist = strings.Split(plats, ",")
	}
	//如果不是1，需要根据权限列获得相应的渠道id
	if len(platlist) != 1 {
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			platlist = ptlist
		}
	}
	datalist := CheckUserRealPlatData(gameid, zoneid, platlist, gameSystem, launchAccount, times, usertype)
	var data []byte
	if len(datalist) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(datalist)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	datalist = nil
	//task.Debug("datalist:%v", datalist)
}

func HandleFormData(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var datalist []*UserDailyData
	var platlist []string
	if plats != "" {
		platlist = strings.Split(plats, ",")
	}

	//如果不是1，需要根据权限列获得相应的渠道id
	if len(platlist) != 1 {
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			platlist = ptlist
		}
	}
	datalist = CheckUserFormData(gameid, zoneid, stime, etime, platlist)
	var data []byte
	if len(datalist) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(datalist)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	datalist = nil
}

// 玩家等级分析
func HandleUserlevelAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1)) //1等级，2vip，3战力区间
	utype := uint32(unibase.Atoi(task.R.FormValue("utype"), 1))   //1注册用户，2活跃用户
	var where string
	if zoneid != 0 {
		where += fmt.Sprintf(" zoneid=%d AND ", zoneid)
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
	if utype == 1 {
		where += fmt.Sprintf(" firstmin>=%d and firstmin<%d", stime/60, (etime+86400)/60)
	} else {
		where += fmt.Sprintf(" last_login_time between %d and %d ", stime, etime)
	}

	//var str string
	tblname := get_user_data_table(gameid)
	tmap := map[uint32]string{
		1: "userlevel",
		2: "viplevel",
		3: "cast(power/10000 as unsigned)",
	}
	//str = fmt.Sprintf("select %s, count(*), avg(onlinemin), avg(gold) from %s where %s group by %s", tmap[optype], tblname, where, tmap[utype])
	str := fmt.Sprintf(`select %s, count(*), avg(onlinemin), avg(gold), avg(viplevel) from (
		select * from (select * from %s where %s order by %s desc) as a group by accid) as b group by %s`,
		tmap[optype], tblname, where, tmap[optype], tmap[optype])

	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserlevelAnalysis error:%s， sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	var total uint32
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var level, num uint32
		var avgonline, avggold, avgvip float32
		if err = rows.Scan(&level, &num, &avgonline, &avggold, &avgvip); err != nil {
			task.Warning("HandleUserlevelAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"level":     level,
			"num":       num,
			"avgonline": fmt.Sprintf("%.2f", avgonline),
			"avggold":   fmt.Sprintf("%.2f", avggold),
			"avgvip":    fmt.Sprintf("%.2f", avgvip),
		}
		total += num
		retl = append(retl, data)
	}
	for _, data := range retl {
		data["percent"] = fmt.Sprintf("%.2f", 100*float32(data["num"].(uint32))/float32(total))
	}
	//task.Debug("HandleUserlevelAnalysis retl:%v", retl)
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

var PowerLevelDescMap = map[uint32]string{
	1:  "0-2000",
	2:  "0.2-0.5万",
	3:  "0.5-1万",
	4:  "1-2万",
	5:  "2-3万",
	6:  "3-4万",
	7:  "4-5万",
	8:  "5-6万",
	9:  "6-8万",
	10: "8-10万",
	11: "10-14万",
	12: "14-20万",
	13: "20-30万",
	14: "30-50万",
	15: "50-100万",
	16: "100-200万",
	17: "200-500万",
	18: "500-1000万",
	19: ">1000万",
}

// 玩家战力分析
func HandleUserPowerAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	utype := uint32(unibase.Atoi(task.R.FormValue("utype"), 1)) //1注册用户，2活跃用户
	var where string
	if zoneid != 0 {
		where += fmt.Sprintf(" zoneid=%d AND ", zoneid)
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
	if utype == 1 {
		where += fmt.Sprintf(" firstmin>=%d and firstmin<%d", stime/60, (etime+86400)/60)
	} else {
		where += fmt.Sprintf(" last_login_time between %d and %d ", stime, etime)
	}

	tblname := get_user_data_table(gameid)
	str := fmt.Sprintf(`select (case when power<=2000 then 1 when power<=5000 then 2 when power<=10000 then 3
		when power<=20000 then 4 when power<=30000 then 5 when power<=40000 then 6 when power<=50000 then 7
		when power<=60000 then 8 when power<=80000 then 9 when power<=100000 then 10 when power<=140000 then 11
		when power<=200000 then 12 when power<=300000 then 13 when power<=500000 then 14 when power<=1000000 then 15
		when power<=2000000 then 16 when power<=5000000 then 17 when power<=10000000 then 18 else 19 end) as level,
		count(*), avg(onlinemin), avg(gold), avg(viplevel) from (
		select * from (select * from %s where %s order by power desc) as a group by accid) as b group by level`,
		tblname, where)

	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserPowerAnalysis error:%s， sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	var total uint32
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var level, num uint32
		var avgonline, avggold, avgvip float32
		if err = rows.Scan(&level, &num, &avgonline, &avggold, &avgvip); err != nil {
			task.Warning("HandleUserPowerAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"level":     PowerLevelDescMap[level],
			"num":       num,
			"avgonline": fmt.Sprintf("%.2f", avgonline),
			"avggold":   fmt.Sprintf("%.2f", avggold),
			"avgvip":    fmt.Sprintf("%.2f", avgvip),
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

// 流失玩家等级在线时长分布
func HandleUserLevelRetainedAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0)) / 60
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0)) / 60

	tbuser := get_user_data_table(gameid)
	where := fmt.Sprintf(" (firstmin between %d and %d) ", stime, etime)
	if zoneid != 0 {
		where = fmt.Sprintf("zoneid=%d AND %s ", zoneid, where)
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
			where += fmt.Sprintf("AND platid in (%s) ", strings.Join(ptlist, ","))
		}
	}

	now := time.Now()
	ts := time.Date(now.Year(), now.Month(), now.Day()-3, 0, 0, 0, 0, now.Location()).Unix()
	str := fmt.Sprintf(`select userlevel, count(*), sum(case when onlinetime=0 then 1 else 0 end),
	sum(case when 0<onlinetime and onlinetime<=60 then 1 else 0 end), sum(case when 60<onlinetime and onlinetime<=300 then 1 else 0 end),
	sum(case when 300<onlinetime and onlinetime<=600 then 1 else 0 end), sum(case when 600<onlinetime and onlinetime<=1200 then 1 else 0 end),
	sum(case when 1200<onlinetime and onlinetime<=1800 then 1 else 0 end), sum(case when 1800<onlinetime and onlinetime<=3600 then 1 else 0 end),
	sum(case when 3600<onlinetime and onlinetime<=10800 then 1 else 0 end), sum(case when 10800<onlinetime and onlinetime<=18000 then 1 else 0 end),
	sum(case when 18000<onlinetime and onlinetime<=28800 then 1 else 0 end), sum(case when onlinetime>=28800 then 1 else 0 end)
	from (select * from (select a.accid, max(userlevel) as userlevel, sum(onlinemin)*60 as onlinetime, max(last_login_time) as login from %s as a where %s group by accid)
	as c where login<?) as d group by userlevel`, tbuser, where)
	//统计等级账号数
	str2 := fmt.Sprintf(`select userlevel, count(*) from (select max(userlevel) as userlevel, a.accid from %s as a where %s group by accid) as c group by userlevel`, tbuser, where)
	retl := make([]map[string]interface{}, 0)
	retm := make(map[uint32]uint32)
	rows, err := db_monitor.Query(str2)
	if err != nil {
		task.Error("HandleUserLevelRetainedAnalysis error:%s， sql:%s", err.Error(), str2)
		goto result
	}
	for rows.Next() {
		var level, num uint32
		if err = rows.Scan(&level, &num); err != nil {
			continue
		}
		retm[level] = num
	}
	rows.Close()

	rows, err = db_monitor.Query(str, ts)
	if err != nil {
		task.Error("HandleUserLevelRetainedAnalysis error:%s， sql:%s", err.Error(), str)
		goto result
	}
	defer rows.Close()

	for rows.Next() {
		var level, num, num1, num2, num3, num4, num5, num6, num7, num8, num9, num10, num11 uint32
		if err = rows.Scan(&level, &num, &num1, &num2, &num3, &num4, &num5, &num6, &num7, &num8, &num9, &num10, &num11); err != nil {
			task.Warning("HandleUserLevelRetainedAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"level": level,
			"total": retm[level],
			"num":   num,
			"num1":  num1,
			"num2":  num2,
			"num3":  num3,
			"num4":  num4,
			"num5":  num5,
			"num6":  num6,
			"num7":  num7,
			"num8":  num8,
			"num9":  num9,
			"num10": num10,
			"num11": num11,
		}
		retl = append(retl, data)
	}

result:
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retm, retl = nil, nil
	return
}

// 玩家流失节点
func HandleUserLossNode(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
	if zoneid != 0 {
		where += fmt.Sprintf(" zoneid=%d AND ", zoneid)
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
	where += fmt.Sprintf(" firstmin>=%d and firstmin<%d", stime/60, etime/60)
	tblname := get_user_data_table(gameid)
	//str := fmt.Sprintf("select from_unixtime(firstmin*60, '%%Y%%m%%d') as daynum, max(isguid), count(*) from %s where %s group by daynum, isguid", tblname, where)
	str := fmt.Sprintf(`select from_unixtime(firstmin*60, '%%Y%%m%%d') as daynum, isguid, count(*) from (select accid,
		max(firstmin) as firstmin, max(isguid) as isguid from %s where %s group by accid) as a group by daynum, isguid`, tblname, where)
	task.Debug("sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserLossNode error:%s， sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	retm := make(map[uint32]uint32)
	for rows.Next() {
		var daynum, isguid, num uint32
		if err = rows.Scan(&daynum, &isguid, &num); err != nil {
			task.Error("HandleUserLossNode error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"daynum": daynum,
			"isguid": isguid,
			"num":    num,
		}
		retl = append(retl, data)
		retm[daynum] += num
	}
	for _, data := range retl {
		data["percent"] = fmt.Sprintf("%.2f", 100*float32(data["num"].(uint32))/float32(retm[data["daynum"].(uint32)]))
	}
	//task.Debug("HandleUserLossNode retl:%v", retl)
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retl, retm = nil, nil
	return
}

// 玩家流失等级
func HandleUserLossLevel(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
	if zoneid != 0 {
		where += fmt.Sprintf(" zoneid=%d AND ", zoneid)
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
	where += fmt.Sprintf(" firstmin>=%d and firstmin<%d", stime/60, etime/60)
	tblname := get_user_data_table(gameid)
	str := fmt.Sprintf("select from_unixtime(firstmin*60, '%%Y%%m%%d') as daynum, isguid, count(*) from %s where %s group by daynum, isguid", tblname, where)
	task.Debug("sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleUserLossNode error:%s， sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	retm := make(map[uint32]uint32)
	for rows.Next() {
		var daynum, isguid, num uint32
		if err = rows.Scan(&daynum, &isguid, &num); err != nil {
			task.Error("HandleUserLossNode error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"daynum": daynum,
			"isguid": isguid,
			"num":    num,
		}
		retl = append(retl, data)
		retm[daynum] += num
	}
	for _, data := range retl {
		data["percent"] = fmt.Sprintf("%.2f", 100*float32(data["num"].(uint32))/float32(retm[data["daynum"].(uint32)]))
	}
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retl, retm = nil, nil
	return
}

func HandleAccoundDaily(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	gameSystem := strings.TrimSpace(task.R.FormValue("gameSystem"))
	launchAccount := strings.TrimSpace(task.R.FormValue("launchAccount"))
	var where string
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

	if gameSystem != "all" {
		where += fmt.Sprintf(` launchid = "%s" AND `, gameSystem)
	}

	if launchAccount != "" {
		where += fmt.Sprintf(`adcode = "%s" AND `, launchAccount)
	}

	where += fmt.Sprintf(" daynum between %d and %d group by daynum order by daynum desc", stime, etime)
	tblname := get_user_account_table(gameid)
	str := fmt.Sprintf(`select daynum, count(*), sum(case when mobilenum > 0 then 1 else 0 end),
		sum(case when firstpaytime=daynum then 1 else 0 end),avg(logintimes),avg(onlinetime),
		sum(case when mobiletime > 0 and datediff(FROM_UNIXTIME(mobiletime),daynum)<3  then 1 else 0 end) from %s where %s`, tblname, where)
	task.Debug("sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleAccoundDaily error:%s， sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var avglogin, avgonline float32
		var daynum, newaccount, banduser, paynum, banduser3 uint32
		if err = rows.Scan(&daynum, &newaccount, &banduser, &paynum, &avglogin, &avgonline, &banduser3); err != nil {
			task.Error("HandleAccoundDaily error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"daynum":    daynum,
			"accnum":    newaccount,
			"banduser":  banduser,
			"paynum":    paynum,
			"avglogin":  fmt.Sprintf("%.2f", avglogin),
			"avgonline": fmt.Sprintf("%.2f", avgonline/60),
			"banduser3": banduser3,
			"bandrate":  "0.00%%",
			"bandrate3": "0.00%%",
		}
		if newaccount > 0 {
			data["bandrate"] = fmt.Sprintf("%.2f%%", 100*float32(banduser)/float32(newaccount))
			data["bandrate3"] = fmt.Sprintf("%.2f%%", 100*float32(banduser3)/float32(newaccount))
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

// 注册充值转化
func HandleRechargeTransform(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	//zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
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
	tblname := get_user_account_table(gameid)
	where += fmt.Sprintf(" daynum between %d and %d group by daynum order by daynum desc", stime, etime)
	str := fmt.Sprintf(`select daynum, count(*), sum(case when firstpaytime=daynum then 1 else 0 end),
		sum(case when datediff(firstpaytime, daynum)<2 then 1 else 0 end),
		sum(case when datediff(firstpaytime, daynum)<3 then 1 else 0 end),
		sum(case when datediff(firstpaytime, daynum)<7 then 1 else 0 end),
		sum(case when datediff(firstpaytime, daynum)<30 then 1 else 0 end) from %s where %s`, tblname, where)
	task.Debug("sql:%s", str)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandleRechargeTransform error:%s， sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var daynum, newnum, pay1, pay2, pay3, pay7, pay30 uint32
		if err = rows.Scan(&daynum, &newnum, &pay1, &pay2, &pay3, &pay7, &pay30); err != nil {
			task.Error("HandleRechargeTransform error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"daynum":    daynum,
			"accnum":    newnum,
			"payrate1":  "0.00%",
			"payrate2":  "0.00%",
			"payrate3":  "0.00%",
			"payrate7":  "0.00%",
			"payrate30": "0.00%",
		}
		if newnum > 0 {
			data["payrate1"] = fmt.Sprintf("%.2f%%", 100*float32(pay1)/float32(newnum))
			data["payrate2"] = fmt.Sprintf("%.2f%%", 100*float32(pay2)/float32(newnum))
			data["payrate3"] = fmt.Sprintf("%.2f%%", 100*float32(pay3)/float32(newnum))
			data["payrate7"] = fmt.Sprintf("%.2f%%", 100*float32(pay7)/float32(newnum))
			data["payrate30"] = fmt.Sprintf("%.2f%%", 100*float32(pay30)/float32(newnum))
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

// 注册充值转化(渠道注册分析)
func HandlePlatAccoundAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	//zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	tblname1 := get_user_account_table(gameid)
	tblname2 := get_user_data_table(gameid)
	str := fmt.Sprintf(`select a.platid, count(distinct a.accid) as account_num, sum(b.pay_all/100) as pay_total from
		%s as a, %s as b where a.accid=b.accid and daynum between %d and %d group by a.platid;`, tblname1, tblname2, stime, etime)
	rows, err := db_monitor.Query(str)
	if err != nil {
		task.Error("HandlePlatAccoundAnalysis error:%s， sql:%s", err.Error(), str)
		return
	}
	defer rows.Close()
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var platid, accountnum uint32
		var payall float32
		if err = rows.Scan(&platid, &accountnum, &payall); err != nil {
			task.Error("HandlePlatAccoundAnalysis error:%s", err.Error())
			continue
		}
		data := map[string]interface{}{
			"platid": platid,
			"accnum": accountnum,
			"payall": payall,
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

func HandleMahjangDaily(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	//zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charnum := uint32(unibase.Atoi(task.R.FormValue("charnum"), 0))
	repnum := uint32(unibase.Atoi(task.R.FormValue("repnum"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
	/*
		if zoneid != 0 {
			where += fmt.Sprintf(" zoneid=%d AND ", zoneid)
		}
	*/
	if charnum != 0 {
		where += fmt.Sprintf(" charnum=%d AND ", charnum)
	}
	if repnum != 0 {
		where += fmt.Sprintf(" repnum=%d AND ", repnum)
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
	tables := make(map[string]int)
	for i := 0; SubDate(int(stime), 0, 0, i) <= int(etime); i++ {
		tblname := get_mahjong_table(gameid, uint32(SubDate(int(stime), 0, 0, i)))
		tables[tblname] = 1
	}
	retl := make([]map[string]interface{}, 0)
	where += fmt.Sprintf(" (daynum between %d and %d) group by daynum, zoneid, charnum, repnum, optype order by daynum desc", stime, etime)
	gameids := make([]interface{}, 0)
	for tblname := range tables {
		if !check_table_exists(tblname) {
			continue
		}
		str := fmt.Sprintf(`select daynum, zoneid, charnum, repnum, optype, count(*), sum(realnum) from %s where %s`, tblname, where)
		rows, err := db_monitor.Query(str)
		if err != nil {
			task.Error("HandleMahjangAnalysis error:%s， sql:%s", err.Error(), str)
			return
		}
		tmpstr := map[uint32]string{0: "房主", 1: "均摊", 2: "大赢家"}
		for rows.Next() {
			var daynum, zoneid, charnum, repnum, optype, roomcount, totalnum uint32
			if err = rows.Scan(&daynum, &zoneid, &charnum, &repnum, &optype, &roomcount, &totalnum); err != nil {
				task.Error("HandleMahjangAnalysis error:%s", err.Error())
				continue
			}
			gameids = append(gameids, zoneid)
			data := map[string]interface{}{
				"daynum":    daynum,
				"gameid":    zoneid,
				"charnum":   charnum,
				"repnum":    repnum,
				"optype":    tmpstr[optype],
				"roomcount": roomcount,
				"totalnum":  totalnum,
			}
			retl = append(retl, data)
		}
		rows.Close()
	}
	gamenames := GetGameName(gameids...)
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		for _, data := range retl {
			data["gamename"] = gamenames[data["gameid"].(uint32)]
		}
		data, _ = json.Marshal(retl)
	}
	gameids, gamenames, retl = nil, nil, nil
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retl = nil
	return
}

func GetGameName(gameids ...interface{}) (retm map[uint32]string) {
	if len(gameids) == 0 {
		return
	}
	tblname := "apps"
	args := strings.Repeat("?,", len(gameids))
	str := fmt.Sprintf("select id, name from %s where id in (%s)", tblname, args[:len(args)-1])
	rows, err := unibase.DBZ.Query(str, gameids...)
	if err != nil {
		return
	}
	retm = make(map[uint32]string)
	for rows.Next() {
		var gameid uint32
		var gamename string
		if err = rows.Scan(&gameid, &gamename); err != nil {
			continue
		}
		retm[gameid] = gamename
	}
	rows.Close()
	return
}

func HandleSearchSubgames(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	ymd := uint32(unitime.Time.YearMonthDay(0))
	tblname := get_mahjong_table(gameid, ymd)
	str := fmt.Sprintf("select distinct zoneid from %s ", tblname)
	rows, err := db_monitor.Query(str)
	if err != nil {
		logging.Error("HandleSearchSubgames error:%s, sql:%s", err.Error(), str)
		task.SendBinary([]byte(`{"data":[]}`))
		return
	}
	gameids := make([]interface{}, 0)
	for rows.Next() {
		var subgameid uint32
		if err = rows.Scan(&subgameid); err != nil {
			continue
		}
		gameids = append(gameids, subgameid)
	}
	rows.Close()
	if len(gameids) == 0 {
		task.SendBinary([]byte(`{"data":[]}`))
		return
	}
	tblname = "apps"
	args := strings.Repeat("?,", len(gameids))
	str = fmt.Sprintf("select id, name from %s where id in (%s)", tblname, args[:len(args)-1])
	rows, err = unibase.DBZ.Query(str, gameids...)
	if err != nil {
		task.SendBinary([]byte(`{"data":[]}`))
		return
	}
	retl := make([]map[string]interface{}, 0)
	for rows.Next() {
		var gameid uint32
		var gamename string
		if err = rows.Scan(&gameid, &gamename); err != nil {
			continue
		}
		data := map[string]interface{}{
			"gameid":   gameid,
			"gamename": gamename,
		}
		retl = append(retl, data)
	}
	rows.Close()
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	gameids, retl = nil, nil
	return
}

func HandleMahjangAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	//zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
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

	tables := make(map[string]int)
	for i := 0; SubDate(int(stime), 0, 0, i) <= int(etime); i++ {
		tblname := get_mahjong_table(gameid, uint32(SubDate(int(stime), 0, 0, i)))
		tables[tblname] = 1
	}
	retl := make([]map[string]interface{}, 0)
	where += fmt.Sprintf(" (daynum between %d and %d) group by daynum, zoneid order by daynum desc", stime, etime)
	gameids := make([]interface{}, 0)
	for tblname := range tables {
		if !check_table_exists(tblname) {
			continue
		}
		str := fmt.Sprintf(`select daynum, zoneid, sum(realnum) as total, sum(case charnum when 2 then realnum else 0 end) as case2,
			sum(case charnum when 3 then realnum else 0 end) as case3, sum(case charnum when 4 then realnum else 0 end) as case4,
			sum(case repnum when 4 then realnum else 0 end) as rep4, sum(case repnum when 8 then realnum else 0 end) as rep8,
			sum(case repnum when 12 then realnum else 0 end) as rep12,sum(case repnum when 16 then realnum else 0 end) as rep16,
			sum(diamond) as diatotal, count(*) from %s where %s`, tblname, where)
		rows, err := db_monitor.Query(str)
		if err != nil {
			task.Error("HandleMahjangAnalysis error:%s， sql:%s", err.Error(), str)
			return
		}
		for rows.Next() {
			var daynum, zoneid, all, case2, case3, case4, rep4, rep8, rep12, rep16, total uint32
			var diamond float32
			if err = rows.Scan(&daynum, &zoneid, &all, &case2, &case3, &case4, &rep4, &rep8, &rep12, &rep16, &diamond, &total); err != nil {
				task.Error("HandleMahjangAnalysis error:%s", err.Error())
				continue
			}
			gameids = append(gameids, zoneid)
			data := map[string]interface{}{
				"daynum":  daynum,
				"gameid":  zoneid,
				"all":     all,
				"case2":   case2,
				"case3":   case3,
				"case4":   case4,
				"rep4":    rep4,
				"rep8":    rep8,
				"rep12":   rep12,
				"rep16":   rep16,
				"diamond": diamond,
				"total":   total,
			}
			retl = append(retl, data)
		}
		rows.Close()
	}
	gamenames := GetGameName(gameids...)
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		for _, data := range retl {
			data["gamename"] = gamenames[data["gameid"].(uint32)]
		}
		data, _ = json.Marshal(retl)
	}
	gameids, gamenames, retl = nil, nil, nil
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	return
}

// 麻将大厅统计
func HandleMjTotalAnalysis(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
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

	tables := make(map[string]int)
	for i := 0; SubDate(int(stime), 0, 0, i) <= int(etime); i++ {
		tblname := get_mahjong_table(gameid, uint32(SubDate(int(stime), 0, 0, i)))
		tables[tblname] = 1
	}
	retl := make([]map[string]interface{}, 0)
	where += fmt.Sprintf(" (daynum between %d and %d) group by daynum order by daynum desc", stime, etime)
	tbaccount := get_user_account_table(gameid)
	for tblname := range tables {
		if !check_table_exists(tblname) {
			continue
		}
		tmps := strings.Split(tblname, "_")
		tbpoint := fmt.Sprintf("user_mjpoint_%s", tmps[len(tmps)-1])
		str := fmt.Sprintf(`select a.daynum, sum(rooms), sum(rounds), sum(diamond) from %s as a inner
			join (select daynum, accid from %s where daynum between %d and %d) as b on
			(a.gameid=%d and a.charid=b.accid and a.daynum=b.daynum) where a.gameid=%d group by daynum`, tbpoint, tbaccount, stime, etime, gameid, gameid)
		rows, err := db_monitor.Query(str)
		if err != nil {
			task.Error("HandleMahjangAnalysis error:%s， sql:%s", err.Error(), str)
			return
		}
		resm := make(map[uint32][3]interface{})
		for rows.Next() {
			var daynum, rooms, rounds uint32
			var diamond float32
			if err = rows.Scan(&daynum, &rooms, &rounds, &diamond); err != nil {
				continue
			}
			resm[daynum] = [3]interface{}{rooms, rounds, diamond}
		}
		rows.Close()

		str = fmt.Sprintf(`select daynum, sum(realnum) as total, sum(case charnum when 2 then realnum else 0 end) as case2,
			sum(case charnum when 3 then realnum else 0 end) as case3, sum(case charnum when 4 then realnum else 0 end) as case4,
			sum(case repnum when 4 then realnum else 0 end) as rep4, sum(case repnum when 8 then realnum else 0 end) as rep8,
			sum(case repnum when 12 then realnum else 0 end) as rep12,sum(case repnum when 16 then realnum else 0 end) as rep16,
			sum(diamond) as diatotal, count(*) from %s where %s`, tblname, where)
		rows, err = db_monitor.Query(str)
		if err != nil {
			task.Error("HandleMahjangAnalysis error:%s， sql:%s", err.Error(), str)
			return
		}
		for rows.Next() {
			var daynum, all, case2, case3, case4, rep4, rep8, rep12, rep16, total uint32
			var diamond float32
			if err = rows.Scan(&daynum, &all, &case2, &case3, &case4, &rep4, &rep8, &rep12, &rep16, &diamond, &total); err != nil {
				task.Error("HandleMahjangAnalysis error:%s", err.Error())
				continue
			}
			data := map[string]interface{}{
				"daynum":     daynum,
				"all":        all,
				"case2":      case2,
				"case3":      case3,
				"case4":      case4,
				"rep4":       rep4,
				"rep8":       rep8,
				"rep12":      rep12,
				"rep16":      rep16,
				"diamond":    diamond,
				"total":      total,
				"rooms":      0,
				"rounds":     0,
				"newdiamond": 0.0,
			}
			if tmp, ok := resm[daynum]; ok {
				data["rooms"] = tmp[0]
				data["rounds"] = tmp[1]
				data["newdiamond"] = tmp[2]
			}
			retl = append(retl, data)
		}
		rows.Close()
		resm = nil
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

func get_mjpoint_table(ymd uint32) string {
	return fmt.Sprintf("user_mjpoint_%d", int(ymd/100))
}

// 麻将积分查询
func HandleMjPointSearch(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	sugameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	charid, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("charid")), 10, 64)
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
	if gameid == 0 || charid == 0 {
		task.SendBinary([]byte(`{"data":[]`))
		return
	}
	where = fmt.Sprintf(" gameid=%d and charid=%d ", gameid, charid)
	if sugameid != 0 {
		where = fmt.Sprintf("%s and subgameid=%d ", where, sugameid)
	}

	tables := make(map[string]int)
	for i := 0; SubDate(int(stime), 0, 0, i) <= int(etime); i++ {
		tblname := get_mjpoint_table(uint32(SubDate(int(stime), 0, 0, i)))
		tables[tblname] = 1
	}
	retl := make([]map[string]interface{}, 0)
	where += fmt.Sprintf(" and (daynum between %d and %d) order by daynum desc", stime, etime)
	gameids := make([]interface{}, 0)
	var total int32
	for tblname := range tables {
		if !check_table_exists(tblname) {
			continue
		}
		str := fmt.Sprintf(`select daynum, subgameid, point from %s where %s`, tblname, where)
		rows, err := db_monitor.Query(str)
		if err != nil {
			task.Error("HandleMjPointSearch error:%s， sql:%s", err.Error(), str)
			return
		}
		for rows.Next() {
			var daynum, subgameid uint32
			var point int32
			if err = rows.Scan(&daynum, &subgameid, &point); err != nil {
				task.Error("HandleMjPointSearch error:%s", err.Error())
				continue
			}
			gameids = append(gameids, subgameid)
			data := map[string]interface{}{
				"daynum":    daynum,
				"subgameid": subgameid,
				"point":     point,
			}
			total += point
			retl = append(retl, data)
		}
		rows.Close()
	}
	gameids = append(gameids, gameid)
	retl = append(retl, map[string]interface{}{"daynum": "总计", "subgameid": gameid, "point": total})
	gamenames := GetGameName(gameids...)
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		for _, data := range retl {
			data["gamename"] = gamenames[data["subgameid"].(uint32)]
		}
		data, _ = json.Marshal(retl)
	}
	gameids, gamenames, retl = nil, nil, nil
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	return
}

func get_user_group_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_group_%d_%d", gameid, ymd/100)
}
func get_user_redpack_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_redpack_%d_%d", gameid, ymd)
}
func get_user_share_table(gameid, ymd uint32) string {
	return fmt.Sprintf("user_share_%d_%d", gameid, ymd/100)
}

func HandleRedpackAnalysis(task *unibase.ChanHttpTask) {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("HandleRedpackAnalysis error:%v", err)
		}
	}()
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	//plats := strings.TrimSpace(task.R.FormValue("platlist"))
	stime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	var where string
	if zoneid != 0 {
		where += fmt.Sprintf(" zoneid=%d AND ", zoneid)
	}

	tables := make(map[string]int)
	for i := 0; SubDate(int(stime), 0, 0, i) <= int(etime); i++ {
		tblname := get_user_group_table(gameid, uint32(SubDate(int(stime), 0, 0, i)))
		tables[tblname] = 1
	}
	retm := make(map[uint32]map[string]uint32)
	where += fmt.Sprintf(" (daynum between %d and %d) group by daynum order by daynum desc", stime, etime)
	for tblname := range tables {
		if !check_table_exists(tblname) {
			continue
		}
		str := fmt.Sprintf("select daynum, count(*) from %s where state=1 and %s", tblname, where)
		rows, err := db_monitor.Query(str)
		if err != nil {
			task.Error("HandleRedpackAnalysis error:%s, sql:%s", err.Error(), str)
			continue
		}
		for rows.Next() {
			var daynum, num uint32
			if err = rows.Scan(&daynum, &num); err != nil {
				task.Error("HandleRedpackAnalysis error:%s", err.Error())
				continue
			}
			data, ok := retm[daynum]
			if !ok {
				data = map[string]uint32{
					"daynum": daynum,
					"gnum":   0,
					"tnum":   0,
					"pnum":   0,
					"snum":   0,
				}
			}
			data["gnum"] = num
			retm[daynum] = data
		}
		rows.Close()
	}
	tables = make(map[string]int)
	for i := 0; SubDate(int(stime), 0, 0, i) <= int(etime); i++ {
		tblname := get_user_share_table(gameid, uint32(SubDate(int(stime), 0, 0, i)))
		tables[tblname] = 1
	}
	for tblname := range tables {
		if !check_table_exists(tblname) {
			continue
		}
		str := fmt.Sprintf("select daynum, count(*) from %s where %s", tblname, where)
		rows, err := db_monitor.Query(str)
		if err != nil {
			task.Error("HandleRedpackAnalysis error:%s, sql:%s", err.Error(), str)
			continue
		}
		for rows.Next() {
			var daynum, num uint32
			if err = rows.Scan(&daynum, &num); err != nil {
				task.Error("HandleRedpackAnalysis error:%s", err.Error())
				continue
			}
			data, ok := retm[daynum]
			if !ok {
				data = map[string]uint32{
					"daynum": daynum,
					"gnum":   0,
					"tnum":   0,
					"pnum":   0,
					"snum":   0,
				}
			}
			data["share"] = num
			retm[daynum] = data
		}
		rows.Close()
	}

	for i := 0; SubDate(int(stime), 0, 0, i) <= int(etime); i++ {
		daynum := uint32(SubDate(int(stime), 0, 0, i))
		tblname := get_user_redpack_table(gameid, daynum)
		if !check_table_exists(tblname) {
			continue
		}
		str := fmt.Sprintf("select sum(case packtype when 1 then 1 else 0 end), sum(case packtype when 2 then 1 else 0 end) from %s ", tblname)
		row := db_monitor.QueryRow(str)
		var num1, num2 uint32
		if err := row.Scan(&num1, &num2); err != nil {
			task.Error("HandleRedpackAnalysis error:%s", err.Error())
			continue
		}
		data, ok := retm[daynum]
		if !ok {
			data = map[string]uint32{
				"daynum": daynum,
				"gnum":   0,
				"tnum":   0,
				"pnum":   0,
				"snum":   0,
			}
		}
		data["tnum"], data["pnum"] = num1, num2
		retm[daynum] = data
	}
	i := 0
	retl := make([]map[string]uint32, len(retm))
	for _, v := range retm {
		retl[i] = v
		i += 1
	}
	var data []byte
	if len(retl) == 0 {
		data = []byte("[]")
	} else {
		data, _ = json.Marshal(retl)
	}
	task.SendBinary([]byte(fmt.Sprintf(`{"data":%s}`, string(data))))
	retl, retm = nil, nil
	return
}

// 用户详细信息
func HandleUserDetailInfo(task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	plats := strings.TrimSpace(task.R.FormValue("platlist"))
	launchid := strings.TrimSpace(task.R.FormValue("launchid"))
	ad_account := strings.TrimSpace(task.R.FormValue("launchAccount"))
	stime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("starttime")), 10, 64)
	etime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("endtime")), 10, 64)
	userid := uint32(unibase.Atoi(task.R.FormValue("userid"), 0))
	latype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	//var where string
	where := fmt.Sprintf(" userlevel >= 0 ")

	////如果输入的货币id为0，默认为元宝
	//if userid == 0 {
	//userid = 8
	//}
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
			if latype == 0 {
				where += fmt.Sprintf(" AND platid in (%s) ", strings.Join(platlist, ","))
			} else {
				where += fmt.Sprintf(" AND curr_platid in (%s) ", strings.Join(platlist, ","))
			}

		}
	} else {

		//如果为空，也要检测下是否有权限设置 ,没有设置就是最大权限
		ptlist := getAccountPlatList(task)
		if len(ptlist) > 0 {
			if latype == 0 {
				where += fmt.Sprintf(" AND platid in (%s) ", strings.Join(ptlist, ","))
			} else {
				where += fmt.Sprintf(" AND curr_platid in (%s) ", strings.Join(ptlist, ","))
			}
		}

	}

	// if zoneid != 0 {
	// 	where += fmt.Sprintf(" AND zoneid=%d ", zoneid)
	// }

	if launchid != "all" {
		// where += fmt.Sprintf(" AND launchid='%s' ", launchid)
		if latype == 0 {
			where += fmt.Sprintf(" AND launchid='%s' ", launchid)
		} else {
			where += fmt.Sprintf(" AND curr_launchid='%s' ", launchid)
		}
	}
	if ad_account != "" {
		// where += fmt.Sprintf(" AND ad_account= '%s' ", ad_account)
		if latype == 0 {
			where += fmt.Sprintf(" AND ad_account='%s' ", ad_account)
		} else {
			where += fmt.Sprintf(" AND curr_ad_account='%s' ", ad_account)
		}
	}

	if userid != 0 {
		where += fmt.Sprintf(" AND userid=%d ", userid)
	} else {
		where += fmt.Sprintf(" AND FROM_UNIXTIME(reg_time , '%%Y%%m%%d') >= %d and FROM_UNIXTIME(reg_time , '%%Y%%m%%d') <= %d ", stime, etime)
	}

	//只查一次
	retl := make([]map[string]interface{}, 0)
	//for i, j := time.Unix(int64(stime), 0), 0; TimeToYmd(&i, j) < edate; j++ {
	tbname := get_user_data_table(gameid)

	if !check_table_exists(tbname) {
		task.Error("HandleUserDetailInfo error: not exist %s table", tbname)
		return
	}
	str := fmt.Sprintf("select zoneid, accid, userid ,launchid , ad_account, username,ip, reg_ip, reg_time, userlevel, viplevel, money, gold, power, pay_all/100,pay_first/100,pay_last/100,pay_first_time, pay_last_time,pay_first_level,last_login_time, last_logout_time , cash_out_all/100 , cash_out_num ,mobile from %s where %s order by reg_time desc", tbname, where)
	rows, err := db_monitor.Query(str)
	task.Debug("sql:%s", str)
	if err != nil {
		task.Error("HandleUserDetailInfo error:%s sql:%s", err.Error(), str)
		return
	}

	//var rank uint32 = 0
	for rows.Next() {
		var zoneid, accid, userid, ip, reg_ip, reg_time, userlevel, viplevel, money, gold, power, pay_first_time, pay_last_time, pay_first_level, last_login_time, last_logout_time, cash_out_num uint32

		var pay_all, pay_first, pay_last, cash_out_all float32
		var launchid, username, ad_account, mobile string

		if err = rows.Scan(&zoneid, &accid, &userid, &launchid, &ad_account, &username, &ip, &reg_ip, &reg_time, &userlevel, &viplevel, &money, &gold, &power, &pay_all, &pay_first, &pay_last, &pay_first_time, &pay_last_time, &pay_first_level, &last_login_time, &last_logout_time, &cash_out_all, &cash_out_num, &mobile); err != nil {
			task.Error("HandleUserMoneyRank error:%s", err.Error())
			return
		}
		data := map[string]interface{}{
			"zoneid":           zoneid,
			"accid":            accid,
			"userid":           userid,
			"launchid":         launchid,
			"ad_account":       ad_account,
			"username":         username,
			"ip":               nettask.GetAddrByIp(ip),
			"reg_ip":           nettask.GetAddrByIp(reg_ip),
			"reg_time":         UnixToYmdhms(uint64(reg_time)),
			"userlevel":        userlevel,
			"viplevel":         viplevel,
			"money":            money,
			"gold":             gold,
			"power":            power,
			"pay_all":          pay_all,
			"pay_first":        pay_first,
			"pay_last":         pay_last,
			"pay_first_day":    UnixToYmdhms(uint64(pay_first_time)),
			"pay_last_day":     UnixToYmdhms(uint64(pay_last_time)),
			"pay_first_level":  pay_first_level,
			"last_login_time":  UnixToYmdhms(uint64(last_login_time)),
			"last_logout_time": UnixToYmdhms(uint64(last_logout_time)),
			"cash_out_all":     cash_out_all,
			"cash_out_num":     cash_out_num,
			"cash_out_ratio":   0.00,
			"mobile":           mobile,
		}
		if pay_all != 0 && cash_out_all != 0 {
			data["cash_out_ratio"], _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(cash_out_all)/float64(pay_all)), 64)
		}
		// task.Debug("new userid=%d,curcoin=%d",userid,curcoin)
		//要查找一下是否有重复的
		var bFind bool = false
		//del_list := []int{}
		//for k, v := range retl {
		//if v["userid"] == userid {
		//task.Debug("重复删除:%d", userid)
		//del_list = append(del_list, k)
		//retl[k] = data
		//bFind = true
		//break
		//}
		//}
		if bFind == false {
			retl = append(retl, data)
		}

	}
	rows.Close()

	//}

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
