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

    FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    INDEX record_name (name),
    INDEX record_type (type),
    INDEX record_zone_id (zone_id)
);
