local _M = {}
local json = require "cjson"

-- entries is a hash table
_M.append = function (topic, entries)
    local res = ngx.location.capture(
        '/append_log_proxy/' .. topic,
        { body = json.encode({topic = topic, entries = entries}) }
    )
    return res
end

return _M
