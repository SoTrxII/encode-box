.PHONY: dapr


dapr:
	dapr run --log-level debug --app-id encode-box --dapr-http-max-request-size="1000" --dapr-grpc-port=50010 --resources-path=dapr/components

mocks:
	mockery

test_with_dapr:
	dapr run --app-id=encode-box  --resources-path ./dapr/components -- go test --tags integration -v ./... -covermode=atomic -coverprofile=coverage.out

build:
	go build cmd/server.go
test:
	go test ./...
profile: build
	OBJECT_STORE_NAME=object-store ./server.exe
# https://github.com/golang/mock
mockgen:
	mockgen -source .\pkg\object-storage\object-storage.go -destination .\internal\mock\object-storage.go
	mockgen -source .\pkg\encode-box\encode-box.go -destination .\internal\mock\encode-box.go