unilight = unilight or {}

--[[
	向指定url请求GET http服务
	resFunc:http请求回调函数
	url:请求http服务器的url
	msg:请求的数据,这时里是一个lua的table
	heads 在这里是一个 map[string]string 选定对应参考与值
]]

unilight.HttpRequestGet = function(resFunc, url, heads, para)
	heads = heads or {}
    para = para or {}
	if type(resFunc) ~= "string" or type(url) ~= "string" or type(heads) ~= "table" then
		unilight.error("unilight.HttpRequestGet params error" .. resFunc .. url)
		return
	end
	reqMsg = json.encode(msg)
	callbackpara = json.encode(para)
	go.httpclient.HttpRequestGet(0, callbackpara, resFunc, url, heads)
end

--[[
	向指定url请求POS http服务
	resFunc:http请求回调函数
	url:请求http服务器的url
	msg:请求的数据,这时里是一个lua的table
	heads 在这里是一个 map[string]string 选定对应参考与值
]]

unilight.HttpRequestPost = function(resFunc, url, body, bodyType, heads, para)
    para = para or {}
	heads = heads or {}
	bodyType = bodyType or "application/x-www-form-urlencoded"
	if type(resFunc) ~= "string" or type(url) ~= "string" or type(body) ~= "table" or type(bodyType) ~= "string"or type(heads) ~= "table" then
		unilight.error("unilight.HttpRequestGet params error" .. resFunc .. url)
		return
	end
	reqMsg = json.encode(body)
	callbackpara = json.encode(para)
	go.httpclient.HttpRequestPost(0, callbackpara, resFunc, url, bodyType, reqMsg, heads)
end

-- deomo: 向指定url发送post请求，其中heads， bodytype都采用缺省方式
function testHttpRequest()
	local req = {
            data = {
                srcImage = "http://98pokerstatic-a.akamaihd.net/BJLSTATIC/head/woman/143.jpg"
            }
		}
	unilight.HttpRequestPost("Echo", "http://14.17.104.56:8888/exchangeurltomyserver", req, nil, nil, {para="1"})
end

Http.Echo = function (cmd, para)
	unilight.info("receive http res" .. table.tostring(cmd))
	unilight.info("receive http para " .. para)
end
