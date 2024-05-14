package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/platcommon"

	apns "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"golang.org/x/net/http2"
)

var (
	TLSDialTimeout    = 15 * time.Second
	HTTPClientTimeout = 15 * time.Second
	DefaultHost       = "https://api.push.apple.com"
	PushMsgFormat     = `{"aps":{"alert":"%s","badge":%d,"sound":"default"}}`
	apnsm             = NewApnsManager()
)

func NewClient(certificate tls.Certificate) *apns.Client {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	if len(certificate.Certificate) > 0 {
		tlsConfig.BuildNameToCertificate()
	}
	transport := &http2.Transport{
		TLSClientConfig: tlsConfig,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return tls.DialWithDialer(&net.Dialer{Timeout: TLSDialTimeout}, network, addr, cfg)
		},
	}
	logging.Debug("NewClient %d", time.Now().Unix())
	return &apns.Client{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   HTTPClientTimeout,
		},
		Certificate: certificate,
		Host:        DefaultHost,
	}
}

type ApnsManager struct {
	m             sync.RWMutex
	canpush       bool
	done          bool
	ConfigMap     map[string]tls.Certificate
	ClientManager *apns.ClientManager
	ChanMsg       chan *ApnsMsg
}

type ApnsMsg struct {
	title string
	desc  string
	ts    uint64
	users []*Pmd.PushUserList
}

func NewApnsMsg(title, desc string, users []*Pmd.PushUserList) *ApnsMsg {
	return &ApnsMsg{
		title: title,
		desc:  desc,
		ts:    uint64(time.Now().Unix()),
		users: users,
	}
}

func (self *ApnsMsg) IsValid() bool {
	if (self.ts + 12) < uint64(time.Now().Unix()) {
		return false
	}
	return true
}

func (self *ApnsMsg) GetTitle() string {
	return self.title
}

func (self *ApnsMsg) GetDesc() string {
	return self.desc
}

func (self *ApnsMsg) GetData() []*Pmd.PushUserList {
	return self.users
}

func NewApnsManager() *ApnsManager {
	manager := &ApnsManager{
		ConfigMap:     make(map[string]tls.Certificate),
		ClientManager: apns.NewClientManager(),
		ChanMsg:       make(chan *ApnsMsg, 200),
		canpush:       false,
		done:          false,
	}
	manager.ClientManager.Factory = NewClient
	return manager
}

func (self *ApnsManager) Reset() {
	self.m.Lock()
	self.ConfigMap = make(map[string]tls.Certificate)
	self.ClientManager = apns.NewClientManager()
	self.canpush = false
	self.m.Unlock()
}

func (self *ApnsManager) AddConfig(bundleid, configname, passwd string) error {
	self.m.Lock()
	cert, err := certificate.FromP12File(configname, passwd)
	if err != nil {
		self.m.Unlock()
		return err
	}
	self.ConfigMap[bundleid] = cert
	self.canpush = true
	self.m.Unlock()
	return nil
}

func (self *ApnsManager) Empty() bool {
	self.m.Lock()
	defer self.m.Unlock()
	return len(self.ConfigMap) == 0
}

func (self *ApnsManager) PushMsg(bundleid, token, msg string) error {
	self.m.Lock()
	defer self.m.Unlock()
	cert, ok := self.ConfigMap[bundleid]
	if !ok {
		return errors.New("PushMsg failed: bundle id not found")
	}
	client := self.ClientManager.Get(cert)
	if client == nil {
		return errors.New("PushMsg failed: get client error")
	}
	notification := &apns.Notification{}
	notification.Topic = bundleid
	notification.DeviceToken = token
	notification.Payload = []byte(msg)
	res, err := client.Production().Push(notification)
	if err != nil {
		return err
	} else if res == nil {
		return errors.New("Pushmsg failed: response is nil")
	} else if res.StatusCode != 200 {
		return errors.New(res.Reason)
	}
	return nil
}

func (self *ApnsManager) AddMsg(title, desc string, datas []*Pmd.PushUserList) {
	if !self.canpush || self.done || len(datas) == 0 {
		return
	}
	apnsmsg := NewApnsMsg(title, desc, datas)
	select {
	case <-time.After(time.Second * 2):
		logging.Debug("ApnsManager AddMsg timeout")
	case self.ChanMsg <- apnsmsg:
		logging.Debug("ApnsManager AddMsg success")
	}
}

func (self *ApnsManager) Loop() {
	for self.canpush {
		select {
		case <-time.After(time.Second * 2):
			logging.Debug("ApnsManager Loop ...")
		case v := <-self.ChanMsg:
			if v.IsValid() {
				self.PushMsgToIOSDevice(v.GetTitle(), v.GetDesc(), v.GetData())
			} else {
				logging.Warning("ApnsManager anpsmsg invalid")
			}
		}
	}
	self.done = true
	logging.Warning("ApnsManager Loop over")
}

