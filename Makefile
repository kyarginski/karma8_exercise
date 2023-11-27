GO=go
GO111MODULE := auto
export GO111MODULE

lint:
	golangci-lint run ./...

test:
	$(GO) test -count=1 -race ./...

build_blog:
	$(GO) build -tags musl -ldflags="-w -extldflags '-static' -X 'main.Version=$(VERSION)'" -o blog check24/cmd/blog

build_admin:
	$(GO) build -tags musl -ldflags="-w -extldflags '-static' -X 'main.Version=$(VERSION)'" -o admin check24/cmd/admin

check-swagger:
	which swagger

swagger: check-swagger
	GO111MODULE=on go mod vendor && GO111MODULE=off swagger generate spec -o ./doc/swagger.json --scan-models

serve-swagger: check-swagger
	swagger serve -F=swagger ./doc/swagger.json
