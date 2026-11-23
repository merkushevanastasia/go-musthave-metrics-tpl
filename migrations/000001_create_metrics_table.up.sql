CREATE TABLE IF NOT EXISTS metrics (
                                       id VARCHAR(255) NOT NULL,
    mtype VARCHAR(50) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    PRIMARY KEY (id, mtype)
    );