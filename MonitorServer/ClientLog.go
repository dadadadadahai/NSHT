package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
)

var (
	clientlogger    = logging.NewLogger()
	clientloghandle *logging.TimeRotationHandler
)

func InitClientLogFile() (err error) {

	clientfilename := config.GetConfigStr("clientlogfilename")

	curpwd, _ := os.Getwd()

	if runtime.GOOS == "windows" {

		if string(clientfilename[1]) != ":" {

			clientfilename = curpwd + "/" + clientfilename
		}
	} else {

		if string(clientfilename[0]) != "/" {

			clientfilename = curpwd + "/" + clientfilename
		}
	}

	clienthandle, err := logging.NewTimeRotationHandler(clientfilename, "060102-15")

	if err != nil {

		logging.Error("log init fail:%s,shutdown", err.Error())
		return err
	} else {

		clientloghandle = clienthandle
		//clienthandle.Async = false
		ResetClientLogLevel()

		clientlogger.AddHandler("CL", clienthandle)
		return nil
	}
	return nil

}

func ResetClientLogLevel() {
	if config.GetConfigStr("clientloglevel") != "" {
		clientloghandle.SetLevelString(config.GetConfigStr("clientloglevel"))

	} else {
		clientloghandle.SetLevel(logging.DEBUG)
	}
}

func WriteClientLog(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	text, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Debug("log err:%s", err.Error())
	}
	switch req.FormValue("level") {
	case "Error":
		clientlogger.Log(logging.ERROR, "%s,%s", req.RemoteAddr, text)
	case "Assert":
	case "Warning":
	case "Exception":
		clientlogger.Log(logging.WARNING, "%s,%s", req.RemoteAddr, text)
	case "Info":
		clientlogger.Log(logging.WARNING, "%s,%s", req.RemoteAddr, text)
	case "Debug":
	case "Log":
	default:
		clientlogger.Log(logging.DEBUG, "%s,%s", req.RemoteAddr, text)
	}
}

func ViewClientLog(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	file, err := os.Open(config.GetConfigStr("clientlogfilename"))
	if err != nil {
		w.Write([]byte("open file err:" + config.GetConfigStr("clientlogfilename")))
		return
	}
	text, _ := ioutil.ReadAll(file)
	w.Write([]byte(text))
}
