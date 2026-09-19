# Mail service

Mail owns email rendering and SMTP delivery. Auth owns decisions, tokens and expiration, and writes a `mail.send.v1` event to its transactional outbox alongside the corresponding state change. This service does not read auth tables and has no database or HTTP API.

## Broker contract

`api/mq` is the private typed contract. Publish persistent JSON messages to durable **direct** exchange `mail`, routing key `mail.send`, with type `mail.send.v1`, a unique message ID and publisher confirms. Start mail before enabling the publisher; use mandatory publication / returned-message handling so a missing binding cannot silently lose messages. The service declares durable queues `mail.send` and `mail.dead`. Provision producer/consumer broker permissions per environment and do not expose this exchange to clients.

Templates: `registration_code`, `password_reset`, `email_change_previous`, `email_change_new`, `password_changed`, `username_changed`, `account_deletion_requested`, `account_deleted`, `new_login`. See `api/mq/mail.go` for fields. Code and link templates require `expirationTimestamp`. Link URLs must use HTTPS. The auth producer owns trusted link construction; message users cannot override From, Reply-To or Subject. Development subjects have `[DEV]`; production subjects have no environment prefix. Default sender is `noreply@chefbook.io`, with no Reply-To.

## Delivery semantics

One message is processed at a time, with up to three SMTP attempts and short bounded backoff. Each attempt has a deadline; cancellation interrupts SMTP network I/O. Expired confirmation messages are acknowledged without sending. Permanent SMTP 5xx rejections are not retried. Invalid payloads and exhausted failures are rejected into `mail.dead` for operator inspection/replay. Shutdown requeues the active delivery. Connection failure exits nonzero for the process supervisor to restart.

Delivery is **at least once**: a crash after SMTP accepts an email and before AMQP acknowledgment can send a duplicate. SMTP has no universal transactional idempotency guarantee. There is deliberately no pretend exactly-once inbox. Retry count is bounded per delivery, not across process crashes. Monitor `mail.dead` depth and restrict retention/access: queued bodies contain personal data and one-time codes. Broker retry/dead-letter settings and retention are deployment responsibilities. Replaying expired verification messages does not send them.

SMTP uses certificate-verified TLS, implicit on 465 and mandatory STARTTLS on other ports. Missing credentials fail startup; there is no silent stub sender. Logs contain only outcomes, never recipients, bodies, SMTP server response text, codes or credentials.

## Run and verification

Copy `deployments/mail.env.example` to an ignored `mail.env` beside the Compose file and insert existing Postbox credentials locally. `AMQP_URL` selects the environment-specific RabbitMQ virtual host; `ENVIRONMENT` must be `develop` or `production`. Do not commit populated environment files. The Compose file connects to an existing RabbitMQ endpoint; it does not provision a broker or join another Compose project's network automatically.

```sh
go test ./...
go build ./cmd/app
docker compose -f deployments/compose.yaml up --build -d
```

Docker build context is this service directory. No messages are sent by tests. The new auth engine must publish through its outbox; legacy auth's direct SMTP worker must only be removed when all runtime paths are switched. This module does not deploy or alter the running environment.

## Postbox and initial cutover

DKIM signing is delegated to Postbox for the verified sender domain; this service does not hold DKIM keys. Existing domain verification and SPF/DKIM/DMARC DNS must be preserved when deploying. These checks were not changed or tested against the live provider in this implementation.

The private `api` module is consumed with a local Go `replace` during development. Before an independent auth image build, include the mail API module in its build context or publish/version it; a sibling-only replacement will not work inside a service-only Docker context. The mail Dockerfile itself needs only this directory, including `api/`. It runs as nonroot with CA certificates from distroless. Source builds and unit tests are verified locally; the container image has not been pulled or deployed.

A broker restart terminates the consumer and the supervisor restarts it. Durable queues and persistent publications retain broker-accepted deliveries, subject to RabbitMQ storage durability. An in-flight SMTP operation can finish before disconnection is noticed; failure to acknowledge then causes redelivery, not an exactly-once guarantee. Review dead-letter retention and operational alerting before enabling public registration.
