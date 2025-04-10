# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	used -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

FRONT_END_BINARY=frontApp
BROKER_BINARY=brokerApp
AUTH_BINARY=authApp
LOGGER_BINARY=loggerApp
QUEUE_BINARY=queueApp
HOSTBILL_BINARY=hostbillApp
DUO_BINARY=duoApp

## up: Starts all containers in the background without forcing build.
.PHONY: up
up:
	@echo "Starting Docker images..."
	docker compose up -d
	@echo "Docker images started!"

## up_build: Stops docker compose (if running), builds all projects and starts docker compose.
.PHONY: up_build
up_build: build_broker build_auth build_logger build_queue build_hostbill build_duo build_front
	@echo "Stopping docker images (if running...)"
	docker compose down
	@echo "Building (when required) and starting docker images..."
	docker compose up --build -d
	@echo "Docker images built and started!"

## down: Stop docker compose.
.PHONY: down
down:
	@echo "Stopping docker compose..."
	docker compose down
	@echo "Done!"

## build_broker: Builds the broker binary as a linux executable.
.PHONY: build_broker
build_broker:
	@echo "Building broker binary..."
	cd ../broker-svc && env GOOS=linux CGO_ENABLED=0 go build -o ${BROKER_BINARY} ./cmd/api
	@echo "Done!"

## build_auth: Builds the auth binary as a linux executable.
.PHONY: build_auth
build_auth:
	@echo "Building auth binary..."
	cd ../auth-svc && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BINARY} ./cmd/api
	@echo "Done!"

## build_logger: Builds the logger binary as a linux executable.
.PHONY: build_logger
build_logger:
	@echo "Building logger binary..."
	cd ../logger-svc && env GOOS=linux CGO_ENABLED=0 go build -o ${LOGGER_BINARY} ./cmd/api
	@echo "Done!"

## build_queue: Builds the queue binary as a linux executable.
.PHONY: build_queue
build_queue:
	@echo "Building queue binary..."
	cd ../queue-svc && env GOOS=linux CGO_ENABLED=0 go build -o ${QUEUE_BINARY} .
	@echo "Done!"


## build_hostbill: Builds the hostbill binary as a linux executable.
.PHONY: build_hostbill
build_hostbill:
	@echo "Building hostbill binary..."
	cd ../hostbill-svc && env GOOS=linux CGO_ENABLED=0 go build -o ${HOSTBILL_BINARY} ./cmd/api
	@echo "Done!"

## build_duo: Builds the duo binary as a linux executable.
.PHONY: build_duo
build_duo:
	@echo "Building duo binary..."
	cd ../duo-svc && env GOOS=linux CGO_ENABLED=0 go build -o ${DUO_BINARY} ./cmd/api
	@echo "Done!"

## build_front: Builds the front end binary.
.PHONY: build_front
build_front:
	@echo "Building front end binary..."
	cd ../front-end && env CGO_ENABLED=0 go build -o ${FRONT_END_BINARY} ./cmd/web
	@echo "Done!"

## start: Starts the front end.
.PHONY: start
start: build_front-svc
	@echo "Starting front end"
	cd ../front-end && ./${FRONT_END_BINARY} &

## stop: Stops the front end.
.PHONY: stop
stop:
	@echo "Stopping front end..."
	@-pkill -SIGTERM -f "./${FRONT_END_BINARY}"
	@echo "Stopped front end!"


# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

