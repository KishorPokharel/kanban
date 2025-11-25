include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/app: run the application with go
.PHONY: run/app
run/app:
	KANBAN_DB_DSN=${KANBAN_DB_DSN} go run .

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${KANBAN_DB_DSN}

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## build/app: build the application
.PHONY: build/app
build/app:
	@echo 'Building app...'
	go build -ldflags='-s' -o=./bin/kanban

## build/run: build and run the app
.PHONY: build/run
build/run: build/app
	KANBAN_DB_DSN=${KANBAN_DB_DSN} ./bin/kanban

## clean: clean build folder
.PHONY: clean
clean:
	rm -rf bin
