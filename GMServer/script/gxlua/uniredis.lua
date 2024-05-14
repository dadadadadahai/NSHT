unilight.initredisdb = function()
	db = 100 
	-- 切换db
	unilight.redis_select = function(dbnum)
		dbnum = dbnum or 0 
		if type(dbnum) ~= "number" then
			unilight.error("redis dbnum type is not number")
			return false
		end
		if dbnum ~= db then
			unilight.info("redis_select from " .. db .."   to " .. dbnum )
			db = dbnum
			unilight.REDISDB.Select(db)
		end
		return true
	end

	--给key设置过期时间
	unilight.redis_setexpire = function(key, seconds, dbnum)
		if type(key) ~= "string" or type(seconds) ~= "number" then
			return  "key or seconds type error"
		end
		if not unilight.redis_select(dbnum) then
			return "db number type err"
		end
		local _, err = unilight.REDISDB.Expire(key, seconds)
		if err ~= nil then
			unilight.error("redissetexpire err key " .. key.. err)
			return err 
		end
	end

	-- 移除key
	unilight.redis_rmkey = function(key, dbnum)
		if type(key) ~= "string" then
			return  "key type error"
		end
		if not unilight.redis_select(dbnum) then
			return "db number type err"
		end
		local _, err = unilight.REDISDB.Del(key)
		if err ~= nil then
			unilight.error("redisrmkey err key " .. key.. err)
			return err 
		end
	end

	-- 设置数据:String类型
	unilight.redis_setdata = function(key, value, dbnum)
		if type(key) ~= "string" or type(value) ~= "string" then
			return  "key or value type error"
		end
		if not unilight.redis_select(dbnum) then
			return "db number type err"
		end
		local err = unilight.REDISDB.Set(key, value, 0, 0, false, false)
		if err ~= nil then
			unilight.error("redissetdata err key " .. key .. "  value  " .. value .. "  ".. err)
			return err 
		end
		return nil
	end

	-- 获取数据:String类型
	unilight.redis_getdata = function(key, dbnum)
		if type(key) ~= "string"  then
			return nil, "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return nil, "db number type err"
		end

		local value, err = unilight.REDISDB.GetRange(key, 0, -1)
		if err ~= nil then
			unilight.error("redis_getdata error " .. err)
			return nil, err 
		end
		return tostring(value)
	end

	-- 设置数据:hash类型
	unilight.redis_sethashdata = function(key, field, value, dbnum)
		if type(key) ~= "string"  or type(field) ~= "string"  then
			return "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  "db number type err"
		end
		local bok, err = unilight.REDISDB.HSet(key, field, value)
		if err ~= nil then
			unilight.error("redis sethash err " .. err)
			return err 
		end

		return nil
	end
	
	unilight.redis_sethashmultdata = function(key, values, dbnum)
		if type(key) ~= "string" or  type(values) ~= "table" then
			return "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  "db number type err"
		end
		local bok, err = unilight.REDISDB.HMSet(key, values)
		if err ~= nil then
			unilight.error("redis sethashmultfield err ") 
			return err 
		end
		return nil
	end

	-- 获取数据:hash类型
	unilight.redis_gethashdata = function(key, field, dbnum)
		if type(key) ~= "string" or type(field) ~= "string" then
			return nil, "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  nil, "db number type err"
		end

		local value, err = unilight.REDISDB.HGet(key, field)
		if err ~= nil then
			unilight.error("redis gethashmultfield err " .. err)
			return nil, err 
		end
		return value
	end

	unilight.redis_gethashdata_Str = function(key, field, dbnum)
		if type(key) ~= "string" or type(field) ~= "string" then
			return nil, "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  nil, "db number type err"
		end

		local value, err = unilight.REDISDB.HGet_Str(key, field)
		if err ~= nil then
			unilight.error("redis gethashmultfield err " .. err)
			return nil, err 
		end
		return value
	end

	unilight.redis_gethashmultdata = function(key, dbnum)
		if type(key) ~= "string" then
			return nil, "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  nil, "db number type err"
		end
		local values, err = unilight.REDISDB.HGetAll(key)
		if err ~= nil then
			unilight.error("redis gethashmultfield err " .. err)
			return nil, err 
		end
		return values 
	end

	unilight.redis_gethashmultdata_Str = function(key, dbnum)
		if type(key) ~= "string" then
			return nil, "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  nil, "db number type err"
		end
		local values, err = unilight.REDISDB.HGetAll_Str(key)
		if err ~= nil then
			unilight.error("redis gethashmultfield err " .. err)
			return nil, err 
		end
		return values 
	end

	--设置数据:list类型
	unilight.redis_setlistdata = function(key, value, dbnum)
		if type(key) ~= "string" or type(value) ~= "string" then
			return nil, "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  nil, "db number type err"
		end

		local value, err = unilight.REDISDB.LPush(key, value)
		if err ~= nil then
			unilight.error("redis LPushmultfield err " .. err)
			return nil, err 
		end
		return value
	end

	--获取数据:list类型
	unilight.redis_getlistdata = function(key, index, dbnum)
		if type(key) ~= "string" then
			return nil, "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  nil, "db number type err"
		end

		local value, err = unilight.REDISDB.LIndex(key,index)

		if err ~= nil then
			unilight.error("redis LIndexmultfield err " .. err)
			return nil, err 
		end
		--return value
		local len = #value
		local str = ""
		for i=1,len do
			local s = string.char(tonumber(value[i]))
			str = str..s
		end
		return str
	end

	--移除数据:list类型
	unilight.redis_rmlistdata = function(key, count, value, dbnum)
		if type(key) ~= "string" or type(value) ~= "string" then
			return nil, "key or value type error"
		end

		if not unilight.redis_select(dbnum) then
			return  nil, "db number type err"
		end

		local value, err = unilight.REDISDB.LRem(key, count, value)

		if err ~= nil then
			unilight.error("redis LRemmultfield err " .. err)
			return nil, err 
		end
	end
end
