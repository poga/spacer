local gateway_def = require "gateway"
local json  = require "cjson"

local G = function (handler)
    local router = require "router"
    local r = router.new()


    -- build router
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

    return r
end

return G
