package routes

import (
	"context"
	"io"
	"net/http"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/logger"
	"github.com/go-chi/chi/v5"
)

type zoneFileQueries interface {
	GetZone(ctx context.Context, zoneId int64) (database.Zone, error)
}

type previewFunc func(ctx context.Context, w io.Writer, zoneInfo database.Zone) error

func AddZoneFileRoutes(r chi.Router, db zoneFileQueries, keystore *mjwt.KeyStore, preview previewFunc) {
	r.Get("/zones/{zone_id:[0-9]+}/zone-file", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
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

		_ = preview(req.Context(), rw, zone)
	}))
}
