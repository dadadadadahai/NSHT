--[[
--解析客户端http发来的参数，构造proto协议转发给GMServer
--
--]]

GmHttp = GmHttp or {}

--十三水控制配置
GmHttp.ParseThirteenWaterControlPmd_CS = function(gmid, taskid, body)
	local params = unilight.getreq(tostring(body))
	params['clientid'] = taskid
	params['gmid'] = gmid
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		gmid = gmid,
		clientid = taskid,
		msg = json.encode({
			["do"] = "Pmd.ThirteenWaterControlPmd_CS",
			data = params
		})
	}
	local resStr = json.encode(encode_repair(data))

	local resProto = go.buildProto("*Pmd.RequestExecGmCommandGmPmd_SC",resStr)
	unilight.debug("十三水控制配置!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--红包数据记录查询
GmHttp.ParseRedPakageActivityRecordPmd_CS = function(gmid, taskid, body)
	local params = unilight.getreq(tostring(body))
	params['clientid'] = taskid
	params['gmid'] = gmid
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		gmid = gmid,
		clientid = taskid,
		msg = json.encode({
			["do"] = "Pmd.RedPakageActivityRecordPmd_CS",
			data = params
		})
	}
	local resStr = json.encode(encode_repair(data))

	local resProto = go.buildProto("*Pmd.RequestExecGmCommandGmPmd_SC",resStr)
	unilight.debug("红包数据记录查询!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end 

--礼品券兑换记录查询
GmHttp.ParseVoucherConvertRecordPmd_CS = function( gmid, taskid, body )
	local params = unilight.getreq(tostring(body))
	params['clientid'] = taskid
	params['gmid'] = gmid
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		gmid = gmid,
		clientid = taskid,
		msg = json.encode({
			["do"] = "Pmd.VoucherConvertRecordPmd_CS",
			data = params
		})
	}
	local resStr = json.encode(encode_repair(data))

	local resProto = go.buildProto("*Pmd.RequestExecGmCommandGmPmd_SC",resStr)
	unilight.debug("礼品券兑换记录查询!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--麻将控制配置
GmHttp.ParseMahjongControlConfigPmd_CS = function( gmid, taskid, body )
	--local params = go.parseMultipartForm(taskid, body)
	local params = unilight.getreq(tostring(body))
	params['clientid'] = taskid
	params['gmid'] = gmid
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		gmid = gmid,
		clientid = taskid,
		msg = json.encode({
			["do"] = "Pmd.MahjongControlConfigPmd_CS",
			data = params
		})
	}
	local resStr = json.encode(encode_repair(data))

	local resProto = go.buildProto("*Pmd.RequestExecGmCommandGmPmd_SC",resStr)
	unilight.debug("麻将控制配置!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--机器人列表
GmHttp.ParseRequestRobotListGmUserPmd_CS = function(gmid, taskid, body)
	--解析http请求的参数form格式的，如果是json格式的，则使用unilight.getreq(tostring(body));格式需自行跟客户端商量
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		room = tonumber(params.room) or 0,
		clientid = taskid,
		gmid = gmid,
	}
	local resStr = json.encode(encode_repair(data)) 
	--根据传入的协议名称和内容构建proto协议包
    local resProto = go.buildProto("*Pmd.RequestRobotListGmUserPmd_CS", resStr)
    unilight.debug("机器人列表!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--添加机器人
GmHttp.ParseOperateRobotGmUserPmd_CS = function (gmid, taskid, body)
	-- body
--	unilight.getreq(tostring(body))
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		optype = tonumber(params.optype) or 0,
		recordid = tonumber(params.id) or 0,
		data = {
			id = tonumber(params.id) or 0,
			num = tonumber(params.num) or 0,
			stime = tonumber(params.stime) or 0,
			etime = tonumber(params.etime) or 0,
			mincoin = tonumber(params.mincoin) or 0,
			maxcoin = tonumber(params.maxcoin) or 0
		},
		clientid = taskid,
		gmid = gmid,
	}	

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.OperateRobotGmUserPmd_CS", resStr)

	unilight.debug("增加机器人!")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--添加区服
