-- name: GetActiveZones :many
SELECT *
FROM zones
WHERE active = 1;

-- name: GetOwnedZones :many
SELECT sqlc.embed(zones), owners.user_id
FROM zones
         INNER JOIN owners ON zones.id = owners.zone_id
WHERE owners.user_id = ?;

-- name: GetZone :one
SELECT *
FROM zones
WHERE id = ?;

-- name: UpdateZoneSerial :exec
UPDATE zones
SET serial =
        IF(LEFT(serial, 8) = DATE_FORMAT(CURDATE(), '%Y%m%d'), serial + 1,
           CAST(DATE_FORMAT(CURDATE(), '%Y%m%d') AS UNSIGNED) * 100 + 1)
WHERE id = ?;

-- name: LookupZone :one
SELECT id
FROM zones
WHERE name = ?;

-- name: UpdateZoneConfig :exec
UPDATE zones
SET refresh = ?,
    retry   = ?,
    expire  = ?,
    ttl     = ?
WHERE id = ?;
