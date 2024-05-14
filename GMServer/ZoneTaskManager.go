package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/servercommon"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/nettask"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

type ZoneTaskManager struct {
	*nettask.TaskManager
	_timer_one_sec *unitime.Timer
	_clock_one_min *unitime.Clocker
}

func NewZoneTaskManager() *ZoneTaskManager {
	return &ZoneTaskManager{
		TaskManager:    nettask.NewTaskManager(),
		_timer_one_sec: unitime.NewTimer(unitime.Time.Now(), 1*1000, true),
		_clock_one_min: unitime.NewClocker(unitime.Time.Now(), 1, 60),
	}
}

func (self *ZoneTaskManager) AddZoneTask(entry *ZoneTask) bool {
	entry.Debug("AddZoneTask")
	return self.AddEntry(entry)
}

func (self *ZoneTaskManager) RemoveZoneTask(entry *ZoneTask) {
	entry.Debug("RemoveEntry")
	self.RemoveEntry(entry)
}

func (self *ZoneTaskManager) GetZoneTaskById(id uint64) *ZoneTask {
	entry := self.GetEntryById(id)
	if entry == nil {
		return nil
	}
	return entry.(*ZoneTask)
}

func (self *ZoneTaskManager) CheckZoneTaskBwPort(gameid uint32) bool {
	bwport := false
	callback := func(v entry.EntryInterface) bool {
		zonetask := v.(*ZoneTask)
		if zonetask.serverState.GetGamezone().GetGameid() == gameid {
			_, bwport = zonetask.NetTaskInterFace.(*nettask.NetTaskBw)
			return false
		}
		return true
	}
	self.EntryManagerId.ExecEvery(callback)
	return bwport
}

func (self *ZoneTaskManager) CheckGameOnline(gameid uint32) bool {
	exists := false
	callback := func(v entry.EntryInterface) bool {
		zonetask := v.(*ZoneTask)
		if zonetask.serverState.GetGamezone().GetGameid() == gameid {
			exists = true
			return false
		}
		return true
	}
	self.EntryManagerId.ExecEvery(callback)
	return exists
}

func (self *ZoneTaskManager) BroadcastToGame(gameid uint32, send proto.Message) {
	callback := func(v entry.EntryInterface) bool {
		task := v.(*ZoneTask)
		if task.serverState.GetGamezone().GetGameid() == gameid {
			task.SendCmd(send)
		}
		return true
	}
	self.ExecEvery(callback)
}

func (self *ZoneTaskManager) ResetServerState(gameid uint32) {
	serverlist, ok := unibase.GlobalZoneListMap[gameid]
	if ok == true {
		for _, server := range serverlist.Zonelist {
			gameTask := zoneTaskManager.GetZoneTaskById(unibase.GetGameZone(serverlist.GetGameid(), server.GetZoneid()))
			if gameTask == nil {
				server.State = Pmd.ZoneState_Shutdown.Enum()
			} else {
				server.State = Pmd.ZoneState_Normal.Enum()
			}
		}
	}
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
	}

	if self._clock_one_min.Check(unitime.Time.Now()) == true {
		self.BroadcastHttpLoop()
	}
	return false
}
func (self *ZoneTaskManager) MonitorLog(format string, v ...interface{}) {
	send := &Pmd.StErrorLogMonitorUserCmd_S{}
	send.Logger = proto.String("GMServer:" + fmt.Sprintf(format, v...))
	self.BroadcastOne(send)
}

func (self *ZoneTaskManager) BroadcastHttpLoop() {
	bSend := false
	zone_list := get_valid_zone()
	for _, zone_info := range zone_list {
		retl := checkNextBroadcast(zone_info.GameId, zone_info.ZoneId)
		for _, data := range retl {
			cmd := &Pmd.BroadcastNewGmUserPmd_C{}
			cmd.Gameid = proto.Uint32(zone_info.GameId)
			cmd.Zoneid = proto.Uint32(zone_info.ZoneId)
			cmd.Gmid = proto.Uint32(1)
			cmd.Data = data
			self.ForwardGmCommand(cmd, zone_info.GameId, zone_info.ZoneId)

		}
		//只发送一个区就可以了
		if false == bSend {
			bSend = true
			send := &Smd.RequestOnlineNumMonitorSmd_S{
				Timestamp: proto.Uint32(uint32(unitime.Time.Sec())),
			}
			self.ForwardGmCommand(send, zone_info.GameId, zone_info.ZoneId)
		}

	}
	logging.Info("[%d]BroadcastHttpLoop", time.Now().Unix())

}
func (self *ZoneTaskManager) ForwardGmCommand(send proto.Message, gameid, zoneid uint32) {
	gamekey, gmlinks := get_game_gmlink_and_key(gameid, zoneid)
	if len(gmlinks) == 0 && gameid != 1001 {
		logging.Error("ForwardGmCommand gmlinks == 0")
		return
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	body, err := json.Marshal(send)
	if err != nil {
		logging.Error("json Marshal error:%s", err.Error())
		return
	}
	sign := md5String(string(body) + gamekey)

	for _, gmlink := range gmlinks {
		url := gmlink + fmt.Sprintf("?sign=%s&do=%s", sign, reflect.TypeOf(send).String()[5:])
		iobody := strings.NewReader(string(body))
		req, err := http.NewRequest("POST", url, iobody)
		logging.Debug("GmTask Http request, url:%s, body:%s", url, string(body))
		if err != nil {
			logging.Error("HttpRequestGet NewRequest err:%s", err.Error())
			return
		}
		response, err := client.Do(req)
		if err == nil {
			defer response.Body.Close()
			var resdata []byte
			resdata, err = ioutil.ReadAll(response.Body)
			logging.Debug("GmTask Http send:%s, response:%s", unibase.GetProtoString(send.String()), resdata)
			var mapMsg map[string]interface{}
			err := json.Unmarshal(resdata, &mapMsg)
				if err != nil {
					logging.Error("json Unmarshal error, send:%s, rev:%s, err:%s", unibase.GetProtoString(send.String()), string(resdata), err.Error())
					return
				}
            if len(mapMsg) == 0{
                return;
            }
            if _, ok := mapMsg["error"]; !ok {
                return;
            }
			var do string;
			errcode := uint32(mapMsg["error"].(float64))
			//更新下区服信息,省去后期人工添加麻烦
			if errcode == 0 {
				do = mapMsg["do"].(string)
				if do == "serverlist"{
					game_id := uint32(mapMsg["gameid"].(float64))
					var lserver_list []interface{}
					lserver_list = mapMsg["data"].([]interface{})
					for _, zone_info := range lserver_list{
						zone_id := uint32(zone_info.(map[string]interface{})["serverid"].(float64))
						zone_name := zone_info.(map[string]interface{})["servername"].(string)
	
						tblname := "gm_zones"
						str := fmt.Sprintf("insert ignore into %s (gameid, zoneid, zonename, status) values (?,?,?,?)", tblname)
						_, err := db_gm.Exec(str, game_id, zone_id, zone_name, 1)
						if err != nil {
							logging.Error("insert gm_zone error:%s, sql:%s", err.Error(), str)
							return 
						}
                        //更新下名字
						str = fmt.Sprintf("update  %s set zonename=? where zoneid=?", tblname)
                        _, err = db_gm.Exec(str, zone_name,  zone_id)
						if err != nil {
							logging.Error("insert gm_zone error:%s, sql:%s", err.Error(), str)
							return 
						}
					}
				}

			}

		} else {
			logging.Error("Http Request error:%s", err.Error())
		}
	}
}