GmHttp.ParseAddNewZoneGmUserPmd_CS = function(gmid, taskid, body)
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		zonename = params.zonename or 0,
		gmlink = params.gmlink or 0,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data)) 
	local resProto = go.buildProto("*Pmd.AddNewZoneGmUserPmd_CS", resStr)

	unilight.debug("添加区服")
	unilight.debug(resStr)
	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end


--添加公告 --原前端函数：cmd:gm_broadcast_add
GmHttp.ParseBroadcastNewGmUserPmd_C = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local allzone = tonumber(params.allzone) or 0
	local zoneid = tonumber(params.zoneid) or 0
	if (allzone > 0) then
		zoneid = 0
	end

	unilight.info(json.encode(encode_repair(params)))
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = zoneid,
		data = {
			gameid = tonumber(params.gameid) or 0,   --协议data里面也需要gameid和zoneid 这两个参数
			["zoneid"] = zoneid,
			btype = tonumber(params.btype) or 0,
			title = params.title or "",
			content = params.content or "",
			starttime = tonumber(params.starttime) or 0,
			endtime = tonumber(params.endtime) or 0,
			intervaltime = tonumber(params.intervaltime or 0) * 60,
			packageid = tonumber(params.packageid) or 0,
			description = params.description or "",
			zonidstart = tonumber(params.zonidstart) or 0,
			zonidend = tonumber(params.zonidend) or 0,
			channelid = tonumber(params.channelid) or 0,
			taskid = tonumber(params.taskid) or 0,
		},
		clientid = taskid, 
		gmid = gmid,
	}

	--unilight.debug(table.tostring(data))

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.BroadcastNewGmUserPmd_C", resStr)

	unilight.debug("普通公告")
	unilight.debug(resStr)
	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--公告列表   --原前端函数：gm_broadcast_list
GmHttp.ParseRequestBroadcastListGmUserPmd_C = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		endtime = tonumber(params.endtime) or 0,
		curpage = tonumber(params.curpage) or 1,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestBroadcastListGmUserPmd_C", resStr)

	unilight.debug("公告列表！")

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--活动控制 
GmHttp.ParseActivitySwitchGmUserPmd_CS = function (gmid, taskid, body)
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		optype = tonumber(params.optype) or 0,
		actid = tonumber(params.actid) or 0,
		actname = params.actname,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		clientid = taskid, 
		gmid = gmid
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.ActivitySwitchGmUserPmd_CS", resStr)
	unilight.debug("活动控制")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--处罚玩家（可以添加成功，但是显示不了）
GmHttp.ParsePunishUserGmUserPmd_C = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		data = {
			gameid = tonumber(params.gameid) or 0,
			zoneid = tonumber(params.zoneid) or 0,
			charid = tonumber(params.charid) or 0,
			ptype = tonumber(params.optype) or 0,
			reason = params.content,
			starttime = tonumber(params.starttime) or 0,
			endtime = tonumber(params.endtime) or 0,
		},
		clientid = taskid, 
		gmid = gmid
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.PunishUserGmUserPmd_C", resStr)
	unilight.debug("处罚玩家")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--运营总览
GmHttp.ParseRequestGameDataGmUserPmd_CS = function (gmid, taskid, body)
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		type = tonumber(params.type) or 1,
		stime = tonumber(params.stime) or 0,
		etime = tonumber(params.etime) or 0,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestGameDataGmUserPmd_CS", resStr)

	unilight.debug("运营总览")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--用户信息（玩家列表）  原前端函数：cmd: "user_list_search"--用户信息（玩家列表）  原前端函数：cmd: "user_list_search"
GmHttp.ParseRequestOnlineUserInfoGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		charname = params.charname,
		isonline = tonumber(params.isonline) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 50,
		accid = tonumber(params.accid) or 0, --账号ID(玩家列表用到)
		isagent = tonumber(params.agent) or 0,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		lstime = tonumber(params.lstime) or 0,
		letime = tonumber(params.letime) or 0,
		mincoin = tonumber(params.mincoin) or 0,
		maxcoin = tonumber(params.maxcoin) or 0,
		mindiamond = tonumber(params.mindiamond) or 0,
		maxdiamond = tonumber(params.maxdiamond) or 0,
		ranktype = tonumber(params.ranktype) or 0,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestOnlineUserInfoGmUserPmd_CS", resStr)

	unilight.debug("用户信息")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end
