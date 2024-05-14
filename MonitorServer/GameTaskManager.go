package main

import (
	"git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type GameTaskManager struct {
	*nettask.TaskManager
	_timer_one_sec *unitime.Timer
}

func NewGameTaskManager() *GameTaskManager {
	return &GameTaskManager{
		TaskManager:    nettask.NewTaskManager(),
		_timer_one_sec: unitime.NewTimer(unitime.Time.Now(), 1*1000, true),
	}
}

func (self *GameTaskManager) AddZoneTaskManager(entry *ZoneTaskManager) bool {
	entry.Debug("AddZoneTaskManager")
	return self.AddEntry(entry)
}

func (self *GameTaskManager) RemoveZoneTaskManager(entry *ZoneTaskManager) {
	entry.Debug("RemoveEntry")
	self.RemoveEntry(entry)
}

func (self *GameTaskManager) GetZoneTaskManagerById(id uint64) *ZoneTaskManager {
	entry := self.GetEntryById(id)
	if entry == nil {
		task := NewZoneTaskManager()
		task.SetId(id)
		self.AddZoneTaskManager(task)
		return task
	}
	return entry.(*ZoneTaskManager)
}

func (self *GameTaskManager) ResetServerState(gameid uint32) {
	serverlist, ok := unibase.GlobalZoneListMap[gameid]
	if ok == true {
		for _, server := range serverlist.Zonelist {
			zoneTask := self.GetZoneTaskManagerById(uint64(serverlist.GetGameid())).GetZoneTaskById(unibase.GetGameZone(serverlist.GetGameid(), server.GetZoneid()))
			if zoneTask == nil {
				server.State = Pmd.ZoneState_Shutdown.Enum()
			} else {
				server.State = Pmd.ZoneState_Normal.Enum()
			}
		}
	}
}

func (self *GameTaskManager) CloseAll() {
	callback := func(v entry.EntryInterface) bool {
		task := v.(*ZoneTaskManager)
		task.CloseAll()
		return true
	}
	self.ExecEvery(callback)
	self.RemoveEntryAll()
}

func (self *GameTaskManager) Loop() bool {
	if self._timer_one_sec.Check(unitime.Time.Now()) == true {
		callback := func(v entry.EntryInterface) bool {
			ZoneTaskManager := v.(*ZoneTaskManager)
			ZoneTaskManager.Loop()
			return true
		}
		self.EntryManagerId.ExecEvery(callback)
	}
	return false
}
