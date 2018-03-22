local make_router = require "make_router"
local gateway_config = require "gateway"

local router = make_router(gateway_config)

local ENV = os.getenv("SPACER_ENV")
local INTERNAL_TOKEN = require "internal_token"

local temp = {}

local run = function (func_path, params)
    temp.func_path = func_path
end

router.handler = function (func_path, params)
    run(func_path, params)
end

local uri = ngx.var.uri
-- remove trailing slash
if string.sub(uri, -1) == '/' then
    uri = string.sub(uri, 0, -2)
end

local params = {}

-- check if this is an internal request
if ngx.req.get_uri_args()["spacer_internal_token"] == INTERNAL_TOKEN then
    -- if it's an internal request, skip router and call the specified function directly
    run(ngx.var.uri, params)

    return temp.func_path
elseif string.sub(ngx.var.uri, 0, 8) == '/private' then
    -- also skip router if it's a private path (protected by nginx)
    run(ngx.var.uri, params)

    return temp.func_path
else
    local ok, errmsg = router:route(
       ngx.var.request_method,
       uri,
       ngx.req.get_uri_args(),  -- all these parameters
       params)         -- into a single "params" table

    if not ok then
        return "not found"
    end

    return temp.func_path
end

