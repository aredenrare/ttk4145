DOCKER = eirare

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/elevator ./src/*.go

.PHONY: build-race
build-race:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -race -o ./bin/elevator ./src/*.go

PHONY: create-docker
create-docker:
	docker build -t "$(DOCKER):latest" .

.PHONY: build-create-docker
build-create-docker: build create-docker


	