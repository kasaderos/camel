
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

lint:
	golangci-lint run --fix --config .golangci.yml

portfolio:
	docker compose run --rm portfolio-manager create --id portfolio1 --csv /app/portfolios/portfolio-1.csv --cash 10000
	docker compose run --rm portfolio-manager create --id portfolio2 --csv /app/portfolios/portfolio-2.csv --cash 10000

portfolio-info:
	docker compose run --rm portfolio-manager info --id portfolio1
	docker compose run --rm portfolio-manager info --id portfolio2

portfolio-score:
	docker compose run --rm portfolio-manager score --id portfolio1

portfolio-rebalance:
	docker compose run --rm portfolio-manager rebalance --id portfolio1
	docker compose run --rm portfolio-manager rebalance --id portfolio2

migrate-drop:
	docker compose run --rm portfolio-manager migrate-drop

migrate-up:
	docker compose run --rm portfolio-manager migrate-up

docker-image:
	docker build --platform linux/amd64 -t kasaderos99/camel:v1 .
