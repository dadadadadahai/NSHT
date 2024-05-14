-- event on server
function StartOver()
	local ql = unilight.startsql().select({"id", "username", "password"}).table("gm_user")
    local res = ql.run()
    for i, v in ipairs(res) do
	    unilight.debug(v.id .. "  " .. v.username.. "  " ..  v.password)

    end
end
