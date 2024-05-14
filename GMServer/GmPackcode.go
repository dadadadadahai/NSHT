package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	Pmd "git.code4.in/mobilegameserver/platcommon"
	"git.code4.in/mobilegameserver/unibase"
)

// 插入礼包码配置
func (self *GmTask) ParsePackcodeInsertGmUserPmd_CS(rev *Pmd.PackcodeInsertGmUserPmd_CS) bool {
	data := rev.GetData()
	rev.Retcode = proto.Uint32(0)
	if len(data) != 0 {
		if _, err := InsertPackcodeType(rev.GetGameid(), uint32(self.GetId()), data); err != nil {
			rev.Retcode = proto.Uint32(1)
			rev.Retdesc = proto.String(err.Error())
		} else {
			content := "添加礼包码配置"
			add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
		}
	}
	self.SendCmd(rev)
	return true
}

func InsertPackcodeType(gameid, gmid uint32, data []*Pmd.ItemInfo) (lastid int64, err error) {
	tbl := get_packcode_type_table()
	cont, err := json.Marshal(data)
	if err != nil {
		logging.Error("InsertPackcodeType Marshal error:%s", err.Error())
		return 0, err
	}
	str := fmt.Sprintf("insert into %s(gameid, content, created_at, gmid) values(?, ?, ?, ?)", tbl)
	result, err := db_gm.Exec(str, gameid, string(cont), time.Now().Unix(), gmid)
	if err == nil {
		lastid, err = result.LastInsertId()
	}
	if err != nil {
		logging.Error("InsertPackcodeType Exec error:%s, gameid:%d, content:%s", err.Error(), gameid, string(cont))
		return 0, err
	}
	return lastid, nil
}

// 查询礼包码配置列表
func (self *GmTask) ParsePackcodeSearchGmUserPmd_CS(rev *Pmd.PackcodeSearchGmUserPmd_CS) bool {
	tbl := get_packcode_type_table()
	str := fmt.Sprintf("select id, content, created_at from %s where gameid=?", tbl)
	rows, err := db_gm.Query(str, rev.GetGameid())
	if err != nil {
		self.Error("ParsePackcodeSearchGmUserPmd_CS Query error:%s, sql:%s, gameid:%d", err.Error(), str, rev.GetGameid())
		self.SendCmd(rev)
		return false
	}
	defer rows.Close()
	rev.Data = make([]*Pmd.PackcodeType, 0)
	for rows.Next() {
		cont, data := "", &Pmd.PackcodeType{}
		if err = rows.Scan(&data.Codetype, &cont, &data.Created); err != nil {
			self.Error("ParsePackcodeSearchGmUserPmd_CS Scan error:%s", err.Error())
			continue
		}
		err = json.Unmarshal([]byte(cont), &data.Items)
		if err != nil {
			self.Error("ParsePackcodeSearchGmUserPmd_CS Unmarshal error:%s, content:%s", err.Error(), cont)
			continue
		}
		rev.Data = append(rev.Data, data)
	}
	rev.Retcode = proto.Uint32(0)
	self.SendCmd(rev)
	return true
}

// 查询礼包码生成批次的历史记录
func (self *GmTask) ParsePackcodeRecordGmUserPmd_CS(rev *Pmd.PackcodeRecordGmUserPmd_CS) bool {
	rev.Retcode = proto.Uint32(0)
	tbl := get_packcode_record_table()
	str := fmt.Sprintf("select id, zoneidmax, zoneidmin, ctype, cnum, packid, platid, stime, etime, climit, cdesc, filename, created_at from %s where gameid=?", tbl)
	rows, err := db_gm.Query(str, rev.GetGameid())
	if err != nil {
		logging.Error("ParsePackcodeRecordGmUserPmd_CS Query error:%s, sql:%s, gameid:%d", err.Error(), str, rev.GetGameid())
		self.SendCmd(rev)
		return false
	}
	defer rows.Close()
	rev.Data = make([]*Pmd.PackcodeRecord, 0)
	for rows.Next() {
		data := &Pmd.PackcodeRecord{}
		if err = rows.Scan(&data.Recordid, &data.Zoneidmax, &data.Zoneidmin, &data.Codetype, &data.Codenum, &data.Packid, &data.Platid, &data.Stime, &data.Etime, &data.Limit, &data.Desc, &data.Filename, &data.Createdat); err != nil {
			logging.Error("ParsePackcodeRecordGmUserPmd_CS Scan error:%s", err.Error())
			continue
		}
		rev.Data = append(rev.Data, data)
	}
	self.SendCmd(rev)
	return true
}

