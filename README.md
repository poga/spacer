# Spacer

Spacer is a serverless function platform for a new way to build businesses around technology.

Spacer is designed to be **minimal** and **simple**. You don't need a Kubernetes cluster to start using Spacer. The architecture is simple enough to run on any typical PaaS such as [Heroku](https://www.heroku.com/).

Spacer provides a fast **edit-save-reload** development cycle. No redeployment or rebuilding image is needed.

Spacer focus on the core benefit of serverless technology: **aggressive code reuse**, **pay per function resource usage**, and **NoOps**.

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

## Develop

```
$ git clone git@github.com:poga/spacer.git
$ make && go install
```

## License

The MIT License

