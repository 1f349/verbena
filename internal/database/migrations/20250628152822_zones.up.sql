CREATE TABLE IF NOT EXISTS zones
(
    id      BIGINT  NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name    TEXT    NOT NULL UNIQUE,
    serial  BIGINT  NOT NULL,
    admin   TEXT    NOT NULL,
    refresh INTEGER NOT NULL,
    retry   INTEGER NOT NULL,
    expire  INTEGER NOT NULL,
    ttl     INTEGER NOT NULL,
    active  BOOLEAN NOT NULL DEFAULT 1,

    INDEX zone_name_index (name)
);
