local token = os.getenv('SPACER_INTERNAL_TOKEN')

if token == "" then
    token = ngx.var.hostname
end

return token
