local json = require "cjson"
local make_router = require "make_router"
local reject = require "reject"
local gateway_config = require "gateway"

local router = make_router(gateway_config)

local ENV = os.getenv("SPACER_ENV")
local INTERNAL_TOKEN = require "internal_token"

local run = function (func_path, params)
    if func_path == nil then return reject(404, {["error"] = "not found"}) end

    local ok, func = pcall(require, func_path)
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

    local ok, ret, err = pcall(func, params)

    if not ok then
        -- unknown exception thown by error()
        ngx.log(ngx.ERR, ret)
        if ENV == 'production' then
            ret = 'We\'re sorry, something went wrong'
        end
        return reject(500, {["error"] = ret})
    end

    -- function returned the second result as error
    if err then
        ngx.log(ngx.ERR, err)
        return reject(400, {["error"] = err})
    end

    ngx.say(json.encode({data = ret}))
end

router.handler = function (func_path, params)
    run(func_path, params)
end

ngx.req.read_body()

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
    run(ngx.var.uri, params)
elseif string.sub(ngx.var.uri, 0, 8) == '/private' then
    -- also skip router if it's a private path (protected by nginx)
    run(ngx.var.uri, params)
else
    local ok, errmsg = router:route(
       ngx.var.request_method,
       uri,
       ngx.req.get_uri_args(),  -- all these parameters
       params)         -- into a single "params" table

    if not ok then
        reject(404, {["error"] = "not found"})
    end
end