// 请求生成礼包码
func (self *GmTask) ParseRequestGenerateCodeGmUserPmd_C(rev *Pmd.RequestGenerateCodeGmUserPmd_C) bool {
	self.Info("gm:%s ParseRequestGenerateCodeGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	if !self.VerifyOk {
		return false
	}
	s := time.Now().Unix()
	rev.Gmid = proto.Uint32(uint32(self.GetId()))
	codenum, filename := GenerateCode(rev)
	send := &Pmd.ReturnGenerateCodeGmUserPmd_S{}
	if codenum > 0 {
		send.Retcode = proto.Uint32(0)
		send.Retdesc = proto.String(fmt.Sprintf("Generate %d code success, filename:%s", codenum, filename))
		content := "添加礼包码，文件为：" + filename
		add_manager_action_record(self.Data.GetGameid(), uint32(self.GetId()), content)
	} else {
		send.Retcode = proto.Uint32(1)
		send.Retdesc = proto.String("Generate code error")
	}
	send.Codenum = proto.Uint32(codenum)
	send.Gmid = proto.Uint32(uint32(self.GetId()))
	self.SendCmd(send)
	e := time.Now().Unix()
	self.Info("ParseRequestGenerateCodeGmUserPmd_C use time:%d", e-s)
	return true
}

// 插入礼包码前，先生成记录
func InsertPackcodeRecord(data *Pmd.RequestGenerateCodeGmUserPmd_C) (lastid int64, err error) {
	tbl := get_packcode_record_table()
	str := fmt.Sprintf("insert into %s(gameid, zoneidmax, zoneidmin, ctype, cnum, packid, platid, climit, stime, etime, cdesc, gmid, created_at) values(?,?,?,?,?,?,?,?,?,?,?,?,?)", tbl)
	result, err := db_gm.Exec(str, data.GetGameid(), data.GetZoneidmax(), data.GetZoneidmin(), data.GetCodetype(), data.GetCodenum(), data.GetPackid(), data.GetPlatid(), data.GetLimit(), data.GetStime(), data.GetEtime(), data.GetDesc(), data.GetGmid(), time.Now().Unix())
	if err == nil {
		lastid, err = result.LastInsertId()
	}
	if err != nil {
		logging.Error("InsertPackcodeRecord error:%s", err.Error())
		return 0, err
	}
	return lastid, nil
}

func UpdatePackcodeRecord(id, num uint32, filename string) (err error) {
	tbl := get_packcode_record_table()
	str := fmt.Sprintf("update %s set filename=?, cnum=? where id=?", tbl)
	if _, err = db_gm.Exec(str, filename, num, id); err != nil {
		logging.Error("UpdatePackcodeRecord error:%s, filename:%s, codenum:%d, id:%d", err.Error(), filename, num, id)
	}
	return
}

