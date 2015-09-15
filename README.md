# Spacer

Spacer is a tool to help you manage your service dependency.

We already have great tools to help us dealing with library dependencies, such as Bundler, NPM, and Cargo. However, in the world of microservices, we still have to create our own microservice infrastructure. Creating a scalable microservice infrastructure is a hard task and need a lot of experiences.

With Spacer, you simply write down the service you need, the version you want.

For example, we need a production-ready open-source spam-filter service. we can write "poga/spam-fighter" in a file named Spacerfile.

Spacer will pull the correct service from github and deploy them to development environment or production-ready IaaS.

And, now we can scale this spam-filter service together. and everyone can benefit from it.
