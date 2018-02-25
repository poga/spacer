local json = require "cjson"

local reject = function (status, body)
    ngx.status = status
    ngx.say(json.encode(body))
    ngx.exit(ngx.HTTP_OK)
end

return reject
