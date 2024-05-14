package main

import (
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type LoginTaskManager struct {
	*nettask.TaskManager
	_timer_one_sec *unitime.Timer
}

func NewLoginTaskManager() *LoginTaskManager {
	return &LoginTaskManager{
		TaskManager:    nettask.NewTaskManager(),
		_timer_one_sec: unitime.NewTimer(unitime.Time.Now(), 1*1000, true),
	}
}

func (self *LoginTaskManager) AddLoginTask(entry *LoginTask) bool {
	entry.Debug("AddLoginTask")
	return self.AddEntry(entry)
}

func (self *LoginTaskManager) RemoveLoginTask(entry *LoginTask) {
	entry.Debug("RemoveEntry")
	self.RemoveEntry(entry)
}

func (self *LoginTaskManager) GetLoginTaskById(id uint64) *LoginTask {
	entry := self.GetEntryById(id)
	if entry == nil {
		return nil
	}
	return entry.(*LoginTask)
}

func (self *LoginTaskManager) CloseAll() {
	callback := func(v entry.EntryInterface) bool {
		task := v.(*LoginTask)
		task.RequestClose("shutdown request")
		task.Debug("shutdown close")
		return true
	}
	self.ExecEvery(callback)
	self.RemoveEntryAll()
}

func (self *LoginTaskManager) Loop() bool {
	if self._timer_one_sec.Check(unitime.Time.Now()) == true {
	}
	return false
}