--获取在线玩家信息
GmHttp.ParseStRequestOnlineListInfoPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.game) or 0,
		charid = tonumber(params.charid) or 0,
		charname = params.charname,
		gametype = tonumber(params.sessions) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 50,
	}
	local gamedata ={
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
	}
	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.StRequestOnlineListInfoPmd_CS", resStr)

	unilight.debug("玩家在线信息")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, gamedata.gameid, gamedata.zoneid, resProto)
end
--获取在vip玩家信息
GmHttp.ParseStRequestVipListInfoPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		
		charid = tonumber(params.charid) or 0,
		charname = params.charname,
		viplevel = tonumber(params.viplevel) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 50,
	}
	local gamedata ={
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
	}
	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.StRequestVipListInfoPmd_CS", resStr)

	unilight.debug("vip玩家信息")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, gamedata.gameid, gamedata.zoneid, resProto)
end

--游戏记录 原前端函数：cmd: "gm_lobby_history"
GmHttp.ParseLobbyGameHistoryGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		starttime = tonumber(params.stime) or 0,
		endtime = tonumber(params.etime) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		curpage = tonumber(params.curpage) or 1,
		perpage = 50,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.LobbyGameHistoryGmUserPmd_CS", resStr)

	unilight.debug("游戏记录")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--解散房间 原前端函数：cmd: "gm_room_dissolve"
GmHttp.ParseRoomDissolveGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		roomid = tonumber(params.roomid) or 0,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RoomDissolveGmUserPmd_CS", resStr)

	unilight.debug("解散房间")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)

end

--[[
--群信息 原前端函数：cmd: "gm_room_dissolve"
GmHttp.ParseGroupInfoGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		ownerid = tonumber(params.ownerid) or 0,
		owner = params.owner,
		stime = tonumber(params.stime) or 0,
		etime = tonumber(params.etime) or 0,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.GroupInfoGmUserPmd_CS", resStr)

	unilight.debug("群信息")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end
--]]

--红包查询、红包统计 原前端函数：cmd: "gm_game_redpacket"
GmHttp.ParseRequestRedPacketsGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		srcuid = tonumber(params.srcuid) or 0,
		desuid = tonumber(params.desuid) or 0,
		packetcode = params.packetid or "",
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		curpage = tonumber(params.curpage) or 0,
		isgmtool = tonumber(params.isgmtool) or 0,
		perpage = 50,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestRedPacketsGmUserPmd_CS", resStr)

	unilight.debug("红包查询、红包统计")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--封红包 原前端函数：cmd: "RequestAddRedPacketGmUserPmd_CS"
GmHttp.ParseRequestAddRedPacketGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		srcuid = tonumber(params.srcuid) or 0,
		money = tonumber(params.money) or 0,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestAddRedPacketGmUserPmd_CS", resStr)

	unilight.debug("封红包")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--领红包 原前端函数：cmd: "RequestAddRedPacketGmUserPmd_CS"
GmHttp.ParseRequestRevRedPacketGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		srcuid = tonumber(params.srcuid) or 0,
		id = params.id,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestRevRedPacketGmUserPmd_CS", resStr)

	unilight.debug("领红包")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--积分兑换报表 原前端函数：cmd: "gm_point_report"
GmHttp.ParseRequestPointReportGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 50,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestPointReportGmUserPmd_CS", resStr)

	unilight.debug("积分兑换报表")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--积分兑换明细 原前端函数：cmd: "gm_point_detail"
GmHttp.ParseRequestPointDetailGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		ptype = tonumber(params.ptype) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 50,
		clientid = taskid, 
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestPointDetailGmUserPmd_CS", resStr)

	unilight.debug("积分兑换明细")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--投注详细信息 原前端函数：cmd: "gm_betting_detail"(数据对不上)
