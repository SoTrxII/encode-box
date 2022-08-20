.PHONY: dapr


dapr:
	dapr run --log-level debug --app-id encode-box --dapr-http-max-request-size="1000" --dapr-http-port=3500 --dapr-grpc-port=50001 --components-path=dapr/components


# https://github.com/golang/mock
mockgen:
	mockgen -source .\pkg\object-storage\object-storage.go -destination .\internal\mock\object-storage.go