# ChefBook backend

ChefBook backend root repository

## Architecture

<p align="center">
    <img src="img/architecture.png" alt="ChefBook backend architecture" width="500" />
</p>

ChefBook uses pretty default microservice architecture. Each service has its own isolated database.
Services communicate each other via sync gRPC or async Message Queue, if it's need.
Server accessible by client via API Gateway, that forwards traffic to target service.
Auth takes place via JWT Token, signed with RSA keypair. API Gateway periodically fetches public key by Auth Service.

## Modules

* `common-lib` - Go common library for infrastructure applications
* `api-gateway` - API Gateway with REST-gRPC conversion
* `service/template` - template for new ChefBook service
* `service/auth` - service provides user authentication/authorization flow
* `secrets` - Helm Chart with sensitive information, encrypted with [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) (obviously, repository is private)