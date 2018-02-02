# Spacer

Spacer is a serverless function platform. It will fundamentally change how you build business around technology and how you code.

## Quickstart

Spacer depends on [OpenResty](https://openresty.org/). Follow the [offical guide](https://openresty.org/en/installation.html) to install.

```
$ go get github.com/poga/spacer
$ spacer init ~/spacer-hello
$ cd ~/spacer-hello
$ ./bin/start.sh
```

Then go to `http://localhost:3000/hello` to see spacer working.

## Develop

```
$ git clone git@github.com:poga/spacer.git
$ make
```

## License

The MIT License
