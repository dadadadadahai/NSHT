GmClient = GmClient or {}

--参数含义taskid GMClient的标识，rev是GMServer回复的proto包
GmClient.ParseRequestModifyPriGmUserPmd_CS = function(taskid, rev)
	local httptaskid = rev.GetClientid()
	--收到GMServer的回复后，直接转发给相应的http客户端
	go.forwardCmd2HttpTask(httptaskid, rev)
end






