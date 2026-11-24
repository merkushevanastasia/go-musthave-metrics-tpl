CREATE TABLE IF NOT EXISTS metrics (
   id BIGSERIAL PRIMARY KEY,
   name VARCHAR(255) NOT NULL,
    mtype VARCHAR(50) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    CONSTRAINT metrics_name_mtype_unique UNIQUE (name, mtype)
    );
