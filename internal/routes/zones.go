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
	"github.com/go-chi/chi"
	"net/http"
	"strconv"
)

type zoneQueries interface {
	GetOwnedZones(ctx context.Context, userID string) ([]database.GetOwnedZonesRow, error)
	GetZone(ctx context.Context, id int64) (database.Zone, error)
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

		outZones := make([]database.Zone, 0, len(zones))
		for _, z := range zones {
			if !b.Claims.Perms.Has("verbena-zone:" + z.Zone.Name) {
				continue
			}
			outZones = append(outZones, z.Zone)
		}

		json.NewEncoder(rw).Encode(outZones)
	}))

	// Show individual zone
	r.Get("/zones/{id:[0-9]+}", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
		zoneIdRaw := chi.URLParam(req, "id")
		zoneId, err := strconv.ParseInt(zoneIdRaw, 10, 64)
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
			logger.Logger.Error("Failed to get owned zones", "err", err)
			http.Error(rw, "Database error occurred", http.StatusInternalServerError)
			return
		}

		if !b.Claims.Perms.Has("verbena-zone:" + zone.Name) {
			http.NotFound(rw, req)
			return
		}
		json.NewEncoder(rw).Encode(zone)
	}))
}
