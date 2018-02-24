# Spacer

A new way to build business around technology and to code. Includes:

* fast **edit-save-reload** development cycle: No redeployment or rebuilding image is needed.

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

The second way is use `flow.call`, which emulate a http request between two function. It's useful when you want to create seperated tracings for two function.

```lua
local flow = require "flow"

local G = function (params, ctx)
  return flow.call("bar", {val = 1}) -- returns 143
end
```

It's called **flow** since the primary usage of it is to trace the flow between functions.

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

## Event, Trigger, and Kappa Architecture

Serverless programming is about events. Fuctions are working together through a series of events.

Spacer use [Kappa Architecture](http://kappa-architecture.com) to provide a unified architecture for both events and data storage.

#### Event and Topic

Events are organized with topics. Topics need to be defined in the config `config/application.yml`.

To emit a event, use the built-in `topics` library.

```lua
local topics = require "topics"

local G = function (params, ctx)
  topics.append("EVENT_NAME", { [EVENT_KEY] = EVENT_PAYLOAD })
end

return G
```

#### Trigger

Triggers are just functions. You can set a function as a trigger by setting them in the config `config/application.yml`.

#### Log, Kappa Architecture, and Replay

Events in spacer are actually permanently store in the specified storage (by default we use PostgreSQL as storage).

## Contribute

To build spacer from source:

```
$ git clone git@github.com:poga/spacer.git
$ make
$ go install  // if you want to put it into your path
```

## License

* `lib/luaunit.lua`: BSD License, Philippe Fremy
* `lib/uuid.lua`: MIT, Thibault Charbonnier
* `lib/router.lua`: MIT
* Everything else: [MIT](./LICENSE)

