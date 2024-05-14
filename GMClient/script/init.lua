require "script/gxlua/unilight"

if Do == nil then Do = {} end
local function init()
	for _,file in pairs(unilight.tablefiles()) do
		unilight.debug("正在加载脚本:"..file)
		dofile(file)
	end
	for _,file in pairs(unilight.scriptfiles()) do
		unilight.debug("正在加载脚本:"..file)
		dofile(file)
	end

	unilight.debug("初始化lua脚本成功")

	-- 覆盖 unilight.lua 中的默认实现
	unilight.response2 = function(w, req)
		req.st = os.time()
		-- local s = json.encode(encode_repair(req))
		local s = json.encode(req)
		w.SendString(s)
		unilight.debug("[send] " .. s)
	end
end

init()
