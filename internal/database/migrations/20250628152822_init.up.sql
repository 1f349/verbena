CREATE TABLE IF NOT EXISTS zones
(
    id     INTEGER     NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name   TEXT UNIQUE NOT NULL,
    serial BIGINT      NOT NULL,
    active BOOLEAN     NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS owners
(
    id      INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    zone_id INTEGER NOT NULL,
    user_id TEXT    NOT NULL,

    FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE TABLE IF NOT EXISTS records
(
    id      INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name    TEXT    NOT NULL,
    zone_id INTEGER NOT NULL,
    ttl     INTEGER NULL,
    type    TEXT    NOT NULL,
    value   TEXT    NOT NULL,
    active  BOOLEAN NOT NULL DEFAULT 1,

    FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE INDEX record_name ON records (name);
CREATE INDEX record_type ON records (type);

CREATE TABLE IF NOT EXISTS staged_records
(
    id        INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    zone_id   INTEGER NOT NULL,
    record_id INTEGER NULL,
    ttl       INTEGER NULL,
    value     TEXT    NOT NULL,
    active    BOOLEAN NOT NULL,

    FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    FOREIGN KEY (record_id) REFERENCES records (id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE INDEX staged_record_record_id ON staged_records (record_id);
CREATE INDEX staged_record_zone_id ON staged_records (record_id);
