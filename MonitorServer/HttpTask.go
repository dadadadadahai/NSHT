package main

import (
	"fmt"

	sjson "github.com/bitly/go-simplejson"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
)

type HttpChanHandlerFunc func(data *sjson.Json) error

var jsmsgMainMap map[string]HttpChanHandlerFunc

func HandleHttpChanCommand(task *unibase.ChanHttpTask) bool {
	defer func() {
		if err := recover(); err != nil {
			task.Error("HandleHttpChanCommand err:%v, path:%s, data:%s", err, task.R.RequestURI, string(task.Rawdata))
		}
	}()
	task.Debug("HandleHttpChanCommand uri:%s, data:%s", task.R.RequestURI, string(task.Rawdata))
	task.R.ParseForm()
	form := task.R.Form
	cmd, sign := form.Get("do"), form.Get("sign")

	task.Debug("HandleHttpChanCommand cmd:%s, sign:%s", cmd, sign)

	data, err := sjson.NewJson(task.Rawdata)
	if err != nil {
		logging.Error("json.Unmarshal error:%s", err.Error())
		task.SendBinary([]byte("参数错误，发送的body非json格式"))
		return true
	}

	msgfun, ok := jsmsgMainMap[cmd]
	var success error
	if cmd != "" && err == nil && ok && checkSign(uint32(data.Get("gameid").MustInt()), sign, string(task.Rawdata)) {
		success = msgfun(data)
	} else {
		success = fmt.Errorf("args error")
	}

	task.JSW = unibase.NewJSResponseWriter(task.W, nil, nil)
	if success == nil {
		task.SendBinary([]byte("success"))
		task.Debug("HandleHttpChanCommand success, path:%s, data:%s", task.R.RequestURI, string(task.Rawdata))
	} else {
		task.SendBinary([]byte("参数或签名错误"))
		task.Error("HandleHttpChanCommand failed, error:%s, path:%s, data:%s", success.Error(), task.R.RequestURI, string(task.Rawdata))
	}
	return true
}

func checkSign(gameid uint32, sign, data string) bool {
	if gameid <= 0 || sign == "" || data == "" {
		return false
	}
	tblname := get_game_table()
	str := fmt.Sprintf("select gamekey from %s where gameid = ?", tblname)
	row := db_monitor.QueryRow(str, gameid)
	gamekey := ""
	if err := row.Scan(&gamekey); err != nil {
		logging.Error("get gamekey error:%s", err.Error())
		return false
	}
	//logging.Warning("key:%s, sign:%s, data:%s, calcsign:%s", gamekey, sign, data, md5String(data+gamekey))
	//logging.Warning("data+gamekey:%s",data+gamekey)
	if sign != md5String(data+gamekey) {
		//logging.Warning("签名cc")
		return false
	}
	return true
}

func InitHttpMsgMapMain() {
	jsmsgMainMap = make(map[string]HttpChanHandlerFunc)
	jsmsgMainMap["DeviceCollect"] = HandleDeviceCount
	jsmsgMainMap["AccountLogin"] = HandleAccountCount
	jsmsgMainMap["UserCreate"] = HandleRoleCreate
	jsmsgMainMap["UserLogin"] = HandleRoleLoginLogout
	jsmsgMainMap["UserLevelup"] = HandleRoleLevelup
	jsmsgMainMap["UserRecharge"] = HandleRoleRecharge
	jsmsgMainMap["UserCashout"] = HandleCashout
	jsmsgMainMap["UserOnline"] = HandleOnlineNum
	jsmsgMainMap["CoinOutputConsumption"] = HandleEconomicProduceConsume
	jsmsgMainMap["GoodsOutputConsumption"] = HandleItemProduceConsume
	jsmsgMainMap["ShopTransaction"] = HandleTransactionRecord
	jsmsgMainMap["RoleTransaction"] = HandleUserTransactionRecord
	jsmsgMainMap["ActionCollect"] = HandleActionRecord
	jsmsgMainMap["GameServerCrash"] = HandleGameServerCrash
}
