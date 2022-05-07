build:
	protoc -I. --go_out=plugins=grpc:. \
	proto/service.proto
