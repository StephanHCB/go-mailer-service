# go-mailer-service

## References

This work was heavily inspired by

- Peter Bourgon's [Blog article about enterprise microservices](https://peter.bourgon.org/go-kit/)
- Erik Lupander's [Blog series about how to implement an enterprise microservice in go](https://callistaenterprise.se/blogg/teknik/2017/02/17/go-blog-series-part1/)
- a talk by Mario-Leander Reimer I recently attended at [GoDays Berlin](https://www.youtube.com/watch?v=x26Q7iGpblw)
- many many excellent libraries and examples, you will find links in the article below
- ... and finally, [Spring Boot](https://spring.io/projects/spring-boot) and [Spring Cloud](https://spring.io/projects/spring-cloud).

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
right next to this repository. If you wish to place your contract specs somewhere else, simply change the
path or URL in `test/contract/producer/setup_ctr_test.go`. 

```
TODO: implement a real world example 
```

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
- The fact that gin insists on using its own context rather than the standard go context makes
  it harder to wire up a lot of standard middlewares such as for authentication.
- Gin uses printf for debug mode logging instead of the provided logger. You can turn it off, thankfully.

In the end, the few lines of code saved are just not worth the additional hassle and complexities
compared to Chi.

## Implementation Experience with Chi

- It's much more low level, for example I needed to write actual code to serve static files,
  see [this example](https://github.com/StephanHCB/go-campaign-service/blob/master/web/controller/swaggerctl/swaggerctl.go)
- Smaller binary, much smaller dependencies footprint
- It relies on standard context, handler and middleware functions, fully compatible with golangs standard
  library. This makes it much easier to use third party middlewares. 

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

_**Update:** And only one day after my own version, I found [alexflint/go-arg](https://github.com/alexflint/go-arg)._

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

### Requirement: Tracing

The chi framework comes with a standard middleware that will parse a `X-Request-Id` header if present, or 
otherwise populate the context with a random string. This is set up by registering the middleware function:

```
server := chi.NewRouter()
server.Use(middleware.RequestID)
```

All you need to do is extract the string from the context and place it in the header again 
when making an outgoing request.

There are both advantages and disadvantages to including the request id on external calls.

_Writing a matching middleware function for the gin framework is easy given the chi implementation 
as a template, but I could not find a ready-made implementation._ 

_This isn't a big loss, because
the chi implementation does not cooperate correctly with 
[Spring Cloud Sleuth](https://spring.io/projects/spring-cloud-sleuth) anyway, both the header and
the exact format of the request ids are different. Again, this is a rather trivial change to make,
and I have left this out here._

### Requirement: Monitoring

#### Liveness and Readiness Probes

With our simple services, liveness and readiness probes can both be provided by the same very simple health endpoint.

Go services start very fast and if configured correctly they will almost immediately be fully available, so this
is not an uncommon approach. In a real-world scenario it would probably be a good idea to check the database
connection pool for an abundance of error states before reporting healthy.

#### Prometheus Integration 

```
TODO
```

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

```
TODO
```

### Requirement: Resilience

[Go Microservices blog series, part 11 - hystrix and resilience](https://callistaenterprise.se/blogg/teknik/2017/09/11/go-blog-series-part11/)

[Don't use go's default http client](https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779)

[A rant that nevertheless has useful tips](https://fasterthanli.me/blog/2020/i-want-off-mr-golangs-wild-ride/)
(but some of the http client points of criticism would have been completely solved by using hystrix as shown
in the first article)

```
TODO
```

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

_One neat thing we came across while researching this subject was [dex](https://github.com/dexidp/dex),
a full fledged OIDC / Oauth2 provider written in go with pluggable connectors written and supported by the coreos team,
which even includes an OpenShift connector and an LDAP connector for easy federation._

##### Security Acceptance and Contract Tests

It is good practice to cover both incoming and outgoing requests with security related tests.

We have implemented a few examples of how to cover incoming requests with acceptance tests that make sure

* no access is granted if no Authorization header is provided
* invalid header values, invalid/expired tokens, or tokens that do not have the necessary claims do not grant access
* valid tokens grant access

In a real world scenario the exposed amount of data might depend on properties
of the token, this should also be tested.

```
TODO implement security acceptance test example
```

Also, we have implemented a single contract test that ensures the outgoing request does not contain
 the token for an external call.

```
TODO add security contract test example
```

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

#### Consumer Driven Contract Tests

This microservice uses [pact-go](https://github.com/pact-foundation/pact-go#installation) for contract tests.

Detailed documentation including some conceptual remarks can be found in the 
[readme for go-campaign-service](https://github.com/StephanHCB/go-campaign-service/blob/master/README.md)

...
