package main

import (
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
)

type GMClientTaskManager struct {
	*nettask.TaskManager
}

func NewGMClientTaskManager() *GMClientTaskManager {
	return &GMClientTaskManager{
		TaskManager: nettask.NewTaskManager(),
	}
}

func (self *GMClientTaskManager) AddGMClientTask(entry *GMClientTask) bool {
	return self.AddEntry(entry)
}

func (self *GMClientTaskManager) RemoveGMClientTask(entry *GMClientTask) {
	self.RemoveEntry(entry)
}

func (self *GMClientTaskManager) GetGMClientTaskById(id uint64) *GMClientTask {
	entry := self.GetEntryById(id)
	if entry == nil {
		return nil
	}
	return entry.(*GMClientTask)
}

// 当停机时,主动通知所有已连接网关关闭连接
func (self *GMClientTaskManager) CloseAll() {
	callback := func(v entry.EntryInterface) bool {
		task := v.(*GMClientTask)
		task.RequestClose("shutdown request")
		return true
	}
	self.ExecEvery(callback)
	self.RemoveEntryAll()
}
