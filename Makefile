.PHONY: dapr


dapr:
	dapr run --log-level debug --app-id encode-box --dapr-http-max-request-size="1000" --dapr-http-port=3500 --dapr-grpc-port=50001 --components-path=dapr/components

build:
	go build cmd/server.go

profile: build
	OBJECT_STORE_NAME=object-store ./server.exe
# https://github.com/golang/mock
mockgen:
	mockgen -source .\pkg\object-storage\object-storage.go -destination .\internal\mock\object-storage.go
	mockgen -source .\pkg\encode-box\encode-box.go -destination .\internal\mock\encode-box.go