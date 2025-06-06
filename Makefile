SHELL := /bin/bash

.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## up: starts through docker the application exposing its HTTP port
up:
	docker-compose up app
	docker-compose down --remove-orphans

## down: stops the application and removes the containers
down:
	docker-compose down --remove-orphans

## clean: clean up all docker containers
clean: down
	docker ps -aq | xargs docker stop | xargs docker rm

## tests: Runs all tests in the project
tests:
	@ echo "Running tests..."
	go clean -testcache && go test -race ./...

## lint: Runs linter for all packages
lint:
	@ docker run  --rm -v "`pwd`:/workspace:cached" -w "/workspace/." golangci/golangci-lint:v2.1-alpine golangci-lint run ./...