// 生成礼包码
func GenerateCode(data *Pmd.RequestGenerateCodeGmUserPmd_C) (uint32, string) {
	lastid, err := InsertPackcodeRecord(data)
	if err != nil {
		return 0, err.Error()
	}

	codemap := make(map[string][]interface{})
	num, gameid, codenum, now := uint32(0), data.GetGameid(), data.GetCodenum(), uint32(time.Now().Unix())
	if codenum > 100000 {
		codenum = 100000
	}

	tmpname := fmt.Sprintf("code_%d_%d_%d_%d.log", gameid, data.GetCodetype(), uint32(lastid), now)
	filename := path.Join(config.GetConfigStr("logdir"), tmpname)
	fd, err := os.Create(filename)
	if err != nil {
		return 0, ""
	}
	defer fd.Close()

	r := NewRand()
	for i := uint32(0); i < codenum; i++ {
		tmpcode := RandStr(10, r)
		tblname := GetPackageCodeTable(tmpcode)
		codemap[tblname] = append(codemap[tblname], tmpcode)

		if len(codemap[tblname]) >= 500 {
			tmpnum := InsertCodeList(tblname, gameid, now, uint32(lastid), codemap[tblname])
			if tmpnum > 0 {
				WriteCodeToLog(fd, codemap[tblname])
				num += tmpnum
			}
			codemap[tblname] = codemap[tblname][:0]
		}

		if codenum >= 200 && i%50 == 0 {
			time.Sleep(time.Duration(1) * time.Nanosecond)
		}
	}

	for tblname, codelist := range codemap {
		if len(codelist) > 0 {
			tmpnum := InsertCodeList(tblname, gameid, now, uint32(lastid), codemap[tblname])
			if tmpnum > 0 {
				WriteCodeToLog(fd, codemap[tblname])
				num += tmpnum
			}
		}
	}
	codemap = nil
	UpdatePackcodeRecord(uint32(lastid), num, tmpname)
	return num, filename
}

// 插入多个礼包码
func InsertCodeList(tblname string, gameid, ctime, recordid uint32, codelist []interface{}) uint32 {
	num := len(codelist)
	if num == 0 {
		return 0
	}
	args := strings.Repeat("(?,?,?,?),", num)
	str := fmt.Sprintf("insert into %s(gameid, code, recordid, created_at) values %s", tblname, args[:len(args)-1])
	vals := make([]interface{}, 0)
	for _, code := range codelist {
		vals = append(vals, gameid, code, recordid, ctime)
	}
	_, err := db_gm.Exec(str, vals...)
	if err != nil {
		args = strings.Repeat("?,", num)
		str = fmt.Sprintf("select count(*) as num from %s where recordid=%d and gameid=%d and code in (%s)", tblname, recordid, gameid, args[:len(args)-1])
		if err = db_gm.QueryRow(str, codelist...).Scan(&num); err != nil {
			logging.Error("InsertCodeList result.Scan err:%s", err.Error())
			return 0
		}
	}
	return uint32(num)
}

// 重置礼包码
func (self *GmTask) ParseRequestOpeartorCodeGmUserPmd_C(rev *Pmd.RequestOpeartorCodeGmUserPmd_C) bool {
	self.Info("gm:%s ParseRequestOpeartorCodeGmUserPmd_C data:%s", self.GetName(), unibase.GetProtoString(rev.String()))
	if !self.VerifyOk {
		return false
	}
	send := &Pmd.ReturnOpreatorCodeGmUserPmd_S{}
	info := CheckoutPackcodeInfo(rev.GetGameid(), rev.GetCode())
	if info == nil {
		send.Retcode = proto.Uint32(1)
		send.Retdesc = proto.String("not found code")
	} else {
		send.Data = info
		send.Retcode = proto.Uint32(0)
	}
	if rev.GetOptype() == 2 && send.GetData().GetFlag() != 0 {
		str := fmt.Sprintf("update %s set uzoneid=0, uuid=0, utime=0, state=0 where id=?", GetPackageCodeTable(rev.GetCode()))
		db_gm.Exec(str, send.GetData().GetId())
		send.Data.Flag = proto.Uint32(0)
	}
	self.SendCmd(send)
	return true
}

func CheckoutPackcodeInfo(gameid uint32, code string) *Pmd.CodeInfo {
	tbl1 := GetPackageCodeTable(code)
	tbl2 := get_packcode_record_table()
	str := fmt.Sprintf("select a.id, code, a.uzoneid, a.uuid, a.utime, a.state, a.created_at, b.id, b.ctype, b.zoneidmax, b.zoneidmin, b.stime, b.etime, b.climit, b.packid, b.platid from %s as a inner join %s as b on a.recordid=b.id where a.gameid=? and code=?", tbl1, tbl2)
	result := db_gm.QueryRow(str, gameid, code)

	info := &Pmd.CodeInfo{}
	if err := result.Scan(&info.Id, &info.Code, &info.Uzoneid, &info.Ucharid, &info.Utime, &info.Flag, &info.Createtime, &info.Recordid, &info.Codetype, &info.Zoneidmax, &info.Zoneidmin, &info.Starttime, &info.Endtime, &info.Limit, &info.Packid, &info.Platid); err != nil {
		return nil
	}
	return info
}

