package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.code4.in/mobilegameserver/config"
	"git.code4.in/mobilegameserver/logging"
	"git.code4.in/mobilegameserver/unibase"
)

type HttpChanHandlerFunc func(task *unibase.ChanHttpTask)
type HttpHandlerFunc func(w http.ResponseWriter, r *http.Request)

var (
	jsmsgMainMap map[string]HttpChanHandlerFunc
	//jsmsgMap map[string]HttpHandlerFunc
	tplMap map[string]*template.Template = make(map[string]*template.Template)
	sm     *SessionManager               = NewSessionManager(3 * 24 * 3600)
)

func InitTable() {
	create_account_table()
	create_zone_table()
	create_game_table()
	create_account_game_table()
	create_permission_item_table()
	careat_defaultinfo() //创建数据库时添加默认信息
}

func careat_defaultinfo() { //创建数据库时添加默认信息
	monitor_account_default_insert() //添加admin用户
	monitor_game_default_insert()    //添加admin用户的游戏信息
	monitor_add_acc_gameinfo()
}
func InitWebApp() bool {
	static, _ := filepath.Abs(config.GetConfigStr("static"))
	tplPath, _ := filepath.Abs(config.GetConfigStr("views"))

	http.Handle("/js/", http.FileServer(http.Dir(static)))
	http.Handle("/css/", http.FileServer(http.Dir(static)))
	http.Handle("/images/", http.FileServer(http.Dir(static)))
	http.Handle("/font/", http.FileServer(http.Dir(static)))
	http.Handle("/layer/", http.FileServer(http.Dir(static)))
	http.Handle("/pageJs/", http.FileServer(http.Dir(static)))
	http.Handle("/json/", http.FileServer(http.Dir(static)))
	http.Handle("/export/", http.FileServer(http.Dir(static)))
	http.Handle("/assets_ace/", http.FileServer(http.Dir(static)))

	InitStaticViews(tplPath)
	http.HandleFunc("/logout", HandleLogout)
	return true
}

func LoginWraper(handler func(w http.ResponseWriter, r *http.Request, s *Session)) HttpHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid, cs := GetSecureCookieEx(r, config.GetConfigStr("secret_key"), "sid")
		session := sm.SessionCheck(sid, cs)
		if session == nil {
			logging.Warning("sid:%s ts:%d, url:%s", sid, cs, r.URL.String())
			http.Redirect(w, r, "/login.html", http.StatusFound)
		} else {
			session.Set("Ip", GetRemoteIP(r.RemoteAddr))
			handler(w, r, session)
		}
	}
}

func GetRemoteIP(remoteadd string) string {
	if remoteadd == "" {
		return ""
	}
	if !strings.Contains(remoteadd, ":") {
		return remoteadd
	}
	addrs := strings.Split(remoteadd, ":")
	return addrs[0]
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	xsrfStr := XSRFFormHTML(w, config.GetConfigStr("secret_key"))
	tplMap["/login.html"].Execute(w, map[string]string{"_xsrf": xsrfStr})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sid := GetSecureCookie(r, config.GetConfigStr("secret_key"), "sid")
	sm.SessionDelete(sid)
	ClearCookie(w, r, "sid")
	http.Redirect(w, r, "/login.html", http.StatusFound)
}

func InitStaticViews(views string) {
	if views[len(views)-1] != '/' {
		views += "/"
	}
	filepath.Walk(views, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		} else if !info.IsDir() {
			basename := "/" + filepath.Base(path)
			t, err := template.ParseFiles(path)
			if err != nil {
				logging.Error("basename:%s, err:%s", basename, err.Error())
				return nil //防止个别文件出错导致其他文件没法解析
			}
			tplMap[basename] = t
			if basename == "/login.html" {
				http.HandleFunc(basename, HandleLogin)
			} else {
				http.HandleFunc(basename, LoginWraper(HandleOther))
			}
		}
		return nil
	})
	if config.GetConfigStr("debug") == "true" {
		http.HandleFunc("/", LoginWraper(HandleOther))
	}
	http.HandleFunc("/favicon.ico", HandlerIcon)
}

func AddHandlerWithTpl(tplPath, url string, hfunc http.HandlerFunc) {
	t, err := template.ParseFiles(tplPath + url + ".html")
	if err != nil {
		return
	}
	tplMap[url] = t
	http.HandleFunc(url, hfunc)
}

func HandlerIcon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/images/"+r.URL.Path, http.StatusFound)
}

