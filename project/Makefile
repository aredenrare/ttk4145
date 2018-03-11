DOCKER = eirare

.PHONY: build
build:
	CGO_ENABLE=0 GOOS=linux GOARCH=arm go build -o ./bin/elevator ./src/*.go

PHONY: create-docker
create-docker:
	docker build -t "$(DOCKER):latest" .

.PHONY: build-create-docker
build-create-docker: build create-docker


	