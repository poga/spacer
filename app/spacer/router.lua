local route = require "gateway"

local router = function (method, path)
    local module = nil
    for i, route in ipairs(route) do
        if route[1] == method and route[2] == path then
            module = route[3]
        end
    end
    return module
end

return router
