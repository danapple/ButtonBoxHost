run: build docker
	docker run docker-bbox:latest --port=${BBOX_PORT}

build:
	go build *.go

docker: build
	docker build --tag docker-bbox .