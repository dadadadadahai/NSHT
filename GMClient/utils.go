package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/unibase"
	"github.com/tealeg/xlsx"
)

func Ip2Int(ipstr string) uint32 {
	if ipstr == "" {
		return 0
	}
	ip := net.ParseIP(strings.Split(ipstr, ":")[0]).To4()
	if len(ip) == 4 {
		return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
	}
	return 0
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

func XSRFFormHTML(w http.ResponseWriter, secret string) string {
	token := unibase.Rand.RandString(32)
	SetSecureCookie(w, secret, "_xsrf", token)
	return token
	//return `<input type="hidden" name="_xsrf" class="_xsrf" value="` + token + `" />`
}

func CheckXSRF(r *http.Request, secret string) bool {
    return true;
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
	return err
}

func md5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
