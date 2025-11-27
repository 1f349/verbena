package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/conf"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/logger"
	"github.com/1f349/verbena/rest"
	"github.com/go-chi/chi/v5"
	"github.com/gobuffalo/nulls"
)

type recordQueries interface {
	GetZoneRecords(ctx context.Context, zoneId int64) ([]database.GetZoneRecordsRow, error)
	GetZoneRecord(ctx context.Context, row database.GetZoneRecordParams) (database.GetZoneRecordRow, error)
	GetZone(ctx context.Context, zoneId int64) (database.Zone, error)
	InsertRecordFromApi(ctx context.Context, row database.InsertRecordFromApiParams) (int64, error)
	UpdateRecordFromApi(ctx context.Context, row database.UpdateRecordFromApiParams) error
	DeleteRecordFromApi(ctx context.Context, row database.DeleteRecordFromApiParams) error
}

func RecordToRestRecord(record database.Record) (rest.Record, error) {
	v, err := rest.ParseRecordValue(record.Type, record.PreValue)
	if err != nil {
		logger.Logger.Debug("Failed to call RecordToRestRecord", "id", record.ID, "zone id", record.ZoneID, "type", record.Type, "pre-value", record.PreValue, "error", err)
		return rest.Record{}, err
	}
	return rest.Record{
		ID:     record.ID,
		Name:   record.Name,
		ZoneID: record.ZoneID,
		Ttl:    record.PreTtl,
		Type:   record.Type,
		Active: record.PreActive,
		Value:  v,
	}, nil
}

func appendRecord(slice []rest.Record, record database.Record) []rest.Record {
	record2, err := RecordToRestRecord(record)
	if err != nil {
		return slice
	}
	return append(slice, record2)
}

