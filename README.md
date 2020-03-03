# go-mailer-service

## Overview

This service is part of my example for [enterprise microservices](https://peter.bourgon.org/go-kit/) in 
[go](https://golang.org/).

The business scenario is rather contrived and not really the point here:
- [go-mailer-service](https://github.com/StephanHCB/go-mailer-service) 
  (this service) offers a REST API to send an email
  given an email address, a subject, and a body. When an email
  is sent, a Kafka message must be sent to inform some hypothetical
  downstream service, but only if a feature toggle is switched on
- [go-campaign-service](https://github.com/StephanHCB/go-campaign-service)
  offers a REST API to plan a campaign (really just a list of email addresses,
  plus a subject and a body) and execute it, 
  using the mailer service

But here's the catch, we are pretending to live in some hypothetical
enterprise that deploys its services as docker containers
into seperate kubernetes installations
(dev, staging, production). 

Also, there are all kinds of compliance rules to follow.

Effectively, that adds lots of nonfunctional requirements concerning:
- **configuration** for multiple environments, with part of the configuration
  injected by the environment 
  - kubernetes provides configmaps and secrets, but we need to implement
    loading yaml configuration and plan for multiple profiles
  - for simplicity we assume that we can accept rolling restarts for
    configuration changes, so no hot reload mechanism must be provided
  - _feature toggles_
- **API docs** (swagger/openapi v3) must be offered and kept current
- **messaging** (Apache Kafka, separate installations per environment)
- **discovery** (done using kubernetes standard features in our example)
- **resilience** 
   - kubernetes provides _scaling_ and _availability_, but for that to
     work, we must support health and readiness probes
   - all cross-service calls must use _circuit breakers_
- **persistence** (mysql, separate installations per environment)
- **security**
   - _transport layer security_
       - exclusively communicate via https between services 
         (except on local developer machines), so outgoing calls
         go to the route endpoint
       - SSL termination for incoming calls done using kubernetes
         (you don't even get to have the certificates for production!)
   - _authentication and authorization_
       - authentication ("who is making the request") is provided by an 
         external [IAM](https://en.wikipedia.org/wiki/Identity_management) solution that includes an 
         [OAuth2](https://oauth.net/2/), [OIDC-compliant](https://openid.net/connect/) Identity Provider. 
         From our perspective, 
         the request includes a signed [JWT](https://jwt.io/) in the `Authorization` header. 
         The signature must be validated for every request and the user information extracted.
         Also, the token must be passed on in outgoing calls, but only to other internal services
         participating in the auth domain.
       - part of the authorization ("are they allowed to do what is requested") 
         is also derived from claims in the JWT
- **logging** must be JSON structured logging to the console, except
  on local developer machines, where this would reduce readability.
  When running in kubernetes, we assume we want to log to console using 
  [Elastic Common Schema](https://www.elastic.co/guide/en/ecs/current),
  so that log output can be shipped directly to logstash.
- **monitoring**
    - _health and readiness end points_ must be provided
    - _prometheus_ endpoints must be available and provide appropriate
      metrics
- **tracing**
    - _OpenTracing support_
    - More specifically: In order to allow _request tracing_ across the 
      many existing Java Spring Boot/Spring Cloud services, every service 
      must propagate (and possibly create)
      a zipkin compatible Request Id, which must be placed in the
      "X-B3-TraceId" header on each response and outgoing call.
- **automated testing**
    - _unit tests_ cover all significant logic (examples only)
    - _acceptance tests_ for all endpoints
        - including negative/positive _security acceptance tests_
    - _consumer driven contract tests_ for all interactions between services
    - _performance testing_
    - _load testing_

The following requirements will not be covered here, either because our 
business scenario does not need them, or because this is after all a 
contrived example:
- we won't do _client side load balancing_ for additional **resilience**,
  mainly because for that to make sense, it would also require to move 
  parts of discovery away from kubernetes standard router features.
  
  In a real world scenario, this could easily happen though. 
  Just think of scaling out the mailer service to some cloud provider, 
  so that some of its instances run in AWS while others are run on premise.
- our use case does not require service node **coordination** beyond 
  sharing the same database, which is simply assumed to be clustered
    - e.g. _partition tolerance_ is not considered here beyond the 
      standard kubernetes features
- we assume no need for **caching** (although this service might 
  benefit drastically from e.g. memcached integration under high load)
- in a real world scenario, there are many additional **security**
  considerations not taken into account here. Some examples:
  - _security logging_
  - _dependency and vulnerability scanning_
  - ...

For each of the requirement areas written in bold face above, you will find 
a section below detailing what go libraries I have looked at, and which
one I have ultimately decided to use, and what I have learned in the
process.

_I have a fair bit of experience with Spring Boot / Spring Cloud, so I will
occasionally make notes of what I am missing coming from that perspective.
If you find I just haven't found a library that does just that or something
similar, do send me a message. I am very happy to learn._

## Developer Instructions

### Development Project Setup

This service uses go modules to provide dependency management, see `go.mod`.

If you place this repository OUTSIDE of your gopath, go build and go test will clone
all required dependencies by default.

### Running on Localhost

TODO

### Running the Automated Tests

TODO

## Selecting a Web Framework/Library

There are three possible approaches:
- use a full fledged RPC/Event Driven Architecture Framework
    - [micro/go-micro](https://github.com/micro/go-micro), which focuses on
      gRPC (google protobuf), and supports discovery via multicast DNS 
      and etcd &rightarrow; not a good fit
    - [NYTimes/gizmo](https://github.com/NYTimes/gizmo), which has
      many of the features we need. Ultimately I decided against it
      because the documentation is just too thin for my taste, and so
      I had a hard time finding my way into it, given my time
      constraints.
- use a Web Service Framework
    - [gin-gonic/gin](https://github.com/gin-gonic/gin), a very active
      MIT licensed framework with lots of good documentation and
      stackoverflow answers. This framework was recommended multiple
      times at a recent go conference I went to, and I could easily
      find example code for most of the requirements that aren't
      directly supported.
    - I found some other frameworks, but they are either outdated,
      or aren't nearly as active as gin. 
- use just a Router
    - [gorilla/mux](https://github.com/gorilla/mux), this is a low
      level choice that only has a http request router, but if you 
      find you are using fewer and fewer standard components from 
      your framework, this might actually be useful.

I chose to go with [gin](https://github.com/gin-gonic/gin). We'll see how that turns out.

## Implementation Experience with Gin

- Easily supports Singleton Controller/Service/Repository by passing down the `gin.Context`  
    - but this allows lower layers direct access to the web request/response, 
        so it takes a little bit more discipline not to abuse this, compared to e.g. gorilla/mux
        handler functions and passing around a `context.Context`. Might want to
        wrap it to give access just to what should be available?
 
...

## Fulfilling the Requirements with Gin 

### Requirement: Configuration

[spf13/viper](https://github.com/spf13/viper) provides configuration file handling
- get config from JSON, YAML, envfile, Java properties file, environment variables, command line flags, defaults
- can live watch for changes (not needed here)
- supports some [remote key/value stores](https://github.com/spf13/viper#remote-keyvalue-store-support) 

[spf13/pflag](https://github.com/spf13/pflag) provides support for parsing GNU style command line flags.

If you need a more complex command line interface, take a look 
at [spf13/cobra](https://github.com/spf13/cobra), which extends pflags.

_And here we come to the first piece that seems lacking a bit. In a Spring Boot web application I would only
have to declare the @configuration classes with the fields I want, everything else is opinionated auto-setup
with zero lines of boilerplate code._

_What I would wish for here is a library that sits on top of viper and pflag, and I just give it a set of
hierarchical structs that represent my configuration (with possible default values coded right in, descriptions
and optional overrides for naming in the various config formats as backtick metadata like `yaml:"server" env:"CONFIG_SERVER"`) 
plus a list of profiles and associated configuration files to load, make a single setup call, and have all
this set up for me._

_Not to be misunderstood, I actually love that no auto-magic is happening. I have spent too much time
hunting down bugs introduced by Spring auto discovery and its confusing precedence rules. 
With what I'm proposing, libraries could offer exported
structs for their configuration (many already do), all I would need to do is reference these structs
somewhere in my configuration, and hey, presto, you can set it all up from outside configuration,
and then I just have to code up a single function that uses these values to configure each library._ 

#### Feature Toggles

If you just have a few feature toggles, using hot reloading of your configuration can be a way
to distribute them via configuration changes. This quickly becomes unwieldy, though.

For more full-fledged solutions, see this 
[overview article for available feature toggle solutions for golang](https://featureflags.io/go-feature-flags/).

Some of these solutions are commercial and either depend on a server you install or on some app
hosted in the cloud. For an enterprise scenario, this could even be reasonable, depending on 
requirements.

I am particularly interested in free, open source solutions that also seem to be under active development.
Two candidates I looked at:
- Decider [vsco/dcdr](https://github.com/vsco/dcdr), MIT license, latest release recent. It is using
  a consul, etcd or Redis backend, and they are working on Zookeeper support. It supports audit trail
  functionality using a git repository, into which it will push your keyspace and all
  changes in JSON format. Comes with a CLI and statsd support.
- [Unleash/unleash](https://github.com/Unleash/unleash), Apache 2 license, very active, released regularly.
  Client implementations exist for Java, Node.js, Go, Python, and more. Although there is a commercial,
  hosted version, it is free if you run your own server. Ready-made docker containers and good documentation
  is linked on github.
  
If this example weren't so basic that I really don't need a feature toggle library, I would try Unleash first.
From the code examples, using it seems straightforward.

### Requirement: API Docs

[go-swagger/go-swagger](https://github.com/go-swagger/go-swagger) provides Swagger / OpenAPI v3 integration.
- can be used to generate stub code from swagger api description, but I prefer working the other way around to avoid
  hassles with code generators. So I like that this package can also [generate a swagger specification from
  annotated code](https://goswagger.io/use/spec.html).
- Note that this does not mean I do not follow api first principles, I just prototype my API in code form using types
  and interfaces in the api package, which I then implement. Much less chance of api docs and code deviating.

After a lot of fiddling, I found two articles on medium.com that were very helpful:
- [generate swagger specification from go source code](https://medium.com/@pedram.esmaeeli/generate-swagger-specification-from-go-source-code-648615f7b9d9)
- [serve swaggerui within your golang application](https://medium.com/@ribice/serve-swaggerui-within-your-golang-application-5486748a5ed4)

Here's what I ended up doing:

In order to compile the `swagger` binary, run this command while inside your project's root directory

`go install github.com/go-swagger/go-swagger/cmd/swagger`

Go now builds a binary called `swagger` and puts it in your `$GOPATH/bin`. You should now be able
to call it, if you have correctly added your `$GOPATH/bin` to your PATH.

Now you can generate swagger.json:

`swagger generate spec -o docs/swagger.json --scan-models`

Serving the swagger ui with the swagger serve command didn't work due to some CORS issues on localhost, and 
besides we want the service to be able to serve swagger-ui and the generated json file anyway (though we'll have
to remember to add security later).

So `git clone https://github.com/swagger-api/swagger-ui` somewhere and copy the LICENSE and the dist files
over into `third_party/swagger_ui`, deleting the .map files to conserve space. Then add static
serve directives for gin.

_I have to say, the documentation for this is very cryptic. Also, I don't like that I seem to be forced to
add extra data types just so I can document the response for a REST api using models. On the other hand
I remember struggling with the documentation for SpringFox, too._

_I like how the swagger spec is generated from godoc comments, which means I don't have to fire up the application
and it can be easily checked in and served statically._

_Statically serving swagger-ui, however, should really be available as a library that I can just reference. I won't
mind having to add that one route in gin, though. Again this makes everything more explicit._

### Requirement: Logging

Although there are many other choices, none of which looks bad, my most promising candidates offer a choice 
between a well aged solution
with a stable api and a very current libary that seems to focus on doing exactly what I need:

- [sirupsen/logrus](https://github.com/sirupsen/logrus), MIT license, is in maintenance mode but actively maintained.
  This might actually be an advantage, as it promises a stable api. It has very nice support for JSON structured
  logging, basic setup can literally be done in a single line, but judging from some of the benchmarks,
  its performance is not the best.
- [rs/zerolog](https://github.com/rs/zerolog), MIT license, very active. This library is focused on JSON structured
  logging, but also has a developer console mode - just what we want. Context integration is built in if you
  use context.Context, including handling of context sensitive fields such as our desired RequestId.
  
I am going to use zerolog for this example, simply because reading the code examples I liked its interface
with the chaining calls and the .Ctx() call, and it does many of the things I want to see by default.

As we are using gin.Context, we will have to implement a small middleware to place a context.Context
in the Keys map of the gin context. This also nicely solves the problem of preventing lazy coders from 
accessing the http request/response in lower layers.

Integration with gin, so it uses zerolog, can be found in [gin-contrib/logger](https://github.com/gin-contrib/logger). 

_One problem I faced was that the field names all had to be adjusted to match the ECS standard. Another thing
that could be a ready-made library, really._

 

### Requirement: Testing

In go, unit tests reside in the package directory, in files called `*_test.go`.

TODO Mocking

#### Acceptance Tests

I place acceptance tests under `test/acceptance`, and possibly use sub-packages if specific interface implementation
mocks are needed (a package cannot provide two implementations of the same interface). I have adopted the convention
of calling files `*_acc_test.go` if they contain actual acceptance tests.

Use the built-in [httptest package](https://golang.org/pkg/net/http/httptest/#Server) for starting up the service
for tests.

[hashicorp/go-memdb](https://github.com/hashicorp/go-memdb) provides an in-memory database. In order to
avoid the dependency on a real database on developer machines, I often just include it among the available 
database at run-time, and make the selection a configuration switch.

[smartystreets/goconvey](https://github.com/smartystreets/goconvey) provides a convenient library for
writing your acceptance tests BDD-style (given/when/then) including in-browser reports.

#### Consumer Driven Contract Tests

...
