![](./github_assets/logo.png)

# Spacer

![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/poga/spacer)](https://goreportcard.com/report/github.com/poga/spacer)

Serverless function platform for Lua

## Features

* High-performance **non-blocking** functions.
* Fast **edit-save-reload** development cycle: No redeployment or rebuilding image is needed.
* **Platform agnostic**: From a simple [PostgreSQL](https://www.postgresql.org/) database to a complex [Kubernetes](https://kubernetes.io/) clusters with [Apache Kafka](https://kafka.apache.org/). Spacer can run on most platforms.

## Synopsis

Here's a minimal spacer project.

You write a function:

```lua
-- app/hello.lua
local G = function (args)
    return "Hello from Spacer!"
end

return G
```

...and define how to expose it to the internet:

```lua
-- app/gateway.lua
local R = {
    -- HTTP Method, Path, Function Name
    {"GET", "/hello", "hello"}
}

return R
```

...then start the project.

```
$ ./bin/dev.sh
```

Test the function.

```
$ curl localhost:3000/hello
{"data":"Hello from Spacer!"}
```

That's it!

## Install

The easiest way would be using the provided [Dockerfile](./Dockerfile).

To install from source:

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

Create a spacer project:

```
$ spacer init ~/spacer-hello
```

Start the development server:

```
$ cd ~/spacer-hello
$ ./bin/dev.sh
```

Open `http://localhost:3000/hello` and you should see spacer working.

### Hello World

Functions in spacer are written in Lua, a simple dynamic langauge. Here's a hello world function:

```lua
-- app/hello.lua
local G = function (args)
    return "Hello from Spacer!"
end

return G
```

Every function takes one argument: `args`. For detail, check the **Functions** section.

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

Functions are the main abstraction in spacer. Functions are composed together to build a complex application.

There are 2 way to invoke other functions from a function. The first is the simplest: just call it like a normal lua function.

```lua
-- app/bar.lua
local G = function (args)
  return params.val + 42
end

-- app/foo.lua
local bar = require "bar"

local G = function (args)
  return 100 + bar({val = 1}) -- returns 143
end
```

The second way is use `flow.call`, which emulate a http request between two function. It's useful when you want to create seperated tracings for two function.

```lua
local flow = require "flow"

local G = function (args)
  return flow.call("bar", {val = 1}) -- returns 143
end
```

It's called **flow** since the primary usage of it is to trace the flow between functions.

### Exposing Functions to the Public

Functions are exposed to the public through a gateway. You can define gateway in `gateway.lua`.

```lua
local _R = {
    -- HTTP Method, Path, Function Name
    {"GET", "/hello", "hello"},

    -- users resources
    {"GET", "/users", "controllers/users/get"},
    {"PUT", "/users/:id", "controllers/users/put"},

    -- teams resources
    {"POST", "/teams", "create_teams"},
    {"GET", "/teams/:id", "get_team"},
    {"PUT", "/teams/:id", "put_team"},
    {"DELETE", "/teams/:id", "delete_team"},

    -- registration
    {"POST", "/register", "controllers/users/create"}
}

return _R
```

When invoking a function via HTTP request, query params, route params, and post body(json) are all grouped into an **args** table and passed to the function.

#### Error handling

An **error** is corresponding to HTTP 4xx status code: something went wrong on the caller side. In this case, return the error as the second returned value

```lua
local G = function (args)
  return nil, "invalid password"
end
```

If an unexpected exception happepend, use the `error()` function to return it. Spacer will return the error with HTTP 500 status code.

```lua
local G = function (args)
  local conn = db.connect()
  if conn == nil then
    error("unable to connect to DB")
  end
end
```

## Event, Trigger, and Kappa Architecture

Serverless programming is about events. Fuctions are working together through a series of events.

Spacer use [Kappa Architecture](http://kappa-architecture.com) to provide a unified architecture for both events and data storage.

#### Event and Topic

Events are organized with topics. Topics need to be defined in the config `config/application.yml`.

To emit a event, use the built-in `topics` library.

```lua
local topics = require "topics"

local G = function (args)
  topics.append("EVENT_NAME", { [EVENT_KEY] = EVENT_PAYLOAD })
end

return G
```

#### Trigger

Triggers are just functions. You can set a function as a trigger by setting them in the config `config/application.yml`.

#### Log, Kappa Architecture, and Replay

Events in spacer are permanently stored in the specified storage (by default we use PostgreSQL as storage).

## Libraries Search Path

You can require lua module under `app/` and `lib/` directly.

```lua
-- requiring app/foo.lua
local m = require "foo"
```

Directories can be used to organize your modules.

```lua
-- requiring app/models/foo.lua
local m = require "models/foo"
```

## Contribute

To build spacer from source:

```
$ git clone git@github.com:poga/spacer.git
$ make
$ go install  // if you want to put it into your path
```

## License

* `lib/luaunit.lua`: BSD. [https://github.com/bluebird75/luaunit](https://github.com/bluebird75/luaunit)
* `lib/uuid.lua`: MIT. [https://github.com/thibaultcha/lua-resty-jit-uuid](https://github.com/thibaultcha/lua-resty-jit-uuid)
* `lib/router.lua`: MIT. [https://github.com/APItools/router.lua](https://github.com/APItools/router.lua)
* `lib/resty/hmac.lua`: BSD. [https://github.com/jkeys089/lua-resty-hmac](https://github.com/jkeys089/lua-resty-hmac)
* `lib/resty/eve.lua, lib/resty/jwt-validators.lua, lib/resty/jwt.lua`: Apache License. [https://github.com/SkyLothar/lua-resty-jwt](https://github.com/SkyLothar/lua-resty-jwt)
* Everything else: [MIT](./LICENSE)

