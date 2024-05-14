unilobby = unilobby or {}
--[[大厅服务器链接进来的回调
Lby.lobby_connect = function(cmd, lobbytask) 
    unilight.info("区服务器回调：新的大厅链接成功 " .. lobbytask.GetGameId() .. ":" .. lobbytask.GetZoneId())
    local lobbytask = unilobby.getlobbytask()
    if lobbytask == nil then
        return 
    end
    local req = {
        ["do"] = "Cmd.UserInfoSynReturnLbyCmd_S",
        ["data"] = {
            resultCode   = 1, 
            desc = "ok"
        }

    }
    unilight.success(lobbytask, req)
end 
--]]

--[[大厅服务器断开的回调
Lby.lobby_disconnect = function(cmd, lobbytask) 
    unilight.info("区服务器回调：与大厅失联了" .. lobbytask.GetGameId() .. ":" .. lobbytask.GetZoneId())
end 
--]]

---[[获取大厅的lobbytask
unilobby.getlobbytask = function()
    return go.lobby
end
--]]

--[[主动向大厅发送消息
local lobbytask = unilobby.getlobbytask()
if lobbytask == nil then
    return 
end
local req = {
    ["do"] = "Cmd.RequestUserinfoLobbyCmd_C",
    ["data"] = {
        uid = 10000,
    },
}
unilight.success(lobbytask, req)
--]]
