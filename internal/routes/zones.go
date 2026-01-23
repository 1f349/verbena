package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/conf"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/logger"
	"github.com/1f349/verbena/rest"
	"github.com/go-chi/chi/v5"
	"github.com/miekg/dns"
)

const oneDaySeconds = 60 * 60 * 24
const oneWeekSeconds = oneDaySeconds * 7
const tenMinutesSeconds = 60 * 10

const refreshMaxOneWeek = oneWeekSeconds
const retryMaxOneWeek = oneWeekSeconds
const expireMax90Days = oneDaySeconds * 90
const ttlMaxOneWeek = oneWeekSeconds

type zoneQueries interface {
	GetOwnedZones(ctx context.Context, userID string) ([]database.GetOwnedZonesRow, error)
	GetZone(ctx context.Context, id int64) (database.Zone, error)
	LookupZone(ctx context.Context, name string) (int64, error)
	UpdateZoneConfig(ctx context.Context, updateZoneConfigParams database.UpdateZoneConfigParams) error
}

func ZoneToRestZone(zone database.Zone, nameservers []string) rest.Zone {
	return rest.Zone{
		ID:      zone.ID,
		Name:    zone.Name,
		Serial:  uint32(zone.Serial),
		Admin:   zone.Admin,
		Refresh: zone.Refresh,
		Retry:   zone.Retry,
		Expire:  zone.Expire,
		Ttl:     zone.Ttl,
		Active:  zone.Active,

		Nameservers: nameservers,
	}
}

func AddZoneRoutes(r chi.Router, db zoneQueries, keystore *mjwt.KeyStore, nameservers conf.NameserverConf) {
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
			if !b.Claims.Perms.Has("domain:owns=" + z.Zone.Name) {
				continue
			}
			outZones = append(outZones, ZoneToRestZone(z.Zone, nameservers.GetNameserversForZone(z.Zone)))
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

		if !b.Claims.Perms.Has("domain:owns=" + zone.Name) {
			http.NotFound(rw, req)
			return
		}
		json.NewEncoder(rw).Encode(ZoneToRestZone(zone, nameservers.GetNameserversForZone(zone)))
	}))

	// Update individual zone
	r.Put("/zones/{zone_id:[0-9]+}", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
		zoneId, err := getZoneId(req)
		if err != nil {
			http.Error(rw, "Invalid zone ID", http.StatusBadRequest)
			return
		}

		type zoneUpdates struct {
			Refresh int32
			Retry   int32
			Expire  int32
			Ttl     int32
		}

		var updates zoneUpdates
		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()
		err = dec.Decode(&updates)
		if err != nil {
			logger.Logger.Error("Failed to decode zone updates", "err", err)
			http.Error(rw, "Invalid zone update configuration", http.StatusBadRequest)
			return
		}

		if updates.Refresh > refreshMaxOneWeek {
			http.Error(rw, "Invalid refresh value, expected less than one week", http.StatusBadRequest)
			return
		}
		if updates.Retry > retryMaxOneWeek {
			http.Error(rw, "Invalid retry value, expected less than one week", http.StatusBadRequest)
			return
		}
		if updates.Expire > expireMax90Days {
			http.Error(rw, "Invalid expire value, expected less than 90 days", http.StatusBadRequest)
			return
		}
		if updates.Ttl > ttlMaxOneWeek {
			http.Error(rw, "Invalid time-to-live value, expected less than one week", http.StatusBadRequest)
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

		if !b.Claims.Perms.Has("domain:owns=" + zone.Name) {
			http.NotFound(rw, req)
			return
		}

		err = db.UpdateZoneConfig(req.Context(), database.UpdateZoneConfigParams{
			Refresh: updates.Refresh,
			Retry:   updates.Retry,
			Expire:  updates.Expire,
			Ttl:     updates.Ttl,
			ID:      zoneId,
		})
		if err != nil {
			logger.Logger.Error("Failed to update zone config", "err", err)
			http.Error(rw, "Database error occurred", http.StatusInternalServerError)
			return
		}
		http.Error(rw, "OK", http.StatusOK)
	}))

	r.Get("/zones/lookup/{zone_name:[a-z0-9-.]+}", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
		zoneName := chi.URLParam(req, "zone_name")

		// Check if zone looks real
		_, validDomain := dns.IsDomainName(zoneName)
		if !validDomain {
			http.Error(rw, "Invalid zone name", http.StatusBadRequest)
			return
		}

		// Check ownership
		if !b.Claims.Perms.Has("domain:owns=" + zoneName) {
			http.NotFound(rw, req)
			return
		}

		// Lookup and respond with zone ID
		zoneId, err := db.LookupZone(req.Context(), zoneName)
		if err != nil {
			return
		}
		json.NewEncoder(rw).Encode(struct {
			ID int64 `json:"id"`
		}{
			ID: zoneId,
		})
	}))
}

func getZoneId(req *http.Request) (int64, error) {
	zoneIdRaw := chi.URLParam(req, "zone_id")
	return strconv.ParseInt(zoneIdRaw, 10, 64)
}