func HandleOther(w http.ResponseWriter, r *http.Request, s *Session) {
	defer r.Body.Close()
	logging.Debug("url:%s", r.URL.String())
	gameid := r.FormValue("gameid")
	if gameid != "" {
		SetCookie(w, "gameid", gameid)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	debug := config.GetConfigStr("debug")
	if debug == "true" {
		views, _ := filepath.Abs(config.GetConfigStr("views"))
		if views[len(views)-1] != '/' {
			views += "/"
		}
		t, err := template.ParseFiles(views + r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		t.Execute(w, s.value)
	} else if tpl, ok := tplMap[r.URL.Path]; ok {
		tpl.Execute(w, s.value)
	} else {
		http.NotFound(w, r)
	}
}

func HandleHttpChanCommand(task *unibase.ChanHttpTask) bool {
	defer task.R.Body.Close()
	task.JSW = unibase.NewJSResponseWriter(task.W, nil, nil)
	task.R.Form = ParseQuery(string(task.Rawdata))
	cmd := task.R.FormValue("cmd")

	if cmd == "user_login" {
		HandleLoginAccount(task)
		return true
	}

	sid := GetSecureCookie(task.R, config.GetConfigStr("secret_key"), "sid")
	if sid == "" {
		task.SendBinary([]byte(`{"retcode":1,"retdesc":"请登录后继续!"`))
		return false
	}
	//lsri, lsti := SessionGet(sid, "LSR"), SessionGet(sid, "LST")
	now := time.Now().Unix()
	//if lsr, ok := lsri.(string); ok {
	//	if lsr == cmd {
	//		if lst, ok := lsti.(int64); ok {
	//			if now-lst < 7 {
	//				task.SendBinary([]byte(`{"retcode":1,"retdesc":"请求太频繁，请稍后!"`))
	//				return false
	//			}
	//		}
	//	}
	//}
	SessionSet(sid, "LSR", cmd)
	SessionSet(sid, "LST", now)

	//todo检查权限
	task.Debug("HandleHttpChanCommand cmd:%s, data:%s", cmd, task.Rawdata)
	if msgfun, ok := jsmsgMainMap[cmd]; ok == true {
		msgfun(task)
	} else {
		task.Error("cmd:%s error, no handler function", cmd)
	}
	return true
}

func InitHttpMsgMapMain() {
	jsmsgMainMap = make(map[string]HttpChanHandlerFunc)

	jsmsgMainMap["user_daily_data"] = HandleDailyData                             //查询每日数据,运营总况， 所有渠道合并成一条数据
	jsmsgMainMap["user_daily_first_data"] = HandleDailyFirstData                  //首冲玩家留存
	jsmsgMainMap["user_month_data"] = HandleMonthData                             //查询报表数据，渠道报表, 时间范围内合每个各一条数据,日报表，月报表
	jsmsgMainMap["user_realtime_data"] = HandleRealData                           //查询实时数据，总况中的实时数据
	jsmsgMainMap["user_form_data"] = HandleFormData                               //查询报表数据，登陆分析、充值玩家，每天每个渠道一条数据,给需要按天显示数据的页面
	jsmsgMainMap["user_level_analysis"] = HandleUserlevelAnalysis                 //玩家等级分析
	jsmsgMainMap["user_power_analysis"] = HandleUserPowerAnalysis                 //玩家战力分析
	jsmsgMainMap["user_activity_analysis"] = HandleUserActAnalysis                //玩法分析，（任务分析，PVE分析，PVP分析，活动分析）
	jsmsgMainMap["user_levelup_analysis"] = HandleUserLevelupAnalysis             //玩家升级时间分析
	jsmsgMainMap["user_level_retain"] = HandleUserLevelRetainedAnalysis           //玩家等级流失分布
	jsmsgMainMap["user_eco_daily"] = HandleUserDailyCoin                          //每日货币
	jsmsgMainMap["user_eco_distribution"] = HandleUserCoinDistribution            //产出、消耗分布
	jsmsgMainMap["user_eco_level_distribution"] = HandleUserCoinLevelDistribution //等级消费分布
	jsmsgMainMap["user_eco_lottery"] = HandleUserLotteryAnalysis                  //抽奖分析
	jsmsgMainMap["user_eco_transaction"] = HandleShopTransactionAnalysis          //商店交易分析
	jsmsgMainMap["user_isguid_analysis"] = HandleUserLossNode                     //流失节点分析
	jsmsgMainMap["user_recharge_detail"] = HandleUserPayAnalysis_new              //充值明细
	jsmsgMainMap["user_recharge_distribution"] = HandleUserPayDistribution        //充值分布
	jsmsgMainMap["user_recharge_firstlevel"] = HandleUserPayLevel                 //首冲等级
	jsmsgMainMap["user_recharge_rank"] = HandleUserPayRank                        //充值排行
	jsmsgMainMap["user_recharge_ltv"] = HandleUserLTVAnalysis                     //ltv分析
	jsmsgMainMap["user_recharge_ltv_cash"] = HandleUserLTVCashAnalysis            //ltv分析，去提现
	jsmsgMainMap["user_recharge_arppu"] = HandleUserARPPUAnalysis                 //arppu分析
	jsmsgMainMap["recovery_data"] = HandleRecoveryData                            // 回收数据
	jsmsgMainMap["recovery_data_roi_cash"] = HandleRecoveryDataCash               // 回收数据（首冲，加提现）
	jsmsgMainMap["user_point_detail"] = HandleUserPointDetail                     //积分兑换明细
	jsmsgMainMap["user_account_daily"] = HandleAccoundDaily                       //每日注册
	jsmsgMainMap["user_recharge_transform"] = HandleRechargeTransform             //注册充值转化
	jsmsgMainMap["user_plataccount_analysis"] = HandlePlatAccoundAnalysis         //渠道注册分析
	jsmsgMainMap["user_mahjong_daily"] = HandleMahjangDaily                       //麻将详情
	jsmsgMainMap["user_mahjong_analysis"] = HandleMahjangAnalysis                 //麻将统计
	jsmsgMainMap["user_mjtotal_analysis"] = HandleMjTotalAnalysis                 //麻将大厅统计
	jsmsgMainMap["user_mjpoint_search"] = HandleMjPointSearch                     //麻将积分查询
	jsmsgMainMap["search_subgames"] = HandleSearchSubgames
	jsmsgMainMap["user_hbq_analysis"] = HandleRedpackAnalysis //红包圈统计
	jsmsgMainMap["user_money_rank"] = HandleUserMoneyRank     //货币排行
	jsmsgMainMap["user_detail_info"] = HandleUserDetailInfo   //用户详细信息

	//完整的渠道列表存放在static里面的json文件中
	jsmsgMainMap["game_plat_list"] = HandleGamePlatList //查询游戏接入的渠道列表
	//新接入的游戏，需要添加游戏信息，否则不予查询
	jsmsgMainMap["game_add"] = HandleAddGame       //添加游戏，用于首页显示
	jsmsgMainMap["game_update"] = HandleGameUpdate //更新游戏状态
	jsmsgMainMap["game_del"] = HandleDelGame       //删除游戏
	jsmsgMainMap["game_list"] = HandleGameList     //查询所有游戏，用于首页显示, 游戏涉及的图片链接通过其他接口上传
	jsmsgMainMap["game_list1"] = HandleGameList1   //查询所有游戏名返回到用户创建页面
	jsmsgMainMap["game_list2"] = HandleGameList2   //查询游戏表 返回到显示游戏列表页面
	jsmsgMainMap["game_list3"] = HandleGameList3

	//添加游戏成功后，再添加区服信息
	jsmsgMainMap["game_zone_add"] = HandleAddGameZone   //添加区服
	jsmsgMainMap["game_zone_del"] = HandleDelGameZone   //删除区服
	jsmsgMainMap["game_zone_list"] = HandleGameZoneList //查询游戏区服列表(需要运营填入区服名称，主要用于网页显示区服名称)
	//账户相关
	jsmsgMainMap["user_add"] = HandleAddAccount         //添加账户
	jsmsgMainMap["user_del"] = HandleDelAccount         //删除账户
	jsmsgMainMap["user_list"] = HandleAccountList       //查询账户列表
	jsmsgMainMap["user_update"] = HandleUpdateAccount   //更新账户权限相关
	jsmsgMainMap["user_login"] = HandleLoginAccount     //用户登录
	jsmsgMainMap["user_getpri"] = HandleGetPri          //获取权限显示菜单
	jsmsgMainMap["user_setting"] = HandleSettingAccount //用户设置

	//汇率相关
	jsmsgMainMap["exchange_rate_list"] = HandleExchangeRateList     //获取汇率列表
	jsmsgMainMap["exchange_rate_update"] = HandleExchangeRateUpdate //修改、添加汇率列表
	//投放渠道相关
	jsmsgMainMap["launch_channel_list"] = HandleLaunchChannelList       //获取渠道账户
	jsmsgMainMap["launch_keywords_action"] = HandleLaunchKeywordsAction //渠道账户操作

	jsmsgMainMap["user_change_info"] = HandleUserChangeInfo //
	jsmsgMainMap["app_number_list"] = HandleAppNumberList   // 获取应用id
	jsmsgMainMap["app_num_action"] = HandleAppNumberAction
	jsmsgMainMap["download_export"] = HandleDownloadexport

	jsmsgMainMap["recovery_examination_list"] = HandleRecoveryExaminationList
	jsmsgMainMap["recovery_examination_list_app"] = HandleRecoveryExaminationListApp
	jsmsgMainMap["recovery_examination"] = HandleRecoveryExamination
	jsmsgMainMap["recovery_examination_app"] = HandleRecoveryExaminationApp
	jsmsgMainMap["recovery_examination_11"] = getadjustreportall
	jsmsgMainMap["update_status_app"] = HandleUpdateStatusApp

}
