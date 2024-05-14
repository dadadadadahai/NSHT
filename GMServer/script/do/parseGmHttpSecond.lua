--[[
--解析客户端http发来的参数，构造proto协议转发给GMServer
--
--]]

GmHttp = GmHttp or {}

--奔驰宝马(四海一家)控制
GmHttp.ParseUniversalOneControlPmd_CS = function(gmid, taskid, body)
	local params = unilight.getreq(tostring(body))
	params['clientid'] = taskid
	params['gmid'] = gmid
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		gmid = gmid,
		clientid = taskid,
		msg = json.encode({
			["do"] = "Pmd.UniversalOneControlPmd_CS",
			data = params
		})
	}
	local resStr = json.encode(encode_repair(data))

	local resProto = go.buildProto("*Pmd.RequestExecGmCommandGmPmd_SC",resStr)
	unilight.debug("四海一家控制配置!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end


--邮件发送
GmHttp.ParseRequestSendMailGmUserPmd_CS = function(gmid, taskid, body)
	--解析http请求的参数form格式的，如果是json格式的，则使用unilight.getreq(tostring(body));格式需自行跟客户端商量
	local params = unilight.getreq(tostring(body))

	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		data = params.data,
		clientid = taskid,
		gmid = gmid,
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.RequestSendMailGmUserPmd_CS", resStr)

	unilight.debug("邮件发送")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end


--切换两个账号信息
GmHttp.ParseChangeUserUidGmUserPmd_CS = function(gmid, taskid, body)
	local params = unilight.getreq(tostring(body))

	local gmdata = params.gmdata
	gmdata['clientid'] = taskid
	gmdata['gmid'] = gmid

	unilight.info(table.tostring(gmdata))
	params.gmdata = gmdata

	local data = {
		gameid = tonumber(params.gmdata.gameid) or 0,
		zoneid = tonumber(params.gmdata.zoneid) or 0,
		gmid = gmid,
		clientid = taskid,
		msg = json.encode({
			["do"] = "Pmd.ChangeUserUidGmUserPmd_CS",
			data = json.encode(params)
		})
	}
	local resStr = json.encode(encode_repair(data))

	local resProto = go.buildProto("*Pmd.RequestExecGmCommandGmPmd_SC",resStr)
	unilight.debug("切换两个账号信息!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end


-- 实物兑换码功能
GmHttp.ParseRedpackCodeSearchGmUserPmd_CS = function(gmid, taskid, body)
	local params = unilight.getreq(tostring(body))
	
	params['clientid'] = taskid
	params['gmid'] = gmid
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		gmid = gmid,
		clientid = taskid,
		msg = json.encode({
			["do"] = "Pmd.RedpackCodeSearchGmUserPmd_CS",
			data = params
		})
	}
	local resStr = json.encode(encode_repair(data))

	local resProto = go.buildProto("*Pmd.RequestExecGmCommandGmPmd_SC",resStr)
	unilight.debug("实物兑换码!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end

--活动控制 代替协议 ActivitySwitchGmUserPmd_CS
GmHttp.ParseActionControlGmUserPmd_CS = function (gmid, taskid, body)
	local params = unilight.getreq(tostring(body))
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		optype = tonumber(params.optype) or 0,
		actid = tonumber(params.actid) or 0,
		rdata = params.rdata,
		clientid = taskid, 
		gmid = gmid
	}

	local resStr = json.encode(encode_repair(data))
	local resProto = go.buildProto("*Pmd.ActionControlGmUserPmd_CS", resStr)
	unilight.debug("活动控制")
	unilight.debug(resStr)

	go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end



-- 请求,修改玩家输赢信息  2017-10-10 
GmHttp.ParseUserWeekMonthWinLoseInfoPmd_CS = function(gmid, taskid, body)
	local params = unilight.getreq(tostring(body))
	
	params['clientid'] = taskid
	params['gmid'] = gmid
	local data = {
		gameid = tonumber(params.gameid) or 0,
		zoneid = tonumber(params.zoneid) or 0,
		gmid = gmid,
		clientid = taskid,
		msg = json.encode({
			["do"] = "Pmd.UserWeekMonthWinLoseInfoPmd_CS",
			data = params
		})
	}
	local resStr = json.encode(encode_repair(data))

	local resProto = go.buildProto("*Pmd.RequestExecGmCommandGmPmd_SC",resStr)
	unilight.debug("请求或修改玩家输赢信息!")
    unilight.debug(resStr)
    --转发协议到GMServer
    go.forwardCmd2GmClient(taskid, data.gameid, data.zoneid, resProto)
end