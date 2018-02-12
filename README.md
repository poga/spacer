# Spacer

Spacer is a serverless function platform for a new way to build businesses around technology.

Spacer is designed to be **minimal** and **simple**. Its architecture is simple enough to run on any typical PaaS such as [Heroku](https://www.heroku.com/) or on a kubernetes cluster.

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
$ ./bin/start.sh
```

Open `http://localhost:3000/hello` and you should see spacer working.

### Function

Functions in spacer are written in Lua, a simple dynamic langauge. Here's a hello world function:

```lua
local G = function (event, ctx)
    return "Hello from Spacer!"
end

return G
```

Every function takes two arguments: `event`, and `ctx`.

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

Check `/test/test_hello.lua` for example.

## Contribute

Build spacer from source:

```
$ git clone git@github.com:poga/spacer.git
$ make && go install
```

## License

The MIT License

