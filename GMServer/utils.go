package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
)

func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandStr(n int, r *rand.Rand) string {
	str := "123456789abcdefghkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ"
	strlen := len(str)
	result := []byte{}

	if r == nil {
		r = NewRand()
	}
	for i := 0; i < n; i++ {
		result = append(result, str[r.Intn(strlen)])
	}

	return string(result)
}

func RandNStr(num, strlen int) []string {
	r := NewRand()
	result := make([]string, num)
	for i := 0; i < num; i++ {
		result[i] = RandStr(strlen, r)
		if num >= 200 && i%50 == 0 {
			time.Sleep(time.Duration(1) * time.Nanosecond)
		}
	}
	return result
}

func SetCookie(w http.ResponseWriter, name, value string) {
	cookie := &http.Cookie{Name: name, Value: value, Path: "/"}
	http.SetCookie(w, cookie)
}

func GetCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func ClearCookie(w http.ResponseWriter, r *http.Request, name string) {
	cookie, err := r.Cookie(name)
	if err != nil || cookie == nil {
		return
	}
	cookie.Value = ""
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

func EncodeSign(secretkey, value string) string {
	t := strconv.FormatInt(time.Now().Unix(), 10)
	s := unibase.Rand.RandString(32)
	v := base64.URLEncoding.EncodeToString([]byte(value))
	h := hmac.New(sha1.New, []byte(secretkey))
	fmt.Fprintf(h, "%s%s%s", v, s, t)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	result := strings.Join([]string{v, s, t, sig}, "|")
	return base64.URLEncoding.EncodeToString([]byte(result))
}

func DecodeSign(secretkey, value string) string {
	if value == "" {
		return value
	}
	data, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return ""
	}
	value = string(data)
	parts := strings.Split(value, "|")
	if len(parts) != 4 {
		return ""
	}
	v, s, t, sig := parts[0], parts[1], parts[2], parts[3]

	h := hmac.New(sha1.New, []byte(secretkey))
	fmt.Fprintf(h, "%s%s%s", v, s, t)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return ""
	}
	vs, _ := base64.URLEncoding.DecodeString(v)
	return string(vs)
}

func SetSecureCookie(w http.ResponseWriter, name, value string) {
	secretkey1 := config.GetConfigStr("secretkey1")
	cookie := EncodeSign(secretkey1, value)
	secretkey2 := config.GetConfigStr("secretkey2")
	cookie = EncodeSign(secretkey2, cookie)
	SetCookie(w, name, cookie)
}

func GetSecureCookie(r *http.Request, name string) string {
	cookie := GetCookie(r, name)
	if cookie == "" {
		return ""
	}
	cookie = DecodeSign(config.GetConfigStr("secretkey2"), cookie)
	cookie = DecodeSign(config.GetConfigStr("secretkey1"), cookie)
	return cookie
}

func GetRemoteIp(r *http.Request) string {
	if ips := r.Header.Get("X-Forwarded-For"); ips != "" {
		host := strings.Split(ips, ",")[0]
		return strings.Split(host, ":")[0]
	}
	ips := strings.Split(r.RemoteAddr, ":")
	if len(ips) > 0 && ips[0] != "[" {
		return ips[0]
	}
	return "127.0.0.1"
}

func Ip2Int(ipstr string) uint32 {
	ip := net.ParseIP(strings.Split(ipstr, ":")[0]).To4()
	if len(ip) == 4 {
		return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
	} else {
		logging.Error("GetRemoteIp err:%s", ipstr)
	}
	return 0
}

func ExportXlsx(filename string, title []string, keys []string, data []map[string]interface{}) error {
	fd := xlsx.NewFile()
	st, err := fd.AddSheet("Sheet1")
	if err != nil {
		return err
	}
	row := st.AddRow()
	for _, v := range title {
		row.AddCell().Value = v
	}
	for _, d := range data {
		row := st.AddRow()
		for _, k := range keys {
			row.AddCell().Value = fmt.Sprintf("%v", d[k])
		}
	}
	err = fd.Save(filename)
	logging.Info("Save excel file:%s, err:%v", filename, err)
	return err
}

func ExportChatMsg(gameid, zoneid, chatdate uint32) error {
	defer func() {
		if err := recover(); err != nil {
			logging.Error("ExportChatMsg error:%v, gameid:%d, zoneid:%d, date:%d", err, gameid, zoneid, chatdate)
		}
	}()
	tblname := get_user_chat_table(gameid, zoneid, chatdate)
	str := fmt.Sprintf("select count(*) from %s", tblname)
	row := db_gm.QueryRow(str)
	var count uint32
	if err := row.Scan(&count); err != nil {
		logging.Error("ExportChatMsg error:%s", err.Error())
		return err
	}
	if count < 200 {
		logging.Warning("ExportChatMsg chat message not enugh, %d", count)
		return nil
	}
	str = fmt.Sprintf("select id, cpid, platid, accid, charid, charname, type, otherid, othername, content, created_at from %s order by id desc limit 50000", tblname)
	rows, err := db_gm.Query(str)
	if err != nil {
		logging.Error("ExportChatMsg error:%s", err.Error())
		return err
	}
	defer rows.Close()
	data := make([]map[string]interface{}, 0)
	for rows.Next() {
		var charname, othername, content string
		var rid, cpid, platid, accid, charid, rtype, otherid, created uint64
		if err := rows.Scan(&rid, &cpid, &platid, &accid, &charid, &charname, &rtype, &otherid, &othername, &content, &created); err != nil {
			logging.Error("ExportChatMsg error:%s", err.Error())
			continue
		}
		d := map[string]interface{}{
			"id":        rid,
			"cpid":      cpid,
			"platid":    platid,
			"accid":     accid,
			"charid":    charid,
			"charname":  charname,
			"type":      rtype,
			"otherid":   otherid,
			"othername": othername,
			"content":   content,
			"created":   created,
		}
		data = append(data, d)
	}
	titles := []string{"序号", "厂商ID", "渠道ID", "账号ID", "角色ID", "角色名", "聊天频道", "聊天对象ID", "聊天对象名称", "聊天内容", "聊天时间"}
	keys := []string{"id", "cpid", "platid", "accid", "charid", "charname", "type", "otherid", "othername", "content", "created"}
	return ExportXlsx(path.Join(config.GetConfigStr("logdir"), tblname+".xlsx"), titles, keys, data)
}
func UnixToYmdhms(ts uint64) string {
	if ts == 0 {
		return "0"
	}
	y, mon, d := time.Unix(int64(ts), 0).Date()
	h, m, s := time.Unix(int64(ts), 0).Clock()
	str := fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d", y, int(mon), d, h, m, s)
	//ret, _ := strconv.Atoi(str)
	return str
}
