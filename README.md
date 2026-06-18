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

1. Following the GettingStarted.md in nekoq-component, located in
    * configure/secret/docs/GettingStarted.md
2. Initialize root keys(e.g. using local unseal provider)
    * `go run github.com/meidoworks/nekoq-component/configure/secret/cmds/initlocal/. > bootstrap.key`
        1. Default unseal key naming(local unseal only): `bootstrap.key`
        2. Default unseal key id(local unseal only): `1`
3. Initialize certs for CA and TLS
    * Prepare init template according to the example file cmd/initsrv/init_template.toml.example
    * run `go run cmd/initsrv/initsrv.go` to initialize
4. Start service
5. Obtain CA certs
    * Run `curl http://127.0.0.1:4011/api/v1/secret/cert/download/root_ca/ROOT-CA-CERT` to get
      CA certs
    * Import cert on windows
        * `certutil -addstore -f "Root" ca-cert.pem`
        * or `Import-Certificate -FilePath "ca-cert.cer" -CertStoreLocation Cert:\LocalMachine\Root`
        * `certutil -addstore -f "CA" intermediate-cert.pem`
        * or `Import-Certificate -FilePath "ca-cert.cer" -CertStoreLocation Cert:\CurrentUser\Root`
    * Import cert on macos
        * `sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca.crt`
    * Import cert on Debian/Ubuntu
        * Step1: `sudo cp your-ca.crt /usr/local/share/ca-certificates/`
        * Step2: `sudo update-ca-certificates`
    * Import cert on RHEL/CentOS/Fedora
        * Step1: `sudo cp your-ca.crt /etc/pki/ca-trust/source/anchors/`
        * Step2: `sudo update-ca-trust extract`

