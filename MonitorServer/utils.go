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
)

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

func GetPlatidFromAccount(platid int, account string) (int, string) {
	// if strings.Contains(account, "::") {
	// 	tmp := strings.SplitN(account, "::", 2)
	// 	if len(tmp) == 2 {
	// 		return unibase.Atoi(tmp[0]), tmp[1]
	// 	}
	// }
	return platid, account
}

func md5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func MaxUint32(args ...uint32) uint32 {
	var tmpuint uint32
	for _, arg := range args {
		if arg > tmpuint {
			tmpuint = arg
		}
	}
	return tmpuint
}

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

func SubDate(ymd, suby, subm, subd int) int {
	y, m, d := ymd/10000, (ymd/100)%100, ymd%100
	dt := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Now().Location()).AddDate(-suby, -subm, -subd)
	st := dt.Format("20060102")
	newymd, _ := strconv.Atoi(st)
	return newymd
}
