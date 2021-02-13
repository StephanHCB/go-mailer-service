# go-mailer-service

## References

This work was heavily inspired by

- Peter Bourgon's [Blog article about enterprise microservices](https://peter.bourgon.org/go-kit/)
- Erik Lupander's [Blog series about how to implement an enterprise microservice in go](https://callistaenterprise.se/blogg/teknik/2017/02/17/go-blog-series-part1/)
- a talk by Mario-Leander Reimer I recently attended at [GoDays Berlin](https://www.youtube.com/watch?v=x26Q7iGpblw)
- many many excellent libraries and examples, you will find links in the article below
- ... and finally, [Spring Boot](https://spring.io/projects/spring-boot) and 
  [Spring Cloud](https://spring.io/projects/spring-cloud).
  
I would also like to thank several individual contributors who have contacted me with
suggestions and improvements.

## Overview

This service is part of my example for **enterprise microservices** in 
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

But here's the catch, I am pretending to work in some hypothetical
enterprise that deploys its services as docker containers
into seperate kubernetes installations
(dev, staging, production). 

Also, there are all kinds of compliance rules to follow.

Effectively, that adds lots of nonfunctional requirements concerning:
- **configuration** for multiple environments, with part of the configuration
  injected by the environment 
  - kubernetes provides configmaps and secrets, but we need to implement
    loading configuration either from yaml files or from
    environment variables
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
         go to an external load balancer and then through an ingress
       - SSL termination for incoming calls is done using said load balancer
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
         is also derived from claims in the JWT. I am going to use a namespaced roles claim
         with the url pointing to this github repository.
- **logging** must be JSON structured logging to the console, except
  on local developer machines, where this would reduce readability.
  When running in kubernetes, I assume we want to log to console using 
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
      a [zipkin](https://zipkin.io/)/[sleuth](https://spring.io/projects/spring-cloud-sleuth) 
      compatible Request Id, which must therefore be picked up from and placed in the
      "X-B3-TraceId" header on each response and outgoing call.
- **automated testing**
    - _unit tests_ cover all significant logic (examples only)
    - _acceptance tests_ for all endpoints
        - including negative/positive _security acceptance tests_
    - _consumer driven contract tests_ for all interactions between services
        - including _security contract tests_ that ensure tokens are not propagated
          where they should not go
    - _performance testing_
    - _load testing_

The following requirements will not be covered here, either because our 
business scenario does not need them, or because this is after all a 
contrived example:
- we won't do _client side load balancing_ for additional **resilience**,
  mainly because for that to make sense, it would also require to move 
  parts of discovery away from kubernetes standard service and ingress
  features.

  In a real world scenario, this could easily happen though. 
  Just think of scaling out the mailer service to some cloud provider, 
  so that some of its instances run there while others are run on premise.
- my use case does not require service node **coordination** beyond 
  sharing the same database, which is simply assumed to be clustered
    - e.g. _partition tolerance_ is not considered here beyond the 
      standard kubernetes features
- we assume no need for **caching** (although this service might 
  benefit drastically from e.g. memcached integration under high load)
- in a real world scenario, there are many additional **security**
  considerations not taken into account here. Some examples:
  - _security logging_
  - _dependency and vulnerability scanning_
  - _access revocation_ in a shorter time span than the expiry time of the token
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

If you change or add a dependency, you will need to do `go build main.go` to clone or update dependencies.

### Running on Localhost

On the command line, `go build main.go` will download all dependencies and build a standalone executable
for you.

The executable expects two configuration files `config.yaml` and `secrets.yaml` in the current directory.
You can override their path locations by passing command line options as follows:

```main --config-path=. --secrets-path=.``` 

Find configuration templates under docs, copy them to the main directory and edit them so they fit your
environment.

### Running the Automated Tests

This service comes with unit, acceptance, and consumer driven contract tests. 

You can run all of these on the command line:

```go test ./...```

In order for the **contract tests** to work, you will need to perform some additional installation:

#### Consumer Driven Contract Tests

This microservice uses [pact-go](https://github.com/pact-foundation/pact-go#installation) for contract tests.

This is the **producer** side.

Before you can run the contract tests in this repository, you need to run the consumer side contract tests
in the [go-campaign-service](https://github.com/StephanHCB/go-campaign-service) to generate
the contract specification. 

You are expected to clone that repository into a directory called `go-campaign-service`
right next to a clone of this repository. If you wish to place your contract specs somewhere else, simply change the
path or URL in `test/contract/producer/setup_ctr_test.go`. 

See the [readme for go-campaign-service](https://github.com/StephanHCB/go-campaign-service/blob/master/README.md) for 
installation instructions for the required tooling.

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
    - [go-chi/chi](https://github.com/go-chi/chi), another very active
      MIT licensed framework that already includes middlewares for
      many of our nonfunctional requirements.
- use just a Router
    - [gorilla/mux](https://github.com/gorilla/mux), this is a low
      level choice that only has a http request router, but if you 
      find you are using fewer and fewer standard components from 
      your framework, this might actually be useful.

I chose to go with [gin](https://github.com/gin-gonic/gin). We'll see how that turns out.

After finding [chi](https://github.com/go-chi/chi), I decided to try that framework
for my [go-campaign-service](https://github.com/StephanHCB/go-campaign-service), the other
side of this example.

## Implementation Experience with Gin

- Easily supports Singleton Controller/Service/Repository by passing down the `gin.Context`  
    - but this allows lower layers direct access to the web request/response, 
        so it takes a little bit more discipline not to abuse this, compared to e.g. gorilla/mux
        handler functions and passing around a `context.Context`. Might want to
        wrap it to give access just to what should be available?
- Has a good example for [including static files in the binary](https://github.com/gin-gonic/examples/tree/master/assets-in-binary) 
- The pre-made request logging library does not retrieve its logger from the context but instead uses
  the global logger instance. This is nice and good, until you want to have your requestIds logged, too.
  Then you are stuck coding up your own request logger middleware again.
- The pre-made gin requestId middleware is completely useless - it doesn't read an existing
  requestId from the relevant header, if present, but instead always generates a new requestId
- The fact that gin insists on using its own context rather than the standard go context makes
  it harder to wire up a lot of standard middlewares such as for authentication.
- Gin uses printf for debug mode logging instead of the provided logger. You can turn it off, thankfully.

In the end, the few lines of code saved are just not worth the additional hassle and complexities
compared to Chi.

## Implementation Experience with Chi

- It's much more low level, for example I needed to write actual code to serve static files,
  see [this example](https://github.com/StephanHCB/go-campaign-service/blob/master/web/controller/swaggerctl/swaggerctl.go)
- Smaller binary, much smaller dependencies footprint
- It relies on standard context, handler and middleware functions, fully compatible with golang's standard
  library. This makes it much easier to use third party middlewares.
- Chi provides a number of [pre-made middlewares](https://github.com/go-chi). They have recently updated their
  documentation and the list is slowly growing. Besides, go-chi handlers are compatible with net/http, so
  a lot of third party middlewares will also work out of the box.

Chi will be the framework of choice for me for future microservices.

## Fulfilling the Requirements

### Requirement: Configuration

[spf13/viper](https://github.com/spf13/viper) provides configuration file handling
- get config from JSON, YAML, envfile, Java properties file, environment variables, command line flags, defaults
- can live watch for changes (not needed here)
- supports some [remote key/value stores](https://github.com/spf13/viper#remote-keyvalue-store-support) 

[spf13/pflag](https://github.com/spf13/pflag) provides support for parsing GNU style command line flags.

If you need a more complex command line interface, take a look 
at [spf13/cobra](https://github.com/spf13/cobra), which extends pflags.

_And here we come to the first piece that seemed lacking a bit. In a Spring Boot web application I would only
have to declare the @configuration classes with the fields I want, everything else is opinionated auto-setup
with zero lines of boilerplate code._

_What I would have wished for here is a library that sits on top of viper and pflag, and I just give it a set of
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

_**Update:** I have written such a library and called 
it [go-autumn-config](https://github.com/StephanHCB/go-autumn-config)._

_This library now even supports structured configuration values. See the README for an example how to use
[mitchellh/mapstructure](https://github.com/mitchellh/mapstructure) to get the structured data out._

_**Update:** I also 
found [alexflint/go-arg](https://github.com/alexflint/go-arg), which provides a very lightweight solution
if all you need are command line arguments and environment variables and need no support for yaml files._

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

After a lot of fiddling and not getting it to work right, I found some articles that were very helpful:
- [generate swagger specification from go source code](https://medium.com/@pedram.esmaeeli/generate-swagger-specification-from-go-source-code-648615f7b9d9)
- [serve swaggerui within your golang application](https://medium.com/@ribice/serve-swaggerui-within-your-golang-application-5486748a5ed4)
- [Create Golang API documentation with SwaggerUI](https://www.ribice.ba/swagger-golang/)

Here's what I ended up doing:

In order to compile the `swagger` binary, run this command while inside your project's root directory

`go install github.com/go-swagger/go-swagger/cmd/swagger`

Go now builds a binary called `swagger` and puts it in your `$GOPATH/bin`. You should now be able
to call it, if you have correctly added your `$GOPATH/bin` to your PATH.

Now you can generate swagger.json:

`swagger generate spec -o docs/swagger.json --scan-models`

_I have to say, the documentation for this is very cryptic. Also, I don't like that I seem to be forced to
add extra data types just so I can document the response for a REST api using models. On the other hand
I remember struggling with the documentation for SpringFox, too._

_I like how the swagger spec is generated from godoc comments, which means I don't have to fire up the application
and it can be easily checked in and served statically._

#### Statically serving swagger-ui

Serving the swagger ui with the swagger serve command didn't work due to some CORS issues on localhost, and 
besides we want the service to be able to serve swagger-ui and the generated json file anyway (though we'll have
to remember to add security later).

So `git clone https://github.com/swagger-api/swagger-ui` somewhere and copy the LICENSE and the dist files
over into `third_party/swagger_ui`, deleting the .map files to conserve space. Then add static
serve directives for gin.

_Statically serving swagger-ui, however, should really be available as a library that I can just reference. I won't
mind having to add that one route in gin, though. Again this makes everything more explicit._

Also see [how to embed static files in go app](https://tech.townsourced.com/post/embedding-static-files-in-go/).
I agree with the author's choice: [shurcooL/vfsgen](https://github.com/shurcooL/vfsgen) is the way to go.

_**Update:** I have written a small library that does it out of the box: 
[go-autumn-web-swagger-ui](https://github.com/StephanHCB/go-autumn-web-swagger-ui)._

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

Integration with chi was a little less straightforward, but I was able to figure out their
logging middleware from looking at the code and 
[implement a replacement logger](https://github.com/StephanHCB/go-campaign-service/blob/master/web/middleware/logfilter/logfilter.go).

_One problem I faced in both cases was that the field names all had to be adjusted to match the ECS standard. Another thing
that could be a ready-made library, really._

_**Update:** I have started a library that configures ECS logging out of the box for zerolog, right now not all fields are configured:
[go-autumn-logging-zerolog](https://github.com/StephanHCB/go-autumn-logging-zerolog)._

### Requirement: Tracing

The chi framework comes with a standard middleware that will parse a `X-Request-Id` header if present, or 
otherwise populate the context with a random string. This is set up by registering the middleware function:

```
server := chi.NewRouter()
server.Use(middleware.RequestID)
```

You can also change the header it will use:

```
middleware.RequestIDHeader = "X-B3-TraceId"
```

Now all you need to do is extract the string from the context and place it in the header again 
when making an outgoing request.

There are both advantages and disadvantages to including the request id on external calls.

_The chi implementation cooperates with 
[Spring Cloud Sleuth](https://spring.io/projects/spring-cloud-sleuth) if you change the name
of the header as shown above, which is conveniently exposed for just this purpose._

_I have implemented a few lines of 
[middleware to always add the request id header to the response](https://github.com/StephanHCB/go-campaign-service/blob/master/web/middleware/requestidinresponse/addresponseheader.go)._

_I am not particularly happy with the way the request id is constructed in this middleware. It constructs it
from the hostname followed by a counter. While fine for logging, if a bit lengthy, this becomes a problem when you wish to expose
the request id to end users or api clients in error messages and responses.
- It exposes your internal host names. If your service runs in a kubernetes pod in production, then this isn't much
of an issue, though.
- It is prone to errors when receiving support requests. End users may misread the number and you have no way
to validate it.
For these reasons, I prefer the request id to be made up of a random hex string of, say, 8 or 12 characters (depending on throughput).
The code is easily adapted, of course._

Another available tracing middleware for chi is [go-chi/httptracer](https://github.com/go-chi/httptracer),
which integrates OpenTracing via [opentracing/opentracing-go](https://github.com/opentracing/opentracing-go).

_This implementation is not compatible with the default configuration of Spring Cloud Sleuth,
so I have not included it, but if Sleuth/Zipkin compatibility is not an issue, you get much more
detailed tracing, including spans for concurrent invocations._ 

_Although gin pretends to support request ids, I could not find a ready-made implementation that really did what's needed.
I was pretty disappointed when I found out that gin's requestId middleware always creates a new
request id for each request instead of taking it from the request header if one is already present.
No distributed tracing for you I guess, just request logging. Instead you are treated to a statistical discussion of the 
probability of collisions in the comments..._

### Requirement: Monitoring

#### Liveness and Readiness Probes

With our simple services, liveness and readiness probes can both be provided by the same very simple health endpoint.

Go services start very fast and if configured correctly they will almost immediately be fully available, so this
is not an uncommon approach. In a real-world scenario it would probably be a good idea to check the database
connection pool for an abundance of error states before reporting healthy.

#### Prometheus Integration 

There are two general ways to interact with [Prometheus](https://prometheus.io/), 
either you let Prometheus pull metrics from you (providing a
metrics endpoint, usually on a separate port that is not exposed), or you regularly push your metrics to Prometheus.

##### Pull Implementation

Following [this guide](https://prometheus.io/docs/guides/go-application/), we have implemented the standard
`/metrics` endpoint on a separate port in 
[go-campaign-service/web/metricsserver.go](https://github.com/StephanHCB/go-campaign-service/blob/master/web/metricsserver.go)
using [github.com/prometheus/client_golang](https://github.com/prometheus/client_golang)'s `promhttp.Handler()` 
ready-made handler function.

_Note how we use a goroutine to start the http server, as http.ListenAndServe() never returns. We have 
not secured this endpoint in this example, but adding that would be trivial, just add the relevant middleware._

The guide also describes how to add custom counters, as it is this only exposes the default metrics. Also take a
look at [766b/chi-prometheus](https://github.com/766b/chi-prometheus) for a basic middleware for the chi framework 
that gives you request metrics. 

_A single histogram for response times is probably too coarse grained, so in a real-world situation I would
probably adapt this code for our needs._ 

##### Push Implementation  

If you wish to go with the push model, there is a ready-made libary called 
[armon/go-metrics](https://github.com/armon/go-metrics).

_See `internal/repository/metricspush/setup.go` for the very simple setup, and `internal/service/emailsrv/emailsrv.go`
for a one-liner that times execution._

### Requirement: Persistence

For relational databases, [jinzhu/gorm](https://github.com/jinzhu/gorm) is the go-to object relational mapper. 
It includes direct support for mysql.
[Oracle support](https://github.com/misterEggroll/gorm) is only available in a fork, currently still being worked on,
and has some limitations.

Further options for nosql or high throughput situations exist, such as [upper.io/db.v3](https://upper.io/db.v3),
but for our purposes, gorm is just fine.

See [go-campaign-service](https://github.com/StephanHCB/go-campaign-service) for further details and an example, as this
service does not need any persistence.

Also read [production ready connection pooling in go](https://making.pusher.com/production-ready-connection-pooling-in-go/).

We have implemented a very basic in-memory persistence option, not at all optimized for performance, which can
be used to demo the service and run some of the lower level tests. For any more complex situations, a
real in-memory-database such as [hashicorp/go-memdb](https://github.com/hashicorp/go-memdb) should be considered.

_On mysql, life is good. Everything is just as easy as it is with Spring Data,
only with better performance and less memory footprint. On the other hand, Oracle support would need some
contributor work._

### Requirement: Messaging

A good introduction to Kafka in golang is 
[Getting Started with Kafka in Golang](https://medium.com/@yusufs/getting-started-with-kafka-in-golang-14ccab5fa26).

[Microservices with kafka and google protobuf](https://medium.com/@self.maurya/building-a-microservice-with-with-golang-kafka-and-dynamodb-part-i-552cc4816ff)
has some brief code snippets, 
[part 2](https://medium.com/@self.maurya/building-a-microservice-with-golang-kafka-and-dynamodb-part-ii-4c2def48a5dc) contains some tips for performance optimization

Although I have not had a chance to try kafka with golang, 
[segmentio/kafka-go(https://github.com/segmentio/kafka-go) looks like the most promising candidate.
The documentation lists the advantages over other approaches:
  - pure golang
  - compat with 0.10.1-2.1.0+
  -	active development

```
TODO implement Kafka integration example
```

Other promising libraries I found for Kafka integration:
  - [Shopify/sarama](https://github.com/Shopify/sarama) MIT licensed, also under active development
    - [Documentation](https://godoc.org/github.com/Shopify/sarama)
    - pure golang
    - https://github.com/bsm/sarama-cluster <- cluster extension, DEPRECATED
  - [confluentinc/confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go)

```
TODO evaluate these
```

### Requirement: Resilience

#### Resilience for Outgoing Requests - Circuit Breaker and Timeouts

[Go Microservices blog series, part 11 - hystrix and resilience](https://callistaenterprise.se/blogg/teknik/2017/09/11/go-blog-series-part11/)

[Don't use go's default http client](https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779) - the
main point being, it has no timeout.

[A rant that nevertheless has useful tips](https://fasterthanli.me/blog/2020/i-want-off-mr-golangs-wild-ride/) -
but most of the points of criticism regarding the http client would have been completely solved by using hystrix.

_I have implemented a small utility package that demonstrates how to use 
[afex/hystrix-go](https://github.com/afex/hystrix-go) and the standard http client safely, providing both
a circuit breaker and a timeout. I have opted not to implement a retry mechanism for this example.
See [the downstreamcall package in go-campaign-service](https://github.com/StephanHCB/go-campaign-service/tree/master/internal/repository/util/downstreamcall).
This also adds the request id to any outgoing requests._

#### Resilience for Incoming Requests - Timeouts

Among other things, [So you want to expose Go on the Internet](https://blog.cloudflare.com/exposing-go-on-the-internet/)
mentions the need to configure request timeouts to allow recovery from overload or even a DOS attack.

```
srv := &http.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
    // TLSConfig:    tlsConfig,
    Addr:         address,
    Handler:      serveMux,
}
srv.ListenAndServe(...)
```

_Unfortunately, with the gin framework this is made needlessly hard, because the ListenAndServe
call is hidden inside gin's `Run()` method, which you will have to duplicate._

_Other than the http server timeouts, chi also comes with a 
[middleware that sets up a timeout](https://pkg.go.dev/github.com/go-chi/chi/middleware#Timeout) on
the context, calling `ctx.Done()` and returning http status 504 to the caller. 
This will not abort processing the request unless you also regularly check the timeout channel set up
on the context provided with the http request in long running operations._

```
func ErrIfTimeout(ctx context.Context) error {
    select {
        case <-ctx.Done():
            return errors.New("operation timed out") 
        default:
            return nil
    }
}
```

_This will only work if you take care that there is a timeout on every blocking 
operation (within reason). For http requests, see the section about circuit breakers. Also remember
to configure timeouts for your database operations (can usually be done at the connection level)
and for things such as messaging._

#### Resilience for Incoming Requests - Throttling and DoS protection

The chi framework offers 
[middleware that can limit the amount of concurrently processed requests](https://pkg.go.dev/github.com/go-chi/chi/middleware#Throttle).
This can help mitigate some overload situations and allow you to open upstream circuit breakers early, 
but for any serious DoS attack, this will not really 
help you. DoS protection needs to be put in place before the requests even reach your service instance.

For this reason, I have opted to omit throttling from this example.

### Requirement: Security

#### Authentication and Authorization

There are ready-made libraries for handling and validating JWTs in golang, but the most complete
and well-documented solution seem to be the MIT licensed open source client libraries provided 
by [auth0](https://auth0.com). These
also help you avoid a number of security pitfalls, like the infamous 
[symmetric cipher key attack](https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/) 
that many implementations used to be vulnerable to.

We used [auth0/go-jwt-middleware](https://github.com/auth0/go-jwt-middleware). With both the chi and gin frameworks, 
you need to write a small amount of code to wrap their functions.
The token is placed in the context in both raw and decoded form, so it becomes easy for handler functions
to check claims to assert the user is logged in, or has a specific role, or for outgoing calls to include
the `Authorization` header.

Note that you will have to take care yourself not to include it on external calls, lest you expose
a valid token to a third party.

```
TODO implement a token passthrough whitelist

in httpcall.go in campaign-service only forward authentication if url is in an address whitelist
```

##### Identity Providers in Golang

_One neat thing we came across while researching this subject was [dex](https://github.com/dexidp/dex),
a full fledged OIDC / Oauth2 provider written in go with pluggable connectors written and supported by the coreos team,
which even includes an OpenShift connector and an LDAP connector for easy federation._

_Another similar project is [hydra](https://www.ory.sh/hydra/docs/),
an identity provider written in go._

##### Security Acceptance and Contract Tests

It is good practice to cover both incoming and outgoing requests with security related tests.

We have implemented a few examples of how to cover incoming requests with acceptance tests that make sure

* no access is granted if no Authorization header is provided
* invalid header values, invalid/expired tokens, or tokens that do not have the necessary claims do not grant access
* valid tokens grant access

In a real world scenario the exposed amount of data might depend on properties
of the token, this should also be tested.

See the [acceptance tests for the campaign controller of go-campaign-service](https://github.com/StephanHCB/go-campaign-service/blob/master/test/acceptance/campaign_acc_test.go)
for an example.

In order to assure that outgoing requests to third party systems do not forward the oauth token, and in order to
test our assumptions about the interface offered by the third party system, we should write contract tests
that only have a consumer side. 

_As we do not make requests to third parties in our contrived example, we have omitted this, but it works
exactly as described in the 
[readme for go-campaign-service](https://github.com/StephanHCB/go-campaign-service/blob/master/README.md),
only that instead of testing for the presence of the authorization header, you test for its absence._

#### TLS Termination

In our scenario, TLS termination is provided by a load balancer or by HAProxy.

If you wish to directly expose golang, [So you want to expose Go on the Internet](https://blog.cloudflare.com/exposing-go-on-the-internet/)
is a must-read. 

### Requirement: Testing

This service comes with unit, acceptance, and consumer driven contract tests. 

You can run all of these on the command line:

```go test ./...```

In order for the **contract tests** to work, you will need to perform some additional installation.

#### Unit Tests

In go, unit tests reside in the package directory, in files called `*_test.go`. 

Go executes tests 
in parallel goroutines, one per package by default. If you keep your packages reasonably small, this
leads to fast test execution while allowing you to set up mock implementation on a per-package basis.

Note how I have wrapped calls to `log.Fatal()` or `os.Exit()` in function pointers kept in public vars. This allows me to
simply swap the function pointers for test runs and obtain full code coverage except for the actual failFunction. 
This is a common pattern used by e.g. pflag and viper, as well as many logging packages I've looked at.

#### Mocking

There are libraries for mocking that use code generators triggered by build stage comments, but I usually
just write my own mocks in testing code. If all components are specified as interfaces which are then
implemented in a package, writing mocks can be assisted by the IDE so much that I don't bother with
specialized mocking packages.

[GoMock](https://github.com/golang/mock) seems the be most common mock code generator if you find you are
spending too much time maintaining your mocks. It's a code generator that can automatically implement interfaces.

[h2non/gock](https://github.com/h2non/gock) is a library for mocking outgoing http connections. In my example
I have so few of those, and they are all calls between services, that I have stuck with pact based consumer
driven contract tests. If you have complex interactions that are hard to represent in the static pact scripts,
this might be a good solution.

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

See the [acceptance tests for the campaign controller of go-campaign-service](https://github.com/StephanHCB/go-campaign-service/blob/master/test/acceptance/campaign_acc_test.go)
for an example, including some example security tests that make sure unauthorized/unauthenticated access
does not work.

_There are, however, some downsides to using goconvey. For one thing, go does not have test dependencies,
so any dependency you pull in for testing ends up increasing the size of your binary. If you do not really
need the reports, or if the service you are writing is especially security critical and you wish to minimize
code footprint, it may be better to just
implement a few small logging functions as I have done in this service in the `docs` package, see `docs/testdocs.go`._

_See the [acceptance tests for the service startup of this service](https://github.com/StephanHCB/go-mailer-service/blob/master/test/acceptance/startup_acc_test.go)
for an example of this hand-coded alternative approach._

#### Consumer Driven Contract Tests

This microservice uses [pact-go](https://github.com/pact-foundation/pact-go#installation) for contract tests.

This is the producer side, see `test/contract/producer/sendmail_ctr_test.go` for the implementation.

Detailed documentation including setup instructions and some conceptual remarks can be found in the 
[readme for go-campaign-service](https://github.com/StephanHCB/go-campaign-service/blob/master/README.md)
