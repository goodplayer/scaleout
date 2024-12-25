# scaleout

Scaleout is a clustered fundamental service suite, supporting various essential capabilities for applications.

This project is focusing on providing scale-out services.

Any management and supporting services are **NOT** included and can be used from other projects that have
compatibilities to `nekoq-component` project.

## Major Features

**Key services**

* [ ] Configure
    * [ ] T-1 Data change notifier
        * [ ] Http Endpoint - :4001
        * [ ] Https Endpoint - :4002
    * [ ] T-1 Data change writer
        * [ ] Http Endpoint - :4003
        * [ ] Https Endpoint - :4004
* [ ] Discovery
    * [ ] T-1 Data storage
    * [ ] T-1 Client keepalive
* [ ] Messaging
    * [ ] T-1 Message broker
* [ ] Secret
    * [ ] T-1 Authority
        * [ ] Https Endpoint
    * [ ] T-1 Validator
        * [ ] Https Endpoint
* [ ] Scheduler
    * [ ] T-1 Controller
    * [ ] T-2

Note1: T-* stands for tier-*, which represents the tier of the purpose while deploying.
Note2: T-1 is the original cluster that handles data source. T-2 is the cluster that interacts with T-1 clusters.

**Fundamental Scale-out methods**

* [ ] Scale out cluster
* [ ] Nested scaling
* [ ] Purpose based scaling
* [ ] Disable specific capabilities

**Supporting surrounding services**

* [ ] Consensus
* [ ] Object/File store
* [ ] Logging/Tracing/Metrics
* [ ] Distributed Transaction
* [ ] Data store
* [ ] Searching
* [ ] Caching
* [ ] Gateway

## Installation

1. Initialize database
2. Initialize root keys(e.g. using local unseal provider)
    1. `go run github.com/meidoworks/nekoq-component/configure/secretcmd/initlocal/.`
        1. Default unseal key naming(local unseal only): `bootstrap.key`
        2. Default unseal key id(local unseal only): `1`
3. Initialize certs for TLS
4. Start service
