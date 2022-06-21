all: vendor test build

build:
	#make vendor
	CGO_ENABLE=0 go build -o app main.go

linux_build:
	#make vendor
	CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o app main.go

test:
	#make vendor
	CGO_ENABLED=1 GOOS=linux go test -race -v ./...

vendor:
	go mod tidy && go mod vendor

docker:
	sh ./docker_build.sh $(tag)

run:
	sh ./deploy.sh $(tag)
