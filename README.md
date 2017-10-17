# docker-microservice-dep

docker-microservice-dep is a dependency manger designed for microservices.

## Synopsis

First, create a file called `spacer.toml` with the following content:
```
[[Service]]
local = "examples/hello"
```

Now we can boot them up and setup a easy-to-use proxy with one command:
```
$ spacer
Cloning git@github.com:poga/spacer.git into services/poga/spacer ...
Building services/poga/docker-microservice-dep/examples/hello/docker-compose.yml ...
...
...
Successfully built 2f880daf95fc

Starting services/poga/docker-microservice-dep/examples/hello/docker-compose.yml ...
Proxying /hello to http://0.0.0.0:32770

Spacer is ready and rocking at 0.0.0.0:9064

$ curl 0.0.0.0:9064/hello
Hello World
```

## Development

1. git clone `git clone git@github.com:poga/docker-microservice-dep.git`
2. Setup Spacerfile
3. `gb build all`
4. `./bin/spacer`

## Todo

- [x] Download and Boot up
- [ ] Scaling service with docker-compose
- [ ] Monitering service with ?
- [ ] Conquer the world with Docker

## Why

We already have great tools to help us dealing with library dependencies,
such as Bundler, NPM, and Cargo.
However, in the world of microservices, we still have to manage our own microservice infrastructure.
Creating a scalable microservice infrastructure is a hard task and need a lot of experiences.

With docker-microservice-dep, you simply write down the service you need, the version you want.
Spacer will take care of all the work of building, monitering, and scaling for you.

For example, we need a production-ready open-source spam-filter service. we can write "poga/spam-fighter" in a file named Spacerfile. Spacer will pull the correct service from github and deploy them to development environment or production-ready IaaS.

Now we can scale this spam-filter service together and everyone can benefit from it. Thanks to Docker.

# Note

This project is an entry of Docker Global Hack Day #3
