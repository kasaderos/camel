CREATE TABLE portfolio_agents (
    id            TEXT PRIMARY KEY,
    portfolio_id  TEXT NOT NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE asset_agents (
    id                    TEXT PRIMARY KEY,

    portfolio_agent_id    TEXT NOT NULL,
    asset_id              TEXT NOT NULL,

    asset_qty             DOUBLE PRECISION NOT NULL DEFAULT 0,
    cash                  DOUBLE PRECISION NOT NULL DEFAULT 0,

    state                 JSONB NOT NULL DEFAULT '{}'::jsonb,

    CONSTRAINT fk_asset_agents_portfolio_agent
        FOREIGN KEY (portfolio_agent_id)
        REFERENCES portfolio_agents(id)
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