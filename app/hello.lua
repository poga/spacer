local bar = require("bar")

local G = function (event, ctx)
    if event.body.bar == 0 then
        ctx.error("bar can't be zero")
    end
    local res = ngx.location.capture('/bar')
    ngx.log(ngx.ERR, res.body)
    return 42 + bar()
end

return G
