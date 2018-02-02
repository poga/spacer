local _M = {}
local INTERNAL_TOKEN = require "internal_token"

local json = require "cjson"

_M.call = function (name, body)
    local res = ngx.location.capture('/' .. name, {args = { spacer_internal_token = INTERNAL_TOKEN}})
    if res.status ~= 200 then
        return error(res.body)
    end
    return json.decode(res.body)["data"]
end

return _M
