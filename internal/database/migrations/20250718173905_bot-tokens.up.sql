CREATE TABLE IF NOT EXISTS bot_tokens
(
    id       BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    owner_id BIGINT NOT NULL,
    zone_id  BIGINT NOT NULL,

    FOREIGN KEY (owner_id) REFERENCES owners (id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE RESTRICT ON UPDATE RESTRICT
);
