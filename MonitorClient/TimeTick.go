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
		for {
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
			case <-ticker_minute.C:
				unitime.Time.SetNow()
				DBPing()
				defaultsm.SessionGC()
			case <-signal.ChanReload:
				unibase.InitConfig("MonitorClient", false, VERSION)
				unibase.ResetLogLevel()
				logging.Info("reload data ok")
			case <-signal.ChanShutdown:
				signal.ChanRunning <- false
				break stop
			}
		}
		logging.Info("TimeTick stop")
	})
}
