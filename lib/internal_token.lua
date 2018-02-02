local token = os.getenv('SPACER_INTERNAL_TOKEN')

if token == nil then
    local uuid = require "uuid"
    uuid.seed()
    token = uuid.generate_v4()
end

return token
