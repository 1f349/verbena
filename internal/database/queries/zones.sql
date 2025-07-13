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
        CASE
            WHEN LEFT(serial, 8) = DATE_FORMAT(CURDATE(), '%Y%m%d') THEN serial + 1
            ELSE CAST(DATE_FORMAT(CURDATE(), '%Y%m%d') AS UNSIGNED) * 100 + 1
            END
WHERE id = ?;