func AddRecordRoutes(r chi.Router, db recordQueries, keystore *mjwt.KeyStore, nameservers conf.NameserverConf) {
	r.Route("/zones/{zone_id:[0-9]+}/records", func(r chi.Router) {
		// List all records
		r.Get("/", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
			zoneId, err := getZoneId(req)
			if err != nil {
				http.Error(rw, "Invalid zone ID", http.StatusBadRequest)
				return
			}

			zone, err := db.GetZone(req.Context(), zoneId)
			if err != nil {
				logger.Logger.Error("Failed to get zone", "err", err)
				http.Error(rw, "Database error occurred", http.StatusInternalServerError)
				return
			}

			if !b.Claims.Perms.Has("domain:owns=" + zone.Name) {
				http.NotFound(rw, req)
				return
			}

			rows, err := db.GetZoneRecords(req.Context(), zoneId)
			if err != nil {
				logger.Logger.Error("Failed to get zone records", "err", err)
				http.Error(rw, "Database error occurred", http.StatusInternalServerError)
				return
			}

			ns := nameservers.GetNameserversForZone(zone)

			records := make([]rest.Record, 0, len(rows)+len(ns))
			for _, ns := range ns {
				records = append(records, rest.Record{
					ID:     -2,
					Name:   "@",
					ZoneID: zoneId,
					Ttl:    nulls.Int32{},
					Type:   "NS",
					Value: rest.RecordValue{
						Target: ns,
					},
					Active: true,
				})
			}
			for _, record := range rows {
				if !b.Claims.Perms.Has("domain:owns=" + record.Name) {
					continue
				}
				records = appendRecord(records, record.Record)
			}

			json.NewEncoder(rw).Encode(records)
		}))

		// List individual record
		r.Get("/{record_id:[0-9]+}", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
			zoneId, err := getZoneId(req)
			if err != nil {
				http.Error(rw, "Invalid zone ID", http.StatusBadRequest)
				return
			}

			recordId, err := getRecordId(req)
			if err != nil {
				http.Error(rw, "Invalid record ID", http.StatusBadRequest)
				return
			}

			row, err := db.GetZoneRecord(req.Context(), database.GetZoneRecordParams{
				RecordID: recordId,
				ZoneID:   zoneId,
			})
			if err != nil {
				logger.Logger.Error("Failed to get zone records", "err", err)
				http.Error(rw, "Database error occurred", http.StatusInternalServerError)
				return
			}

			if !b.Claims.Perms.Has("domain:owns=" + row.Name) {
				http.NotFound(rw, req)
				return
			}

			record, err := RecordToRestRecord(row.Record)
			if err != nil {
				http.Error(rw, "Server error occurred", http.StatusInternalServerError)
				return
			}

			json.NewEncoder(rw).Encode(record)
		}))

		// Create record
		r.Post("/", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
			var record struct {
				Name  string           `json:"name"`
				Ttl   nulls.Int32      `json:"ttl"`
				Type  string           `json:"type"`
				Value rest.RecordValue `json:"value"`
			}

			err := json.NewDecoder(req.Body).Decode(&record)
			if err != nil {
				http.Error(rw, "Invalid request body", http.StatusBadRequest)
				return
			}

			if record.Ttl.Valid && record.Ttl.Int32 <= 200 {
				http.Error(rw, "Invalid time to live, expected 'ttl <= 200'", http.StatusBadRequest)
				return
			}

			if !record.Value.IsValidForType(record.Type) {
				http.Error(rw, "Invalid value for type", http.StatusBadRequest)
				return
			}

			zoneId, err := getZoneId(req)
			if err != nil {
				http.Error(rw, "Invalid zone ID", http.StatusBadRequest)
				return
			}

			zone, err := db.GetZone(req.Context(), zoneId)
			if err != nil {
				http.Error(rw, "Invalid zone ID", http.StatusBadRequest)
				return
			}

			if !b.Claims.Perms.Has("domain:owns=" + zone.Name) {
				http.NotFound(rw, req)
				return
			}

			genId, err := db.InsertRecordFromApi(req.Context(), database.InsertRecordFromApiParams{
				Name:      record.Name,
				ZoneID:    zoneId,
				Type:      record.Type,
				PreTtl:    record.Ttl,
				PreValue:  record.Value.ToValueString(record.Type),
				PreActive: true,
			})
			if err != nil {
				logger.Logger.Debug("Failed to insert record from API", "err", err)
				http.Error(rw, "Database error occurred", http.StatusInternalServerError)
				return
			}

			json.NewEncoder(rw).Encode(rest.Record{
				ID:     genId,
				Name:   record.Name,
				ZoneID: zoneId,
				Ttl:    record.Ttl,
				Type:   record.Type,
				Active: true,
				Value:  record.Value,
			})
		}))

		// Update record
		r.Put("/{record_id:[0-9]+}", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
			var record struct {
				Ttl    nulls.Int32      `json:"ttl"`
				Active bool             `json:"active"`
				Value  rest.RecordValue `json:"value"`
			}

			err := json.NewDecoder(req.Body).Decode(&record)
			if err != nil {
				http.Error(rw, "Invalid request body", http.StatusBadRequest)
				return
			}

			if record.Ttl.Valid && record.Ttl.Int32 <= 200 {
				http.Error(rw, "Invalid time to live, expected 'ttl <= 200'", http.StatusBadRequest)
				return
			}

			zoneId, err := getZoneId(req)
			if err != nil {
				http.Error(rw, "Invalid zone ID", http.StatusBadRequest)
				return
			}

			recordId, err := getRecordId(req)
			if err != nil {
				http.Error(rw, "Invalid record ID", http.StatusBadRequest)
				return
			}

			originalRecord, err := db.GetZoneRecord(req.Context(), database.GetZoneRecordParams{
				RecordID: recordId,
				ZoneID:   zoneId,
			})
			if err != nil {
				logger.Logger.Debug("Failed to get zone record", "err", err)
				http.Error(rw, "Database error occurred", http.StatusInternalServerError)
				return
			}

			if !record.Value.IsValidForType(originalRecord.Record.Type) {
				http.Error(rw, "Invalid value for type", http.StatusBadRequest)
				return
			}

			if !b.Claims.Perms.Has("domain:owns=" + originalRecord.Name) {
				http.NotFound(rw, req)
				return
			}

			err = db.UpdateRecordFromApi(req.Context(), database.UpdateRecordFromApiParams{
				PreTtl:    record.Ttl,
				PreValue:  record.Value.ToValueString(originalRecord.Record.Type),
				PreActive: record.Active,
				ID:        recordId,
				ZoneID:    zoneId,
			})
			if err != nil {
				logger.Logger.Debug("Failed to update record from API", "err", err)
				http.Error(rw, "Database error occurred", http.StatusInternalServerError)
				return
			}

			json.NewEncoder(rw).Encode(rest.Record{
				ID:     recordId,
				Name:   originalRecord.Record.Name,
				ZoneID: zoneId,
				Ttl:    record.Ttl,
				Type:   originalRecord.Record.Type,
				Active: record.Active,
				Value:  record.Value,
			})
		}))

		// Delete record
		r.Delete("/{record_id:[0-9]+}", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
			zoneId, err := getZoneId(req)
			if err != nil {
				http.Error(rw, "Invalid zone ID", http.StatusBadRequest)
				return
			}

			recordId, err := getRecordId(req)
			if err != nil {
				http.Error(rw, "Invalid record ID", http.StatusBadRequest)
				return
			}

			originalRecord, err := db.GetZoneRecord(req.Context(), database.GetZoneRecordParams{
				RecordID: recordId,
				ZoneID:   zoneId,
			})

			if !b.Claims.Perms.Has("domain:owns=" + originalRecord.Name) {
				http.NotFound(rw, req)
				return
			}

			err = db.DeleteRecordFromApi(req.Context(), database.DeleteRecordFromApiParams{
				RecordID: recordId,
				ZoneID:   zoneId,
			})
			if err != nil {
				logger.Logger.Debug("Failed to delete record from API", "err", err)
				http.Error(rw, "Database error occurred", http.StatusInternalServerError)
				return
			}

			rw.WriteHeader(http.StatusOK)
		}))
	})
}

func getRecordId(req *http.Request) (int64, error) {
	recordIdRaw := chi.URLParam(req, "record_id")
	return strconv.ParseInt(recordIdRaw, 10, 64)
}
