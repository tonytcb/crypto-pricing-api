SHELL := /bin/bash

.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## up: starts through docker the application exposing its HTTP port
up:
	docker-compose up app
	docker-compose down
