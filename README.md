# CloudKey Platform Service Hub

This monorepo contains the possible micro services solution to power CloudKey
Platform automation. Services, include:

- Front End "Service"
- Broker Service
- Authentication Service
- Listener Service
- Logger Service

More services will be added as needed, here is a diagram for the general
overview and state of the project.

![Diagram](./docs/images/logical-layer-micro-services.png)

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

#### Hostbill Web Hook Validation

Some notes from Hostbill docs regarding web hook authentication and validation.

> Each request is signed with your secret key, this allows you to validate that
> the events were sent by your HostBill installation, not by a third party.
>
> Here are the steps required to validate request signature.
>
> Step 1: Obtain timestamp and signature from response headers, the HB-Signature
> header contains signature that you want to verify and HB-Timestamp contains
> timestamp used to generate that signature.
>
> Step 2: Prepare the payload string by concatenating:
>
> Step 3: Compute a HMAC with the SHA256 hash function. Use secret as the key,
> and use the payload string as the message.
>
> Step 4: Compare the signature in the HB-Signature header to the one computed
> in step 3. If it matches, compute the difference between the current timestamp
> and the received timestamp, then decide if the difference is within your
> tolerance.

See the provided PHP snippet for details.

```php
<?php

$secret = ''; // paste your Secret here

//fetch request body
$data = file_get_contents('php://input');
$payload = $_SERVER["HTTP_HB_TIMESTAMP"] . $data;

$signature = hash_hmac('sha256', $payload, $secret);

//compare signature in header with the one computed above
if($signature !== $_SERVER["HTTP_HB_SIGNATURE"])
    die('invalid signature')

// signature valid, verify timestamp
if($_SERVER["HTTP_HB_TIMESTAMP"] < time() - 60)
    die('timestamp older than 60 sec')
```

### Listener Service

The Listener service will run RabbitMQ with gRPC. It will enable perfomant
internal communication between all services.

### Logger Service

Probably integrate this service with ARIA logging (custom ELK stack) as
mentioned by Ryan.

We can use a DB for this if we need.

### Other Services

Other services will be added as needed.
