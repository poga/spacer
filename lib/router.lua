local route = require "gateway"

local router = function (method, path)
    local module = nil
    -- remove trailing slash
    if string.sub(path, -1) == '/' then
        path = string.sub(path, 0, -2)
    end

    for i, route in ipairs(route) do
        if route[1] == method and route[2] == path then
            module = route[3]
        end
    end
    return module
end

return router
