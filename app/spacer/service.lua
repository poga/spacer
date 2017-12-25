local _M = {}

_M.set = function (name, body)
    local res = ngx.location.capture('/' .. name)
end

return _M
