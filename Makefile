PACKAGES := $(shell go list ./...)
# name := $(shell basename ${PWD})

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## init: initialize project (make init module=github.com/user/project)
.PHONY: init
init:
	go install github.com/air-verse/air@latest
	asdf reshim golang

## vet: vet code
.PHONY: vet
vet:
	go vet $(PACKAGES)

## test: run unit tests
.PHONY: test
test:
	go test -race -cover $(PACKAGES)

## migrate-create: create a new migration
.PHONY: migrate-create
migrate-create:
	go run cmd/migrate/main.go create "$(name)"

## migrate-up: migrate up
.PHONY: migrate-up
migrate-up:
	go run cmd/migrate/main.go up

## migrate-down: migrate down
.PHONY: migrate-down
migrate-down:
	go run cmd/migrate/main.go down

## migrate-status: migrate status
.PHONY: migrate-status
migrate-status:
	go run cmd/migrate/main.go status

## migrate-version: migrate version
.PHONY: migrate-version
migrate-version:
	go run cmd/migrate/main.go version