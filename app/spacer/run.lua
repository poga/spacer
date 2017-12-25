local json = require "cjson"

ngx.req.read_body()

-- TODO: not safe at all
local path = ngx.var.uri
local ok, func = pcall(require, path)

-- ngx.log(ngx.ERR, "running " .. path .. " " , ngx.is_subrequest, " ", ngx.req.get_headers()["X-SPACER-TRACE-TOKEN"])

if not ok then
    ngx.log(ngx.ERR, res)
    return ngx.exit(500)
end

local body = ngx.req.get_body_data()
local event = {}
if body then
    event.body = json.decode(body)
end

local context = {}

function context.error (err)
    error({t = "error", err = err})
end

function context.fatal (err)
    error(err)
end

local ok, ret = pcall(func, event, context)

if not ok then
    if ret.t == "error" then -- user error
        ngx.say(json.encode({["error"] = ret.err}))
        ngx.log(ngx.ERR, ret.err)
        return ngx.exit(200)
    else -- unknown exception
        -- TODO: don't return exception in production
        ngx.say(json.encode({["error"] = ret}))
        ngx.log(ngx.ERR, ret)
        return ngx.exit(200)
    end
end

ngx.say(json.encode({data = ret}))
