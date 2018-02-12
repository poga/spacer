local lu = require 'luaunit'
local m = require 'hello'

testT = {}

function testT:test_ret ()
    lu.assertEquals(m(), "Hello from Spacer!")
end

os.exit(lu.LuaUnit.run())
