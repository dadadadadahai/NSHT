
CreateClass("UniTimerClass")
CreateClass("UniEventClass")


function NewUniTimerClass(callback, msec, ...)
	local timer = {
		nextmsec=unitimer.now,
		callback = callback,
		tick = msec,
		params = arg,
	}
	UniTimerClass:New(timer)
	return timer
end
function NewUniEventClass(callback, msec,maxtimes, ...)
	local event = {
		nextmsec=unitimer.now+msec,
		callback = callback,
		tick = msec,
		maxtimes = maxtimes,
		params = arg,
	}
	UniEventClass:New(event)
	return event
end
function UniTimerClass:GetId()
	return self.tick
end
function UniTimerClass:GetName()
	return self.tick
end
function UniTimerClass:Stop()
	self.stop = true
	unitimer.removetimer(self)
end
function UniTimerClass:Check(now)
	if self.nextmsec <= now then
		--unilight.error("UniEventClass:"..unpack(self.params))
		self.callback(unpack(self.params))
		self.nextmsec = now + self.tick
		return true
	end
	--duaration=0
	--first=true
	--pause=false
	--stop=false
	return false
end
function UniEventClass:GetId()
	return self.tick
end
function UniEventClass:GetName()
	return self.maxtimes
end
function UniEventClass:Check(now)
	if self.nextmsec <= now then
		self.callback(unpack(self.params))
		self.nextmsec = now + self.tick
		self.maxtimes = self.maxtimes - 1
		return true
	end
	return false
end
function UniEventClass:Stop()
	unitimer.removeevent(self)
end
unitimer={
	timermap={},
	eventmap={},
}
function unitimer.init(msec) --最小精度的定时器
	unitimer.now = go.time.Msec()
	if unitimer.ticktimer ~= nil then
		unilight.error("unitimer.init已经初始化过了:"..unitimer.tickmsec)
		return false
	end
	unitimer.tickmsec = msec --保存最小精度
	unitimer.ticktimer = unilight.addtimermsec("unitimer.loop", msec)
	return true
end
function unitimer.loop()
	unitimer.now = go.time.Msec()
	if next(unitimer.timermap) ~= nil then
		for k,v in pairs(unitimer.timermap) do
			if v:Check(unitimer.now) == true then
				if v.stop == true then
					unitimer.removetimer(v)
				end
			end
		end
	end
	if next(unitimer.eventmap) ~= nil then
		for k,v in pairs(unitimer.eventmap) do
			if v:Check(unitimer.now) == true then
				if v.maxtimes <= 0 then
					--v:Debug("事件结束")
					v:Stop()
				end
			end
		end
	end
end
function unitimer.addtimermsec(callback, msec, ...)
	local timer = NewUniTimerClass(callback, msec, ...)
	--timer:Debug("unitimer.addtimermsec:")
	unitimer.timermap[timer]=timer
	return timer
end
function unitimer.addtimer(callback, msec, ...)
	return unitimer.addtimermsec(callback, msec*1000, ...)
end
function unitimer.removetimer(timer)
	--timer:Debug("unitimer.removetimer:")
	unitimer.timermap[timer] = nil
end
function unitimer.addevent(callback, msec,maxtimes, ...)
	return unitimer.addeventmsec(callback, msec*1000,maxtimes, ...)
end
function unitimer.addeventmsec(callback, msec,maxtimes, ...)
	maxtime = maxtimes or 1
	local event = NewUniEventClass(callback, msec,maxtimes, ...)
	--event:Debug("设置事件:"..msec)
	unitimer.eventmap[event]=event
	return event
end
function unitimer.removeevent(event)
	unitimer.eventmap[event] = nil
end
