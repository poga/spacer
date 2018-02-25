local gateway_def = require "gateway"
local json = require "cjson"
local router = require "router"

local _M = {}

-- build router
local r = router.new()

local handler = function (func_path)
    return function (params)
        _M.handler(func_path, params)
    end
end

for i, route in ipairs(gateway_def) do
    local method = route[1]
    local route_ptrn = route[2]
    local func_path = route[3]

    if     method == "GET"   then r:get(route_ptrn, handler(func_path))
    elseif method == "POST"  then r:post(route_ptrn, handler(func_path))
    elseif method == "DELTE" then r:delete(route_ptrn, handler(func_path))
    elseif method == "PUT"   then r:put(route_ptrn, handler(func_path))
    end
end

_M.router = r

function _M:route(method, uri, uri_args, params)
    return _M.router:execute(method, uri, uri_args, params)
end

return _M
