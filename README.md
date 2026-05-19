# Portfolio Manager

A simple portfolio management CLI powered by Docker Compose.

## Requirements

* [Docker](https://www.docker.com/?utm_source=chatgpt.com)
* [Docker Compose](https://docs.docker.com/compose/?utm_source=chatgpt.com)
* `make`

---

### Run Database Migrations

Apply migrations:

```bash
make migrate-up
```

Equivalent command:

```bash
docker compose run --rm portfolio-manager migrate-up
```

Drop migrations:

```bash
make migrate-drop
```

Equivalent command:

```bash
docker compose run --rm portfolio-manager migrate-drop
```

---

## Available Commands

### Create Portfolios

Creates two portfolios from CSV files with an initial cash balance of `10000`.

```bash
make portfolio
```

Equivalent commands:

```bash
docker compose run --rm portfolio-manager create --id portfolio1 --csv /app/portfolios/portfolio-1.csv --cash 10000

docker compose run --rm portfolio-manager create --id portfolio2 --csv /app/portfolios/portfolio-2.csv --cash 10000
```

---

### Show Portfolio Info

Displays information about the portfolios.

```bash
make portfolio-info
```

Equivalent commands:

```bash
docker compose run --rm portfolio-manager info --id portfolio1

docker compose run --rm portfolio-manager info --id portfolio2
```

---

### Rebalance Portfolios

Runs portfolio rebalancing for both portfolios.

```bash
make portfolio-rebalance
```

Equivalent commands:

```bash
docker compose run --rm portfolio-manager rebalance --id portfolio1

docker compose run --rm portfolio-manager rebalance --id portfolio2
```
---

## Example Workflow

```bash
# Apply migrations
make migrate-up

# Create portfolios
make portfolio

# View portfolio info
make portfolio-info

# Rebalance portfolios
make portfolio-rebalance
```

---

## Project Structure

```text
.
├── Makefile
├── docker-compose.yml
├── portfolios/
│   ├── portfolio-1.csv
│   └── portfolio-2.csv
```

