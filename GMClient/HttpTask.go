package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

	sjson "github.com/bitly/go-simplejson"

	"git.code4.in/mobilegameserver/config"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

var (
	jsmsgMainMap map[string]GmHandlerFunc
)

type GmHandlerFunc func(gmid uint32, task *unibase.ChanHttpTask)

func HandleGMLogin(gmid uint32, task *unibase.ChanHttpTask) {
	username := strings.TrimSpace(task.R.FormValue("username"))
	password := strings.TrimSpace(task.R.FormValue("password"))
	if username == "" || password == "" {
		task.Error("HandleGMLogin error, no username or password: (%s, %s)", username, password)
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"no username or password"}`))
	} else {
		ret := CheckXSRF(task.R, config.GetConfigStr("secret_key"))
		if !ret {
			task.Error("HandleGMLogin error, CheckXSRF failed: (%s, %s)", username, password)
			task.SendBinary([]byte(`{"retcode":1,"retdesc":"CheckXSRF failed"}`))
			return
		}
		InitGMClientTask(task.GetId(), username, password, Ip2Int(GetRemoteIp(task.R)))
	}
}

func HandleChangePassword(gmid uint32, task *unibase.ChanHttpTask) {
	oldpasswd := strings.TrimSpace(task.R.FormValue("oldpasswd"))
	newpasswd1 := strings.TrimSpace(task.R.FormValue("newpasswd1"))
	newpasswd2 := strings.TrimSpace(task.R.FormValue("newpasswd2"))
	if (newpasswd1 != newpasswd2) || !(newpasswd1 != "") {
		task.Error("HandleChangePassword error: password error!")
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"password error"}`))
		return
	}
	send := &Pmd.SetPasswordGmUserPmd_CS{}
	send.Oldpassword = proto.String(oldpasswd)
	send.Newpassword = proto.String(newpasswd1)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleChangeUserPassword(gmid uint32, task *unibase.ChanHttpTask) {
	// account := strings.TrimSpace(task.R.FormValue("account"))
	// passwd := strings.TrimSpace(task.R.FormValue("passwd"))
	// extdata := strings.TrimSpace(task.R.FormValue("extdata"))
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// if !(gameid != 0 && account != "" && passwd != "" && extdata != "") {
	// 	task.Error("HandleChangeUserPassword error: %s, %s, %s!", account, passwd, extdata)
	// 	task.SendBinary([]byte(`{"retcode":1,"retdesc":"password error"}`))
	// 	return
	// }
	// send := &Pmd.LobbyChgUserPwdGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())

	// send.Account = proto.String(account)
	// send.Passwd = proto.String(md5String(passwd))
	// send.Extdata = proto.String(extdata)
	// ForwardGmCommand(task, send, 0, 0, false)
}

// 添加GM账号
func HandleAddGmAccount(gmid uint32, task *unibase.ChanHttpTask) {
	//账号、密码、权限、游戏ID、绑定的IP地址
	name := strings.TrimSpace(task.R.FormValue("name"))
	passwd := strings.TrimSpace(task.R.FormValue("passwd"))
	pri, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("pri")), 10, 64)
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	//zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	if !(name != "" && passwd != "") || len(passwd) < 6 {
		task.Error("HandleAddGmAccount error: name or password error")
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"name or passwd error"}`))
		return
	}
	ip := net.ParseIP(strings.TrimSpace(task.R.FormValue("bindip"))).To4()
	var bindip uint32 = 0
	if len(ip) == 4 {
		bindip = uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
	}

	send := &Pmd.AddNewGmUserPmd_CS{}
	send.Clientid = proto.Uint64(task.GetId())

	send.Data = &Pmd.GmUserInfo{}
	send.Data.Name = proto.String(name)
	send.Data.Password = proto.String(passwd)
	send.Data.Pri = proto.Uint64(pri)
	send.Data.Gameid = proto.Uint32(gameid)
	//send.Data.Zoneid = proto.Uint32(zoneid)
	send.Data.Bindip = proto.Uint32(bindip)
	ForwardGmCommand(task, send, 0, 0, false)
}

// 删除GM账号
func HandleDelGmAccount(gmid uint32, task *unibase.ChanHttpTask) {
	accid := uint32(unibase.Atoi(task.R.FormValue("accid"), 0))
	send := &Pmd.RequestDelGmUserPmd_CS{}
	send.Gmid = proto.Uint32(accid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

// 删除游戏
func HandleDelGame(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	send := &Pmd.RequestDelGamePmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

// 添加游戏
func HandleGmAddGame(gmid uint32, task *unibase.ChanHttpTask) {
	createzoneid := uint32(unibase.Atoi(task.R.FormValue("createzoneid"), 0))
	createzonename := strings.TrimSpace(task.R.FormValue("createzonename"))
	creategameid := uint32(unibase.Atoi(task.R.FormValue("creategameid"), 0))
	creategamename := strings.TrimSpace(task.R.FormValue("creategamename"))
	gamelink := strings.TrimSpace(task.R.FormValue("gamelink"))
	//createusername := strings.TrimSpace(task.R.FormValue("createusername"))
	//createpassword := strings.TrimSpace(task.R.FormValue("createpassword"))
	gmlink := strings.TrimSpace(task.R.FormValue("gmlink"))
	gamekey := strings.TrimSpace(task.R.FormValue("gamekey"))
	gametype := uint32(unibase.Atoi(task.R.FormValue("gametype"), 0))

	send := &Pmd.AddNewGameGmUserPmd_CS{}
	send.Creategameid = proto.Uint32(creategameid)
	send.Creategamename = proto.String(creategamename)
	send.Gamelink = proto.String(gamelink)
	send.Gmlink = proto.String(gmlink)
	send.Gamekey = proto.String(gamekey)
	send.Createzoneid = proto.Uint32(createzoneid)
	send.Createzonename = proto.String(createzonename)

	//send.Data = &Pmd.GmUserInfo{}
	//send.Data.Name = proto.String(createusername)
	//send.Data.Password = proto.String(createpassword)
	//send.Data.Gameid = proto.Uint32(creategameid)

	//Gamekey status 前端没有传参 后面再加
	send.Gametype = proto.Uint32(gametype)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleAddZone(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	zonename := strings.TrimSpace(task.R.FormValue("zonename"))
	gmlink := strings.TrimSpace(task.R.FormValue("gmlink"))
	send := &Pmd.AddNewZoneGmUserPmd_CS{}
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Zoneid = proto.Uint32(zoneid)
	send.Gameid = proto.Uint32(gameid)
	send.Zonename = proto.String(zonename)
	send.Gmlink = proto.String(gmlink)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleModifyGmPri(gmid uint32, task *unibase.ChanHttpTask) {
	pri, err := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("pri")), 10, 64)
	if err != nil {
		task.Error("HandleModifyGmPri error:%s", err.Error())
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"pri error"}`))
		return
	}
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	send := &Pmd.RequestModifyPriGmUserPmd_CS{}
	send.Gmid = proto.Uint32(uint32(gmid))
	send.Pri = proto.Uint64(pri)
	send.Gameid = proto.Uint32(gameid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSelectGamezone(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.SelectGamezoneGmUserPmd_SC{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchUserInfo(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("charid")), 10, 64)
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	pid := strings.TrimSpace(task.R.FormValue("pid"))

	send := &Pmd.RequestUserInfoGmUserPmd_C{}
	send.Pid = proto.String(pid)
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Charid = proto.Uint64(uint64(charid))

	// if charname != "" {
	// 	str := fmt.Sprintf(`select id from channel_accounts where plataccount="%s"`, charname)
	// 	var id uint64

	// 	rows := unibase.DBZ.QueryRow(str)

	// 	err := rows.Scan(&id)
	// 	if err != nil {
	// 		var data []byte
	// 		data, _ = json.Marshal(map[string]string{"retdesc": "查询玩家失败"})
	// 		task.SendBinary(data)
	// 	}
	// 	send.Charid = proto.Uint64(uint64(id))
	// }
	send.Charname = proto.String(charname)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleUserInfoUpdatePassword(gmid uint32, task *unibase.ChanHttpTask) {
	charid, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("charid")), 10, 64)
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	password := strings.TrimSpace(task.R.FormValue("password"))

	where := ""

	var data []byte
	if charid != 0 && charname != "" {
		where += fmt.Sprintf(`  plataccount="%s" and id = %d `, charname, charid)
	} else if charid != 0 {
		where += fmt.Sprintf("  id=%d ", charid)
	} else if charname != "" {
		where += fmt.Sprintf(`  plataccount="%s" `, charname)
	}

	str := fmt.Sprintf("update channel_accounts set password=md5(?) where %s ", where)

	result, err := unibase.DBZ.Exec(str, password)
	if err == nil {
		_, err = result.RowsAffected()
	}
	if err != nil {
		data, _ = json.Marshal(map[string]string{"retdesc": "修改玩家密码失败"})
	}
	data, _ = json.Marshal(map[string]string{"retdesc": "修改玩家密码成功"})

	task.SendBinary(data)
}
func HandleSearchUserList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("charid")), 10, 64)
	regip := strings.TrimSpace(task.R.FormValue("regip"))
	isonline := uint32(unibase.Atoi(task.R.FormValue("isonline"), 0))
	accid, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("accid")), 10, 64)
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	lstime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("lstime")), 10, 64)
	letime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("letime")), 10, 64)
	mincoin, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("mincoin")), 10, 64)
	maxcoin, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("maxcoin")), 10, 64)
	minlevel := uint32(unibase.Atoi(task.R.FormValue("minlevel"), 0))
	maxlevel := uint32(unibase.Atoi(task.R.FormValue("maxlevel"), 0))
	agent := uint32(unibase.Atoi(task.R.FormValue("agent"), 0))
	ranktype := uint32(unibase.Atoi(task.R.FormValue("ranktype"), 0))
	usertype := uint32(unibase.Atoi(task.R.FormValue("usertype"), 0))
	regflag := uint32(unibase.Atoi(task.R.FormValue("regflag"), 0))
	phonenum := strings.TrimSpace(task.R.FormValue("mobile"))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	account := strings.TrimSpace(task.R.FormValue("account"))
	cpf := strings.TrimSpace(task.R.FormValue("cpf"))

	send := &Pmd.RequestOnlineUserInfoGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Charid = proto.Uint64(charid)
	send.Regip = proto.String(regip)
	send.Isonline = proto.Uint32(isonline)
	send.Accid = proto.Uint64(accid)
	send.Minlevel = proto.Uint32(minlevel)
	send.Maxlevel = proto.Uint32(maxlevel)
	send.Starttime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Isagent = proto.Uint32(agent)
	send.Lstime = proto.Uint64(lstime)
	send.Letime = proto.Uint64(letime)
	send.Mincoin = proto.Uint64(mincoin)
	send.Maxcoin = proto.Uint64(maxcoin)
	send.Ranktype = proto.Uint32(ranktype)
	send.Usertype = proto.Uint32(usertype)
	send.Regflag = proto.Uint32(regflag)
	send.Subplatid = proto.Uint32(platid)
	send.Phonenum = proto.String(phonenum)
	send.Account = proto.String(account)
	send.Cpf = proto.String(cpf)

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	send.Ranktype = proto.Uint32(uint32(unibase.Atoi(task.R.FormValue("ranktype"), 0)))
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleModifyUserInfo(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid := strings.TrimSpace(task.R.FormValue("charid"))
	content := strings.TrimSpace(task.R.FormValue("content"))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	chtype := uint32(unibase.Atoi(task.R.FormValue("changetype"), 0))

	send := &Pmd.RequestModifyUserInfoGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Charids = proto.String(charid)
	send.Optype = proto.Uint32(optype)
	send.Changetype = proto.Uint32(chtype)
	// if optype == 1 {
	// 	send.Charname = proto.String(content)
	// } else if optype == 11 {
	// 	tmps := strings.SplitN(content, "|", 2)
	// 	if len(tmps) == 2 {
	// 		send.Ext = proto.String(tmps[1])
	// 	}
	// 	send.Opnum = proto.Int32(int32(unibase.Atoi(tmps[0], 0)))
	// } else {
	// 	opnum := int32(unibase.Atoi(content, 0))
	// 	send.Opnum = proto.Int32(opnum)
	// }
	if optype == 3 || optype == 7 || optype == 22 || optype == 23 || optype == 24 {
		opnum := int32(unibase.Atoi(content, 0))
		send.Opnum = proto.Int32(opnum)
	} else {

		send.Ext = proto.String(content)
		//if optype == 17 || optype == 18 || optype == 19 {
		send.Opnum = proto.Int32(1)
		//}
	}
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchUserRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("charid")), 10, 64)
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	recordtime := uint32(unibase.Atoi(task.R.FormValue("recordtime"), 0))

	send := &Pmd.RequestUserRecordGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Optype = proto.Uint32(1)
	send.Charid = proto.Uint64(uint64(charid))
	send.Charname = proto.String(charname)
	send.Recordtime = proto.Uint32(recordtime)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

// 系统设置
func HandleSystemSet(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	if optype < 1 || optype > 3 {
		task.SendBinary([]byte(`{"ret":1,retcode":1,"retdesc":"optype error"}`))
		return
	}
	if optype == 1 {
		send := &Pmd.RequestHotRestartGmUserPmd_CS{}
		send.Gameid = proto.Uint32(gameid)
		send.Zoneid = proto.Uint32(zoneid)
		send.Gmid = proto.Uint32(gmid)
		send.Clientid = proto.Uint64(task.GetId())
		ForwardGmCommand(task, send, 0, 0, false)
	} else if optype == 2 {
		send := &Pmd.RequesetScriptUpdateGmUserPmd_CS{}
		send.Gameid = proto.Uint32(gameid)
		send.Zoneid = proto.Uint32(zoneid)
		send.Gmid = proto.Uint32(gmid)
		send.Clientid = proto.Uint64(task.GetId())
		ForwardGmCommand(task, send, 0, 0, false)
	} else {
		send := &Pmd.RequesetRefreshGatewaylistGmUserPmd_CS{}
		send.Gameid = proto.Uint32(gameid)
		send.Zoneid = proto.Uint32(zoneid)
		send.Gmid = proto.Uint32(gmid)
		send.Clientid = proto.Uint64(task.GetId())
		ForwardGmCommand(task, send, 0, 0, false)
	}

}

