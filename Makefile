GO111MODULE := auto
export GO111MODULE

lint:
	golangci-lint run ./...

test:
	go test -count=1 -race ./...

build_service_a:
	mkdir -p /var/tmp/service_a/cache && chmod -R 777 /var/tmp/service_a/cache
	go build -tags musl -ldflags="-w -extldflags '-static' -X 'main.Version=$(VERSION)'" -o service_a karma8/cmd/service_a

build_service_b:
	go build -tags musl -ldflags="-w -extldflags '-static' -X 'main.Version=$(VERSION)'" -o service_b karma8/cmd/service_b

check-swagger:
	which swagger

swagger: check-swagger
	GO111MODULE=on go mod vendor && GO111MODULE=off swagger generate spec -o ./doc/swagger.json --scan-models

serve-swagger: check-swagger
	swagger serve -F=swagger ./doc/swagger.json
