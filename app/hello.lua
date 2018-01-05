local service = require "service"

local G = function (event, ctx)
    local ret = service.call("lib/foo")
    return "Hello Spacer " .. ret
end

return G
