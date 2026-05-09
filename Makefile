
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

portfolio:
	docker compose run --rm portfolio-manager create --csv /app/portfolios/portfolio-1.csv

portfolio-info:
	docker compose run --rm portfolio-manager info --id portfolio-agent-1

portfolio-rebalance:
	docker compose run --rm portfolio-manager rebalance --id portfolio-agent-1
