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

func InitTimeTick() {
	routine := goroutine.NewGoRoutine("TimeTick")
	routine.Start(func() {
		ticker_minute := time.NewTicker(time.Minute)
		ticker_msec50 := time.NewTicker(time.Millisecond * 50)
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
				HandleHttpGm(msg)
			case msg := <-ChanHttpRecive_api:
				HandleGmApi(msg)
			case <-ticker_msec50.C:
				unitime.Time.SetNow()
				zoneTaskManager.Loop()
				defaultsm.SessionGC()
			case <-ticker_minute.C:
				zoneTaskManager.PrintStatistics("ZoneTask")
				gmTaskManager.PrintStatistics("GmTask")
				unitime.Time.SetNow()
				DBPing()
				//unibase.ResetZoneList()
			case <-signal.ChanReload:
				unibase.InitConfig("GMServer", false, VERSION)
				unibase.ResetLogLevel()
				InitPushmsgConfig()
				logging.Info("reload data ok")
			case <-signal.ChanShutdown:
				signal.ChanRunning <- false
				break stop
			}
		}
		logging.Info("TimeTick stop")
	})
}
