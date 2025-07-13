-- name: GetZoneActiveRecords :many
SELECT *
FROM records
WHERE active = 1
  AND zone_id = ?;

-- name: GetZoneRecords :many
SELECT sqlc.embed(records), zones.name
FROM records
         INNER JOIN zones ON records.zone_id = zones.id
WHERE zone_id = ?;

-- name: GetZoneRecord :one
SELECT sqlc.embed(records), zones.name
FROM records
         INNER JOIN zones ON records.zone_id = zones.id
WHERE records.id = sqlc.arg(record_id)
  AND zones.id = sqlc.arg(zone_id);

-- name: InsertRecordFromApi :execlastid
INSERT INTO records (name, zone_id, ttl, type, value, active, pre_ttl, pre_value, pre_active)
VALUES (?, ?, 0, ?, "", 0, ?, ?, ?);

-- name: UpdateRecordFromApi :exec
UPDATE records
SET pre_ttl    = ?,
    pre_value  = ?,
    pre_active = ?
WHERE id = ?
  AND zone_id = ?;

-- name: DeleteRecordFromApi :exec
UPDATE records
SET pre_delete = TRUE
WHERE id = sqlc.arg(record_id)
  AND zone_id = sqlc.arg(zone_id);

-- name: CommitZoneRecords :execrows
UPDATE records
SET ttl    = pre_ttl,
    value  = pre_value,
    active = pre_active
WHERE zone_id = ?
  AND (
    ttl != pre_ttl
        OR (ttl IS NULL) != (pre_ttl IS NULL)
        OR (`value` != pre_value)
        OR (active != pre_active)
    );
