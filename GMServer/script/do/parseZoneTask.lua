ZoneTask = ZoneTask or {}

--[[
--该模块用以处理接收到的游戏服消息，主要是检索数据库返回给游戏服(比如游戏服查询当前公告)、回复GMClient发送的请求
--主要使用函数有sendCmd2ZoneTask(taskid, cmd)，回复ZoneTask的请求； sendCmd2GmTask(taskid, cmd),此处的taskid即gmid
--]]

--[[
ZoneTask.ParseRequestUserInfoGmUserPmd_S = function(taskid, cmd)   
	local taskid = cmd.GetGmid()
    go.sendCmd2GmTask(taskid, cmd) 
end
--]]



