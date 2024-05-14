package main

import (
	"time"

	"code.google.com/p/goprotobuf/proto"

	// "git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	Smd "git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type ZoneTaskManager struct {
	*entry.Entry
	*nettask.TaskManager
	_timer_one_sec    *unitime.Timer
	_clock_one_min    *unitime.Clocker
	_clock_five_min   *unitime.Clocker //五分钟定时器，主要用来计算区服实时数据
	_clock_table_init *unitime.Clocker //23:47分创建第二天的数据表
	_clock_zero       *unitime.Clocker //0点执行报表计算
}

func NewZoneTaskManager() *ZoneTaskManager {
	return &ZoneTaskManager{
		Entry:             &entry.Entry{},
		TaskManager:       nettask.NewTaskManager(),
		_timer_one_sec:    unitime.NewTimer(unitime.Time.Now(), 1*1000, true),
		_clock_one_min:    unitime.NewClocker(unitime.Time.Now(), 1, 60),
		_clock_five_min:   unitime.NewClocker(unitime.Time.Now(), 1, 5*60),
		_clock_table_init: unitime.NewClocker(unitime.Time.Now(), (47*60 + 23*3600), 24*3600),
		_clock_zero:       unitime.NewClocker(unitime.Time.Now(), 0, 24*3600),
	}
}

func (self *ZoneTaskManager) AddZoneTask(entry *ZoneTask) bool {
	entry.Debug("AddZoneTask")
	return self.AddEntry(entry)
}

func (self *ZoneTaskManager) RemoveZoneTask(entry *ZoneTask) {
	entry.Debug("RemoveEntry")
	//下线的时候要设置下玩家的在线状态
	entry.ResetAllUserShutdownOnline()
	self.RemoveEntry(entry)
}

func (self *ZoneTaskManager) GetZoneTaskById(id uint64) *ZoneTask {
	entry := self.GetEntryById(id)
	if entry == nil {
		return nil
	}
	return entry.(*ZoneTask)
}

func (self *ZoneTaskManager) CloseAll() {
	callback := func(v entry.EntryInterface) bool {
		task := v.(*ZoneTask)
		task.RequestClose("shutdown request")
		task.Debug("shutdown close")
		return true
	}
	self.ExecEvery(callback)
	self.RemoveEntryAll()
}

func (self *ZoneTaskManager) Loop() bool {
	if self._timer_one_sec.Check(unitime.Time.Now()) == true {
		callback := func(v entry.EntryInterface) bool {
			zonetask := v.(*ZoneTask)
			zonetask.Loop()
			return true
		}
		self.EntryManagerId.ExecEvery(callback)
		if self._clock_one_min.Check(unitime.Time.Now()) == true {
			send := &Smd.RequestOnlineNumMonitorSmd_S{
				Timestamp: proto.Uint32(uint32(unitime.Time.Sec())),
			}
			logging.Info("[%d-%d]请求在线人数", self.GetId(), time.Now().Unix())
			self.Broadcast(send)
		}
		//定时五分钟计算实时数据
		// if self._clock_five_min.Check(unitime.Time.Now()) == true && config.GetConfigStr("realtimedata") == "true" {
		// 	callback := func(v entry.EntryInterface) bool {
		// 		zonetask := v.(*ZoneTask)
		// 		go zonetask.CalcRealtimeData()
		// 		return true
		// 	}
		// 	logging.Info("[%d-%d]计算实时数据", self.GetId(), time.Now().Unix())
		// 	self.EntryManagerId.ExecEvery(callback)
		// 	go self.CalcRealtimeData()
		// }
		//每天23：47分创建第二天的数据表
		if self._clock_table_init.Check(unitime.Time.Now()) == true {
			go self.CheckResetDatabase(1)
		}
		//0点报表计算
		// if self._clock_zero.Check(unitime.Time.Now()) == true {
		// 	callback := func(v entry.EntryInterface) bool {
		// 		zonetask := v.(*ZoneTask)
		// 		zonetask.ClockZero()
		// 		return false
		// 	}
		// 	logging.Info("[%d-%d]计算游戏报表", self.GetId(), time.Now().Unix())
		// 	go self.EntryManagerId.ExecEvery(callback)
		// }
	}
	return false
}

func (self *ZoneTaskManager) CheckResetDatabase(offset int) {
	for emptyday := -5; emptyday >= -30; emptyday-- {
		if check_empty_user_economic(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(emptyday))) == false {
			break
		}
		if check_empty_item_table(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(emptyday))) == false {
			break
		}
		if check_empty_action_record(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(emptyday))) == false {
			break
		}
		if check_empty_user_online(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(emptyday))) == false {
			break
		}
		if check_empty_user_detail(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(emptyday))) == false {
			break
		}
	}
	create_user_login(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset-1)))
	create_user_online(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset-1)))
	create_user_detail(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset-1)))
	create_user_online(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset)))
	create_user_detail(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset)))
	create_user_online_plat(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset-1)))
	create_user_online_plat(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset)))

	create_action_record(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset)))
	create_user_economic(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset)))
	create_user_item(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset)))

	create_realtime_data_table(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset-1)))
	create_realtime_data_table(uint32(self.GetId()), uint32(unitime.Time.YearMonthDay(offset)))
}

func (self *ZoneTaskManager) CalcRealtimeData() {
	CalcGameRealtimeData(uint32(self.GetId()))
}
