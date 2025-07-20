package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/logger"
	"github.com/1f349/verbena/rest"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type zoneQueries interface {
	GetOwnedZones(ctx context.Context, userID string) ([]database.GetOwnedZonesRow, error)
	GetZone(ctx context.Context, id int64) (database.Zone, error)
}

func ZoneToRestZone(zone database.Zone) rest.Zone {
	return rest.Zone{
		ID:      zone.ID,
		Name:    zone.Name,
		Serial:  zone.Serial,
		Admin:   zone.Admin,
		Refresh: zone.Refresh,
		Retry:   zone.Retry,
		Expire:  zone.Expire,
		Ttl:     zone.Ttl,
		Active:  zone.Active,
	}
}

func AddZoneRoutes(r chi.Router, db zoneQueries, keystore *mjwt.KeyStore) {
	// List all zones
	r.Get("/zones", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
		zones, err := db.GetOwnedZones(req.Context(), b.Subject)
		if err != nil {
			logger.Logger.Error("Failed to get owned zones", "err", err)
			http.Error(rw, "Database error occurred", http.StatusInternalServerError)
			return
		}

		outZones := make([]rest.Zone, 0, len(zones))
		for _, z := range zones {
			if !b.Claims.Perms.Has("verbena-zone:" + z.Zone.Name) {
				continue
			}
			outZones = append(outZones, ZoneToRestZone(z.Zone))
		}

		json.NewEncoder(rw).Encode(outZones)
	}))

	// Show individual zone
	r.Get("/zones/{zone_id:[0-9]+}", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
		zoneId, err := getZoneId(req)
		if err != nil {
			http.Error(rw, "Invalid zone ID", http.StatusBadRequest)
			return
		}

		zone, err := db.GetZone(req.Context(), zoneId)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(rw, req)
			return
		case err != nil:
			logger.Logger.Error("Failed to get zone", "err", err)
			http.Error(rw, "Database error occurred", http.StatusInternalServerError)
			return
		}

		if !b.Claims.Perms.Has("verbena-zone:" + zone.Name) {
			http.NotFound(rw, req)
			return
		}
		json.NewEncoder(rw).Encode(ZoneToRestZone(zone))
	}))
}

func getZoneId(req *http.Request) (int64, error) {
	zoneIdRaw := chi.URLParam(req, "zone_id")
	return strconv.ParseInt(zoneIdRaw, 10, 64)
}
