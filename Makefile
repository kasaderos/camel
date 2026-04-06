
.PHONY: deps
deps:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go get -tool google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go get -tool connectrpc.com/connect/cmd/protoc-gen-connect-go@latest

.PHONY: proto
proto:
	buf dep update
	buf lint
	buf generate