func HandleOrderList(gmid uint32, task *unibase.ChanHttpTask) {

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	begintime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	status := uint32(unibase.Atoi(task.R.FormValue("status"), 0))
	rechargetype := uint32(unibase.Atoi(task.R.FormValue("rechargetype"), 0))
	gameorder := strings.TrimSpace(task.R.FormValue("gameorder"))
	platorder := strings.TrimSpace(task.R.FormValue("platorder"))
	regflag := uint32(unibase.Atoi(task.R.FormValue("regflag"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	send := &Pmd.GameOrderListGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)
	send.Begintime = proto.String(begintime)
	send.Endtime = proto.String(endtime)
	send.Status = proto.Uint32(status)
	send.Rechargetype = proto.Uint32(rechargetype)
	send.Gameorder = proto.String(gameorder)
	send.Platorder = proto.String(platorder)
	send.Regflag = proto.Uint32(regflag)
	send.Subplatid = proto.Uint32(platid)

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)

	fmt.Println(send)

	ForwardGmCommand(task, send, 0, 0, false)

}

func HandlePunishUser(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid := strings.TrimSpace(task.R.FormValue("charid"))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	pid := strings.TrimSpace(task.R.FormValue("pid"))
	content := task.R.FormValue("content")
	pointnum := strings.TrimSpace(task.R.FormValue("pointnum"))
	multiple := uint32(unibase.Atoi(task.R.FormValue("multiple"), 0))
	send := &Pmd.PunishUserGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Charids = proto.String(charid)

	send.Data = &Pmd.PunishUserInfo{}
	send.Data.Pid = proto.String(pid)
	send.Data.Gameid = proto.Uint32(gameid)
	send.Data.Zoneid = proto.Uint32(zoneid)
	send.Data.Gmid = proto.Uint32(gmid)
	send.Data.Ptype = proto.Uint32(optype)
	send.Data.Reason = proto.String(content)
	send.Data.Starttime = proto.Uint64(starttime)
	send.Data.Endtime = proto.Uint64(endtime)
	send.Data.Punishvalue = proto.String(pointnum)
	send.Data.Multiple = proto.Uint32(multiple)

	ForwardGmCommand(task, send, 0, 0, false)
}

func HandlePunishList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	ptype := uint32(unibase.Atoi(task.R.FormValue("ptype"), 0))
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	pid := strings.TrimSpace(task.R.FormValue("pid"))
	punishvalue := strings.TrimSpace(task.R.FormValue("punishvalue"))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))

	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send := &Pmd.RequestPunishListGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Pid = proto.String(pid)
	send.Charid = proto.Uint64(charid)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Ptype = proto.Uint32(ptype)
	send.State = proto.Uint32(state)
	send.Punishvalue = proto.String(punishvalue)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandlePunishDelete(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	taskid := uint32(unibase.Atoi(task.R.FormValue("taskid"), 0))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	ptype := uint32(unibase.Atoi(task.R.FormValue("ptype"), 0))
	pid := strings.TrimSpace(task.R.FormValue("pid"))
	send := &Pmd.DeletePunishUserGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Ptype = proto.Uint32(ptype)

	send.Pid = proto.String(pid)
	send.Taskid = proto.Uint32(taskid)
	send.Charid = proto.Uint64(charid)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleFeedbackList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	ftype := uint32(unibase.Atoi(task.R.FormValue("feedbacktype"), 0))
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send := &Pmd.RequestFeedbackListGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Platid = proto.Uint32(platid)
	send.Charid = proto.Uint64(charid)
	send.Feedbacktype = proto.Uint32(ftype)
	send.State = proto.Uint32(state)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Charname = proto.String(charname)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleDealFeedback(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	recordid := uint32(unibase.Atoi(task.R.FormValue("recordid"), 0))
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	subject := strings.TrimSpace(task.R.FormValue("subject"))
	content := strings.TrimSpace(task.R.FormValue("content"))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)

	send := &Pmd.RequestDealFeedbackGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Recordid = proto.Uint32(recordid)
	send.State = proto.Uint32(state)
	send.Charid = proto.Uint64(charid)
	send.Subject = proto.String(subject)
	send.Reply = proto.String(content)
	//task.Info("HandleDealFeedback charid:%d", charid)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleGMLoginSuccess(task *unibase.ChanHttpTask, rev *Pmd.ReturnLoginGmUserPmd_S) {
	if rev.GetRetcode() == 0 {
		SetSecureCookie(task.W, config.GetConfigStr("secret_key"), "id", strconv.FormatUint(task.GetId(), 10))
		SetCookie(task.W, "username", task.Name)
	}
	b, _ := json.Marshal(rev)
	task.SendBinary(b)
}

