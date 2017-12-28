local json = require "cjson"
local routes = require "routes"

local env = os.getenv('SPACER_ENV')

ngx.req.read_body()

local path = ngx.var.uri
local method = ngx.var.request_method

local module = nil

for i, route in ipairs(routes) do
    if route[1] == method and route[2] == path then
        module = route[3]
    end
end

if module == nil then
    ngx.status = 404
    ngx.say(json.encode({["error"] = "not found"}))
    return ngx.exit(ngx.HTTP_OK)
end

local ok, func = pcall(require, module)
if not ok then
    -- `func` will be the error message if error occured
    ngx.log(ngx.ERR, func)
    if string.find(func, "not found") then
        ngx.status = 404
    else
        ngx.status = 500
    end
    if env == 'production' then
        func = 'Internal Server Error'
    end
    ngx.say(json.encode({["error"] = func}))
    return ngx.exit(ngx.HTTP_OK)
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
        ngx.status = 400
        ngx.say(json.encode({["error"] = ret.err}))
        ngx.log(ngx.ERR, ret.err)
        return ngx.exit(ngx.HTTP_OK)
    else -- unknown exception
        -- TODO: don't return exception in production
        if env == 'production' then
            ret = 'We\'re sorry, something went wrong'
        end
        ngx.status = 500
        ngx.say(json.encode({["error"] = ret}))
        ngx.log(ngx.ERR, ret)
        return ngx.exit(ngx.HTTP_OK)
    end
end

ngx.say(json.encode({data = ret}))
