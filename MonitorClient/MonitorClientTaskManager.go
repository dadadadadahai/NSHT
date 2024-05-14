package main

import (
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
)

type MonitorClientTaskManager struct {
	*nettask.TaskManager
}

func NewMonitorClientTaskManager() *MonitorClientTaskManager {
	return &MonitorClientTaskManager{
		TaskManager: nettask.NewTaskManager(),
	}
}

func (self *MonitorClientTaskManager) AddMonitorClientTask(entry *MonitorClientTask) bool {
	return self.AddEntry(entry)
}

func (self *MonitorClientTaskManager) RemoveMonitorClientTask(entry *MonitorClientTask) {
	self.RemoveEntry(entry)
}

func (self *MonitorClientTaskManager) GetMonitorClientTaskById(id uint64) *MonitorClientTask {
	entry := self.GetEntryById(id)
	if entry == nil {
		return nil
	}
	return entry.(*MonitorClientTask)
}

// 当停机时,主动通知所有已连接网关关闭连接
func (self *MonitorClientTaskManager) CloseAll() {
	callback := func(v entry.EntryInterface) bool {
		task := v.(*MonitorClientTask)
		task.RequestClose("shutdown request")
		return true
	}
	self.ExecEvery(callback)
	self.RemoveEntryAll()
}