func HandleExecFunc(gmid uint32, task *unibase.ChanHttpTask) {
	command := task.R.FormValue("command")
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	args := task.R.FormValue("commandargs")

	var argsmap map[string]string = make(map[string]string)
	argslist := strings.Fields(args)
	for _, arg := range argslist {
		tmp := strings.Split(arg, "=")
		if len(tmp) != 2 {
			continue
		}
		argsmap[tmp[0]] = tmp[1]
	}

	sj := sjson.New()
	sj.Set("data", argsmap)
	sj.Set("do", command)
	data, _ := sj.MarshalJSON()

	send := &Pmd.RequestExecGmCommandGmPmd_SC{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Msg = proto.String(string(data))
	ForwardGmCommand(task, send, gameid, zoneid, false)
}

func HandleExecStr(gmid uint32, task *unibase.ChanHttpTask) {
	command := task.R.FormValue("command")
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	send := &Pmd.RequestExecScriptGmPmd_S{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Script = proto.String(command)
	ForwardGmCommand(task, send, gameid, zoneid, false)
}

// func HandlePointReport(gmid uint32, task *unibase.ChanHttpTask) {
// 	now := uint32(unitime.Time.Sec())
// 	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
// 	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
// 	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
// 	endtime := uint32(unibase.Atoi(task.R.FormValue("endtime"), int(now)))
// 	starttime := uint32(unibase.Atoi(task.R.FormValue("starttime"), int(endtime-6*3600)))

// 	send := &Pmd.RequestPointReportGmUserPmd_CS{}
// 	send.Gameid = proto.Uint32(gameid)
// 	send.Zoneid = proto.Uint32(zoneid)
// 	send.Gmid = proto.Uint32(gmid)
// 	send.Clientid = proto.Uint64(task.GetId())

// 	send.Charid = proto.Uint64(charid)
// 	send.Starttime = proto.Uint32(starttime)
// 	send.Endtime = proto.Uint32(endtime)

// 	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
// 	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
// 	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
// 	send.Curpage = proto.Uint32(curpage)
// 	send.Maxpage = proto.Uint32(maxpage)
// 	send.Perpage = proto.Uint32(perpage)
// 	ForwardGmCommand(task, send, 0, 0, false)
// }

// func HandlePointDetail(gmid uint32, task *unibase.ChanHttpTask) {
// 	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
// 	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
// 	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
// 	ptype := uint32(unibase.Atoi(task.R.FormValue("ptype"), 0))
// 	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
// 	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
// 	if endtime == 0 {
// 		endtime = uint64(unitime.Time.Sec())
// 	}
// 	if starttime == 0 {
// 		starttime = endtime - 6*3600
// 	}

// 	send := &Pmd.RequestPointDetailGmUserPmd_CS{}
// 	send.Gameid = proto.Uint32(gameid)
// 	send.Zoneid = proto.Uint32(zoneid)
// 	send.Gmid = proto.Uint32(gmid)
// 	send.Clientid = proto.Uint64(task.GetId())

// 	send.Charid = proto.Uint64(charid)
// 	send.Ptype = proto.Uint32(ptype)
// 	send.Starttime = proto.Uint64(starttime)
// 	send.Endtime = proto.Uint64(endtime)

// 	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
// 	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
// 	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
// 	send.Curpage = proto.Uint32(curpage)
// 	send.Maxpage = proto.Uint32(maxpage)
// 	send.Perpage = proto.Uint32(perpage)
// 	ForwardGmCommand(task, send, 0, 0, false)
// }

func HandleSubgameList(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// send := &Pmd.RequestedSubgameListGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())
	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleStockData(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// subgame := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	// send := &Pmd.RequestStockInfoGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())
	// send.Subgameid = proto.Uint32(subgame)
	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleModStockData(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// subgame := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	// stockid := uint64(unibase.Atoi(task.R.FormValue("stockid"), 0))
	// stock := uint64(unibase.Atoi(task.R.FormValue("stock"), 0))
	// threshold := uint64(unibase.Atoi(task.R.FormValue("threshold"), 0))
	// lottery := uint64(unibase.Atoi(task.R.FormValue("lottery"), 0))

	// send := &Pmd.RequestModStockInfoGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())

	// send.Subgameid = proto.Uint32(subgame)
	// send.Data = &Pmd.StockData{}
	// send.Data.Id = proto.Uint64(stockid)
	// send.Data.Stock = proto.Uint64(stock)
	// send.Data.Threshold = proto.Uint64(threshold)
	// send.Data.Lottery = proto.Uint64(lottery)
	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchWinningList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	timestamp, _ := strconv.ParseUint(task.R.FormValue("timestamp"), 10, 64)

	send := &Pmd.RequestWinningListGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Subgameid = proto.Uint32(subgame)
	send.Timestamp = proto.Uint64(timestamp)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchBlackWhiteList(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// subgame := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	// id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	// charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)

	// send := &Pmd.RequestBlackWhitelistGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())

	// send.Subgameid = proto.Uint32(subgame)
	// send.Charid = proto.Uint64(charid)
	// send.Id = proto.Uint32(id)

	// curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	// maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	// perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	// send.Curpage = proto.Uint32(curpage)
	// send.Maxpage = proto.Uint32(maxpage)
	// send.Perpage = proto.Uint32(perpage)
	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleBlackWhitelistMod(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// subgame := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	// id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	// charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	// charname := task.R.FormValue("charname")
	// setchips := uint32(unibase.Atoi(task.R.FormValue("setchips"), 0))
	// curchips := uint32(unibase.Atoi(task.R.FormValue("curchips"), 0))
	// state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	// btype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	// winrate := uint32(unibase.Atoi(task.R.FormValue("winrate"), 0))
	// settimes := uint32(unibase.Atoi(task.R.FormValue("settimes"), 0))
	// intervaltimes := uint32(unibase.Atoi(task.R.FormValue("intervaltimes"), 0))

	// send := &Pmd.ModBlackWhitelistGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())

	// send.Subgameid = proto.Uint32(subgame)
	// send.Data = &Pmd.BlackWhitelistInfo{}
	// send.Data.Id = proto.Uint32(id)
	// send.Data.Charid = proto.Uint64(charid)
	// send.Data.Charname = proto.String(charname)
	// send.Data.Subgameid = proto.Uint32(subgame)
	// send.Data.Setchips = proto.Uint32(setchips)
	// send.Data.Curchips = proto.Uint32(curchips)
	// send.Data.State = proto.Uint32(state)
	// send.Data.Type = proto.Uint32(btype)
	// send.Data.Winrate = proto.Uint32(winrate)
	// send.Data.Settimes = proto.Uint32(settimes)
	// send.Data.Intervaltimes = proto.Uint32(intervaltimes)

	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleBlackWhitelistAdd(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// subgame := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	// charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	// charname := task.R.FormValue("charname")
	// setchips := uint32(unibase.Atoi(task.R.FormValue("setchips"), 0))
	// curchips := uint32(unibase.Atoi(task.R.FormValue("curchips"), 0))
	// state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	// btype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	// winrate := uint32(unibase.Atoi(task.R.FormValue("winrate"), 0))
	// settimes := uint32(unibase.Atoi(task.R.FormValue("settimes"), 0))
	// intervaltimes := uint32(unibase.Atoi(task.R.FormValue("intervaltimes"), 0))

	// send := &Pmd.AddBlackWhitelistGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())

	// send.Subgameid = proto.Uint32(subgame)
	// data := &Pmd.BlackWhitelistInfo{}
	// data.Charid = proto.Uint64(charid)
	// data.Charname = proto.String(charname)
	// data.Subgameid = proto.Uint32(subgame)
	// data.Setchips = proto.Uint32(setchips)
	// data.Curchips = proto.Uint32(curchips)
	// data.State = proto.Uint32(state)
	// data.Type = proto.Uint32(btype)
	// data.Winrate = proto.Uint32(winrate)
	// data.Settimes = proto.Uint32(settimes)
	// data.Intervaltimes = proto.Uint32(intervaltimes)
	// send.Data = append(send.Data, data)

	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleBlackWhitelistDel(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// subgame := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	// ids := task.R.FormValue("ids")
	// send := &Pmd.DelBlackWhitelistGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())
	// send.Subgameid = proto.Uint32(subgame)

	// for _, tmpid := range strings.Split(ids, ",") {
	// 	id := uint32(unibase.Atoi(tmpid, 0))
	// 	if id == 0 {
	// 		continue
	// 	}
	// 	send.Ids = append(send.Ids, id)
	// }
	// if len(send.Ids) == 0 {
	// 	send.Retcode = proto.Uint32(1)
	// 	send.Retdesc = proto.String("删除的ID为空")
	// 	data, _ := json.Marshal(send)
	// 	task.SendBinary(data)
	// 	return
	// }
	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSendGmMail(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	mtype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	content := strings.TrimSpace(task.R.FormValue("content"))
	gold := uint32(unibase.Atoi(task.R.FormValue("gold"), 0))
	pid := strings.TrimSpace(task.R.FormValue("pid"))

	send := &Pmd.RequestSendMailGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	gmail := &Pmd.GmMailInfo{}
	gmail.Gameid = proto.Uint32(gameid)
	gmail.Zoneid = proto.Uint32(zoneid)
	gmail.Type = proto.Uint32(mtype)
	gmail.Pid = proto.String(pid)
	gmail.Content = proto.String(content)
	gmail.Gold = proto.Uint32(gold)

	send.Data = gmail
	ForwardGmCommand(task, send, 0, 0, false)
}

// 发送批量条件式邮件
func HandleSendGmMails(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subject := strings.TrimSpace(task.R.FormValue("subject"))
	content := strings.TrimSpace(task.R.FormValue("content"))
	attachments := strings.TrimSpace(task.R.FormValue("attachments"))
	gold := uint32(unibase.Atoi(task.R.FormValue("gold"), 0))
	goldbind := uint32(unibase.Atoi(task.R.FormValue("goldbind"), 0))
	money := uint32(unibase.Atoi(task.R.FormValue("money"), 0))

	minlevel := uint32(unibase.Atoi(task.R.FormValue("minlevel"), 0))
	maxlevel := uint32(unibase.Atoi(task.R.FormValue("maxlevel"), 0))
	minvip := uint32(unibase.Atoi(task.R.FormValue("minvip"), 0))
	maxvip := uint32(unibase.Atoi(task.R.FormValue("maxvip"), 0))
	minlogin := uint32(unibase.Atoi(task.R.FormValue("minlogin"), 0))
	maxlogin := uint32(unibase.Atoi(task.R.FormValue("maxlogin"), 0))
	online := uint32(unibase.Atoi(task.R.FormValue("online"), 0))
	ext := strings.TrimSpace(task.R.FormValue("ext"))

	send := &Pmd.RequestSendMailExGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	if subject == "" || content == "" {
		send.Retcode = proto.Uint32(1)
		send.Retdesc = proto.String("邮件标题、内容错误")
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	gmail := &Pmd.GmMailInfoEx{}
	gmail.Gameid = proto.Uint32(gameid)
	gmail.Zoneid = proto.Uint32(zoneid)
	gmail.Subject = proto.String(subject)
	gmail.Content = proto.String(content)
	gmail.Gold = proto.Uint32(gold)
	gmail.Goldbind = proto.Uint32(goldbind)
	gmail.Money = proto.Uint32(money)
	gmail.Maxlevel = proto.Uint32(maxlevel)
	gmail.Minlevel = proto.Uint32(minlevel)
	gmail.Maxvip = proto.Uint32(maxvip)
	gmail.Minvip = proto.Uint32(minvip)
	gmail.Maxlogin = proto.Uint32(maxlogin)
	gmail.Minlogin = proto.Uint32(minlogin)
	gmail.Online = proto.Uint32(online)
	gmail.Type = proto.Uint32(1)
	if strings.Contains(ext, "_") {
		gmail.Type = proto.Uint32(0)
	}
	for _, tmpatt := range strings.Split(attachments, ",") {
		att := strings.Split(tmpatt, "*")
		if len(att) != 4 {
			continue
		}
		itype := uint32(unibase.Atoi(att[0], 0))
		itemid := uint32(unibase.Atoi(att[1], 0))
		itemnum := uint32(unibase.Atoi(att[2], 0))
		bind := uint32(unibase.Atoi(att[3], 0))
		if itemid == 0 || itemnum == 0 {
			continue
		}
		data := &Pmd.ItemInfo{}
		data.Itemtype = proto.Uint32(itype)
		data.Itemid = proto.Uint32(itemid)
		data.Itemnum = proto.Uint32(itemnum)
		data.Bind = proto.Uint32(bind)
		gmail.Attachment = append(gmail.Attachment, data)
	}
	send.Data = gmail
	send.Extdata = proto.String(ext)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchItem(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.RequestItemTypeInfoGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		send.Retcode = proto.Uint32(1)
		send.Retdesc = proto.String("参数错误")
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchLoginRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.RequestLoginRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	accid, _ := strconv.ParseUint(task.R.FormValue("accid"), 10, 64)
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send.Charid = proto.Uint64(charid)
	send.Accid = proto.Uint64(accid)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchConsumeRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	btype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	send := &Pmd.RequestConsumeRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 || btype == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	itemid := uint32(unibase.Atoi(task.R.FormValue("itemid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	actionid := uint32(unibase.Atoi(task.R.FormValue("actionid"), 0))
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send.Charid = proto.Uint64(charid)
	send.Type = proto.Uint32(btype)
	send.Itemid = proto.Uint32(itemid)
	send.Optype = proto.Uint32(optype)
	send.Actionid = proto.Uint32(actionid)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchActionRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	btype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	send := &Pmd.RequestActionRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 || btype == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	acttype := uint32(unibase.Atoi(task.R.FormValue("acttype"), 0))
	actionid := uint32(unibase.Atoi(task.R.FormValue("actionid"), 0))
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send.Charid = proto.Uint64(charid)
	send.Type = proto.Uint32(btype)
	send.Acttype = proto.Uint32(acttype)
	send.Actionid = proto.Uint32(actionid)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchStrengthenRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.RequestStrengthenRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	optarget := uint32(unibase.Atoi(task.R.FormValue("optarget"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send.Charid = proto.Uint64(charid)
	send.Optarget = proto.Uint32(optarget)
	send.Optype = proto.Uint32(optype)
	send.State = proto.Uint32(state)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchMailRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	send := &Pmd.RequestMailRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	recvid, _ := strconv.ParseUint(task.R.FormValue("recvid"), 10, 64)
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}

	send.Charid = proto.Uint64(charid)
	send.Recvid = proto.Uint64(recvid)
	send.Optype = proto.Uint32(optype)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)

	if optype == 1 {
		data := &Pmd.RequestMailRecordGmUserPmd_CS_MailData{}

		data.Id = proto.Uint32(id)
		data.Recvid = proto.Uint64(recvid)

		send.Data = append(send.Data, data)
	}
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchRankRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.RequestRankRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	ranktype := uint32(unibase.Atoi(task.R.FormValue("ranktype"), 0))
	rankname := task.R.FormValue("rankname")
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send.Ranktype = proto.Uint32(ranktype)
	send.Rankname = proto.String(rankname)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchBossRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.RequestBossRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	bosstype := uint32(unibase.Atoi(task.R.FormValue("bosstype"), 0))
	bossname := task.R.FormValue("bossname")
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send.Bosstype = proto.Uint32(bosstype)
	send.Bossname = proto.String(bossname)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchRenameRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.RequestRenameRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	oldname := task.R.FormValue("oldname")
	newname := task.R.FormValue("newname")
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}
	send.Charid = proto.Uint64(charid)
	send.Oldname = proto.String(oldname)
	send.Newname = proto.String(newname)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchItemRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.RequestUserItemsHistoryListGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	send.Optype = proto.Uint32(uint32(unibase.Atoi(task.R.FormValue("cointype"), 0)))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	send.Charid = proto.Uint64(charid)
	send.Charname = proto.String(charname)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleMatchAwardSearch(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	send := &Pmd.MatchAwardSearchGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	send.Charid = proto.Uint64(charid)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleMatchAward(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	recordid, _ := strconv.ParseUint(task.R.FormValue("recordid"), 10, 64)
	send := &Pmd.MatchAwardGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.State = proto.Uint32(1)
	if gameid == 0 || zoneid == 0 || recordid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}

	send.Recordid = proto.Uint64(recordid)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleMatchSignupSearch(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	matchtype := uint32(unibase.Atoi(task.R.FormValue("matchtype"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 100))
	send := &Pmd.MatchSignupSearchGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	send.Charid = proto.Uint64(charid)
	send.Matchtype = proto.Uint32(matchtype)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleMatchSignupOperator(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	matchtype := uint32(unibase.Atoi(task.R.FormValue("matchtype"), 0))
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 1))
	send := &Pmd.MatchSignupOperatorGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.State = proto.Uint32(state)
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}

	send.Charid = proto.Uint64(charid)
	send.Matchtype = proto.Uint32(matchtype)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleMessagePush(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// title := strings.TrimSpace(task.R.FormValue("title"))
	// desc := strings.TrimSpace(task.R.FormValue("content"))

	// send := &Pmd.GmRequestPushMessageUserListGmUserPmd_CS{}
	// tmp := strings.TrimSpace(task.R.FormValue("charid"))
	// charids := strings.Split(tmp, ",")
	// for _, tmpcharid := range charids {
	// 	if charid, err := strconv.ParseUint(tmpcharid, 10, 64); err == nil {
	// 		send.Charids = append(send.Charids, charid)
	// 	}
	// }
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())
	// if gameid == 0 || zoneid == 0 || desc == "" || len(send.Charids) == 0 {
	// 	data, _ := json.Marshal(send)
	// 	task.SendBinary(data)
	// 	return
	// }

	// send.Timestamp = proto.Uint32(uint32(unitime.Time.Sec()))
	// //todo 添加其他查询条件，目前只添加了角色ID
	// send.Title = proto.String(title)
	// send.Desc = proto.String(desc)
	// send.Script = proto.String("")
	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleLobbyGameHistory(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	stime, _ := strconv.ParseUint(task.R.FormValue("stime"), 10, 64)
	etime, _ := strconv.ParseUint(task.R.FormValue("etime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))

	send := &Pmd.LobbyGameHistoryGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Subgameid = proto.Uint32(subgameid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	if etime == 0 {
		etime = uint64(unitime.Time.Sec())
	}
	if stime == 0 {
		stime = etime - 6*3600
	}
	send.Charid = proto.Uint64(charid)
	send.Starttime = proto.Uint64(stime)
	send.Endtime = proto.Uint64(etime)
	send.Perpage = proto.Uint32(perpage)
	send.Curpage = proto.Uint32(curpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleLobbyGameDetailHistory(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	// charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	// groomid := uint32(unibase.Atoi(task.R.FormValue("groomid"), 0))
	// stime, _ := strconv.ParseUint(task.R.FormValue("stime"), 10, 64)
	// etime, _ := strconv.ParseUint(task.R.FormValue("etime"), 10, 64)
	// curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	// perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))

	// send := &Pmd.LobbyGameDetailHistoryGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Subgameid = proto.Uint32(subgameid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())
	// if gameid == 0 || zoneid == 0 {
	// 	data, _ := json.Marshal(send)
	// 	task.SendBinary(data)
	// 	return
	// }
	// if etime == 0 {
	// 	etime = uint64(unitime.Time.Sec())
	// }
	// if stime == 0 {
	// 	stime = etime - 6*3600
	// }
	// send.Charid = proto.Uint64(charid)
	// send.Groomid = proto.Uint32(groomid)
	// send.Starttime = proto.Uint64(stime)
	// send.Endtime = proto.Uint64(etime)
	// send.Perpage = proto.Uint32(perpage)
	// send.Curpage = proto.Uint32(curpage)
	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchRelationShip(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	stime, _ := strconv.ParseUint(task.R.FormValue("stime"), 10, 64)
	etime, _ := strconv.ParseUint(task.R.FormValue("etime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))

	send := &Pmd.RelationShipInfoGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Charid = proto.Uint64(charid)
	send.Charname = proto.String(charname)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	if etime == 0 {
		etime = uint64(unitime.Time.Sec())
	}
	if stime == 0 {
		stime = etime - 6*3600
	}
	send.Starttime = proto.Uint64(stime)
	send.Endtime = proto.Uint64(etime)
	send.Perpage = proto.Uint32(perpage)
	send.Curpage = proto.Uint32(curpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchGroupInfo(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid, _ := strconv.ParseUint(task.R.FormValue("charid"), 10, 64)
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	stime, _ := strconv.ParseUint(task.R.FormValue("stime"), 10, 64)
	etime, _ := strconv.ParseUint(task.R.FormValue("etime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	ownerid, _ := strconv.ParseUint(task.R.FormValue("ownerid"), 10, 64)
	owner := strings.TrimSpace(task.R.FormValue("owner"))
	groupid, _ := strconv.ParseUint(task.R.FormValue("groupid"), 10, 64)
	groupname := strings.TrimSpace(task.R.FormValue("groupname"))

	send := &Pmd.GroupInfoGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Charid = proto.Uint64(charid)
	send.Charname = proto.String(charname)
	send.Clientid = proto.Uint64(task.GetId())
	send.Ownerid = proto.Uint64(ownerid)
	send.Owner = proto.String(owner)
	send.Groupid = proto.Uint64(groupid)
	send.Groupname = proto.String(groupname)

	if gameid == 0 || zoneid == 0 {
		data, _ := json.Marshal(send)
		task.SendBinary(data)
		return
	}
	if etime == 0 {
		etime = uint64(unitime.Time.Sec())
	}
	if stime == 0 {
		stime = etime - 6*3600
	}
	send.Starttime = proto.Uint64(stime)
	send.Endtime = proto.Uint64(etime)
	send.Perpage = proto.Uint32(perpage)
	send.Curpage = proto.Uint32(curpage)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleModGroupInfo(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	groupid, _ := strconv.ParseUint(task.R.FormValue("groupid"), 10, 64)
	groupname := strings.TrimSpace(task.R.FormValue("groupname"))
	memlimit := uint32(unibase.Atoi(task.R.FormValue("memlimit"), 0))
	level := uint32(unibase.Atoi(task.R.FormValue("level"), 0))

	send := &Pmd.ModGroupInfoGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Groupid = proto.Uint64(groupid)
	send.Groupname = proto.String(groupname)
	send.Memlimit = proto.Uint32(memlimit)
	send.Level = proto.Uint32(level)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleManageGroup(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	groupid, _ := strconv.ParseUint(task.R.FormValue("groupid"), 10, 64)
	optype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	charids := strings.TrimSpace(task.R.FormValue("charids"))

	send := &Pmd.ManageGroupGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Groupid = proto.Uint64(groupid)

	if len(charids) == 0 || optype < 1 || optype > 4 {
		send.Retcode = proto.Uint32(1)
		send.Retdesc = proto.String("参数错误")
		data, _ := json.Marshal(send)
		task.SendBinary([]byte(data))
		return
	}
	tmpids := strings.Split(charids, ",")
	for _, tmpid := range tmpids {
		charid, err := strconv.ParseUint(tmpid, 10, 64)
		if err != nil {
			continue
		}
		send.Charids = append(send.Charids, charid)
	}

	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchChatMessage(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// charid, _ := strconv.ParseUint(task.R.FormValue("groupid"), 10, 64)
	// export := uint32(unibase.Atoi(task.R.FormValue("export"), 0))
	// ctype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	// cdate := uint32(unibase.Atoi(task.R.FormValue("chatdate"), 0))

	// send := &Pmd.RequestChatMessageGmUserPmd_CS{}
	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)
	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())
	// send.Charid = proto.Uint64(charid)
	// send.Download = proto.Uint32(export)
	// send.Type = proto.Uint32(ctype)
	// send.Chatdate = proto.Uint32(cdate)
	// ForwardGmCommand(task, send, 0, 0, false)
}

func HandleActControl(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	actid := uint32(unibase.Atoi(task.R.FormValue("actid"), 0))
	actname := strings.TrimSpace(task.R.FormValue("actname"))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	stime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	etime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)

	send := &Pmd.ActivitySwitchGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Actid = proto.Uint32(actid)
	send.Actname = proto.String(actname)
	send.Optype = proto.Uint32(optype)
	send.Starttime = proto.Uint64(stime)
	send.Endtime = proto.Uint64(etime)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleCPay(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	payplatid := uint32(unibase.Atoi(task.R.FormValue("payplatid"), 0))
	state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	packid := uint32(unibase.Atoi(task.R.FormValue("packageid"), 0))
	minlevel := uint32(unibase.Atoi(task.R.FormValue("minlevel"), 0))
	maxlevel := uint32(unibase.Atoi(task.R.FormValue("maxlevel"), 0))
	minmoney := uint32(unibase.Atoi(task.R.FormValue("minmoney"), 0))
	maxmoney := uint32(unibase.Atoi(task.R.FormValue("maxmoney"), 0))
	stime := uint32(unibase.Atoi(task.R.FormValue("stime"), 0))
	etime := uint32(unibase.Atoi(task.R.FormValue("etime"), 0))
	recordid, _ := strconv.ParseUint(task.R.FormValue("recordid"), 10, 64)

	send := &Pmd.CPayGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Data = &Pmd.CPayData{}
	send.Data.Recordid = proto.Uint64(recordid)
	send.Data.Platid = proto.Uint32(platid)
	send.Data.Payplatid = proto.Uint32(payplatid)
	send.Data.Minlevel = proto.Uint32(minlevel)
	send.Data.Maxlevel = proto.Uint32(maxlevel)
	send.Data.Minmoney = proto.Uint32(minmoney)
	send.Data.Maxmoney = proto.Uint32(maxmoney)
	send.Data.Stime = proto.Uint32(stime)
	send.Data.Etime = proto.Uint32(etime)
	send.Data.State = proto.Uint32(state)
	send.Data.Packageid = proto.Uint32(packid)

	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSetWinOrLose(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	settings := strings.TrimSpace(task.R.FormValue("settings"))
	send := &Pmd.SetWinOrLoseGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Settings = proto.String(settings)
	send.Subgameid = proto.Uint32(subgameid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleCPayList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	recordid, _ := strconv.ParseUint(task.R.FormValue("recordid"), 10, 64)
	//platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	//payplatid := uint32(unibase.Atoi(task.R.FormValue("payplatid"), 0))
	//state := uint32(unibase.Atoi(task.R.FormValue("state"), 0))
	//packid := uint32(unibase.Atoi(task.R.FormValue("packageid"), 0))

	send := &Pmd.CPaylistGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Recordid = proto.Uint64(recordid)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleAddRedPacket(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	srcuid, _ := strconv.ParseUint(task.R.FormValue("srcuid"), 10, 64)
	money, _ := strconv.ParseUint(task.R.FormValue("money"), 10, 64)

	send := &Pmd.RequestAddRedPacketGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Srcuid = proto.Uint64(srcuid)
	send.Money = proto.Uint64(money)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleRevRedPacket(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	srcuid, _ := strconv.ParseUint(task.R.FormValue("srcuid"), 10, 64)
	codeid := strings.TrimSpace(task.R.FormValue("id"))

	send := &Pmd.RequestRevRedPacketGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Srcuid = proto.Uint64(srcuid)
	send.Id = proto.String(codeid)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleGameData(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	stime, _ := strconv.ParseUint(task.R.FormValue("stime"), 10, 64)
	etime, _ := strconv.ParseUint(task.R.FormValue("etime"), 10, 64)

	send := &Pmd.RequestGameDataGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Type = proto.Uint32(optype)
	send.Stime = proto.Uint64(stime)
	send.Etime = proto.Uint64(etime)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleRoomDissolve(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	roomid, _ := strconv.ParseUint(task.R.FormValue("roomid"), 10, 64)
	send := &Pmd.RoomDissolveGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Roomid = proto.Uint64(roomid)
	ForwardGmCommand(task, send, 0, 0, false)
}

func CheckGmHasLogin(gmname string) bool {
	var res bool = false
	callback := func(v entry.EntryInterface) bool {
		task := v.(*GMClientTask)
		if task.Data.GetName() == gmname {
			res = true
			return false
		}
		return true
	}
	GMCM.ExecEvery(callback)
	return res
}

func GetGMClientTask(r *http.Request, clietid uint64) (gmtask *GMClientTask) {
	var cookie, key string
	key = GetCookie(r, "key")
	cookie = GetSecureCookie(r, config.GetConfigStr("secret_key"), "id")

	debug := config.GetConfigStr("debug")
	if debug == "true" && cookie == "" {
		cookie = r.FormValue("id")
	}
	id, err := strconv.ParseUint(cookie, 10, 64)
	if err != nil {
		return nil
	}
	gmtask = GMCM.GetGMClientTaskById(id)
	if gmtask != nil && (gmtask.LastSecretKey == key || debug == "true") {
		gmtask.LastChanTaskId = clietid
		return gmtask
	}
	return
}

func HandleGmPriviliege(gmid uint32, task *unibase.ChanHttpTask) {
	gmtask := GetGMClientTask(task.R, task.GetId())
	if gmtask != nil {
		priviliege := gmtask.Data.GetPri()
		task.SendBinary([]byte(fmt.Sprintf(`{"retcode":0,"data": %d}`, priviliege)))
	} else {
		task.Warning("id or secret key error")
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"id or secret key error"}`))
	}
}

func HandleGmZonelist(gmid uint32, task *unibase.ChanHttpTask) {
	gmtask := GetGMClientTask(task.R, task.GetId())
	if gmtask != nil {
		gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
		for _, game := range gmtask.GameList {
			if gameid == game.GameId {
				ret := map[string]interface{}{
					"retcode": 1,
					"data":    game.ZoneList,
				}
				data, _ := json.Marshal(ret)
				task.SendBinary(data)
				return
			}
		}
	}
	task.SendBinary([]byte(`{"retcode":1,"retdesc":"游戏未上线或cookie已过期"}`))
}

func ForwardGmCommand(task *unibase.ChanHttpTask, send proto.Message, gameid, zoneid uint32, wrap bool) bool {
	gmtask := GetGMClientTask(task.R, task.GetId())
	if gmtask != nil {
		if wrap {
			data, _ := json.Marshal(send)
			sj := sjson.New()
			sj.Set("data", string(data))
			sj.Set("do", reflect.TypeOf(send).String()[1:])
			data, _ = sj.MarshalJSON()
			sendwrap := &Pmd.RequestExecGmCommandGmPmd_SC{}
			sendwrap.Gameid = proto.Uint32(gameid)
			sendwrap.Zoneid = proto.Uint32(zoneid)
			sendwrap.Msg = proto.String(string(data))
			sendwrap.Clientid = proto.Uint64(task.GetId())
			gmtask.SendCmd(sendwrap)
			task.Debug("ForwardGmCommand:%s", unibase.GetProtoString(sendwrap.String()))
		} else {
			gmtask.SendCmd(send)
			task.Debug("ForwardGmCommand:%s", unibase.GetProtoString(send.String()))
		}
	} else {
		task.Warning("id or secret key error")
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"id or secret key error"}`))
		return false
	}
	return true
}

func ForwardGmProto(protoname string, task *unibase.ChanHttpTask) bool {
	rawdata := task.Rawdata
	task.Debug("Handle %s, %s", task.R.RequestURI, string(rawdata))
	task.JSW = unibase.NewJSResponseWriter(task.W, nil, nil)
	sj, err := sjson.NewJson(rawdata)
	if err != nil {
		task.Error("HandleHttpGm ParseJson error:%s, data:%s", err.Error(), string(rawdata))
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return false
	}
	gmdata, ok := sj.CheckGet("gmdata")
	if !ok {
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return false
	}
	gmdata.Set("clientid", task.GetId())

	data, err := sj.MarshalJSON()
	if err != nil {
		task.Debug("ForwardGmProto MarshalJSON error:%s, proto:%s", err.Error(), protoname)
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return false
	}
	send := nettask.GetCmdByFullName("*Pmd."+protoname, data)
	if send == nil {
		task.Debug("ForwardGmCommand GetCmdByFullName proto:%s", protoname)
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"参数错误"}`))
		return false
	}
	gmtask := GetGMClientTask(task.R, task.GetId())
	if gmtask != nil {
		gmtask.SendCmd(send)
		task.Debug("ForwardGmProto:%s", unibase.GetProtoString(send.String()))
	}
	return true
}

func ParseDrawcashListGmUserPmd_CS(gmid uint32, task *unibase.ChanHttpTask) {
	ForwardGmProto("DrawcashListGmUserPmd_CS", task)
}

func ParseDrawcashManagerGmUserPmd_CS(gmid uint32, task *unibase.ChanHttpTask) {
	ForwardGmProto("DrawcashManagerGmUserPmd_CS", task)
}

func HandleGmCommand(task *unibase.ChanHttpTask) bool {
	defer task.R.Body.Close()
	defer func() {
		if err := recover(); err != nil {
			task.Error("HandleRedpackAnalysis error:%v", err)
		}
	}()

	task.R.Form = ParseQuery(string(task.Rawdata))
	task.JSW = unibase.NewJSResponseWriter(task.W, nil, nil)

	cmd := task.R.FormValue("cmd")
	islogin := task.R.FormValue("islogin") //这个是给其它后台加的可以直接使用,不判断登陆
	task.Debug("HandleGmCommand cmd:%s, data:%s, istestlogin=%d", cmd, task.Rawdata, islogin)

	fmt.Println("cmd:", cmd)

	if cmd == "gmlogin" {
		HandleGMLogin(0, task)
	} else if islogin == "ok" {
		if _, ok := CallSdkLuaFunc("GmHttp.Parse"+cmd, "", 10001, task.GetId(), string(task.Rawdata)); !ok {
			if msgfun, ok := jsmsgMainMap[cmd]; ok == true {
				msgfun(10001, task)
			} else {
				task.SendBinary([]byte(`{"retcode":1,"retdesc":"暂不支持，请联系后台"}`))
			}
		}
	} else if gmclient := GetGMClientTask(task.R, task.GetId()); gmclient != nil {
		if _, ok := CallSdkLuaFunc("GmHttp.Parse"+cmd, "", gmclient.Data.GetGmid(), task.GetId(), string(task.Rawdata)); !ok {
			if msgfun, ok := jsmsgMainMap[cmd]; ok == true {
				msgfun(gmclient.Data.GetGmid(), task)
			} else {
				task.SendBinary([]byte(`{"retcode":1,"retdesc":"暂不支持，请联系后台"}`))
			}
		}
	} else {
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"请先登录!"}`))
	}
	return true
}
func HandleSearchGmUserListByGameId_C(gmid uint32, task *unibase.ChanHttpTask) {
	curtype := uint32(unibase.Atoi(task.R.FormValue("curtype"), 0))
	username := strings.TrimSpace(task.R.FormValue("username"))
	id := uint64(unibase.Atoi(task.R.FormValue("id"), 0))
	Password := strings.TrimSpace(task.R.FormValue("Password"))
	priviliege := uint64(unibase.Atoi(task.R.FormValue("priviliege"), 0))
	Bbool := strings.TrimSpace(task.R.FormValue("Bbool"))
	Pbool := strings.TrimSpace(task.R.FormValue("Pbool"))
	gamestr := strings.TrimSpace(task.R.FormValue("gamestr"))

	ip := net.ParseIP(strings.TrimSpace(task.R.FormValue("bindip"))).To4()

	var bindip uint32 = 0

	if len(ip) == 4 {
		bindip = uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
	}

	send := &Pmd.SetPriGameGmUserPmd_CS{}
	curpage := uint32(unibase.Atoi(task.R.FormValue("Curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("Perpage"), 20))

	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)

	send.Id = proto.Uint64(id)
	send.Priviliege = proto.Uint64(priviliege)
	send.Curtype = proto.Uint32(curtype)
	send.Username = proto.String(username)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Gamestr = proto.String(gamestr)

	send.Bindip = proto.Uint32(bindip)

	send.Bbool = proto.String(Bbool)
	send.Pbool = proto.String(Pbool)
	send.Password = proto.String(Password)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleSearchGmGameList_CS(gmid uint32, task *unibase.ChanHttpTask) {
	Curtype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	send := &Pmd.GmGameOrZonelistGet_CS{}
	send.Curtype = proto.Uint32(Curtype)
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(gameid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleModifyGmGameInfo_CS(gmid uint32, task *unibase.ChanHttpTask) {
	Gameid := uint32(unibase.Atoi(task.R.FormValue("Gameid"), 0))
	Curtype := uint32(unibase.Atoi(task.R.FormValue("Curtype"), 0))

	modifygamename := strings.TrimSpace(task.R.FormValue("modifygamename"))
	modifygamelink := strings.TrimSpace(task.R.FormValue("modifygamelink"))
	modifygmlink := strings.TrimSpace(task.R.FormValue("modifygmlink"))
	modifygametype := uint32(unibase.Atoi(task.R.FormValue("modifygametype"), 0))
	modifygamekey := strings.TrimSpace(task.R.FormValue("modifygamekey"))
	modifystatus := uint32(unibase.Atoi(task.R.FormValue("modifystatus"), 0))

	send := &Pmd.ModifyGameretInfolist_CS{}
	send.Curtype = proto.Uint32(Curtype)
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(Gameid)
	send.Clientid = proto.Uint64(task.GetId())

	curpage := uint32(unibase.Atoi(task.R.FormValue("Curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("Perpage"), 20))

	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)

	send.Modifygamename = proto.String(modifygamename)
	send.Modifygamelink = proto.String(modifygamelink)
	send.Modifygmlink = proto.String(modifygmlink)
	send.Modifygametype = proto.Uint32(modifygametype)
	send.Modifygamekey = proto.String(modifygamekey)
	send.Modifystatus = proto.Uint32(modifystatus)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleZoneOperation_CS(gmid uint32, task *unibase.ChanHttpTask) {
	Gameid := uint32(unibase.Atoi(task.R.FormValue("Gameid"), 0))
	Zoneid := uint32(unibase.Atoi(task.R.FormValue("Zoneid"), 0))
	Curtype := uint32(unibase.Atoi(task.R.FormValue("Curtype"), 0))
	modifygzonename := strings.TrimSpace(task.R.FormValue("Modifygzonename"))
	modifygmlink := strings.TrimSpace(task.R.FormValue("Modifygmlink"))
	modifystatus := uint32(unibase.Atoi(task.R.FormValue("Modifystatus"), 0))
	send := &Pmd.RequestZoneOperation_CS{}
	send.Curtype = proto.Uint32(Curtype)
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(Gameid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Zoneid = proto.Uint32(Zoneid)
	send.Modifygzonename = proto.String(modifygzonename)
	send.Modifygmlink = proto.String(modifygmlink)
	send.Modifystatus = proto.Uint32(modifystatus)
	curpage := uint32(unibase.Atoi(task.R.FormValue("Curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("Perpage"), 20))

	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Perpage = proto.Uint32(perpage)
	//task.Error("gameid:%d,curtype:%d",Gameid,Curtype)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGetUseritemList_CS(gmid uint32, task *unibase.ChanHttpTask) {
	Gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	Zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	pid := strings.TrimSpace(task.R.FormValue("pid"))
	send := &Pmd.RequestGetUserItemList_C{}
	send.Pid = proto.String(pid)
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(Gameid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Zoneid = proto.Uint32(Zoneid)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleModifyUseritem_CS(gmid uint32, task *unibase.ChanHttpTask) {
	Gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	Zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	item_id := uint32(unibase.Atoi(task.R.FormValue("item_id"), 0))
	modify_num := uint32(unibase.Atoi(task.R.FormValue("modify_num"), 0))
	pid := strings.TrimSpace(task.R.FormValue("pid"))
	send := &Pmd.ModifyUserItem_C{}
	send.Pid = proto.String(pid)
	send.Gmid = proto.Uint32(gmid)
	send.Itemid = proto.Uint32(item_id)
	send.ModifyNum = proto.Uint32(modify_num)
	send.Gameid = proto.Uint32(Gameid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Zoneid = proto.Uint32(Zoneid)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleEmptyUseritem_CS(gmid uint32, task *unibase.ChanHttpTask) {
	Gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	Zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	pid := strings.TrimSpace(task.R.FormValue("pid"))
	send := &Pmd.EmptyUserItem_C{}
	send.Pid = proto.String(pid)
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(Gameid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Zoneid = proto.Uint32(Zoneid)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleAddModifyTypes_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	typeid := uint32(unibase.Atoi(task.R.FormValue("typeid"), 0))
	typename := strings.TrimSpace(task.R.FormValue("typename"))
	send := &Pmd.AddModifyTypes_CS{}
	send.Typename = proto.String(typename)
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(gameid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Typeid = proto.Uint32(typeid)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGetModifyTypesList_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	curtype := uint32(unibase.Atoi(task.R.FormValue("curtype"), 0))
	send := &Pmd.GetModifyTypesList_CS{}
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(gameid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Curtype = proto.Uint32(curtype)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleModifyModifyTypesMessage_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	typeid := uint32(unibase.Atoi(task.R.FormValue("typeid"), 0))
	typename := strings.TrimSpace(task.R.FormValue("typename"))
	send := &Pmd.ModifyModifyTypesMessage_CS{}
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(gameid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Typeid = proto.Uint32(typeid)
	send.Typename = proto.String(typename)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleDeleteModifyTypesMessage_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	typeid := uint32(unibase.Atoi(task.R.FormValue("typeid"), 0))
	send := &Pmd.DeleteModifyTypesMessage_CS{}
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(gameid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Typeid = proto.Uint32(typeid)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmManagerActionRecordSearch_CS(gmid uint32, task *unibase.ChanHttpTask) {

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	managerid := uint64(unibase.Atoi(task.R.FormValue("managerid"), 0))
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))

	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}

	send := &Pmd.GetManagerActionRecordList_CS{}
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(gameid)
	send.Managerid = proto.Uint64(managerid)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmLimitIporcode_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	id := uint64(unibase.Atoi(task.R.FormValue("id"), 0))
	limittype := uint32(unibase.Atoi(task.R.FormValue("limittype"), 0))
	typeid := uint32(unibase.Atoi(task.R.FormValue("typeid"), 1))
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	content := task.R.FormValue("content")
	code := task.R.FormValue("code")

	send := &Pmd.ModifyLimitIporcode_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Gmid = proto.Uint32(gmid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Id = proto.Uint64(id)
	send.Clientid = proto.Uint64(task.GetId())
	send.Typeid = proto.Uint32(typeid)
	send.Optype = proto.Uint32(optype)
	send.Limittype = proto.Uint32(limittype)
	send.Code = proto.String(code)
	send.Content = proto.String(content)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Retcode = proto.Uint32(0)
	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmLimitIporcodeRecord_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	limittype := uint32(unibase.Atoi(task.R.FormValue("limittype"), 0))
	code := task.R.FormValue("code")
	starttime, _ := strconv.ParseUint(task.R.FormValue("starttime"), 10, 64)
	endtime, _ := strconv.ParseUint(task.R.FormValue("endtime"), 10, 64)
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))

	if endtime == 0 {
		endtime = uint64(unitime.Time.Sec())
	}
	if starttime == 0 {
		starttime = endtime - 6*3600
	}

	send := &Pmd.GetLimitIporcodeList_CS{}
	send.Gmid = proto.Uint32(gmid)
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Optype = proto.Uint32(optype)
	send.Limittype = proto.Uint32(limittype)
	send.Code = proto.String(code)
	send.Starttime = proto.Uint64(starttime)
	send.Endtime = proto.Uint64(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleUserOnlineSearch_CS(gmid uint32, task *unibase.ChanHttpTask) {

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	subgameid := uint32(unibase.Atoi(task.R.FormValue("game"), 0))

	charid := uint32(unibase.Atoi(task.R.FormValue("charid")))
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	gametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	regflag := uint32(unibase.Atoi(task.R.FormValue("regflag"), 0))
	rechargeflag := uint32(unibase.Atoi(task.R.FormValue("rechargeflag"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	send := &Pmd.StRequestOnlineListInfoPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)
	send.Gametype = proto.Uint32(gametype)
	send.Subgameid = proto.Uint32(subgameid)
	send.Regflag = proto.Uint32(regflag)
	send.Rechargeflag = proto.Uint32(rechargeflag)
	send.Subplatid = proto.Uint32(platid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleUserListVip_CS(gmid uint32, task *unibase.ChanHttpTask) {

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid")))
	viplevel := int32(unibase.Atoi(task.R.FormValue("viplevel"), -1))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StRequestVipListInfoPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Charid = proto.Uint32(charid)
	send.Viplevel = proto.Int32(viplevel)
	send.Optype = proto.Uint32(optype)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleCashOutAudit_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))
	orderid := uint64(unibase.Atoi(task.R.FormValue("orderid"), 0))
	status := uint32(unibase.Atoi(task.R.FormValue("status"), 0))
	ordertype := uint32(unibase.Atoi(task.R.FormValue("ordertype"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	charids := strings.TrimSpace(task.R.FormValue("charids"))
	opvalue := uint32(unibase.Atoi(task.R.FormValue("opvalue"), 0))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	regflag := uint32(unibase.Atoi(task.R.FormValue("regflag"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	realname := strings.TrimSpace(task.R.FormValue("realname"))
	cpf := strings.TrimSpace(task.R.FormValue("cpf"))

	send := &Pmd.StRequestConvertVerifyPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Charid = proto.Uint32(charid)
	send.Optype = proto.Uint32(optype)
	send.Orderid = proto.Uint64(orderid)
	send.Status = proto.Uint32(status)
	send.Ordertype = proto.Uint32(ordertype)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Regflag = proto.Uint32(regflag)
	send.Subplatid = proto.Uint32(platid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Charids = proto.String(charids)
	send.Opvalue = proto.Uint32(opvalue)
	send.Realname = proto.String(realname)
	send.Cpf = proto.String(cpf)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGameSlotsList_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("sub_game"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("type"), 1))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StRequestSlotsGameParamPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Subgameid = proto.Uint32(subgame)
	send.Subgametype = proto.Uint32(subgametype)
	send.Optype = proto.Uint32(optype)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	if optype == 2 {

		data := &Pmd.StRequestSlotsGameParamPmd_CS_SlotsInfo{}

		addrealpoolper := uint32(unibase.Atoi(task.R.FormValue("pool_ratio"), 0))
		bomblooptime := uint32(unibase.Atoi(task.R.FormValue("pool_cycle"), 0))
		standardchips := uint32(unibase.Atoi(task.R.FormValue("pool_gold"), 0))
		fakepoolmin := uint32(unibase.Atoi(task.R.FormValue("pool_ratio_lower"), 0))
		fakepoolmax := uint32(unibase.Atoi(task.R.FormValue("pool_ratio_upper"), 0))
		rebatevalue := uint32(unibase.Atoi(task.R.FormValue("reward"), 0))
		limitlow := uint32(unibase.Atoi(task.R.FormValue("lowest_gold"), 0))
		realpoolchips := uint32(unibase.Atoi(task.R.FormValue("realpoolchips"), 0))
		poolId := uint32(unibase.Atoi(task.R.FormValue("poolId"), 0))
		norechargertp := uint32(unibase.Atoi(task.R.FormValue("norechargertp"), 0))
		lowrechargertp := uint32(unibase.Atoi(task.R.FormValue("lowrechargertp"), 0))

		data.Subgameid = proto.Uint32(subgame)
		data.Subgametype = proto.Uint32(subgametype)

		data.Addrealpoolper = proto.Uint32(addrealpoolper)
		data.Bomblooptime = proto.Uint32(bomblooptime)
		data.Standardchips = proto.Uint32(standardchips)
		data.Fakepoolmin = proto.Uint32(fakepoolmin)
		data.Fakepoolmax = proto.Uint32(fakepoolmax)
		data.Rebatevalue = proto.Uint32(rebatevalue)
		data.Limitlow = proto.Uint32(limitlow)
		data.Realpoolchips = proto.Uint32(realpoolchips)
		data.PoolId = proto.Uint32(poolId)
		data.Norechargertp = proto.Uint32(norechargertp)
		data.Lowrechargertp = proto.Uint32(lowrechargertp)

		send.Datas = append(send.Datas, data)
	}
	if optype == 3 {
		sysrtp := uint32(unibase.Atoi(task.R.FormValue("system_rtp"), 0))

		regrtp := uint32(unibase.Atoi(task.R.FormValue("system_rtp_nolaunch"), 0))

		send.Sysrtp = proto.Uint32(sysrtp)
		send.Regrtp = proto.Uint32(regrtp)
	}

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleHundredGameList_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("sub_game"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype")))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StRequestHundredGameParamPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Subgameid = proto.Uint32(subgame)
	send.Subgametype = proto.Uint32(subgametype)
	send.Optype = proto.Uint32(optype)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	if optype == 2 {

		data := &Pmd.StRequestHundredGameParamPmd_CS_GameInfo{}

		tarstock := uint32(unibase.Atoi(task.R.FormValue("target_stock"), 0))
		srcstock := uint32(unibase.Atoi(task.R.FormValue("real_stock"), 0))
		cutper := uint32(unibase.Atoi(task.R.FormValue("pumping_ratio"), 0))
		decaytime := uint32(unibase.Atoi(task.R.FormValue("attenuation_form"), 0))
		decayratio := uint32(unibase.Atoi(task.R.FormValue("attenuation_ratio"), 0))
		limitchips := uint32(unibase.Atoi(task.R.FormValue("lowest_gold"), 0))
		decaytype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))

		data.Subgameid = proto.Uint32(subgame)
		data.Subgametype = proto.Uint32(subgametype)
		data.Tarstock = proto.Uint32(tarstock)
		data.Srcstock = proto.Uint32(srcstock)
		data.Cutper = proto.Uint32(cutper)
		data.Decaytime = proto.Uint32(decaytime)
		data.Decayratio = proto.Uint32(decayratio)
		data.Limitchips = proto.Uint32(limitchips)
		data.Decaytype = proto.Uint32(decaytype)

		send.Datas = append(send.Datas, data)
	}
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmStandAloneList_CS(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("sub_game"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype")))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StRequestSignleGameParamPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Subgameid = proto.Uint32(subgame)
	send.Subgametype = proto.Uint32(subgametype)
	send.Optype = proto.Uint32(optype)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	if optype == 2 {

		data := &Pmd.StRequestSignleGameParamPmd_CS_GameInfo{}

		tarstock := uint32(unibase.Atoi(task.R.FormValue("target_stock"), 0))
		srcstock := uint32(unibase.Atoi(task.R.FormValue("real_stock"), 0))
		cutper := uint32(unibase.Atoi(task.R.FormValue("pumping_ratio"), 0))
		decaytime := uint32(unibase.Atoi(task.R.FormValue("attenuation_form"), 0))
		decayratio := uint32(unibase.Atoi(task.R.FormValue("attenuation_ratio"), 0))
		addrealpoolper := uint32(unibase.Atoi(task.R.FormValue("pool_ratio"), 0))
		bomblooptime := uint32(unibase.Atoi(task.R.FormValue("pool_cycle"), 0))
		standardchips := uint32(unibase.Atoi(task.R.FormValue("pool_gold"), 0))
		fakepoolmin := uint32(unibase.Atoi(task.R.FormValue("pool_ratio_lower"), 0))
		fakepoolmax := uint32(unibase.Atoi(task.R.FormValue("pool_ratio_upper"), 0))
		rebatevalue := uint32(unibase.Atoi(task.R.FormValue("reward"), 0))
		realpoolchips := uint32(unibase.Atoi(task.R.FormValue("realpoolchips"), 0))
		poolId := uint32(unibase.Atoi(task.R.FormValue("poolId"), 0))
		decaytype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))

		limitchips := uint32(unibase.Atoi(task.R.FormValue("lowest_gold"), 0))

		data.Subgameid = proto.Uint32(subgame)
		data.Subgametype = proto.Uint32(subgametype)
		data.Tarstock = proto.Uint32(tarstock)
		data.Srcstock = proto.Uint32(srcstock)
		data.Cutper = proto.Uint32(cutper)
		data.Decaytime = proto.Uint32(decaytime)
		data.Decayratio = proto.Uint32(decayratio)
		data.Addrealpoolper = proto.Uint32(addrealpoolper)
		data.Bomblooptime = proto.Uint32(bomblooptime)
		data.Standardchips = proto.Uint32(standardchips)
		data.Fakepoolmin = proto.Uint32(fakepoolmin)
		data.Fakepoolmax = proto.Uint32(fakepoolmax)
		data.Rebatevalue = proto.Uint32(rebatevalue)
		data.Realpoolchips = proto.Uint32(realpoolchips)
		data.PoolId = proto.Uint32(poolId)
		data.Decaytype = proto.Uint32(decaytype)

		data.Limitchips = proto.Uint32(limitchips)

		send.Datas = append(send.Datas, data)
	}
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogWinlose(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("sub_game"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))

	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	usertype := uint32(unibase.Atoi(task.R.FormValue("usertype"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	send := &Pmd.StWinLoseRecordGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Subgameid = proto.Uint32(subgame)
	send.Subgametype = proto.Uint32(subgametype)
	send.Usertype = proto.Uint32(usertype)
	send.Subplatid = proto.Uint32(platid)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmLogGoldChange(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("sub_game"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("uid"), 0))

	starttime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	endtime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StGameEnterGoldRecordGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Subgameid = proto.Uint32(subgame)
	send.Subgametype = proto.Uint32(subgametype)
	send.Charid = proto.Uint32(charid)

	send.Begintime = proto.Uint32(starttime)
	send.Endtime = proto.Uint32(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleGmLogInout(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("sub_game"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("uid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StGameEnterOutRecordGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Subgameid = proto.Uint32(subgame)
	send.Subgametype = proto.Uint32(subgametype)
	send.Charid = proto.Uint32(charid)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleGmLogMatchmaking(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("sub_game"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("uid"), 0))

	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	// if endtime == 0 {
	// 	endtime = uint32(unitime.Time.Sec())
	// }
	// if starttime == 0 {
	// 	starttime = endtime - 6*3600
	// }

	send := &Pmd.StGameHundredRecordGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Subgameid = proto.Uint32(subgame)
	send.Subgametype = proto.Uint32(subgametype)
	send.Charid = proto.Uint32(charid)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmLogRemedies(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	charid := uint32(unibase.Atoi(task.R.FormValue("uid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("nickname"))

	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	send := &Pmd.StBenefitRecordGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Subplatid = proto.Uint32(platid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogDesposit(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	charid := uint32(unibase.Atoi(task.R.FormValue("uid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("nickname"))

	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StPiggyBankRecordGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogSignin(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	charid := uint32(unibase.Atoi(task.R.FormValue("uid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("nickname"))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StSignInRecordGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmOrderManger(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	reqtype := uint32(unibase.Atoi(task.R.FormValue("reqtype"), 1))

	send := &Pmd.StReqRechargePlatInfoGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Optype = proto.Uint32(optype)
	send.Reqtype = proto.Uint32(reqtype)
	if optype == 2 || optype == 4 {
		id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
		status := uint32(unibase.Atoi(task.R.FormValue("status"), 0))

		send.Id = proto.Uint32(id)
		send.Status = proto.Uint32(status)
	}
	if optype == 3 {
		cashoutauto := uint32(unibase.Atoi(task.R.FormValue("cashoutauto"), 0))

		send.Cashoutauto = proto.Uint32(cashoutauto)
	}
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleUserSpreadInfo(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))

	// childcharid := uint32(unibase.Atoi(task.R.FormValue("childcharid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))

	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	// level := uint32(unibase.Atoi(task.R.FormValue("level"), 0))

	send := &Pmd.StPlayerPromoteInfoGmUserPmd_CS{}

	send.Charid = proto.Uint32(charid)
	// send.Childcharid = proto.Uint32(childcharid)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	// send.Level = proto.Uint32(level)

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmLogSpread(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	charname := strings.TrimSpace(task.R.FormValue("charname"))

	send := &Pmd.StPromoteLogGmUserPmd_CS{}

	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmLogSpreadCash(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))

	reqtype := uint32(unibase.Atoi(task.R.FormValue("reqtype"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.StWithdrawLogGmUserPmd_CS{}

	send.Uid = proto.Uint32(charid)
	send.Reqtype = proto.Uint32(reqtype)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleLuckyTurntableAction(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	optype := uint32(unibase.Atoi(task.R.FormValue("optype")))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StTurntableParamPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Optype = proto.Uint32(optype)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	if optype == 2 {

		data := &Pmd.StTurntableParamPmd_CS_GameInfo{}

		tarstock := uint32(unibase.Atoi(task.R.FormValue("target_stock"), 0))
		srcstock := uint32(unibase.Atoi(task.R.FormValue("real_stock"), 0))
		decaytime := uint32(unibase.Atoi(task.R.FormValue("attenuation_form"), 0))
		decayratio := uint32(unibase.Atoi(task.R.FormValue("attenuation_ratio"), 0))
		decaytype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))

		data.Tarstock = proto.Uint32(tarstock)
		data.Srcstock = proto.Uint32(srcstock)
		data.Decaytime = proto.Uint32(decaytime)
		data.Decayratio = proto.Uint32(decayratio)
		data.Decaytype = proto.Uint32(decaytype)

		send.Data = append(send.Data, data)
	}
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleWeekInfoList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("charname"))

	send := &Pmd.StRequestWeekInfoGmPmd_CS{}

	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmLogFootballUse(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	lotteryid := uint32(unibase.Atoi(task.R.FormValue("lotteryid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.StFootbalCouponLogGmUserPmd_CS{}

	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Lotteryid = proto.Uint32(lotteryid)
	send.Optype = proto.Uint32(optype)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogCurrency(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	charid := uint64(unibase.Atoi(task.R.FormValue("charid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	changetype := uint32(unibase.Atoi(task.R.FormValue("changetype"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	starttime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	endtime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))
	querytype := uint32(unibase.Atoi(task.R.FormValue("querytype"), 1))

	send := &Pmd.RequestUserItemsHistoryListGmUserPmd_CS{}

	send.Charid = proto.Uint64(charid)
	send.Charname = proto.String(charname)
	send.Begintime = proto.Uint32(starttime)
	send.Endtime = proto.Uint32(endtime)
	send.Querytype = proto.Uint32(querytype)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Changetype = proto.Uint32(changetype)
	send.Optype = proto.Uint32(optype)

	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleGmLogTiger(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	cointype := uint32(unibase.Atoi(task.R.FormValue("cointype"), 0))
	starttime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	endtime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))

	send := &Pmd.StSlotsLotterylogGmUserPmd_CS{}

	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)
	send.Begintime = proto.Uint32(starttime)
	send.Endtime = proto.Uint32(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Cointype = proto.Uint32(cointype)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogVipReword(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	gettype := uint32(unibase.Atoi(task.R.FormValue("gettype"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.StVipRewardlogGmUserPmd_CS{}

	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Gettype = proto.Uint32(gettype)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogWeekcardGet(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	charname := strings.TrimSpace(task.R.FormValue("charname"))
	gettype := uint32(unibase.Atoi(task.R.FormValue("gettype"), 0))
	starttime := uint32(unibase.Atoi(task.R.FormValue("starttime"), 0))
	endtime := uint32(unibase.Atoi(task.R.FormValue("endtime"), 0))

	send := &Pmd.StWeekCardlogGmUserPmd_CS{}

	send.Charid = proto.Uint32(charid)
	send.Charname = proto.String(charname)
	send.Begintime = proto.Uint32(starttime)
	send.Endtime = proto.Uint32(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Gettype = proto.Uint32(gettype)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleEarlyWarning(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	charname := strings.TrimSpace(task.R.FormValue("charname"))

	optype := uint32(unibase.Atoi(task.R.FormValue("optype")))
	controlvalue := uint32(unibase.Atoi(task.R.FormValue("controlvalue"), 0))
	rechargetype := uint32(unibase.Atoi(task.R.FormValue("rechargetype"), 0))
	controltype := uint32(unibase.Atoi(task.R.FormValue("controltype"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StAutoPunishControlWarnGmUserPmd_CS{}

	if optype == uint32(1) {

		charid := strings.TrimSpace(task.R.FormValue("charid"))

		charids := strings.Split(charid, ",")

		for _, v := range charids {

			a, _ := strconv.ParseUint(v, 10, 32)

			if a != 0 {
				data := &Pmd.StAutoPunishControlWarnGmUserPmd_CS_DataInfo{}

				data.Charid = proto.Uint32(uint32(a))

				data.Controlvalue = proto.Uint32(controlvalue)

				send.Datas = append(send.Datas, data)

				fmt.Println("send.datas:", send.Datas)

			}
		}

	} else {
		charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
		send.Charid = proto.Uint32(charid)
	}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Controlvalue = proto.Uint32(controlvalue)
	send.Rechargetype = proto.Uint32(rechargetype)
	send.Charname = proto.String(charname)
	send.Controltype = proto.Uint32(controltype)
	send.Subplatid = proto.Uint32(platid)

	send.Optype = proto.Uint32(optype)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogWarning(gmid uint32, task *unibase.ChanHttpTask) {
	// gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	// zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	// charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))

	// charname := strings.TrimSpace(task.R.FormValue("charname"))

	// controlvalue := uint32(unibase.Atoi(task.R.FormValue("controlvalue"), 0))
	// rechargetype := uint32(unibase.Atoi(task.R.FormValue("rechargetype"), 0))

	// curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	// perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	// send := &Pmd.StAutoPunishControlLogGmUserPmd_CS{}

	// send.Gameid = proto.Uint32(gameid)
	// send.Zoneid = proto.Uint32(zoneid)

	// send.Charid = proto.Uint32(charid)
	// send.Controlvalue = proto.Uint32(controlvalue)
	// send.Rechargetype = proto.Uint32(rechargetype)
	// send.Charname = proto.String(charname)

	// send.Gmid = proto.Uint32(gmid)
	// send.Clientid = proto.Uint64(task.GetId())

	// send.Curpage = proto.Uint32(curpage)
	// send.Perpage = proto.Uint32(perpage)

	// ForwardGmCommand(task, send, 0, 0, false)
}
func HandleControllOrder(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	id := uint32(unibase.Atoi(task.R.FormValue("id")))

	optype := uint32(unibase.Atoi(task.R.FormValue("optype")))
	controlvalue := uint32(unibase.Atoi(task.R.FormValue("controlvalue"), 0))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StAutoControlInfoGmUserPmd_CS{}

	if optype <= uint32(4) {

		if optype > uint32(1) {

			data := &Pmd.StAutoControlInfoGmUserPmd_CS_DataInfo{}

			data.Id = proto.Uint32(id)

			if optype == uint32(2) || optype == uint32(4) {

				rechargemin := uint32(unibase.Atoi(task.R.FormValue("rechargemin")))
				rechargemax := uint32(unibase.Atoi(task.R.FormValue("rechargemax")))

				data.Autocontrolvalue = proto.Uint32(controlvalue)
				data.Rechargemin = proto.Uint32(rechargemin)
				data.Rechargemax = proto.Uint32(rechargemax)

			}

			send.Datas = append(send.Datas, data)

		}

	} else if optype <= uint32(8) {
		if optype > uint32(5) {

			data := &Pmd.StAutoControlInfoGmUserPmd_CS_DataInfo{}

			data.Id = proto.Uint32(id)

			if optype == uint32(6) || optype == uint32(8) {

				for i := 1; i < 15; i++ {
					name := fmt.Sprintf("rechargemul%d", i)

					rechargemul, _ := strconv.ParseFloat(task.R.FormValue(name), 32)

					rechargemuls := proto.Float32(float32(rechargemul))

					data.Rechargemuls = append(data.Rechargemuls, *rechargemuls)
				}

				data.Autocontrolvalue = proto.Uint32(controlvalue)
				regflag := uint32(unibase.Atoi(task.R.FormValue("regflag"), 0))
				data.Regflag = proto.Uint32(regflag)

			}

			send.Datas = append(send.Datas, data)

		}
	} else {
		if optype > uint32(9) {
			data := &Pmd.StAutoControlInfoGmUserPmd_CS_DataInfo{}

			data.Id = proto.Uint32(id)

			rechargemin := uint32(unibase.Atoi(task.R.FormValue("goldin")))
			rechargemax := uint32(unibase.Atoi(task.R.FormValue("goldout")))

			data.Rechargemin = proto.Uint32(rechargemin)
			data.Rechargemax = proto.Uint32(rechargemax)

			send.Datas = append(send.Datas, data)
		}
	}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Id = proto.Uint32(id)
	send.Optype = proto.Uint32(optype)

	send.Optype = proto.Uint32(optype)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)

	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleGmOrderCashoutRank(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	regflag := uint32(unibase.Atoi(task.R.FormValue("regflag"), 0))
	ordertype := uint32(unibase.Atoi(task.R.FormValue("ordertype"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	send := &Pmd.StRechargeWithdrawRankGmUserPmd_CS{}

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Optype = proto.Uint32(optype)
	send.Regflag = proto.Uint32(regflag)
	send.Ordertype = proto.Uint32(ordertype)
	send.Subplatid = proto.Uint32(platid)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleUpdateUrl(gmid uint32, task *unibase.ChanHttpTask) {

	kf := strings.TrimSpace(task.R.FormValue("kf"))
	tg1 := strings.TrimSpace(task.R.FormValue("tg1"))
	tg2 := strings.TrimSpace(task.R.FormValue("tg2"))
	url := strings.TrimSpace(task.R.FormValue("url"))

	file, err := os.OpenFile(url, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)

	var data []byte
	if err != nil {
		data, _ = json.Marshal(map[string]string{"retdesc": "文件打开失败"})
	}

	con := fmt.Sprintf("{\"kf\":\"%s\" , \"tg1\":\"%s\" ,\"tg2\":\"%s\" }", kf, tg1, tg2)

	_, err = io.WriteString(file, con)
	if err != nil {
		data, _ = json.Marshal(map[string]string{"retdesc": "文件写入失败"})
	} else {
		data, _ = json.Marshal(map[string]string{"retdesc": "修改成功"})
	}

	task.SendBinary(data)

}
func HandleCustomerService(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	url := strings.TrimSpace(task.R.FormValue("url"))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))

	send := &Pmd.StCustomerServicePmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Optype = proto.Uint32(optype)

	if optype == 2 {
		data := &Pmd.StCustomerServicePmd_CS_Data{}
		data.Platid = proto.Uint32(platid)
		data.Url = proto.String(url)

		send.Data = append(send.Data, data)
	}
	ForwardGmCommand(task, send, 0, 0, false)

}

func HandleGmOrderCashList(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	regflag := uint32(unibase.Atoi(task.R.FormValue("regflag"), 0))
	ignorerechargechips := uint32(unibase.Atoi(task.R.FormValue("ignorerechargechips"), 0))
	rechargemin := uint32(unibase.Atoi(task.R.FormValue("rechargemin"), 0))
	rechargemax := uint32(unibase.Atoi(task.R.FormValue("rechargemax"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	reqtype := uint32(unibase.Atoi(task.R.FormValue("reqtype"), 0))

	send := &Pmd.StWithdrawPercentGmUserPmd_CS{}

	if starttime == "" && endtime == "" {
		// starttime = time.Now().AddDate(0, 0, -6).Format("2006-01-02")
		starttime = time.Now().Format("2006-01-02")
		endtime = time.Now().Format("2006-01-02")
	}

	starttime = fmt.Sprintf("%s 00:00:00", starttime)
	endtime = fmt.Sprintf("%s 23:59:59", endtime)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Regflag = proto.Uint32(regflag)
	send.Subplatid = proto.Uint32(platid)
	send.Reqtype = proto.Uint32(reqtype)
	send.Ignorerechargechips = proto.Uint32(ignorerechargechips)
	send.Rechargemin = proto.Uint32(rechargemin)
	send.Rechargemax = proto.Uint32(rechargemax)

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleInventoryControll(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("subgametype"), 1))

	send := &Pmd.StSlotsStockXSGmUserPmd_CS{}

	if optype == 2 {

		data := &Pmd.StSlotsStockXSGmUserPmd_CS_DataInfo{}

		if subgametype == 5 {
			for i := 1; i < 15; i++ {
				name := fmt.Sprintf("rechargemul%d", i)

				rechargemul, _ := strconv.ParseFloat(task.R.FormValue(name), 32)

				rechargemuls := proto.Float32(float32(rechargemul))

				data.Rechargemuls = append(data.Rechargemuls, *rechargemuls)
			}
			controlvalue := uint32(unibase.Atoi(task.R.FormValue("controlvalue"), 0))
			data.Controlvalue = proto.Uint32(controlvalue)
		}

		if subgametype == 4 {
			rechargemin := uint32(unibase.Atoi(task.R.FormValue("rechargemin"), 0))
			rechargemax := uint32(unibase.Atoi(task.R.FormValue("rechargemax"), 0))
			rtpxs := uint32(unibase.Atoi(task.R.FormValue("rtpxs"), 0))

			data.Rechargemin = proto.Uint32(rechargemin)
			data.Rechargemax = proto.Uint32(rechargemax)
			data.Rtpxs = proto.Uint32(rtpxs)
		}
		if subgametype < 4 {

			stockmin := uint32(unibase.Atoi(task.R.FormValue("rechargemin"), 0))
			stockmax := uint32(unibase.Atoi(task.R.FormValue("rechargemax"), 0))
			rtpxs := uint32(unibase.Atoi(task.R.FormValue("rtpxs"), 0))

			data.Stockmin = proto.Uint32(stockmin)
			data.Stockmax = proto.Uint32(stockmax)
			data.Rtpxs = proto.Uint32(rtpxs)
		}
		id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
		data.Id = proto.Uint32(id)
		send.Datas = append(send.Datas, data)

	}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Optype = proto.Uint32(optype)
	send.Subgametype = proto.Uint32(subgametype)

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleInventoryControlllist(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("subgametype"), 1))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))

	send := &Pmd.StSlotsStockTaxGmUserPmd_CS{}
	if optype == 2 {
		data := &Pmd.StSlotsStockTaxGmUserPmd_CS_DataInfo{}
		subgametype := uint32(unibase.Atoi(task.R.FormValue("subgametype"), 1))
		subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
		id := uint32(unibase.Atoi(task.R.FormValue("id")))
		taxpercent := uint32(unibase.Atoi(task.R.FormValue("taxpercent"), 0))
		curstocknum := uint32(unibase.Atoi(task.R.FormValue("curstocknum"), 0))

		data.Subgametype = proto.Uint32(subgametype)
		data.Taxpercent = proto.Uint32(taxpercent)
		data.Curstocknum = proto.Uint32(curstocknum)
		data.Subgameid = proto.Uint32(subgameid)
		data.Id = proto.Uint32(id)

		send.Datas = append(send.Datas, data)
	}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Optype = proto.Uint32(optype)
	send.Subgametype = proto.Uint32(subgametype)
	send.Subgameid = proto.Uint32(subgameid)

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleInventoryControlllistlog(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 9999999))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("subgametype"), 1))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	if starttime == "" && endtime == "" {
		h, _ := time.ParseDuration("-3h")
		starttime = time.Now().Add(h).Format("2006-01-02 15:04:05")
		endtime = time.Now().Format("2006-01-02 15:04:05")
	}

	send := &Pmd.StSlotsStockLogGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Subgametype = proto.Uint32(subgametype)
	send.Subgameid = proto.Uint32(subgameid)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogPump(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 9999999))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("subgametype"), 1))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	send := &Pmd.StSlotsTaxLogGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Subgametype = proto.Uint32(subgametype)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Subplatid = proto.Uint32(platid)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogRetained(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 9999999))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.StRechargeRetentionLogGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Endtime = proto.String(endtime)
	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmOrderCashoutLog(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	if starttime == "" && endtime == "" {
		h, _ := time.ParseDuration("-168h")
		starttime = time.Now().Add(h).Format("2006-01-02 15:04:05")
		endtime = time.Now().Format("2006-01-02 15:04:05")
	}
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))

	send := &Pmd.RechargeExchangeLogGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Clientid = proto.Uint64(task.GetId())
	send.Optype = proto.Uint32(optype)
	send.Charid = proto.Uint32(charid)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Subplatid = proto.Uint32(platid)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleExclusiveRewards(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	opvalue := strings.TrimSpace(task.R.FormValue("opvalue"))
	interval := strings.TrimSpace(task.R.FormValue("interval"))
	begintime := strings.TrimSpace(task.R.FormValue("begintime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.RechargeRewardGmUserPmd_CS{}
	send.Optype = proto.Uint32(optype)

	strarr := strings.Split(opvalue, ",")

	for _, value := range strarr {

		n, _ := strconv.ParseUint(value, 10, 32)

		if n != 0 {
			send.Opvalue = append(send.Opvalue, uint32(n))
		}
	}

	send.Interval = proto.String(interval)
	send.Begintime = proto.String(begintime)
	send.Endtime = proto.String(endtime)

	if optype == 4 {
		data := &Pmd.RechargeRewardGmUserPmd_CS_DataInfo{}
		id := uint32(unibase.Atoi(task.R.FormValue("id")))

		chipsmin := uint32(unibase.Atoi(task.R.FormValue("chipsmin"), 0))
		chipsmax := uint32(unibase.Atoi(task.R.FormValue("chipsmax"), 0))

		data.Chipsmin = proto.Uint32(chipsmin)
		data.Chipsmax = proto.Uint32(chipsmax)
		data.Id = proto.Uint32(id)

		send.Data = append(send.Data, data)
	}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleExclusiveRewardslog(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.RechargeRewardLogGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleExclusiveRewardsgetlog(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	uid := uint32(unibase.Atoi(task.R.FormValue("uid")))
	isget := uint32(unibase.Atoi(task.R.FormValue("isget"), 0))
	interval := uint32(unibase.Atoi(task.R.FormValue("interval"), 1))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.PlayerRechargeRewardLogGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Charid = proto.Uint32(uid)
	send.Gettype = proto.Uint32(isget)
	send.Rewardid = proto.Uint32(interval)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Clientid = proto.Uint64(task.GetId())
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleCallbackUser(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	begintime := strings.TrimSpace(task.R.FormValue("begintime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	phones := strings.TrimSpace(task.R.FormValue("phones"))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))

	phondata := strings.Split(phones, " ")

	send := &Pmd.SmsPlayerReturnInfoGmUserPmd_CS{}

	send.Phonenum = phondata
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Begintime = proto.String(begintime)
	send.Endtime = proto.String(endtime)
	send.Clientid = proto.Uint64(task.GetId())
	send.Curpage = proto.Uint32(1)
	send.Perpage = proto.Uint32(perpage)
	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleGmLogFlowingWater(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	begintime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	gametype := uint32(unibase.Atoi(task.R.FormValue("gametype"), 0))
	subgameid := uint32(unibase.Atoi(task.R.FormValue("subgameid"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("subgametype"), 0))
	stocktype := uint32(unibase.Atoi(task.R.FormValue("stocktype"), 0))

	send := &Pmd.StWinLoseStatisticsGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Begintime = proto.String(begintime)
	send.Endtime = proto.String(endtime)
	send.Clientid = proto.Uint64(task.GetId())
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Gametype = proto.Uint32(gametype)
	send.Subgameid = proto.Uint32(subgameid)
	send.Subgametype = proto.Uint32(subgametype)
	send.Stocktype = proto.Uint32(stocktype)
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogGold(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	begintime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 500))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))

	send := &Pmd.GameDayChipsStatisicGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Begintime = proto.String(begintime)
	send.Endtime = proto.String(endtime)
	send.Clientid = proto.Uint64(task.GetId())
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmOrderAgain(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	begintime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))

	send := &Pmd.StDayRechargeStatisticGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Begintime = proto.String(begintime)
	send.Endtime = proto.String(endtime)
	send.Clientid = proto.Uint64(task.GetId())
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmFalseGold(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("type"), 0))
	datastr := strings.TrimSpace(task.R.FormValue("datastr"))
	id := uint32(unibase.Atoi(task.R.FormValue("id"), 0))
	reqtype := uint32(unibase.Atoi(task.R.FormValue("reqtype"), 1))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))

	send := &Pmd.StExcelHotupGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)

	send.Clientid = proto.Uint64(task.GetId())
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Optype = proto.Uint32(optype)
	send.Reqtype = proto.Uint32(reqtype)

	if optype == 2 {
		data := &Pmd.StExcelHotupGmUserPmd_CS_Data{}

		data.Id = proto.Uint32(id)
		dataarr := strings.Split(datastr, ",")

		for _, v := range dataarr {

			a, _ := strconv.ParseUint(v, 10, 64)

			va := proto.Uint64(uint64(a))

			data.Intdatas = append(data.Intdatas, *va)

		}
		send.Data = append(send.Data, data)
	}

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogJackpot(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	subgame := uint32(unibase.Atoi(task.R.FormValue("sub_game"), 0))
	subgametype := uint32(unibase.Atoi(task.R.FormValue("sessions"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("uid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StUserJackpotLogGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Subgameid = proto.Uint32(subgame)
	send.Subgametype = proto.Uint32(subgametype)
	send.Uid = proto.Uint32(charid)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLogUpload(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	status := uint32(unibase.Atoi(task.R.FormValue("status"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))
	imagetype := uint32(unibase.Atoi(task.R.FormValue("imagetype"), 1))

	send := &Pmd.StUserImageUploadGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Uid = proto.Uint32(charid)
	send.Optype = proto.Uint32(optype)
	send.Status = proto.Uint32(status)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Imagetype = proto.Uint32(imagetype)

	if optype == 1 {
		charids := strings.TrimSpace(task.R.FormValue("charids"))

		dataarr := strings.Split(charids, ",")

		for _, v := range dataarr {

			a, _ := strconv.ParseUint(v, 10, 64)

			if a != 0 {
				data := &Pmd.StUserImageUploadGmUserPmd_CS_Data{}
				data.Uid = proto.Uint64(a)
				send.Data = append(send.Data, data)
			}
		}

	}

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmMailrecordSend(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	uid := uint32(unibase.Atoi(task.R.FormValue("uid"), 0))

	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 0))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 50))

	send := &Pmd.StMailLogGmUserPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Uid = proto.Uint32(uid)

	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleUserSpreadRank(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))

	// starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	// endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	// optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))

	send := &Pmd.StRequestPromotionPmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	// send.Begintime = proto.String(starttime)
	// send.Endtime = proto.String(endtime)
	send.Charid = proto.Uint32(charid)
	// send.Optype = proto.Uint32(optype)

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandlleUserSpreadTurntable(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))

	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))

	send := &Pmd.RequestTurntablePmd_CS{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Optype = proto.Uint32(optype)

	send.Charid = proto.Uint32(charid)
	if optype == 2 {
		starttime := strings.TrimSpace(task.R.FormValue("starttime"))
		endtime := strings.TrimSpace(task.R.FormValue("endtime"))

		send.Begintime = proto.String(starttime)
		send.Endtime = proto.String(endtime)

	}

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleRedEnvelopeRain(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))
	period_start := uint32(unibase.Atoi(task.R.FormValue("period_start"), 0))
	period_end := uint32(unibase.Atoi(task.R.FormValue("period_end"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.RequestRedEnvelopeRainPmd_C{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Periodstart = proto.Uint32(period_start)
	send.Periodend = proto.Uint32(period_end)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	if optype == 2 {
		charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
		send.Charid = proto.Uint32(charid)
	}
	send.Optype = proto.Uint32(optype)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleBettingLevel(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))

	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.RequestBettingLevelPmd_C{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Charid = proto.Uint32(charid)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleRedemptionCode(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	batch := strings.TrimSpace(task.R.FormValue("batch"))
	expiretime := strings.TrimSpace(task.R.FormValue("expiretime"))
	codenum := uint32(unibase.Atoi(task.R.FormValue("codenum"), 0))
	coderepeatcount := uint32(unibase.Atoi(task.R.FormValue("coderepeatcount"), 0))
	batchtype := uint32(unibase.Atoi(task.R.FormValue("batchtype"), 0))
	gold := uint32(unibase.Atoi(task.R.FormValue("gold"), 0))

	codetype := uint32(unibase.Atoi(task.R.FormValue("codetype"), 1))
	childnum := uint32(unibase.Atoi(task.R.FormValue("childnum"), 0))
	actchildnum := uint32(unibase.Atoi(task.R.FormValue("actchildnum"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))

	send := &Pmd.RequestRedemptioncodePmd_C{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Batch = proto.String(batch)
	send.Expiretime = proto.String(expiretime)
	send.Codenum = proto.Uint32(codenum)
	send.Coderepeatcount = proto.Uint32(coderepeatcount)
	send.Batchtype = proto.Uint32(batchtype)
	send.Gold = proto.Uint32(gold)
	send.Codetype = proto.Uint32(codetype)
	send.Childnum = proto.Uint32(childnum)
	send.Actchildnum = proto.Uint32(actchildnum)
	send.Optype = proto.Uint32(optype)

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleRedemptionCodeSearch(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))

	batch := strings.TrimSpace(task.R.FormValue("batch"))
	code := strings.TrimSpace(task.R.FormValue("code"))

	send := &Pmd.RequestRedemptioncodeListPmd_C{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Batch = proto.String(batch)
	send.Code = proto.String(code)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleRedemptionCodeUseLog(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))

	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	code := strings.TrimSpace(task.R.FormValue("code"))

	send := &Pmd.RequestRedemptioncodeUsedPmd_C{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Charid = proto.Uint32(charid)
	send.Code = proto.String(code)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmLossRebateLog(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))

	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	rebatetype := uint32(unibase.Atoi(task.R.FormValue("rebatetype"), 0))

	send := &Pmd.RequestLossRebatePmd_C{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Charid = proto.Uint32(charid)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Rebatetype = proto.Uint32(rebatetype)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGmRewardActivitiesLog(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))

	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	receivetype := uint32(unibase.Atoi(task.R.FormValue("receivetype"), 0))

	send := &Pmd.RequestRewardActivitiesPmd_C{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	send.Charid = proto.Uint32(charid)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)
	send.Receivetype = proto.Uint32(receivetype)

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandelLuckyUser(gmid uint32, task *unibase.ChanHttpTask) {

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))

	batch := strings.TrimSpace(task.R.FormValue("batch"))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))
	lowcharge := uint32(unibase.Atoi(task.R.FormValue("lowcharge"), 0))
	topcharge := uint32(unibase.Atoi(task.R.FormValue("topcharge"), 0))
	usernum := uint32(unibase.Atoi(task.R.FormValue("usernum"), 0))

	lowgold := uint32(unibase.Atoi(task.R.FormValue("lowgold"), 0))
	topgold := uint32(unibase.Atoi(task.R.FormValue("topgold"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))

	codetype := uint32(unibase.Atoi(task.R.FormValue("codetype"), 0))

	lowchildnum := uint32(unibase.Atoi(task.R.FormValue("lowchildnum"), 0))
	topchildnum := uint32(unibase.Atoi(task.R.FormValue("topchildnum"), 0))
	lowactchildnum := uint32(unibase.Atoi(task.R.FormValue("lowactchildnum"), 0))
	topactchildnum := uint32(unibase.Atoi(task.R.FormValue("topactchildnum"), 0))

	send := &Pmd.RequestLuckyUserPmd_C{}

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)

	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Optype = proto.Uint32(optype)

	if optype == 0 || optype == 1 {
		data := &Pmd.RequestLuckyUserPmd_C_Storedata{}
		data.Starttime = proto.String(starttime)
		data.Endtime = proto.String(endtime)
		data.Lowcharge = proto.Uint32(lowcharge)
		data.Topcharge = proto.Uint32(topcharge)
		data.Codetype = proto.Uint32(codetype)
		data.Lowchildnum = proto.Uint32(lowchildnum)
		data.Topchildnum = proto.Uint32(topchildnum)
		data.Lowactchildnum = proto.Uint32(lowactchildnum)
		data.Topactchildnum = proto.Uint32(topactchildnum)

		if optype == 1 {
			data.Batch = proto.String(batch)
			data.Usernum = proto.Uint32(usernum)
			data.Lowgold = proto.Uint32(lowgold)
			data.Topgold = proto.Uint32(topgold)
		}

		send.Storedata = append(send.Storedata, data)
	} else {
		send.Batch = proto.String(batch)
		send.Charid = proto.Uint32(charid)
		send.Begintime = proto.String(starttime)
		send.Endtime = proto.String(endtime)
	}

	ForwardGmCommand(task, send, 0, 0, false)
}
func HandleGeneralActivites(gmid uint32, task *unibase.ChanHttpTask) {

	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	charid := uint32(unibase.Atoi(task.R.FormValue("charid"), 0))
	activename := uint32(unibase.Atoi(task.R.FormValue("activename"), 0))
	curpage := uint32(unibase.Atoi(task.R.FormValue("curpage"), 1))
	maxpage := uint32(unibase.Atoi(task.R.FormValue("Maxpage"), 10))
	perpage := uint32(unibase.Atoi(task.R.FormValue("perpage"), 10))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))
	starttime := strings.TrimSpace(task.R.FormValue("starttime"))
	endtime := strings.TrimSpace(task.R.FormValue("endtime"))

	send := &Pmd.RequestGeneralActivitiesPmd_C{}

	send.Charid = proto.Uint32(charid)
	send.Activity = proto.Uint32(activename)
	send.Optype = proto.Uint32(optype)
	send.Curpage = proto.Uint32(curpage)
	send.Perpage = proto.Uint32(perpage)
	send.Maxpage = proto.Uint32(maxpage)
	send.Begintime = proto.String(starttime)
	send.Endtime = proto.String(endtime)

	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandlePhonenumUpload(gmid uint32, task *unibase.ChanHttpTask) {
	file := strings.TrimSpace(task.R.FormValue("file"))
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))

	files := strings.Split(*proto.String(file), ",")

	send := &Pmd.StPhonenumUploadPmd_C{}

	for _, v := range files {
		a, _ := strconv.ParseUint(strings.TrimSpace(v), 10, 64)

		if a != 0 {
			data := &Pmd.StPhonenumUploadPmd_C_Data{}
			data.XId = proto.Uint64(a)

			send.Data = append(send.Data, data)

		}

	}
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)

}
func HandleUserControllList(gmid uint32, task *unibase.ChanHttpTask) {
	charid := strings.TrimSpace(task.R.FormValue("charid"))
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	zoneid := uint32(unibase.Atoi(task.R.FormValue("zoneid"), 0))
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 1))

	send := &Pmd.RequestControllerUserPmd_C_CS{}

	send.Charid = proto.String(charid)
	send.Optype = proto.Uint32(optype)
	send.Gameid = proto.Uint32(gameid)
	send.Zoneid = proto.Uint32(zoneid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}

func InitHttpMsgMapMain() {
	jsmsgMainMap = make(map[string]GmHandlerFunc)
	jsmsgMainMap["gm_game_priviliege"] = HandleGmPriviliege
	jsmsgMainMap["gmlogin"] = HandleGMLogin
	jsmsgMainMap["gamezone"] = HandleSelectGamezone
	jsmsgMainMap["gm_exec_str"] = HandleExecStr           //脚本指令
	jsmsgMainMap["gm_exec_func"] = HandleExecFunc         //内置命令
	jsmsgMainMap["gm_info_modify"] = HandleChangePassword //修改密码
	jsmsgMainMap["gm_change_userpwd"] = HandleChangeUserPassword
	jsmsgMainMap["gm_info_add"] = HandleAddGmAccount //添加GM账号
	jsmsgMainMap["gm_info_del"] = HandleDelGmAccount //删除GM账号
	jsmsgMainMap["user_info_search"] = HandleSearchUserInfo
	jsmsgMainMap["user_info_modify"] = HandleModifyUserInfo
	jsmsgMainMap["user_record_search"] = HandleSearchUserRecord
	jsmsgMainMap["gm_system_set"] = HandleSystemSet
	jsmsgMainMap["gm_code_generate"] = HandleCodeGenerate //礼包码
	jsmsgMainMap["gm_code_operator"] = HandleCodeOperator
	jsmsgMainMap["gm_broadcast_add"] = HandleBroadcastAdd
	jsmsgMainMap["gm_broadcast_list"] = HandleBroadcastList
	jsmsgMainMap["gm_order_list"] = HandleOrderList
	jsmsgMainMap["gm_punish_user"] = HandlePunishUser
	jsmsgMainMap["gm_punish_record"] = HandlePunishList
	jsmsgMainMap["gm_punish_delete"] = HandlePunishDelete
	jsmsgMainMap["gm_feedback_record"] = HandleFeedbackList //问卷调查反馈
	jsmsgMainMap["gm_feedback_deal"] = HandleDealFeedback   //删除问卷
	jsmsgMainMap["gm_broadcast_delete"] = HandleBroadcastDelete
	// jsmsgMainMap["gm_point_report"] = HandlePointReport
	// jsmsgMainMap["gm_point_detail"] = HandlePointDetail

	jsmsgMainMap["gm_game_stock"] = HandleStockData
	jsmsgMainMap["gm_game_stock_mod"] = HandleModStockData

	jsmsgMainMap["gm_subgame_list"] = HandleSubgameList
	jsmsgMainMap["user_list_search"] = HandleSearchUserList
	jsmsgMainMap["gm_winning_list"] = HandleSearchWinningList
	jsmsgMainMap["gm_blackwithe_list"] = HandleSearchBlackWhiteList
	jsmsgMainMap["gm_blackwhitelist_add"] = HandleBlackWhitelistAdd
	jsmsgMainMap["gm_blackwhitelist_mod"] = HandleBlackWhitelistMod
	jsmsgMainMap["gm_blackwhitelist_del"] = HandleBlackWhitelistDel
	jsmsgMainMap["gm_lobby_history"] = HandleLobbyGameHistory
	jsmsgMainMap["gm_lobby_history_detail"] = HandleLobbyGameDetailHistory
	jsmsgMainMap["gm_mail_send"] = HandleSendGmMail                           //发送邮件
	jsmsgMainMap["gm_itemtype_search"] = HandleSearchItem                     //查找游戏的道具类型信息，用来发送邮件附件
	jsmsgMainMap["gm_loginrecord_research"] = HandleSearchLoginRecord         //登陆日志查询
	jsmsgMainMap["gm_consumerecord_research"] = HandleSearchConsumeRecord     //消耗类日志查询
	jsmsgMainMap["gm_actionrecord_research"] = HandleSearchActionRecord       //活动类日志查询
	jsmsgMainMap["gm_strengthenrecord_search"] = HandleSearchStrengthenRecord //强化类日志查询
	jsmsgMainMap["gm_mailrecord_search"] = HandleSearchMailRecord             //邮件日志查询
	jsmsgMainMap["gm_rankrecord_search"] = HandleSearchRankRecord             //排行榜日志查询
	jsmsgMainMap["gm_bossrecord_search"] = HandleSearchBossRecord             //boss日志查询
	jsmsgMainMap["gm_renamerecord_search"] = HandleSearchRenameRecord         //重命名日志查询
	jsmsgMainMap["gm_itemrecord_search"] = HandleSearchItemRecord             //查询道具、货币日志,棋牌
	jsmsgMainMap["gm_msg_push"] = HandleMessagePush                           //消息推送
	jsmsgMainMap["gm_relationship_search"] = HandleSearchRelationShip
	jsmsgMainMap["gm_groupinfo_search"] = HandleSearchGroupInfo
	jsmsgMainMap["gm_groupinfo_mod"] = HandleModGroupInfo
	jsmsgMainMap["gm_groupinfo_manage"] = HandleManageGroup
	jsmsgMainMap["gm_zone_list"] = HandleGmZonelist       //查询账号拥有权限的区服列表
	jsmsgMainMap["gm_mails_send"] = HandleSendGmMails     //发送批量条件式邮件
	jsmsgMainMap["gm_chat_msg"] = HandleSearchChatMessage //查询最近200条聊天记录
	jsmsgMainMap["gm_act_control"] = HandleActControl     //活动开关控制
	jsmsgMainMap["gm_act_control_ex"] = HandleActionControl
	jsmsgMainMap["gm_act_control_search"] = HandleActionControlSearch
	jsmsgMainMap["gm_server_list"] = HandleServerListSearch
	jsmsgMainMap["gm_broadcast_shutdown"] = HandleBroadcastShutdown
	jsmsgMainMap["gm_cpay"] = HandleCPay          //添加修改切支付
	jsmsgMainMap["gm_cpay_list"] = HandleCPayList //切支付配置列表
	jsmsgMainMap["DrawcashListGmUserPmd_CS"] = ParseDrawcashListGmUserPmd_CS
	jsmsgMainMap["DrawcashManagerGmUserPmd_CS"] = ParseDrawcashManagerGmUserPmd_CS
	jsmsgMainMap["gm_room_dissolve"] = HandleRoomDissolve //解散房间
	jsmsgMainMap["RedpackCodeSearchGmUserPmd_CS"] = HandleRedpackCodeSearch
	jsmsgMainMap["RedPackCodeOperateGmUserPmd_CS"] = HandleRedpackCodeOperate
	jsmsgMainMap["MatchAwardSearchGmUserPmd_CS"] = HandleMatchAwardSearch
	jsmsgMainMap["MatchAwardGmUserPmd_CS"] = HandleMatchAward
	jsmsgMainMap["MatchSignupSearchGmUserPmd_CS"] = HandleMatchSignupSearch
	jsmsgMainMap["MatchSignupOperatorGmUserPmd_CS"] = HandleMatchSignupOperator
	jsmsgMainMap["PackcodeInsertGmUserPmd_CS"] = HandlePackcodeInsert
	jsmsgMainMap["PackcodeSearchGmUserPmd_CS"] = HandlePackcodeSearch
	jsmsgMainMap["PackcodeRecordGmUserPmd_CS"] = HandlePackcodeRecord
	jsmsgMainMap["DownloadFile"] = HandleDownloadFile
	jsmsgMainMap["SetWinOrLoseGmUserPmd_CS"] = HandleSetWinOrLose
	jsmsgMainMap["RequestAddRedPacketGmUserPmd_CS"] = HandleAddRedPacket
	jsmsgMainMap["RequestRevRedPacketGmUserPmd_CS"] = HandleRevRedPacket
	jsmsgMainMap["AddNewZoneGmUserPmd_CS"] = HandleAddZone
	jsmsgMainMap["RequestGameDataGmUserPmd_CS"] = HandleGameData
	jsmsgMainMap["SystemClassifyControlGmUserPmd_CS"] = HandleSystemClassifyControl
	jsmsgMainMap["CurrentBetInfoGmUserPmd_CS"] = HandleSearchCurrentBetinfo
	jsmsgMainMap["ClassifyProfitExportGmUserPmd_CS"] = HandleExportClassifyProfit
	jsmsgMainMap["RobotBetControlGmUserPmd_CS"] = HandleRobotControl
	jsmsgMainMap["OperateRobotGmUserPmd_CS"] = HandleOperateRobot
	jsmsgMainMap["RequestRobotListGmUserPmd_CS"] = HandleSearchRobotList
	jsmsgMainMap["gm_game_add"] = HandleGmAddGame                           //GM用户添加游戏
	jsmsgMainMap["gm_game_del"] = HandleDelGame                             //GM用户删除游戏
	jsmsgMainMap["gm_userlist"] = HandleSearchGmUserListByGameId_C          //查找当前游戏用户通过gameid
	jsmsgMainMap["gm_get_gamelist"] = HandleSearchGmGameList_CS             //通过类型查找所有游戏和大区
	jsmsgMainMap["gm_modify_game"] = HandleModifyGmGameInfo_CS              //通过类型查找所有游戏和大区
	jsmsgMainMap["gm_zoneoperation"] = HandleZoneOperation_CS               //当前游戏大区操作
	jsmsgMainMap["user_backpack_item_list"] = HandleGetUseritemList_CS      //获取角色背包道具信息
	jsmsgMainMap["modify_user_backpack_item"] = HandleModifyUseritem_CS     //修改角色背包道具信息
	jsmsgMainMap["empty_user_backpack_item"] = HandleEmptyUseritem_CS       //清空角色背包道具信息
	jsmsgMainMap["add_modify_types"] = HandleAddModifyTypes_CS              //添加修改类型信息
	jsmsgMainMap["get_modify_typeslist"] = HandleGetModifyTypesList_CS      //获取修改类型信息列表
	jsmsgMainMap["modify_typesmessage"] = HandleModifyModifyTypesMessage_CS //修改  修改类型信息
	jsmsgMainMap["delete_modifytypes"] = HandleDeleteModifyTypesMessage_CS  //删除  修改类型信息

	jsmsgMainMap["gm_manager_action_record_search"] = HandleGmManagerActionRecordSearch_CS //查找管理员操作日志
	jsmsgMainMap["gm_limit_iporcode"] = HandleGmLimitIporcode_CS                           //添加限制ip/机器码
	jsmsgMainMap["gm_limit_iporcode_record"] = HandleGmLimitIporcodeRecord_CS              //获取限制ip/机器码记录
	jsmsgMainMap["online_list_info"] = HandleUserOnlineSearch_CS                           //获取在线玩家所在房间
	jsmsgMainMap["user_list_vip"] = HandleUserListVip_CS                                   //获取vip玩家信息
	jsmsgMainMap["cash_out_audit"] = HandleCashOutAudit_CS                                 //获取兑换列表
	jsmsgMainMap["game_slots_list"] = HandleGameSlotsList_CS                               // slots配置
	jsmsgMainMap["game_noslots_list"] = HandleHundredGameList_CS
	jsmsgMainMap["gm_stand_alone_list"] = HandleGmStandAloneList_CS
	jsmsgMainMap["gm_log_winlose"] = HandleGmLogWinlose         //输赢统计
	jsmsgMainMap["gm_log_gold_change"] = HandleGmLogGoldChange  //游戏中金币变化明细
	jsmsgMainMap["gm_log_inout"] = HandleGmLogInout             // 游戏进出日志
	jsmsgMainMap["gm_log_matchmaking"] = HandleGmLogMatchmaking //对局日志
	jsmsgMainMap["gm_log_remedies"] = HandleGmLogRemedies       // 救济金日志
	jsmsgMainMap["gm_log_desposit"] = HandleGmLogDesposit       //存钱罐日志
	jsmsgMainMap["gm_log_signin"] = HandleGmLogSignin           //签到 日志
	jsmsgMainMap["gm_order_manger"] = HandleGmOrderManger       // 支付,提现管理

	jsmsgMainMap["user_spread_info"] = HandleUserSpreadInfo            // 推广信息
	jsmsgMainMap["gm_log_spread"] = HandleGmLogSpread                  // 推广日志
	jsmsgMainMap["gm_log_spread_cash"] = HandleGmLogSpreadCash         // 推广日志
	jsmsgMainMap["user_spread_turntable"] = HandlleUserSpreadTurntable //邀请转盘

	jsmsgMainMap["lucky_turntable_action"] = HandleLuckyTurntableAction // 幸运转盘管理

	jsmsgMainMap["week_info_list"] = HandleWeekInfoList          //周卡列表
	jsmsgMainMap["gm_log_football_use"] = HandleGmLogFootballUse // 足球抽奖券日志
	jsmsgMainMap["gm_log_currency"] = HandleGmLogCurrency        // 货币日志
	jsmsgMainMap["gm_log_tiger"] = HandleGmLogTiger              //老虎机抽奖日志
	jsmsgMainMap["gm_log_vip_reword"] = HandleGmLogVipReword     // VIP奖励领取日志
	jsmsgMainMap["gm_log_weekcard_get"] = HandleGmLogWeekcardGet // 周卡奖励领取日志

	jsmsgMainMap["early_warning"] = HandleEarlyWarning   // 自动点控玩家
	jsmsgMainMap["gm_log_warning"] = HandleGmLogWarning  // 自动点控日志
	jsmsgMainMap["controll_order"] = HandleControllOrder //自动点控管理

	jsmsgMainMap["gm_order_cashout_rank"] = HandleGmOrderCashoutRank //充值提现排行
	jsmsgMainMap["update_server_url"] = HandleUpdateUrl              //修改客服地址
	jsmsgMainMap["customer_service"] = HandleCustomerService
	jsmsgMainMap["gm_order_cash_list"] = HandleGmOrderCashList

	//库存
	jsmsgMainMap["inventory_controll"] = HandleInventoryControll
	jsmsgMainMap["inventory_controll_list"] = HandleInventoryControlllist
	jsmsgMainMap["inventory_controll_log_list"] = HandleInventoryControlllistlog
	//抽水日志
	jsmsgMainMap["gm_log_pump"] = HandleGmLogPump
	jsmsgMainMap["gm_log_retained"] = HandleGmLogRetained

	jsmsgMainMap["gm_order_cashout_log"] = HandleGmOrderCashoutLog
	jsmsgMainMap["exclusive_rewards"] = HandleExclusiveRewards
	jsmsgMainMap["exclusive_rewards_log"] = HandleExclusiveRewardslog
	jsmsgMainMap["exclusive_rewards_get_log"] = HandleExclusiveRewardsgetlog
	jsmsgMainMap["callback_user"] = HandleCallbackUser
	jsmsgMainMap["gm_log_flowing_water"] = HandleGmLogFlowingWater
	jsmsgMainMap["gm_log_gold"] = HandleGmLogGold
	jsmsgMainMap["gm_order_again"] = HandleGmOrderAgain
	jsmsgMainMap["gm_false_gold"] = HandleGmFalseGold
	jsmsgMainMap["gm_log_jackpot"] = HandleGmLogJackpot
	jsmsgMainMap["gm_log_upload"] = HandleGmLogUpload
	jsmsgMainMap["gm_mailrecord_send"] = HandleGmMailrecordSend
	jsmsgMainMap["user_info_update_password"] = HandleUserInfoUpdatePassword
	jsmsgMainMap["user_spread_rank"] = HandleUserSpreadRank
	jsmsgMainMap["red_envelope_rain"] = HandleRedEnvelopeRain
	jsmsgMainMap["gm_betting_level_log"] = HandleBettingLevel
	jsmsgMainMap["redemption_code"] = HandleRedemptionCode
	jsmsgMainMap["redemption_code_search"] = HandleRedemptionCodeSearch
	jsmsgMainMap["redemption_code_use_log"] = HandleRedemptionCodeUseLog
	jsmsgMainMap["gm_loss_rebate_log"] = HandleGmLossRebateLog
	jsmsgMainMap["gm_reward_activities_log"] = HandleGmRewardActivitiesLog
	jsmsgMainMap["gm_lucky_user"] = HandelLuckyUser
	jsmsgMainMap["gm_general_activites_log"] = HandleGeneralActivites
	jsmsgMainMap["gm_phonenum_upload"] = HandlePhonenumUpload
	jsmsgMainMap["user_controll_list"] = HandleUserControllList
}
