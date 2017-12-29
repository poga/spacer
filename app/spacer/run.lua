local json = require "cjson"
local gateway = require "gateway"
local env = os.getenv('SPACER_ENV')

local reject = function (status, body)
    ngx.status = status
    ngx.say(json.encode(body))
    ngx.exit(ngx.HTTP_OK)
end

ngx.req.read_body()

local module = gateway(ngx.var.request_method, ngx.var.uri)

if module == nil then return reject(404, {["error"] = "not found"}) end

local ok, func = pcall(require, module)
if not ok then
    -- `func` will be the error message if error occured
    ngx.log(ngx.ERR, func)
    local status = nil
    if string.find(func, "not found") then
        ngx.status = 404
    else
        ngx.status = 500
    end
    if env == 'production' then
        func = 'Internal Server Error'
    end

    return reject(status, {["error"] = func})
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
        ngx.log(ngx.ERR, ret.err)
        return reect(400, {["error"] = ret.err})
    else -- unknown exception
        ngx.log(ngx.ERR, ret)
        if env == 'production' then
            ret = 'We\'re sorry, something went wrong'
        end
        return reect(500, {["error"] = ret})
    end
end

ngx.say(json.encode({data = ret}))