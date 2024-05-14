package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strconv"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
	"git.code4.in/mobilegameserver/unibase/entry"
	"git.code4.in/mobilegameserver/unibase/luaginx"
)

type HttpHandlerFunc func(w http.ResponseWriter, r *http.Request, gmid string)

var (
	Handlers = map[string]HttpHandlerFunc{}
	TplMap   = make(map[string]*template.Template)
)

func InitWebService() {
	InitHandlerMap()
	InitStaticFiles()
	InitStaticHandlers()
}

func InitHandlerMap() {
	Handlers["/games"] = HandleGames
	Handlers["/index"] = HandleIndex // 暂时没用
	Handlers["/login"] = HandleLogin
}

func InitStaticFiles() bool {
	static, _ := filepath.Abs(config.GetConfigStr("static"))
	http.Handle("/js/", http.FileServer(http.Dir(static)))
	http.Handle("/css/", http.FileServer(http.Dir(static)))
	http.Handle("/images/", http.FileServer(http.Dir(static)))
	http.Handle("/fonts/", http.FileServer(http.Dir(static)))
	http.HandleFunc("/favicon.ico", HandlerIcon)
	http.HandleFunc("/logout", HandleLogout)
	return true
}

func HandlerIcon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/images/"+r.URL.Path, http.StatusFound)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ClearCookie(w, r, "gmid")
	ClearCookie(w, r, "name")
	ClearCookie(w, r, "gameid")
	http.Redirect(w, r, "/login", http.StatusFound)
}

func InitStaticHandlers() error {
	viewpath, _ := filepath.Abs(config.GetConfigStr("views"))
	basetpls := GetBaseTpls()
	tmpBase := make([]string, len(basetpls))
	for i := range basetpls {
		tmpBase[i] = path.Join(viewpath, basetpls[i])
	}
	for p, tpls := range GetTplPaths() {
		tmps := make([]string, len(tpls))
		for i := range tpls {
			tmps[i] = path.Join(viewpath, tpls[i])
		}
		tmps = append(tmps, tmpBase...)
		t, err := template.ParseFiles(tmps...)
		if err != nil {
			logging.Error("InitStaticHandlers error:%s", err.Error())
			return err
		}
		TplMap[p] = t
		if handler, ok := Handlers[p]; ok {
			logging.Debug("add handler with path:%s", p)
			http.HandleFunc(p, HandlerLoginWrap(handler))
		} else {
			logging.Debug("add default handler with path:%s", p)
			http.HandleFunc(p, HandlerLoginWrap(HandleDefault))
		}
	}
	return nil
}

//http装饰器，必须登录才能打开除登录以外的页面
func HandlerLoginWrap(httpfunc HttpHandlerFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gmid, debug := GetSecureCookie(r, "gmid"), config.GetConfigStr("debug")
		logging.Debug("Handle Request:%s, debug mode:%s, gmid:%s", r.RequestURI, debug, gmid)
		if debug != "true" && r.URL.Path != "/login" && SessionGet(gmid, "login") != "1" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if debug != "true" && !CheckPermission(gmid, r.URL.Path) {
			http.Error(w, "对不起，您没有访问该页面的权限", http.StatusUnauthorized)
			return
		}
		httpfunc(w, r, gmid)
	}
}

func CheckPermission(gmid, pathstr string) bool {
	return true
}

type MenuZone struct {
	Zoneid   int
	Zonename string
}

type MenuGame struct {
	Gameid   int
	Gamename string
	Zones    []*MenuZone
}

type Menu struct {
	Menuid   int
	Menuname string
	Menulink string
	Active   int
	Invalid  int
	Submenus []*Menu
}

var Games = []*MenuGame{
	{Gamename: "广东麻将", Gameid: 9009, Zones: []*MenuZone{{Zoneid: 116, Zonename: "代理商后台"}, {Zoneid: 115, Zonename: "广东鸡平胡"}}},
}

type SortedMenuGame []*MenuGame

func (self SortedMenuGame) Len() int {
	return len(self)
}

func (self SortedMenuGame) Less(i, j int) bool {
	return self[i].Gameid < self[j].Gameid
}

