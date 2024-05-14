package main

//日期：2017-06-15
//描述：礼包码功能接口

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
)

// 礼包码管理
func HandleCodeGenerate(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	codenum := uint32(unibase.Atoi(task.R.FormValue("codenum"), 0))
	codetype := uint32(unibase.Atoi(task.R.FormValue("codetype"), 0))
	packageid := uint32(unibase.Atoi(task.R.FormValue("packageid"), 0))
	platid := uint32(unibase.Atoi(task.R.FormValue("platid"), 0))
	limit := uint32(unibase.Atoi(task.R.FormValue("limit"), 0))
	zoneidmin := uint32(unibase.Atoi(task.R.FormValue("zoneidmin"), 0))
	zoneidmax := uint32(unibase.Atoi(task.R.FormValue("zoneidmax"), 0))
	desc := strings.TrimSpace(task.R.FormValue("desc"))
	stime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("stime")), 10, 64)
	etime, _ := strconv.ParseUint(strings.TrimSpace(task.R.FormValue("etime")), 10, 64)

	send := &Pmd.RequestGenerateCodeGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Codenum = proto.Uint32(codenum)
	send.Codetype = proto.Uint32(codetype)
	send.Packid = proto.Uint32(packageid)
	send.Platid = proto.Uint32(platid)
	send.Desc = proto.String(desc)
	send.Limit = proto.Uint32(limit)
	send.Zoneidmin = proto.Uint32(zoneidmin)
	send.Zoneidmax = proto.Uint32(zoneidmax)
	send.Stime = proto.Uint64(stime)
	send.Etime = proto.Uint64(etime)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleCodeOperator(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	code := task.R.FormValue("code")
	optype := uint32(unibase.Atoi(task.R.FormValue("optype"), 0))
	send := &Pmd.RequestOpeartorCodeGmUserPmd_C{}
	send.Gameid = proto.Uint32(gameid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	send.Code = proto.String(code)
	send.Optype = proto.Uint32(optype)
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandlePackcodeInsert(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	items := strings.TrimSpace(task.R.FormValue("items"))
	if items == "" || gameid == 0 {
		task.SendBinary([]byte(`{"ret":1,retcode":1,"retdesc":"参数错误"}`))
		return
	}
	send := &Pmd.PackcodeInsertGmUserPmd_CS{}
	send.Data = make([]*Pmd.ItemInfo, 0)
	if err := json.Unmarshal([]byte(items), &send.Data); err != nil {
		logging.Error("HandlePackcodeInsert Unmarshal error:%s, items:%s", err.Error(), items)
		task.SendBinary([]byte(`{"ret":1,retcode":1,"retdesc":"参数错误"}`))
		return
	}
	send.Gameid = proto.Uint32(gameid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandlePackcodeSearch(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	send := &Pmd.PackcodeSearchGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())

	ForwardGmCommand(task, send, 0, 0, false)
}

func HandlePackcodeRecord(gmid uint32, task *unibase.ChanHttpTask) {
	gameid := uint32(unibase.Atoi(task.R.FormValue("gameid"), 0))
	send := &Pmd.PackcodeRecordGmUserPmd_CS{}
	send.Gameid = proto.Uint32(gameid)
	send.Gmid = proto.Uint32(gmid)
	send.Clientid = proto.Uint64(task.GetId())
	ForwardGmCommand(task, send, 0, 0, false)
}

func HandleDownloadFile(gmid uint32, task *unibase.ChanHttpTask) {
	filename := strings.TrimSpace(task.R.FormValue("filename"))
	if filename == "" {
		task.SendBinary([]byte("文件不存在!"))
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		task.Error("HandleDownloadFile Open error:%s, filename:%s", err.Error(), filename)
		task.SendBinary([]byte("下载文件出错!"))
		return
	}
	defer file.Close()

	filename = path.Base(filename)
	filename = url.QueryEscape(filename)
	task.W.Header().Set("Content-Type", "application/octet-stream")
	task.W.Header().Set("content-disposition", "attachment; filename=\""+filename+"\"")
	if _, err := io.Copy(task.W, file); err != nil {
		http.Error(task.W, err.Error(), http.StatusInternalServerError)
	}
}

func HandleDownloadFileHttp(w http.ResponseWriter, r *http.Request) {
	task := GetGMClientTask(r, 0)
	if task == nil {
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}
	filename := strings.TrimSpace(r.FormValue("filename"))
	if filename == "" || !strings.Contains(filename, "code_") {
		w.Write([]byte("文件不存在!"))
		return
	}
	logdir := config.GetConfigStr("downloadir")
	if logdir == "" {
		logdir = config.GetConfigStr("logdir")
	}
	fmt.Println("logdir", logdir)
	filename = path.Join(logdir, filename)
	file, err := os.Open(filename)
	if err != nil {
		logging.Error("HandleDownloadFile Open error:%s, filename:%s", err.Error(), filename)
		w.Write([]byte("下载文件出错!"))
		return
	}
	defer file.Close()

	filename = path.Base(filename)
	filename = url.QueryEscape(filename)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("content-disposition", "attachment; filename=\""+filename+"\"")
	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		task.Info("GM: %s download file: %s!", task.Name, filename)
	}
}
