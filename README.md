# Micro Service Hub

This monorepo contains micro services to power a cloud platform. This is
currently using the Chi router in most places, refactoring to use only the Go
standard library is planned. Services, include:

- Front End "Service"
- Broker Service
- Authentication Service
- Listener Service
- Logger Service

More services will be added as needed, here is a diagram for the general
overview and state of the project.

## Getting Started

### Prerequisites

- Go
- Docker
- Make

### Install and Run

Docker is used extensively, and can be used in conjunction with Kubernetes or
Docker Swarm.

To get started, clone the repo and `cd` into the project root, then into
`config/`. Check out Makefile comments for more info, here are some key Makefile
commands.

1. Compile and run back end - `make up_build`
2. Compile and run front end - `make start`
3. Stop front end - `make stop`
4. Stop back end - `make down`
5. Repeat as necessary

Visit `http://localhost` to access the service testing page.

## Services

### Front End "Service"

This service is strictly for testing communication between services. Currently,
only the front end and broker services are operational.

### Broker Service (API Gateway)

The broker service connects to the front end testing "service" and and other
clients like Hostbill.

Each addon service, like Veeam and Zerto, will be its own micro service. The
broker service will focus on middleware and serving as the entrypoint. Renaming
it to "API Gateway" may better reflect its role.

### Authentication Service

This runs a Postgres DB container with a `users` table.

For now, the authentication service could have limited responsibilities.
Hostbill web hook API keys can be loaded from Github secrets, or we can save
those credentials in the DB instead.

Advanced features for managing keys and credentials can be added here in the
future.

### Listener Service

The Listener service will run RabbitMQ with gRPC. It will enable perfomant
internal communication between all services.

### Logger Service

Probably integrate this service with ARIA logging (custom ELK stack) as
mentioned by Ryan.

We can use a DB for this if we need.

### Other Services

Other services will be added as needed.
