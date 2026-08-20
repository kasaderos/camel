CREATE TABLE portfolios (
    id          TEXT PRIMARY KEY,
    cash        DOUBLE PRECISION NOT NULL,
    cost        DOUBLE PRECISION NOT NULL DEFAULT 0,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE portfolio_stocks (
    portfolio_id  TEXT NOT NULL,
    stock_id      TEXT NOT NULL,
    entry_price   DOUBLE PRECISION NOT NULL DEFAULT 0,
    quantity      DOUBLE PRECISION NOT NULL DEFAULT 0,

    PRIMARY KEY (portfolio_id, stock_id),

    CONSTRAINT fk_portfolio_stocks_portfolio_id
        FOREIGN KEY (portfolio_id)
        REFERENCES portfolios(id)
        ON DELETE CASCADE
);

CREATE TABLE rebalance_tasks (
    id              BIGSERIAL PRIMARY KEY,
    portfolio_id    TEXT NOT NULL,
    stock_id        TEXT NOT NULL,
    quantity        DOUBLE PRECISION NOT NULL,
    side            TEXT NOT NULL,
    status          TEXT NOT NULL,

    order_id        TEXT,
    avg_fill_price  DOUBLE PRECISION NOT NULL DEFAULT 0,
    filled_qty      DOUBLE PRECISION NOT NULL DEFAULT 0,
    submitted_at    TIMESTAMPTZ,
    filled_at       TIMESTAMPTZ,
    error_message   TEXT NOT NULL DEFAULT '',

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_rebalance_tasks_portfolio_id
        FOREIGN KEY (portfolio_id)
        REFERENCES portfolios(id)
        ON DELETE SET NULL
);

CREATE TABLE asset_bars (
    asset_id    TEXT NOT NULL,
    date        DATE NOT NULL,

    open        DOUBLE PRECISION NOT NULL,
    high        DOUBLE PRECISION NOT NULL,
    low         DOUBLE PRECISION NOT NULL,
    close       DOUBLE PRECISION NOT NULL,
    volume      BIGINT NOT NULL DEFAULT 0,

    PRIMARY KEY (asset_id, date)
);
