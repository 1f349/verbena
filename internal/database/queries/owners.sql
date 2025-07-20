-- name: GetOwnerByUserIdAndZone :one
SELECT sqlc.embed(owners), sqlc.embed(zones)
FROM owners
         INNER JOIN zones ON owners.zone_id = zones.id
WHERE user_id = ?
  AND zones.name = ?;
