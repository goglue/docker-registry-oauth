dep:
	go get ./...

test:
	go test -v -bench=. -benchmem -race ./...

build: dep
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -a -tags netgo \
	-o docker-registry-oauth cmd/server.go

build-image:
	docker build . -t ${DOCKER_ORG}/docker-registry-oauth:${TRAVIS_TAG} \
	-t ${DOCKER_ORG}/docker-registry-oauth:latest

deploy-image:
	docker push ${DOCKER_ORG}/docker-registry-oauth:${TRAVIS_TAG} &&\
	docker push ${DOCKER_ORG}/docker-registry-oauth:latest