// 使用礼包码
func UsePackageCode(gameid, zoneid uint32, userid uint64, code string, platid, packid uint32) (uint32, uint32) {
	info := CheckoutPackcodeInfo(gameid, code)
	if info == nil {
		return 1, 0
	}
	if info.GetFlag() == 1 {
		logging.Warning("UsePackageCode code:%s have been used", code)
		return 3, 0
	}
	stime, etime, now := info.GetStarttime(), info.GetEndtime(), uint32(time.Now().Unix())
	if etime > 0 && (etime <= now || stime > now) {
		logging.Warning("UsePackageCode code:%s stime:%d, now:%d, etime:%d", code, stime, now, etime)
		return 2, 0
	}
	zonemax, zonemin, limit := info.GetZoneidmax(), info.GetZoneidmin(), info.GetLimit()
	if ((zonemax != 0 && zonemax < zoneid) || (zonemin != 0 && zonemin > zoneid)) || (limit != 0 && limit <= CheckCodetypeUsed(gameid, userid, info.GetRecordid())) {
		return 4, 0
	}
	if (info.GetPlatid() != 0 && info.GetPlatid() != platid) || (info.GetPackid() != 0 && info.GetPackid() != packid) {
		return 5, 0
	}
	str := fmt.Sprintf("update %s set uzoneid=?, uuid=?, utime=?, state=1 where id=? and state=0", GetPackageCodeTable(code))
	_, err := db_gm.Exec(str, zoneid, userid, now, info.GetId())
	if err != nil {
		logging.Warning("UpdatePackageCode error :%s", err.Error())
		return 3, 0
	}
	return 0, info.GetCodetype()
}

func CheckoutItemsByCodetype(gameid, codetype uint32) (retl []*Pmd.ItemInfo) {
	if gameid == 0 || codetype == 0 {
		return
	}
	tbl := get_packcode_type_table()
	str := fmt.Sprintf("select content from %s where id=%d and gameid=%d", tbl, codetype, gameid)
	row := db_gm.QueryRow(str)
	var cont string = ""
	if err := row.Scan(&cont); err != nil {
		logging.Error("CheckoutItemsByCodetype no result, err:%s", err.Error())
		return
	}
	if cont != "" {
		retl = make([]*Pmd.ItemInfo, 0)
		if err := json.Unmarshal([]byte(cont), &retl); err != nil {
			logging.Error("CheckoutItemsByCodetype Unmarshal error:%s, content:%s", err.Error(), cont)
		}
	}
	return
}

func CheckCodetypeUsed(gameid uint32, userid uint64, recordid uint32) uint32 {
	var num uint32 = 0
	for i := 0; i < 10; i++ {
		tblname := fmt.Sprintf("gm_packcode_%d", i)
		str := fmt.Sprintf("select count(*) from %s where uuid=? and recordid=? and gameid=?", tblname)
		var count uint32 = 0
		db_gm.QueryRow(str, userid, recordid, gameid).Scan(&count)
		num += count
	}
	return num
}

func WriteCodeToLog(fd *os.File, codelist []interface{}) {
	for _, code := range codelist {
		line := fmt.Sprintf("%v\n", code)
		fd.WriteString(line)
	}
}

var PackcodeTblMap = map[int]string{
	0: "gm_packcode_0",
	1: "gm_packcode_1",
	2: "gm_packcode_2",
	3: "gm_packcode_3",
	4: "gm_packcode_4",
	5: "gm_packcode_5",
	6: "gm_packcode_6",
	7: "gm_packcode_7",
	8: "gm_packcode_8",
	9: "gm_packcode_9",
}

func GetPackageCodeTable(code string) string {
	var i int = 0
	for _, b := range []byte(code) {
		i += int(b)
	}
	return PackcodeTblMap[i%10]
}