func (self *ApnsManager) PushMsgToIOSDevice(title, content string, userlist []*Pmd.PushUserList) {
	for _, data := range userlist {
		imei := strings.TrimSpace(data.GetImei())
		if imei == "" || !strings.Contains(imei, ":") {
			continue
		}
		tmpstr := strings.SplitN(imei, ":", 2)
		if len(tmpstr) != 2 || tmpstr[1] == "" {
			continue
		}
		bundleid := tmpstr[1]
		if strings.Contains(bundleid, "_") {
			bundleid = strings.SplitN(bundleid, "_", 2)[0]
		}
		if len([]rune(content)) > 12 {
			content = string([]rune(content)[:12]) + "..."
		}
		msgnum := data.GetMsgnum()
		if msgnum == 0 {
			msgnum = 1
		}
		msg := fmt.Sprintf(PushMsgFormat, content, msgnum)
		err := self.PushMsg(bundleid, tmpstr[0], msg)
		if err != nil {
			logging.Error("PushMsgToIOSDevice error:%s, charid:%d, token:%s", err.Error(), data.GetCharid(), tmpstr[0])
		} else {
			logging.Info("PushMsgToIOSDevice success, charid:%d, token:%s", data.GetCharid(), tmpstr[0])
		}
	}
}

func InitPushmsgConfig() {
	configstr := config.GetConfigStr("pushmsgconfig")
	if configstr == "" {
		logging.Error("InitPushmsgConfig falied: bundle config not set")
		return
	}
	succ, err := InitBundleConfig(configstr)
	if succ == 0 || err != nil {
		logging.Error("InitPushmsgConfig falied")
	} else {
		logging.Info("InitPushmsgConfig success")
	}
}

func InitBundleConfig(bundleconfig string) (succ int, err error) {
	bundleconfig = strings.TrimSpace(bundleconfig)
	if bundleconfig == "" {
		return 0, errors.New("bundle config string is none")
	}
	bundlelist := strings.Split(bundleconfig, ",")
	apnsm.Reset()
	num := 0
	for _, bundles := range bundlelist {
		bundles = strings.TrimSpace(bundles)
		if bundles == "" {
			continue
		}
		bl := strings.Split(bundles, "|")
		if len(bl) != 3 {
			continue
		}
		err := apnsm.AddConfig(bl[0], bl[1], bl[2])
		if err == nil {
			num += 1
		}
	}
	go apnsm.Loop()
	return num, nil
}

func get_msgcrontab_table(gameid uint32) string {
	return fmt.Sprintf("gm_msgcrontab_%d", gameid)
}

func create_msgcrontab_table(gameid uint32) {
	tblname := get_msgcrontab_table(gameid)
	if check_table_exists(tblname) == true {
		return
	}
	//自增Id， 创建日期，区id，角色Id，角色名称，账号Id，imei, 类型(1单次推送，2周期性推送)，消息内容id，开始时间戳，结束时间戳，间隔时间，
	//循环次数,下次推送时间，成功推送次数，失败推送次数，当前状态(默认0，1成功，2失败, 3取消)
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		daynum int(10) NOT NULL default '0',
		zoneid int(10) NOT NULL default '0',
		charid bigint(20) NOT NULL default '0',
		charname varchar(64) NOT NULL default '',
		accid bigint(20) NOT NULL default '0',
		imei varchar(128) NOT NULL default '',
		type int(10) NOT NULL default '0',
		msgid bigint(20) NOT NULL default '0',
		startat int(10) NOT NULL default '0',
		endat int(10) NOT NULL default '0',
		interval int(10) NOT NULL default '0',
		looptimes int(10) NOT NULL default '0',
		nexttime int(10) NOT NULL default '0',
		succtimes int(10) NOT NULL default '0',
		failtimes int(10) NOT NULL default '0',
		status int(10) NOT NULL default '0',
		createdat int(10) NOT NULL default '0',
		updatedat int(10) NOT NULL default '0',
		primary key (id),
		key index_type (type),
		key index_daynum (daynum),
		key index_charid (zoneid,charid),
		key index_status (status),
		key index_nexttime (nexttime)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func get_msgcontent_table() string {
	return "gm_msgcontent"
}

func create_msgcontent_table() {
	tblname := get_msgcontent_table()
	if check_table_exists(tblname) == true {
		return
	}
	str := fmt.Sprintf(`
	create table IF NOT EXISTS %s (
		id bigint(20) NOT NULL AUTO_INCREMENT,
		title varchar(128) NOT NULL default '',
		content varchar(256) NOT NULL default '',
		primary key (id),
		unique key index_account (username)
	)engine=MyISAM default charset=utf8 collate=utf8_unicode_ci;
	`, tblname)
	_, err := db_gm.Exec(str)
	if err != nil {
		logging.Warning("db_gm create table err:%s, %s", err.Error(), str)
	} else {
		logging.Info("db_gm create table:%s", tblname)
	}
}

func save_pushmsg(title, content string) (int64, error) {
	str := fmt.Sprintf("insert into %s(title,content) values (?, ?)", get_msgcontent_table())
	result, err := db_gm.Exec(str, title, content)
	var msgid int64
	if err == nil {
		msgid, err = result.LastInsertId()
	}
	return msgid, err
}

func get_pushmsg(msgid int64) (title string, content string) {
	str := fmt.Sprintf("select title, content from %s where id=?", get_msgcontent_table())
	row := db_gm.QueryRow(str, msgid)
	if err := row.Scan(&title, &content); err != nil {
		logging.Error("get_pushmsg error:%s, msgid:%d", err.Error(), msgid)
	}
	return
}
