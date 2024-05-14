package main

import (
	"sync"
	"time"

	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type Session struct {
	sid        string                 //session id
	createtime int64                  //last access time
	value      map[string]interface{} //session store
	lock       sync.RWMutex
}

func NewSession(sid string, createtime int64) *Session {
	return &Session{sid: sid, createtime: createtime, value: make(map[string]interface{})}
}

// Set value to memory session
func (self *Session) Set(key string, value interface{}) error {
	self.lock.Lock()
	self.value[key] = value
	self.lock.Unlock()
	return nil
}

// Get value from memory session by key
func (self *Session) Get(key string) interface{} {
	self.lock.RLock()
	defer self.lock.RUnlock()
	if v, ok := self.value[key]; ok {
		return v
	}
	return nil
}

// Delete in memory session by key
func (self *Session) Delete(key string) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.value, key)
	return nil
}

// Flush clear all values in memory session
func (self *Session) Flush() error {
	self.lock.Lock()
	self.createtime = time.Now().Unix()
	self.value = make(map[string]interface{})
	self.lock.Unlock()
	return nil
}

// SessionID get this id of memory session store
func (self *Session) SessionID() string {
	return self.sid
}

type SessionManager struct {
	lock        sync.RWMutex // locker
	sessions    map[string]*Session
	maxlifetime int64
	clock_zero  *unitime.Clocker
}

var (
	defaultsm = NewSessionManager(3600 * 24 * 2)
)

func SessionSet(sid, key, value string) {
	if sid == "" {
		return
	}
	session := defaultsm.SessionInit(sid)
	session.Set(key, value)
}

func SessionGet(sid, key string) string {
	if sid == "" {
		return ""
	}
	session := defaultsm.SessionInit(sid)
	value := session.Get(key)
	if value == nil {
		return ""
	}
	if value_str, ok := value.(string); ok {
		return value_str
	}
	return ""
}

func SessionFlush(sid string) {
	if sid == "" {
		return
	}
	session := defaultsm.SessionInit(sid)
	session.Flush()
}

func DeleteSession(sid string) {
	defaultsm.SessionDelete(sid)
}

func NewSessionManager(maxlifetime int64) *SessionManager {
	return &SessionManager{maxlifetime: maxlifetime, sessions: make(map[string]*Session), clock_zero: unitime.NewClocker(unitime.Time.Now(), 0, 24*3600)}
}

func (self *SessionManager) SessionRead(sid string) *Session {
	self.lock.RLock()
	defer self.lock.RUnlock()
	if session, ok := self.sessions[sid]; ok {
		return session
	}
	return nil
}

func (self *SessionManager) SessionInit(sid string) *Session {
	self.lock.RLock()
	if session, ok := self.sessions[sid]; ok {
		self.lock.RUnlock()
		return session
	}
	self.lock.RUnlock()
	self.lock.Lock()
	session := NewSession(sid, time.Now().Unix())
	self.sessions[sid] = session
	self.lock.Unlock()
	return session
}

func (self *SessionManager) SessionGC() {
	if !self.clock_zero.Check(unitime.Time.Now()) {
		return
	}
	now := time.Now().Unix()
	logging.Debug("session gc at %d", now)
	for sid, session := range self.sessions {
		if session.createtime+self.maxlifetime < now {
			self.lock.Lock()
			delete(self.sessions, sid)
			self.lock.Unlock()
		}
	}
}

func (self *SessionManager) SessionDelete(sid string) {
	self.lock.Lock()
	delete(self.sessions, sid)
	self.lock.Unlock()
}

func (self *SessionManager) SessionCheck(sid string, createtime int64) *Session {
	session := self.SessionRead(sid)
	if session != nil && session.createtime == createtime {
		return session
	}
	return nil
}
