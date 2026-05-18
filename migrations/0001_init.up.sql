CREATE TABLE portfolios (
    id          TEXT PRIMARY KEY,
    cash        DOUBLE PRECISION NOT NULL,
    weights     JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE portfolio_agents (
    id            TEXT PRIMARY KEY,
    portfolio_id  TEXT NOT NULL,
    asset_id      TEXT NOT NULL,
    asset_qty     DOUBLE PRECISION NOT NULL DEFAULT 0,
    score         DOUBLE PRECISION NOT NULL DEFAULT 0,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_portfolio_agents_portfolio_id
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

    PRIMARY KEY (asset_id, date)
);
