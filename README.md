# Spacer

Spacer is a serverless function platform for a new way to build business around technology.

## Quickstart

#### Install Dependencies

* [OpenResty](https://openresty.org/)
* [librdkafka](https://github.com/edenhill/librdkafka)

#### Install Spacer

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
