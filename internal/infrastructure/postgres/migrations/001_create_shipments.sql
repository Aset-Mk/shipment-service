CREATE TABLE IF NOT EXISTS shipments (
    id             TEXT        PRIMARY KEY,
    reference      TEXT        NOT NULL UNIQUE,
    origin         TEXT        NOT NULL,
    destination    TEXT        NOT NULL,
    status         TEXT        NOT NULL DEFAULT 'pending',
    driver_name    TEXT        NOT NULL DEFAULT '',
    driver_license TEXT        NOT NULL DEFAULT '',
    unit_id        TEXT        NOT NULL DEFAULT '',
    unit_type      TEXT        NOT NULL DEFAULT '',
    amount         NUMERIC(12,2) NOT NULL DEFAULT 0,
    driver_revenue NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL
);
