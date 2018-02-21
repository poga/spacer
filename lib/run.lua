local json = require "cjson"
local route = require "route"
local context = require "context"

local ENV = os.getenv('SPACER_ENV')
local INTERNAL_TOKEN = require "internal_token"

local reject = function (status, body)
    ngx.status = status
    ngx.say(json.encode(body))
    ngx.exit(ngx.HTTP_OK)
end

ngx.req.read_body()

local tmp = {func_path = nil, params = nil}
local R = route(function(func_path)
    tmp.func_path = func_path
    return function (params)
        tmp.params = params
    end
end)

local uri = ngx.var.uri
-- remove trailing slash
if string.sub(uri, -1) == '/' then
    uri = string.sub(uri, 0, -2)
end

-- parse request json body
local body = ngx.req.get_body_data()
local params = {}
if body then
    params = json.decode(body)
end

-- check if this is an internal request
if ngx.req.get_uri_args()["spacer_internal_token"] == INTERNAL_TOKEN then
    -- if it's an internal request, skip router and call the specified function directly
    tmp.func_path = ngx.var.uri
else
    local ok, errmsg = R:execute(
       ngx.var.request_method,
       uri,
       ngx.req.get_uri_args(),  -- all these parameters
       params)         -- into a single "params" table

    if not ok then
        reject(404, {["error"] = "not found"})
    end
end

if tmp.func_path == nil then return reject(404, {["error"] = "not found"}) end

local ok, func = pcall(require, tmp.func_path)
if not ok then
    -- `func` will be the error message if error occured
    ngx.log(ngx.ERR, func)
    local status = nil
    if string.find(func, "not found") then
        status = 404
    else
        status = 500
    end
    if ENV == 'production' then
        func = 'Internal Server Error'
    end

    return reject(status, {["error"] = func})
end

local ok, ret = pcall(func, tmp.params, context)

if not ok then
    if ret.t == "error" then -- user error
        ngx.log(ngx.ERR, ret.err)
        return reject(400, {["error"] = ret.err})
    else -- unknown exception
        ngx.log(ngx.ERR, ret)
        if ENV == 'production' then
            ret = 'We\'re sorry, something went wrong'
        end
        return reject(500, {["error"] = ret})
    end
end

ngx.say(json.encode({data = ret}))
