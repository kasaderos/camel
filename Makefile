
.PHONY: deps
deps:

lint:
	golangci-lint run --fix --config .golangci.yml

portfolio:
	docker compose run --rm portfolio-manager create --id portfolio3 --csv /app/portfolios/portfolio-3.csv --cash 10000

portfolio-info:
	docker compose run --rm portfolio-manager info --id portfolio3

portfolio-score:
	docker compose run --rm portfolio-manager score --id portfolio3

portfolio-plan:
	docker compose run --rm portfolio-manager plan --id portfolio3

portfolio-rebalance:
	docker compose run --rm portfolio-manager rebalance --id portfolio3

migrate-drop:
	docker compose run --rm portfolio-manager migrate-drop

migrate-up:
	docker compose run --rm portfolio-manager migrate-up

docker-image:
	docker build --platform linux/amd64 -t kasaderos99/camel:v1 .

# docker exec -it eae31436815c psql -U camel -d camel
