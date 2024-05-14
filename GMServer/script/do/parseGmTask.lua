GmTask = GmTask or {}

--[[
--该模块用以处理接收到的GMClient的消息，主要是操作数据库，转发消息到ZoneTask(游戏服务器)
--主要使用函数有forwardCmd2ZoneTask(gmid, gameid, zoneid, send, recvcmd),最后这个参数主要用以兼容http请求，并将内容序列化成proto返回
--]]

--[[
GmTask.ParseRequestUserInfoGmUserPmd_C = function(taskid, cmd)   
    local protoname = "*Pmd.RequestUserInfoGmUserPmd_S"
    local data = {} 
    go.forwardCmd2ZoneTask(taskid, cmd.GetGameid(), cmd.GetZoneid(), cmd, go.buildProto(protoname, data)) 
end
--]]



