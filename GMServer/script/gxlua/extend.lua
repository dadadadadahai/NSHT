-- 通用lua扩充函数，与任何游戏逻辑无关

-- from http://snippets.luacode.org/?p=snippets/String_to_Hex_String_68
string.tohex = function(str, spacer)
	return  string.gsub(str,"(.)", function (c) return string.format("%02X%s", string.byte(c), spacer or "") end)
end

string.trim = function(str)
	return string.gsub(str, "^%s*(.-)%s*$", "%1")
end
string.trimbegin = function(str)
	return string.gsub(str, "^%s*(.-)$", "%1")
end
string.trimend = function(str)
	return string.gsub(str, "^(.-)%s*$", "%1")
end

string.padleft = function(str, totalWidth, paddingChar)
	local len = #str
	if len >= totalWidth then
		return str
	else
		paddingChar = paddingChar or ' '
		assert(#paddingChar == 1)
		return string.rep(paddingChar, totalWidth - len) .. str
	end
end
string.padright = function(str, totalWidth, paddingChar)
	local len = #str
	if len >= totalWidth then
		return str
	else
		paddingChar = paddingChar or ' '
		assert(#paddingChar == 1)
		return str .. string.rep(paddingChar, totalWidth - len)
	end
end

string.split = function(szFullString, szSeparator)
	local FindStartIndex = 1
	local SplitArray = {}
	while true do
		local FindLastIndex = string.find(szFullString, szSeparator, FindStartIndex, true)
		if not FindLastIndex then
			table.insert(SplitArray, string.sub(szFullString, FindStartIndex, string.len(szFullString)))
			break
		end
		table.insert(SplitArray, string.sub(szFullString, FindStartIndex, FindLastIndex-1))
		FindStartIndex = FindLastIndex + string.len(szSeparator)
	end
	return SplitArray
end

table.find = function(this, value)
	for k,v in pairs(this) do
		if v == value then return k end
	end
end

table.tostring = function(data, _indent)
	local visited = {}
	local function dump(data, prefix)
		local str = tostring(data)
		if table.find(visited, data) ~= nil then return str end
		table.insert(visited, data)

		local prefix_next = prefix .. "  "
		str = str .. "\n" .. prefix .. "{"
		for k,v in pairs(data) do
			if type(k) == "number" then
				str = str .. "\n" .. prefix_next .. "[" .. tostring(k) .. "] = "
			else
				str = str .. "\n" .. prefix_next .. tostring(k) .. " = "
			end
			if type(v) == "table" then
				str = str .. dump(v, prefix_next)
			elseif type(v) == "string" then
				str = str .. '"' .. v .. '"'
			else
				str = str .. tostring(v)
			end
		end
		str = str .. "\n" .. prefix .. "}"
		return str
	end
	return dump(data, _indent or "")
end

table.merge = function(base, delta)
	if type(delta) ~= "table" then return end
	for k,v in pairs(delta) do
		base[k] = v
	end
end

table.extend = function(base, delta)
	if type(delta) ~= "table" then return end
	for i,v in ipairs(delta) do
		table.insert(base, v)
	end
end

table.len = function(tbl)
	if type(tbl) ~= "table" then return 0 end
	local n = 0
	for k,v in pairs(tbl) do n = n + 1 end
	return n
end

table.empty = function(tbl)
	if tbl == nil then return true end
	assert(type(tbl) == "table")
	if #tbl > 0 then return false end
	for k,v in pairs(tbl) do return false end
	return true
end

-- http://snippets.luacode.org/?p=snippets/Deep_copy_of_a_Lua_Table_2
table.clone = function(t)
	if type(t) ~= 'table' then return t end
	local mt = getmetatable(t)
	local res = {}
	for k,v in pairs(t) do
		if type(v) == 'table' then
			v = table.clone(v)
		end
		res[k] = v
	end
	setmetatable(res,mt)
	return res
end

-- http://snippets.luacode.org/?p=snippets/Table_Slice_116
table.slice = function(values,i1,i2)
	local res = {}
	local n = #values
	i1 = i1 or 1
	i2 = i2 or n
	if i2 < 0 then
		i2 = n + i2 + 1
	elseif i2 > n then
		i2 = n
	end
	if i1 < 1 or i1 > n then
		return {}
	end
	local k = 1
	for i = i1,i2 do
		res[k] = values[i]
		k = k + 1
	end
	return res
end

table.reverse = function(tab)
	local size = #tab
	local newTable = {}
	for i,v in ipairs(tab) do
		newTable[size+1-i] = v
	end
	return newTable
end

-- math.random({0.7, 0.1, 0.2}, {'A', 'B', 'C'})
math.random = function(m, n)
	if type(m) == "table" and #m == #n then
		-- 标准化概率表
		local sum = 0
		for _,v in ipairs(m) do sum = sum + v end
		local sm = {}
		for k,v in ipairs(m) do sm[k] = v / sum end
		-- 得到下标
		local r = go.rand.Random()
		for k,v in ipairs(sm) do
			if r <= v then return n[k]
			else r = r - v end
		end
		assert(false)
	end

	if m == nil then return go.rand.Random() end
	local _random = function(m, n)
		m, n = math.min(m, n), math.max(m, n)
		local mi, mf = math.modf(m)
		local ni, nf = math.modf(n)
		if mf == 0 and nf == 0 then
			return go.rand.RandBetween(m, n)
		else
			return m + go.rand.Random() * (n - m)
		end
	end
	if n == nil then return _random(1, m) end
	return _random(m, n)
end

-- http://www.cplusplus.com/reference/algorithm/random_shuffle/
-- http://stackoverflow.com/questions/17119804/lua-array-shuffle-not-working
math.shuffle = function(array)
	local counter = #array
	while counter > 1 do
		local index = math.random(counter)
		array[index], array[counter] = array[counter], array[index]
		counter = counter - 1
	end
	return array
end
