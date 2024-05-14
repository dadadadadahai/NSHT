package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	sjson "github.com/bitly/go-simplejson"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/unitime"
)

// 过期时间，获取到access_token的时间+有效期
var expires_in int64 = 0

// 推送消息需要的key
var access_token string = ""

// 群发消息接口,每个账户每月只能接收4条，每个公众号每日只能发送100次
var send_notice_url string = "https://api.weixin.qq.com/cgi-bin/message/mass/sendall?access_token=%s"

// 客服下发消息接口,每日限额500000
var send_custom_url string = "https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s"

// 获取access_token接口
var get_access_token_url string = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"

func checkAccessToken() {
	return //不再推送微信公众号消息
	now := time.Now().Unix()
	if now+120 >= expires_in {
		getAccessToken()
	}
}

func getAccessToken() {
	appid := strings.TrimSpace(config.GetConfigStr("appid"))
	appsecret := strings.TrimSpace(config.GetConfigStr("appsecret"))
	url := fmt.Sprintf(get_access_token_url, appid, appsecret)
	res, err := http.Get(url)
	if err != nil {
		logging.Error("getAccessToken Error:%s url:%s", err.Error(), url)
		return
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	sjs, err := sjson.NewJson(body)
	if err != nil {
		logging.Error("getAccessToken NewJson error:%s", err.Error())
		return
	}
	new_expires_in := sjs.Get("expires_in").MustInt()
	new_access_token := sjs.Get("access_token").MustString()

	if new_access_token != "" {
		access_token = new_access_token
		expires_in = time.Now().Unix() + int64(new_expires_in)
		logging.Debug("new access_token:%s, new expires_in:%d", access_token, expires_in)
	} else {
		logging.Error("getAccessToken res:%s", body)
	}
}

// 客服接口
func sendCustomMessage(msg string) error {
	openids := config.GetConfigStr("openid")
	if openids == "" {
		logging.Warning("config not set openids")
		return nil
	}
	openid_list := strings.Split(openids, ",")
	for _, openid := range openid_list {
		err := _sendCustomMessage(openid, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func _sendCustomMessage(openid, msg string) error {
	params := map[string]interface{}{
		"touser":  openid,
		"msgtype": "text",
		"text":    map[string]string{"content": msg},
	}
	return postDataToTencent(send_custom_url, params, "_sendCustomMessage")
}

// 群发接口
func sendMassMessage(msg string) error {
	params := map[string]interface{}{
		"filter":  map[string]interface{}{"is_to_all": false, "group_id": 100},
		"msgtype": "text",
		"text":    map[string]string{"content": msg},
	}
	return postDataToTencent(send_notice_url, params, "sendMassMessage")
}

func postDataToTencent(url string, params map[string]interface{}, errmsg string) error {
	checkAccessToken()
	data, err := json.Marshal(params)
	if err != nil {
		logging.Error("%s params error:%s", errmsg, err.Error())
		return err
	}
	body := bytes.NewBuffer([]byte(data))
	res, err := http.Post(fmt.Sprintf(url, access_token), "application/json;charset=utf-8", body)
	if err != nil {
		logging.Error("%s Post error:%s", errmsg, err.Error)
		return err
	}
	defer res.Body.Close()
	data, _ = ioutil.ReadAll(res.Body)
	logging.Debug("%s result:%s", errmsg, data)
	sjs, err := sjson.NewJson(data)
	if err != nil {
		return err
	}
	errcode := sjs.Get("errcode").MustInt()
	if errcode != 0 {
		logging.Error("%s failed, errcode:%d", errmsg, errcode)
		return errors.New("errcode not 0")
	}
	return nil
}

func sendSMMessage(msg, phone_list string) error {
	if phone_list == "" {
		phone_list = config.GetConfigStr("phone_list")
	}
	if phone_list == "" {
		logging.Error("no phone in config")
		return nil
	}
	msg = url.QueryEscape(url.QueryEscape(msg))
	params := "3.0:" + config.GetConfigStr("apikey") + ":" + strconv.FormatInt(unitime.Time.Sec()*1000000, 10)
	tmp := []string{
		config.GetConfigStr("passinfo"),
		config.GetConfigStr("gameid"),
		config.GetConfigStr("channel"),
		phone_list,
		msg,
		config.GetConfigStr("sendtype"),
		config.GetConfigStr("remark"),
		params,
	}
	sign := md5String(strings.Join(tmp, "") + config.GetConfigStr("secretkey"))
	send_url := config.GetConfigStr("sm_url") + strings.Join(tmp, "/") + "/" + sign

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", send_url, nil)
	if err != nil {
		logging.Error("sendSMMessage Error:%s", err.Error())
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		logging.Error("sendSMMessage Error:%s", err.Error())
		return err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	logging.Debug("sendSMMessage result:%s", string(body))
	sjs, err := sjson.NewJson(body)
	if err != nil {
		return err
	}
	errcode := sjs.Get("status").MustInt()
	if errcode == 201 {
		logging.Error("sendSMMessage failed, errcode:%d", errcode)
		return errors.New("errcode 201")
	}
	return nil
}

func md5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
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

func SetSecureCookie(w http.ResponseWriter, secret, name, value string) {
	t := strconv.FormatInt(time.Now().Unix(), 10)
	s := unibase.Rand.RandString(32)
	v := base64.URLEncoding.EncodeToString([]byte(value))
	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s%s", v, s, t)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	cookie := strings.Join([]string{v, s, t, sig}, "|")
	SetCookie(w, name, cookie)
}

// 主要是为了获取session的生成时间，超过这个时间需要重新登录
func SetSecureCookieEx(w http.ResponseWriter, secret, name, value string, createtime int64) {
	t := strconv.FormatInt(createtime, 10)
	s := unibase.Rand.RandString(32)
	v := base64.URLEncoding.EncodeToString([]byte(value))
	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s%s", v, s, t)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	cookie := strings.Join([]string{v, s, t, sig}, "|")
	SetCookie(w, name, cookie)
}

func GetSecureCookie(r *http.Request, secret, name string) string {
	cookie := GetCookie(r, name)
	if cookie == "" {
		return ""
	}
	parts := strings.Split(cookie, "|")
	if len(parts) != 4 {
		return ""
	}
	v, s, t, sig := parts[0], parts[1], parts[2], parts[3]

	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s%s", v, s, t)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return ""
	}
	vs, _ := base64.URLEncoding.DecodeString(v)
	return string(vs)
}

func GetSecureCookieEx(r *http.Request, secret, name string) (value string, createtime int64) {
	cookie := GetCookie(r, name)
	if cookie == "" {
		return
	}
	parts := strings.Split(cookie, "|")
	if len(parts) != 4 {
		return
	}
	v, s, t, sig := parts[0], parts[1], parts[2], parts[3]

	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s%s", v, s, t)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return
	}
	vs, _ := base64.URLEncoding.DecodeString(v)
	value = string(vs)
	createtime, _ = strconv.ParseInt(t, 10, 64)
	return
}

func XSRFFormHTML(w http.ResponseWriter, secret string) string {
	token := unibase.Rand.RandString(32)
	SetSecureCookie(w, secret, "_xsrf", token)
	return token
	//return `<input type="hidden" name="_xsrf" class="_xsrf" value="` + token + `" />`
}

func CheckXSRF(r *http.Request, secret string) bool {
	token := r.FormValue("_xsrf")
	if token == "" {
		return false
	}
	cookieToken := GetSecureCookie(r, secret, "_xsrf")
	if cookieToken == "" {
		return false
	}
	return token == cookieToken
}

func ParseQuery(rawData string) url.Values {
	newValues, _ := url.ParseQuery(rawData)
	if newValues == nil {
		newValues = make(url.Values)
	}
	return newValues
}

func check_table_exists(tblname string) bool {
	var count int

	config := config.GetConfigStr("mysql_dbname")

	configarr := strings.Split(config, "?")

	str := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.TABLES where TABLE_SCHEMA='%s' and TABLE_NAME='%s'", configarr[0], tblname)
	db_monitor.QueryRow(str).Scan(&count)

	if count > 0 {
		return true
	}

	return false
}

func check_column_exists(tblname, colname string) bool {
	var count int
	config := config.GetConfigStr("mysql_dbname")

	configarr := strings.Split(config, "?")

	str := fmt.Sprintf("select count(*) from INFORMATION_SCHEMA.COLUMNS where TABLE_SCHEMA='%s' and TABLE_NAME='%s' and COLUMN_NAME='%s'", configarr[0], tblname, colname)
	db_monitor.QueryRow(str).Scan(&count)
	if count > 0 {
		return true
	}
	return false
}

func UnixToYmd(ts uint64) int {
	y, m, d := time.Unix(int64(ts), 0).Date()
	str := fmt.Sprintf("%04d%02d%02d", y, int(m), d)
	ret, _ := strconv.Atoi(str)
	return ret
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

func TimeToYmd(t *time.Time, offset int) int {
	y, m, d := t.AddDate(0, 0, offset).Date()
	str := fmt.Sprintf("%04d%02d%02d", y, int(m), d)
	ret, _ := strconv.Atoi(str)
	return ret
}

func SubDate(ymd, suby, subm, subd int) int {
	y, m, d := ymd/10000, (ymd/100)%100, ymd%100
	dt := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Now().Location()).AddDate(suby, subm, subd)
	st := dt.Format("20060102")
	newymd, _ := strconv.Atoi(st)
	return newymd
}