GmHttp.ParseRequestBettingDetailGmUserPmd_CS = function(gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		dealerid = tonumber(params.dealerid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 50,
		opensource = tonumber(params.opensource) or 0,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestBettingDetailGmUserPmd_CS", resStr)

	unilight.debug("投注详细信息")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--库存信息 原前端函数：cmd: "gm_game_stock"
GmHttp.ParseRequestStockInfoGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestStockInfoGmUserPmd_CS", resStr)

	unilight.debug("库存信息")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--输赢排行榜 原前端函数：cmd: "gm_winning_list"
GmHttp.ParseRequestWinningListGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		timestamp = tonumber(params.timestamp) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 500,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestWinningListGmUserPmd_CS", resStr)

	unilight.debug("输赢排行榜")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--黑名单信息查询 原前端函数：cmd: "gm_blackwithe_list"
GmHttp.ParseRequestBlackWhitelistGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		curpage = tonumber(params.curpage) or 1,
		id = tonumber(params.id) or 0,
		perpage = 50,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestBlackWhitelistGmUserPmd_CS", resStr)

	unilight.debug("黑名单查询")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--[[
--黑名单添加 原前端函数：cmd: "gm_blackwhitelist_add"
GmHttp.ParseAddBlackWhitelistGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		data = {
			id = tonumber(params.id) or 0,
			charid = tonumber(params.charid) or 0,
			charname = params.charname,
			setchips = tonumber(params.setchips) or 0,
			curchips = tonumber(params.curchips) or 0,
			state = tonumber(params.state) or 0,
			type = tonumber(params.type) or 0,
			settimes = tonumber(params.settimes) or 0,
			winrate = tonumber(params.winrate) or 0,
			intervaltimes = tonumber(params.intervaltimes) or 0,
			gameid = tonumber(params.gameid) or 0,
			zoneid = tonumber(params.zoneid) or 0,
		},
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.AddBlackWhitelistGmUserPmd_CS", resStr)

	unilight.debug("黑名单添加")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end
--]]

--黑名单删除 原前端函数：cmd: "gm_blackwhitelist_del"
GmHttp.ParseDelBlackWhitelistGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		ids = tonumber(params.ids) or 0,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.DelBlackWhitelistGmUserPmd_CS", resStr)

	unilight.debug("黑名单删除")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--金币日志 原前端函数：cmd: "gm_itemrecord_search"
GmHttp.ParseRequestUserItemsHistoryListGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		charname = params.charname or '',
		curpage = tonumber(params.curpage) or 1,
		perpage = tonumber(params.perpage) or 100,
		optype = tonumber(params.cointype) or 0,
		clientid = taskid,
		acttype = tonumber(params.acttype) or 0,
		stime = tonumber(params.stime) or 0,
		etime = tonumber(params.etime) or 0,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestUserItemsHistoryListGmUserPmd_CS", resStr)

	unilight.debug("金币日志")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--角色查找(座位管理) 原前端函数：cmd: "user_info_search"
GmHttp.ParseRequestUserInfoGmUserPmd_C = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		charname = params.charname,
		optype = tonumber(params.cointype) or 0,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestUserInfoGmUserPmd_C", resStr)

	unilight.debug("角色查找")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--信息修改 原前端函数：cmd: "user_info_modify"
GmHttp.ParseRequestModifyUserInfoGmUserPmd_C = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		optype = tonumber(params.optype) or 0,
		changetype = tonumber(params.changetype) or 0,
		opnum = tonumber(params.content) or 0,
		charname = params.content,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestModifyUserInfoGmUserPmd_C", resStr)

	unilight.debug("信息修改")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--修改密码 原前端函数：cmd: "gm_info_modify"
GmHttp.ParseSetPasswordGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		oldpassword = params.oldpasswd,
		newpassword = params.newpasswd1,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.SetPasswordGmUserPmd_CS", resStr)

	unilight.debug("密码修改")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--牛牛配置 原前端函数：cmd: "SetWinOrLoseGmUserPmd_CS"
GmHttp.ParseSetWinOrLoseGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		subgameid = tonumber(params.subgameid) or 0,
		settings = params.settings,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.SetWinOrLoseGmUserPmd_CS", resStr)

	unilight.debug("牛牛配置")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--登录日志 原前端函数：cmd: "gm_loginrecord_search" （日志部分没做）
GmHttp.ParseRequestLoginRecordGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		accid = tonumber(params.accid) or 0,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 500,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestLoginRecordGmUserPmd_CS", resStr)

	unilight.debug("登录日志")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--物品日志（货币日志） 原前端函数：cmd: "gm_consumerecord_search" 
GmHttp.ParseRequestConsumeRecordGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		itemid = tonumber(params.itemid) or 0,
		optype = tonumber(params.optype) or 0,
		actionid = tonumber(params.actionid) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 500,
		type = tonumber(params.type) or 0,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestConsumeRecordGmUserPmd_CS", resStr)

	if data.type == 1 then 
		unilight.debug("物品日志")
	else
		unilight.debug("货币日志")
	end
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--任务日志 原前端函数：cmd: "gm_actionrecord_search" 
GmHttp.ParseRequestActionRecordGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		acttype = tonumber(params.acttype) or 0,
		state = tonumber(params.state) or 0,
		type = tonumber(params.type) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 500,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestActionRecordGmUserPmd_CS", resStr)

	unilight.debug("任务日志")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--礼包生成码 原前端函数：cmd: "gm_code_generate" 
GmHttp.ParseRequestGenerateCodeGmUserPmd_C = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		codetype = tonumber(params.codetype) or 0,
		codenum = tonumber(params.codenum) or 0,
		platid = tonumber(params.platid) or 0,
		packid = tonumber(params.packageid) or 0,
		limit = tonumber(params.limit) or 0,
		zoneidmin = tonumber(params.zoneidmin) or 0,
		zoneidmax = tonumber(params.zoneidmax) or 0,
		desc = params.desc,
		stime = tonumber(params.stime) or 0,
		etime = tonumber(params.etime) or 0,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestGenerateCodeGmUserPmd_C", resStr)

	unilight.debug("礼包码生成")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--礼包码查询 原前端函数：cmd: "gm_code_operator" 
GmHttp.ParseRequestOpeartorCodeGmUserPmd_C = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		code = params.code,
		optype = tonumber(params.optype) or 0,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestOpeartorCodeGmUserPmd_C", resStr)

	unilight.debug("礼包码查询")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--礼包码配置 原前端函数：cmd: "PackcodeSearchGmUserPmd_CS" 
GmHttp.ParsePackcodeSearchGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 500,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.PackcodeSearchGmUserPmd_CS", resStr)

	unilight.debug("礼包码配置")
	
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--[[
--新增礼包码配置
GmHttp.ParsePackcodeInsertGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	--unilight.debug(json.encode(encode_repair(params)))
	unilight.debug("******************")
--	local temp = json.encode(encode_repair(params))
--	unilight.debug(json.encode(encode_repair(params.items)))

	unilight.debug(params.items)
	local temp = loadstring("return " .. params.items)()
	unilight.debug(type(params.items))

	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		data = {
			itemid = tonumber(params.items.itemid) or 0,
			itemname = params.items.itemname,
			itemnum = tonumber(params.items.itemnum) or 0,
			itemtype = tonumber(params.items.itemtype) or 0,
			bind = tonumber(params.items.bind) or 0
		},
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.PackcodeInsertGmUserPmd_CS", resStr)

	unilight.debug("新增礼包码配置")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end
--]]

--礼包码历史 原前端函数：cmd: "PackcodeRecordGmUserPmd_CS" 
GmHttp.ParsePackcodeRecordGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 500,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.PackcodeRecordGmUserPmd_CS", resStr)

	unilight.debug("礼包码历史")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--问卷反馈(用户反馈) 原前端函数：cmd: "gm_feedback_record" 
GmHttp.ParseRequestFeedbackListGmUserPmd_C = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		starttime = tonumber(params.starttime) or 0,
		endtime = tonumber(params.endtime) or 0,
		feedbacktype = tonumber(params.feedbacktype) or 0,
		state = tonumber(params.state) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 500,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestFeedbackListGmUserPmd_C", resStr)

	if feedbacktype == 0 then 
		unilight.debug("问卷反馈")
	else
		unilight.debug("用户反馈")
	end
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--聊天监控 原前端函数：cmd: "gm_chat_msg"
GmHttp.ParseRequestChatMessageGmUserPmd_CS = function (gmid, taskid, body)
	-- body
	local params = go.parseMultipartForm(taskid, body)
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		charid = tonumber(params.charid) or 0,
		type = tonumber(params.type) or 0,
		chatdate = tonumber(params.chatdate) or 0,
		curpage = tonumber(params.curpage) or 0,
		perpage = 500,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestChatMessageGmUserPmd_CS", resStr)

	unilight.debug("聊天信息")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end
