-- name: GetActiveZones :many
SELECT *
FROM zones
WHERE active = 1;
