module github.com/mephistolie/chefbook-backend-mail

go 1.26.2

require (
	github.com/mephistolie/chefbook-backend-mail/api v0.0.0
	github.com/rabbitmq/amqp091-go v1.11.0
)

replace github.com/mephistolie/chefbook-backend-mail/api => ./api
