

function GetClassName(name)
	for k,v in pairs(_G) do
		if name==v then
			return k
		end 
	end
end
function CreateClass(name)
	if _G[name] ~= nil then
		unilight.error("CreateClass被意外全局初始化,这里强制重置成类:"..name)
		return nil
	end
	_G[name] = {}
	local class = _G[name]
	class.__classname = name
	function class:New(initclass)
		local new = initclass or {}
		setmetatable(new, { __index = self})
		return new
	end
	function class:SetClassName(cname)
		self.__classname = cname or self.__classname
	end
	function class:GetLogPrefix()
		local id = nil
		local name = ""
		if self.id then
			id = self.id
		elseif self.Id then
			id = self.Id
		elseif self.tempid then
			id = self.tempid
		elseif self.Tempid then
			id = self.Tempid
		elseif self.GetId then
			id = self:GetId()
		end
		if self.name then
			name = self.name
		elseif self.Name then
			name = self.Name
		elseif self.GetName then
			name = self:GetName()
		end
		local id = id or ""
		local name = name or ""
		return self.__classname .. "[" .. id .."," ..name.. "] " 
	end
	function class:Debug(...)
		unilight.debug(self:GetLogPrefix() .. unpack(arg))
	end
	function class:Info(...)
		unilight.info(self:GetLogPrefix() .. unpack(arg))
	end
	function class:Warn(...)
		unilight.warn(self:GetLogPrefix() .. unpack(arg))
	end
	function class:Error(...)
		unilight.error(self:GetLogPrefix() .. unpack(arg))
	end
	return class
end
--Class = CreateClass("Class")
--new = Class:New()
--new:Debug("whj")