func (self SortedMenuGame) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func GetOnlineGames(gmid string) []*MenuGame {
	gamemap := make(map[uint32]*MenuGame)
	gameids := get_gameids_by_accid(uint32(unibase.Atoi(gmid, 0)))

	callback := func(v entry.EntryInterface) bool {
		task, ok := v.(*ZoneTask)
		if ok {
			tgameid := task.serverState.GetGamezone().GetGameid()
			tzoneid := task.serverState.GetGamezone().GetZoneid()
			if (tzoneid == 1) && int(tgameid/1000) == 9 && (len(gameids) == 0 || InSlice(tgameid, gameids)) {
				game, ok := gamemap[tgameid]
				if !ok {
					game = &MenuGame{}
					game.Gameid = int(tgameid)
					game.Gamename = task.serverState.GetGamezone().GetGamename()
					game.Zones = make([]*MenuZone, 0)
				}
				zone := &MenuZone{}
				zone.Zoneid = int(task.serverState.GetGamezone().GetZoneid())
				zone.Zonename = task.serverState.GetGamezone().GetZonename()
				game.Zones = append(game.Zones, zone)
				gamemap[tgameid] = game
			}
		}
		return true
	}
	logging.Info("GMWEB: %d", len(gameids))
	zoneTaskManager.ExecEvery(callback)
	games := make([]*MenuGame, 0)
	for _, game := range gamemap {
		games = append(games, game)
		logging.Info("GMWEB: Gameid:%d, Gamename:%s", game.Gameid, game.Gamename)
	}
	if len(games) == 0 && config.GetConfigStr("debug") == "true" {
		return Games
	}
	sort.Sort(SortedMenuGame(games))
	return games
}

func GetGames(gmid string) map[string]interface{} {
	games := GetOnlineGames(gmid)
	return map[string]interface{}{
		"Games":     games,
		"Firstpage": GetFirstMenu(gmid),
	}
}

func GetFirstMenu(gmid string) (pathstr string) {
	menus := GetMenuInfos()
	if len(menus) > 0 {
		if len(menus[0].Submenus) > 0 {
			return menus[0].Submenus[0].Menulink
		}
		return menus[0].Menulink
	}
	logging.Error("no menu set")
	return ""
}

func GetMenus(gmid string) map[string]interface{} {
	active := SessionGet(gmid, "active")
	find := false
	menus := GetMenuInfos()
	menuids := get_menuids(uint64(unibase.Atoi(gmid, 0)))
	for _, menu := range menus {
		if active == menu.Menulink {
			menu.Active = 1
			find = true
		} else {
			menu.Active = 0
		}
		if InSlice(uint32(menu.Menuid), menuids) {
			menu.Invalid = 1
		}
		for _, submenu := range menu.Submenus {
			if active == submenu.Menulink {
				submenu.Active = 1
				menu.Active = 1
				find = true
			} else {
				submenu.Active = 0
			}
			if InSlice(uint32(submenu.Menuid), menuids) {
				submenu.Invalid = 1
			}
		}
	}
	if find == false && len(menus) > 0 {
		menu := menus[0]
		menu.Active = 1
		if len(menu.Submenus) > 0 {
			menu.Submenus[0].Active = 1
		}
	}
	return map[string]interface{}{
		"Menus":    menus,
		"CurGame":  GetCurGame(gmid),
		"Nickname": SessionGet(gmid, "name"),
	}
}

func SortZones(zoneid int, zones []*MenuZone) {
	if len(zones) <= 1 {
		return
	}
	for i := 0; i < (len(zones) - 1); i++ {
		var t = i
		for j := 1; j < len(zones); j++ {
			if zones[t].Zoneid > zones[j].Zoneid {
				t = j
			}
		}
		if t != i {
			zones[i], zones[t] = zones[t], zones[i]
		}
	}
	for i := range zones {
		if zones[i].Zoneid == zoneid {
			zones[0], zones[i] = zones[i], zones[0]
			break
		}
	}
}

func GetCurGame(gmid string) *MenuGame {
	gameid := int(unibase.Atoi(SessionGet(gmid, "gameid"), 0))
	var tmpgame *MenuGame
	for _, game := range GetOnlineGames(gmid) {
		if game.Gameid == gameid {
			tmpgame = game
			break
		}
	}
	zoneid := int(unibase.Atoi(SessionGet(gmid, "zoneid"), 0))
	if tmpgame != nil {
		SortZones(zoneid, tmpgame.Zones)
	}
	return tmpgame
}

