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
         The signature must be validated for every request and the user information extracted
       - part of the authorization ("are they allowed to do what is requested") 
         is also derived from claims in the JWT
- **logging** must be JSON structured logging to the console, except
  on local developer machines, where this would reduce readability.
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

## Fulfilling the Requirements with Gin 

### Requirement: Configuration

...
