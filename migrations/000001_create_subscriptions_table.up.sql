CREATE TABLE IF NOT EXISTS subscriptions (
    id           UUID        PRIMARY KEY,
    service_name TEXT        NOT NULL,
    price        BIGINT      NOT NULL,
    user_id      UUID        NOT NULL,
    start_date   DATE        NOT NULL,
    end_date     DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id      ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions (LOWER(service_name));
CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date   ON subscriptions (start_date);
