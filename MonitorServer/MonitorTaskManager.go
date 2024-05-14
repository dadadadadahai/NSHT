package main

import (
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
)

var (
	monitor_task_tempid = uint64(100000)
)

type MonitorTaskManager struct {
	*nettask.TaskManager
}

func NewMonitorTaskManager() *MonitorTaskManager {
	return &MonitorTaskManager{
		TaskManager: nettask.NewTaskManager(),
	}
}

func (self *MonitorTaskManager) AddMonitorTask(entry *MonitorTask) bool {
	entry.Debug("AddMonitorTask")
	return self.AddEntry(entry)
}

func (self *MonitorTaskManager) RemoveMonitorTask(entry *MonitorTask) {
	entry.Debug("RemoveMonitorTask")
	self.RemoveEntry(entry)
}

func (self *MonitorTaskManager) GetMonitorTaskById(id uint64) *MonitorTask {
	entry := self.GetEntryById(id)
	if entry == nil {
		return nil
	}
	return entry.(*MonitorTask)
}

func (self *MonitorTaskManager) CloseAll() {
	callback := func(v entry.EntryInterface) bool {
		task := v.(*MonitorTask)
		task.Close()
		task.Debug("shutdown close")
		return true
	}
	self.EntryManagerId.ExecEvery(callback)
}
