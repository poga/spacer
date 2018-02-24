local _M = {}
local json = require "cjson"
local flow = require "flow"

-- entries is a hash table
_M.append = function (topic, entries)
    return flow.call(
        "append_log_proxy/" .. topic,
        {topic = topic, entries = entries}
    )
end

return _M
