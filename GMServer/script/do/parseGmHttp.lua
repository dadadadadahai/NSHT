GmHttp = GmHttp or {}

-- 基础模板文件
TableBaseTpls = {"base_header_meta.tpl", "base_menu_bar.tpl", "base_top_bar.tpl", "base_footer_meta.tpl"}

-- 路径、模板文件对应表
TableTplPaths = {
	["/login"] 							= { "login.tpl" },
	["/games"] 							= { "games.tpl" },
	["/mj_agentRequirement"] 			= { "mj_agentRequirement.tpl" },
	["/mj_agentRequirementList"] 		= { "mj_agentRequirementList.tpl" },
	["/mj_agentManage"] 				= { "mj_agentManage.tpl" },
	["/mj_agentApplication"] 			= { "mj_agentApplication.tpl" },
	["/mj_agentRechargeRecord"] 		= { "mj_agentRechargeRecord.tpl" },
	["/mj_rechargeItems"] 				= { "mj_rechargeItems.tpl" },
	["/mj_rechargeRecord"] 				= { "mj_rechargeRecord.tpl" },
	["/mj_agentOpraLog"] 				= { "mj_agentOpraLog.tpl" },
	["/mj_agentPromotion"]				= { "mj_agentPromotion.tpl" },
	["/mj_agentExtensionSettlement"] 	= {"mj_agentExtensionSettlement.tpl"},
	["/mj_agentExtension"]				= { "mj_agentExtension.tpl" },	
	["/mj_agentExtensionApplication"]	= { "mj_agentExtensionApplication.tpl" },
	["/mj_agentRecharge"]				= { "mj_agentRecharge.tpl" },	
	["/mj_generalAgentRecharge"]	= { "mj_generalAgentRecharge.tpl" },
	["/mj_agentAchievement"]				= { "mj_agentAchievement.tpl" },	
	["/mj_agentOperationaldata"]	= { "mj_agentOperationaldata.tpl" },
	["/mj_agentRegion"]	= { "mj_agentRegion.tpl" },
	["/mj_agentRegiondata"]	= { "mj_agentRegiondata.tpl" },
	["/mj_agentSettlementSearch"] = {"mj_agentSettlementSearch.tpl"},
	["/mj_agentdataSearch"] = {"mj_agentdataSearch.tpl"},
	["/mj_agentBindSearch"] = {"mj_agentBindSearch.tpl"},
	["/mj_agentSeniorAchievement"] = {"mj_agentSeniorAchievement.tpl"},
	["/mj_agentBindSearchWZ"] = {"mj_agentBindSearchWZ.tpl"},
	["/mj_agentSeniorAchievementWZ"] = {"mj_agentSeniorAchievementWZ.tpl"},
	
	-- ["/gm_a_rolesearch"] = {"gm_a_rolesearch.tpl"},
	-- ["/gm_a_gameplayer"] = {"gm_a_gameplayer.tpl"},
	-- ["/gm_a_info_modify"] = {"gm_a_info_modify.tpl"},
	-- ["/gm_a_record_search"] = {"gm_a_record_search.tpl"},
	-- ["/gm_a_act_control"] = {"gm_a_act_control.tpl"},
	-- ["/gm_b_punish_user"] = {"gm_b_punish_user.tpl"},
	-- ["/gm_b_punish_record"] = {"gm_b_punish_record.tpl"},
	-- ["/gm_c_broadcast_add"] = {"gm_c_broadcast_add.tpl"},
	-- ["/gm_c_broadcast_shutdown"] = {"gm_c_broadcast_shutdown.tpl"},
	-- ["/gm_c_broadcast_list"] = {"gm_c_broadcast_list.tpl"},
	-- ["/gm_d_gmcommand"] = {"gm_d_gmcommand.tpl"},
	-- ["/gm_d_order_list"] = {"gm_d_order_list.tpl"},
	-- ["/gm_d_payment_list"] = {"gm_d_payment_list.tpl"},
	-- ["/gm_d_pw_modify"] = {"gm_d_pw_modify.tpl"},
	-- ["/gm_d_system_set"] = {"gm_d_system_set.tpl"},
	-- ["/gm_d_user_add"] = {"gm_d_user_add.tpl"},
	-- ["/gm_e_code_generate"] = {"gm_e_code_generate.tpl"},
	-- ["/gm_e_code_list"] = {"gm_e_code_list.tpl"},
	-- ["/gm_f_user_feedback"] = {"gm_f_user_feedback.tpl"},
	-- ["/gm_f_feedback_list"] = {"gm_f_feedback_list.tpl"},
	-- ["/gm_f_email_single"] = {"gm_f_email_single.tpl"},
	-- ["/gm_f_email_batched"] = {"gm_f_email_batched.tpl"},
	-- ["/gm_f_user_chatmonitor"] = {"gm_f_user_chatmonitor.tpl"},
}

