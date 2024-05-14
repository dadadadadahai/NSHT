package main

import (
	"time"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/goroutine"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/signal"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

var (
	TimeTickStoped = true
)

func InitTimeTick() {
	routine := goroutine.NewGoRoutine("TimeTick")
	routine.Start(func() {
		ticker_minute := time.NewTicker(time.Minute)
		ticker_minute10 := time.NewTicker(time.Minute * 10)
		ticker_msec50 := time.NewTicker(time.Millisecond * 50)
		TimeTickStoped = false
	stop:
		for signal.ServerRunning {
			routine.SetRunning()
			unitime.Time.SetNow()
			select {
			case msg := <-nettask.ChanRecive:
				if msg.IsEof() {
					msg.RemoveMe()
				} else {
					msg.Parse()
				}
			case msg := <-unibase.ChanHttpRecive:
				HandleHttpChanCommand(msg)
			case <-ticker_msec50.C:
				unitime.Time.SetNow()
				gameTaskManager.Loop()
				HttpLoop() // 联运游戏，通过http发送数据
			case <-ticker_minute10.C:
				gameTaskManager.PrintStatistics("zoneTask")
				monitorTaskManager.PrintStatistics("monitorTask")
			case <-ticker_minute.C:
				unitime.Time.SetNow()
				db_monitor.Ping()
				unibase.ResetZoneList()
			case <-signal.ChanReload:
				unibase.InitConfig("MonitorServer", false, VERSION)
				unibase.ResetLogLevel()
				ResetClientLogLevel()
				logging.Info("reload data ok")
			case <-signal.ChanShutdown:
				signal.ChanRunning <- false
				break stop
			}
		}
		TimeTickStoped = true
		logging.Info("TimeTick stop")
	})
}
