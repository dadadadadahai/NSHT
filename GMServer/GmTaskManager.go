package main

import (
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
)

type GmTaskManager struct {
	*nettask.TaskManager
}

func NewGmTaskManager() *GmTaskManager {
	return &GmTaskManager{
		TaskManager: nettask.NewTaskManager(),
	}
}

func (self *GmTaskManager) AddGmTask(entry *GmTask) bool {
	entry.Debug("AddGmTask")
	return self.AddEntry(entry)
}

func (self *GmTaskManager) RemoveGmTask(entry *GmTask) {
	entry.Debug("RemoveGmTask")
	self.RemoveEntry(entry)
}

func (self *GmTaskManager) GetGmTaskById(id uint64) *GmTask {
	entry := self.GetEntryById(id)
	if entry == nil {
		return nil
	}
	return entry.(*GmTask)
}

func (self *GmTaskManager) CloseAll() {
	callback := func(v entry.EntryInterface) bool {
		task := v.(*GmTask)
		task.RequestClose("shutdown close")
		task.Debug("shutdown close")
		return true
	}
	self.EntryManagerId.ExecEvery(callback)
}