func SetActive(w http.ResponseWriter, r *http.Request, gmid string) {
	SessionSet(gmid, "active", r.URL.Path)
	gameid := int(unibase.Atoi(r.FormValue("gameid"), 0))
	gameids := get_gameids_by_accid(uint32(unibase.Atoi(gmid, 0)))
	if gameid != 0 && (len(gameids) == 0 || InSlice(uint32(gameid), gameids)) {
		SessionSet(gmid, "gameid", strconv.Itoa(gameid))
		SetCookie(w, "gameid", strconv.Itoa(gameid))
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request, gmid string) {
	defer r.Body.Close()
	t, err := GetTemplate(r.URL.Path)
	if err != nil {
		logging.Error("HandleGames error, path:%s, err:%s", r.URL.Path, err.Error())
	}
	if t != nil && err == nil {
		t.ExecuteTemplate(w, r.URL.Path[1:], nil)
	} else {
		http.NotFound(w, r)
	}
}

func HandleGames(w http.ResponseWriter, r *http.Request, gmid string) {
	defer r.Body.Close()
	t, err := GetTemplate(r.URL.Path)
	if err != nil {
		logging.Error("HandleGames error, path:%s, err:%s", r.URL.Path, err.Error())
	}
	if t != nil && err == nil {
		t.ExecuteTemplate(w, r.URL.Path[1:], GetGames(gmid))
	} else {
		http.NotFound(w, r)
	}
}

func HandleIndex(w http.ResponseWriter, r *http.Request, gmid string) {
	defer r.Body.Close()
	t, err := GetTemplate(r.URL.Path)
	if err != nil {
		logging.Error("HandleIndex error, path:%s, err:%s", r.URL.Path, err.Error())
	}
	if t != nil && err == nil {
		SetActive(w, r, gmid)
		t.ExecuteTemplate(w, r.URL.Path[1:], GetMenus(gmid))
	} else {
		http.NotFound(w, r)
	}
	return
}

func HandleDefault(w http.ResponseWriter, r *http.Request, gmid string) {
	defer r.Body.Close()
	t, err := GetTemplate(r.URL.Path)
	if err != nil {
		logging.Error("HandleDefault error, path:%s, err:%s", r.URL.Path, err.Error())
	}
	if t != nil && err == nil {
		SetActive(w, r, gmid)
		t.ExecuteTemplate(w, r.URL.Path[1:], GetMenus(gmid))
	} else {
		http.NotFound(w, r)
	}
}

func GetTemplate(pathstr string) (t *template.Template, err error) {
	if config.GetConfigStr("debug") != "true" {
		t = TplMap[pathstr]
	} else {
		viewpath, _ := filepath.Abs(config.GetConfigStr("views"))
		basetpls := GetBaseTpls()
		tmps := make([]string, len(basetpls)+1)
		for i := range basetpls {
			tmps[i+1] = path.Join(viewpath, basetpls[i])
		}
		tmps[0] = path.Join(viewpath, pathstr+".tpl")
		t, err = template.ParseFiles(tmps...)
	}
	return
}

func GetBaseTpls() []string {
	luafun := "GmHttp.GetBaseTpls"
	if luaginx.LuaIsFunction(luafun) {
		basetpls, err := luaginx.LuaDoFunc(luafun)
		if err != nil {
			logging.Error("%s error: %s", luafun, err.Error())
			return nil
		}
		retl := make([]string, len(basetpls.([]interface{})))
		for i, e := range basetpls.([]interface{}) {
			retl[i] = e.(string)
		}
		return retl
	}
	logging.Error("lua GmHttp.GetBaseTpls error")
	return nil
}

func GetTplPaths() map[string][]string {
	luafun := "GmHttp.GetTplPaths"
	if luaginx.LuaIsFunction(luafun) {
		tplpaths, err := luaginx.LuaDoFunc(luafun)
		if err != nil {
			logging.Error("%s error: %s", luafun, err.Error())
			return nil
		}
		retm := make(map[string][]string, len(tplpaths.(map[string]interface{})))
		for k, vs := range tplpaths.(map[string]interface{}) {
			retm[k] = make([]string, len(vs.([]interface{})))
			for i, v := range vs.([]interface{}) {
				retm[k][i] = v.(string)
			}
		}
		return retm
	}
	logging.Error("lua GmHttp.GetTplPaths error")
	return nil
}

func GetMenuInfos() []*Menu {
	luafun := "GmHttp.GetMenusInfo"
	if luaginx.LuaIsFunction(luafun) {
		tmpmenu, err := luaginx.LuaDoFunc(luafun)
		if err != nil {
			logging.Error("%s error: %s", luafun, err.Error())
			return nil
		}
		retl := make([]*Menu, 0)
		err = json.Unmarshal([]byte(tmpmenu.(string)), &retl)
		if err != nil {
			logging.Error("%s, %s error: %s", luafun, tmpmenu, err.Error())
			return nil
		}
		return retl
	}
	logging.Error("lua GmHttp.GetMenusInfo error")
	return nil
}
