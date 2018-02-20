# Spacer

Spacer provides a new way to build businesses around technology.

Spacer provides a fast **edit-save-reload** development cycle. No redeployment or rebuilding image is needed.

## Install

1. Install Dependencies

* [OpenResty](https://openresty.org/)
* [librdkafka](https://github.com/edenhill/librdkafka)

If you're on a Mac, `brew install openresty/brew/openresty librdkafka` should get everything you need.

2. Install Spacer

Spacer is written in [Go](https://golang.org/). It can be installed via `go get`.

```
$ go get -u github.com/poga/spacer
```

## Quick Start

```
$ spacer init ~/spacer-hello
$ cd ~/spacer-hello
$ ./bin/dev.sh
```

Open `http://localhost:3000/hello` and you should see spacer working.

### Hello World

Functions in spacer are written in Lua, a simple dynamic langauge. Here's a hello world function:

```lua
-- app/hello.lua
local G = function (params, context)
    return "Hello from Spacer!"
end

return G
```

Every function takes two arguments: `params` and `context`. For detail, check the **Functions** section.

### Test

Spacer have built-in test framework. Run all tests with command `./bin/test.sh`.

```
$ ./bin/test.sh
1..1
# Started on Mon Feb 12 17:46:48 2018
# Starting class: testT
ok     1	testT.test_ret
# Ran 1 tests in 0.000 seconds, 1 success, 0 failures
```

See `test/test_hello.lua` for example.

## Functions

Code in spacer are organized by functions. For now, spacer only support [Lua](https://www.lua.org/) as the programming language.

There are 2 way to invoke a function. The first is the simplest: just call it like a normal lua function.

```lua
-- app/bar.lua
local G = function (params, ctx)
  return params.val + 42
end

-- app/foo.lua
local bar = require "bar"

local G = function (params, ctx)
  return 100 + bar({val = 1}) -- returns 143
end
```

The second way is use `service.call`, which emulate a http request between two function. It's useful when you want to create seperated tracings for two function.

```lua
local service = require "service"

local G = function (params, ctx)
  return service.call("bar", {val = 1}) -- returns 143
end
```

#### Error handling

There are two kind of error in spacer: **error** and **Fatal**.

An **error** is corresponding to http 4xx status code: something went wrong on the caller side. To return an error, just call `ctx.error`.

```lua
local G = function (params, ctx)
  ctx.error("Invalid Password")
end
```

A **fatal** is corresponding too http 5xx status code: the function itself goes wrong (and it's not recoverable). call `ctx.fatal` to return a fatal.
```lua
local G = function (params, ctx)
  ctx.fatal("DB not available")
end
```

All uncaught exceptions are **fatal**.

#### Async

With Lua, you can write non-blocking, asynchronous function in synchronous flavor. Thanks to Lua's coroutine and OpenResty's cosocket.

```lua
local G = function (params, ctx)
  local http = require "resty.http"
  local httpc = http.new()

  -- this is actually non-blocking
  local res, err = httpc:request_uri("http://example.com/helloworld")
  -- the second request will wait for the first request to finish
  local res, err = httpc:request_uri("http://example.com/helloworld2")
end
```

## Contribute

To build spacer from source:

```
$ git clone git@github.com:poga/spacer.git
$ make
$ go install  // if you want to put it into your path
```

## License

* `lib/luaunit.lua`: BSD License, Copyright (c) 2005-2014, Philippe Fremy <phil at freehackers dot org>
* `lib/uuid.lua`: MIT, Thibault Charbonnier
* Everything else: [MIT](./LICENSE)

