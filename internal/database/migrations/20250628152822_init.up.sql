CREATE TABLE IF NOT EXISTS zones
(
    id     BIGINT  NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name   TEXT    NOT NULL UNIQUE,
    serial BIGINT  NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS owners
(
    id      BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    zone_id BIGINT NOT NULL,
    user_id TEXT   NOT NULL,

    FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE TABLE IF NOT EXISTS records
(
    id         BIGINT  NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name       TEXT    NOT NULL,
    zone_id    BIGINT  NOT NULL,
    ttl        INTEGER NULL,
    type       TEXT    NOT NULL,
    value      TEXT    NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT 1,

    pre_ttl    INTEGER NULL,
    pre_value  TEXT    NOT NULL,
    pre_active BOOLEAN NOT NULL,
    pre_delete BOOLEAN NOT NULL,

    FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE INDEX record_name ON records (name);
CREATE INDEX record_type ON records (type);
CREATE INDEX record_zone_id ON records (zone_id);
