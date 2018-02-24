local _M = {}
local json = require "cjson"
local service = require "service"

-- entries is a hash table
_M.append = function (topic, entries)
    return service.call(
        "append_log_proxy/" .. topic,
        {topic = topic, entries = entries}
    )
end

return _M