-- 菜单配置表
TableMenuInfos = {
	{
		["Menuid"]=1,["Menuname"]="麻将代理",["Menulink"]="",["Active"]=0,["Submenus"]={
			{["Menuid"]=4,["Menuname"]="代理条件",["Menulink"]="/mj_agentRequirement",["Active"]=0},
			{["Menuid"]=11,["Menuname"]="晋升福利",["Menulink"]="/mj_agentPromotion",["Active"]=0},
			{["Menuid"]=5,["Menuname"]="代理管理",["Menulink"]="/mj_agentManage",["Active"]=0},
			{["Menuid"]=6,["Menuname"]="代理申请",["Menulink"]="/mj_agentApplication",["Active"]=0},
			{["Menuid"]=7,["Menuname"]="代充记录",["Menulink"]="/mj_agentRechargeRecord",["Active"]=0},
			{["Menuid"]=14,["Menuname"]="推广管理",["Menulink"]="/mj_agentExtension",["Active"]=0},
			{["Menuid"]=15,["Menuname"]="推广申请",["Menulink"]="/mj_agentExtensionApplication",["Active"]=0},
			{["Menuid"]=16,["Menuname"]="配置列表",["Menulink"]="/mj_agentRequirementList",["Active"]=0},
			{["Menuid"]=17,["Menuname"]="区域配置",["Menulink"]="/mj_agentRegion",["Active"]=0},
		}
	},
	{
		["Menuid"]=2,["Menuname"]="充值管理",["Menulink"]="",["Active"]=0,["Submenus"]={
			{["Menuid"]=8,["Menuname"]="充值面额",["Menulink"]="/mj_rechargeItems",["Active"]=0},
			{["Menuid"]=9,["Menuname"]="充值记录",["Menulink"]="/mj_rechargeRecord",["Active"]=0},			
			{["Menuid"]=29,["Menuname"]="代理充值",["Menulink"]="/mj_agentRecharge",["Active"]=0},
			{["Menuid"]=28,["Menuname"]="总代理充值",["Menulink"]="/mj_generalAgentRecharge",["Active"]=0},
		}
	},
	{
		["Menuid"]=3,["Menuname"]="系统日志",["Menulink"]="",["Active"]=0,["Submenus"]={
			{["Menuid"]=10,["Menuname"]="操作日志",["Menulink"]="/mj_agentOpraLog",["Active"]=0}
		}
	},
	{
		["Menuid"]=12,["Menuname"]="结算管理",["Menulink"]="",["Active"]=0,["Submenus"]={
			{["Menuid"]=13,["Menuname"]="推广结算",["Menulink"]="/mj_agentExtensionSettlement",["Active"]=0},
			{["Menuid"]=121,["Menuname"]="结算查询",["Menulink"]="/mj_agentSettlementSearch",["Active"]=0},
		}
	},
	{
		["Menuid"]=512,["Menuname"]="数据查询",["Menulink"]="",["Active"]=0,["Submenus"]={
			{["Menuid"]=513,["Menuname"]="代理业绩查询",["Menulink"]="/mj_agentAchievement",["Active"]=0},
			{["Menuid"]=514,["Menuname"]="运营数据查询",["Menulink"]="/mj_agentOperationaldata",["Active"]=0},
			{["Menuid"]=515,["Menuname"]="区域数据查询",["Menulink"]="/mj_agentRegiondata",["Active"]=0},
			{["Menuid"]=516,["Menuname"]="代理数据查询",["Menulink"]="/mj_agentdataSearch",["Active"]=0},
			{["Menuid"]=517,["Menuname"]="代理绑定查询",["Menulink"]="/mj_agentBindSearch",["Active"]=0},
			{["Menuid"]=518,["Menuname"]="高代业绩查询",["Menulink"]="/mj_agentSeniorAchievement",["Active"]=0},
			{["Menuid"]=519,["Menuname"]="代理绑定查询(婺州)",["Menulink"]="/mj_agentBindSearchWZ",["Active"]=0},
			{["Menuid"]=520,["Menuname"]="高代业绩查询(婺州)",["Menulink"]="/mj_agentSeniorAchievementWZ",["Active"]=0},
		}
	},
	-- {["Menuid"]=13,["Menuname"]="账号管理",["Menulink"]="",["Active"]=0,["Submenus"]={{["Menuid"]=14,["Menuname"]="角色查询",["Menulink"]="/gm_a_rolesearch",["Active"]=0},{["Menuid"]=15,["Menuname"]="玩家列表",["Menulink"]="/gm_a_gameplayer",["Active"]=0},{["Menuid"]=16,["Menuname"]="信息修改",["Menulink"]="/gm_a_info_modify",["Active"]=0},{["Menuid"]=17,["Menuname"]="挑战记录",["Menulink"]="/gm_a_record_search",["Active"]=0},{["Menuid"]=18,["Menuname"]="活动控制",["Menulink"]="/gm_a_act_control",["Active"]=0}}},
	-- {["Menuid"]=19,["Menuname"]="奖惩管理",["Menulink"]="",["Active"]=0,["Submenus"]={{["Menuid"]=20,["Menuname"]="处罚玩家",["Menulink"]="/gm_b_punish_user",["Active"]=0},{["Menuid"]=21,["Menuname"]="最近处罚",["Menulink"]="/gm_b_punish_record",["Active"]=0}}},
	-- {["Menuid"]=22,["Menuname"]="公告管理",["Menulink"]="",["Active"]=0,["Submenus"]={{["Menuid"]=23,["Menuname"]="添加公告",["Menulink"]="/gm_c_broadcast_add",["Active"]=0},{["Menuid"]=24,["Menuname"]="停机公告",["Menulink"]="/gm_c_broadcast_shutdown",["Active"]=0},{["Menuid"]=25,["Menuname"]="历史公告",["Menulink"]="/gm_c_broadcast_list",["Active"]=0}}},
	-- {["Menuid"]=26,["Menuname"]="系统管理",["Menulink"]="",["Active"]=0,["Submenus"]={{["Menuid"]=27,["Menuname"]="GM指令",["Menulink"]="/gm_d_gmcommand",["Active"]=0},{["Menuid"]=28,["Menuname"]="订单详情",["Menulink"]="/gm_d_order_list",["Active"]=0},{["Menuid"]=29,["Menuname"]="支付管理",["Menulink"]="/gm_d_payment_list",["Active"]=0},{["Menuid"]=30,["Menuname"]="修改密码",["Menulink"]="/gm_d_pw_modify",["Active"]=0},{["Menuid"]=31,["Menuname"]="添加用户",["Menulink"]="/gm_d_user_add",["Active"]=0},{["Menuid"]=32,["Menuname"]="服务器管理",["Menulink"]="/gm_d_system_set",["Active"]=0}}},
	-- {["Menuid"]=33,["Menuname"]="礼包管理",["Menulink"]="",["Active"]=0,["Submenus"]={{["Menuid"]=34,["Menuname"]="生成礼包码",["Menulink"]="/gm_e_code_generate",["Active"]=0},{["Menuid"]=35,["Menuname"]="礼包码查询",["Menulink"]="/gm_e_code_list",["Active"]=0}}},
	-- {["Menuid"]=36,["Menuname"]="反馈管理",["Menulink"]="",["Active"]=0,["Submenus"]={{["Menuid"]=37,["Menuname"]="问卷反馈",["Menulink"]="/gm_f_user_feedback",["Active"]=0},{["Menuid"]=38,["Menuname"]="用户反馈",["Menulink"]="/gm_f_feedback_list",["Active"]=0},{["Menuid"]=39,["Menuname"]="发送邮件",["Menulink"]="/gm_f_email_single",["Active"]=0},{["Menuid"]=40,["Menuname"]="批量邮件",["Menulink"]="/gm_f_email_batched",["Active"]=0},{["Menuid"]=41,["Menuname"]="聊天监控",["Menulink"]="/gm_f_user_chatmonitor",["Active"]=0}}},
}
-- 角色权限
--TableRoleMenus = {
--	{["Roleid"]=1,["Rolename"]="超级管理员",["Pri"]={["0"]={1,2,3,4,5,6,7,8,9,10}}},
--	{["Roleid"]=2,["Rolename"]="管理员",["Pri"]={[9017]={1,2,3,4,5,6,7,8,9,10},[9011]={1,2,3,4,5,6,7,8,9,10},[9014]={1,2,3,4,5,6,7,8,9,10},[9021]={1,2,3,4,5,6,7,8,9,10}}},
--	{["Roleid"]=3,["Rolename"]="年会专用",["Pri"]={[0]={11,}}},
--}

GmHttp.GetBaseTpls = function()
	return TableBaseTpls
end

GmHttp.GetTplPaths = function()
	return TableTplPaths
end

GmHttp.GetMenusInfo = function()
	return json.encode(TableMenuInfos)
end



