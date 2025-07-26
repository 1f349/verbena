package routes

import (
	"context"
	"database/sql"
	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type zoneTestQueries struct {
}

func (z *zoneTestQueries) GetOwnedZones(ctx context.Context, userID string) ([]database.GetOwnedZonesRow, error) {
	if userID != "1234" {
		return []database.GetOwnedZonesRow{}, nil
	}

	return []database.GetOwnedZonesRow{
		{
			Zone: database.Zone{
				ID:      3456,
				Name:    "example.com",
				Serial:  2025062801,
				Admin:   "admin.example.com",
				Refresh: 10,
				Retry:   11,
				Expire:  12,
				Ttl:     13,
				Active:  true,
			},
			UserID: "1234",
		},
	}, nil
}

func (z *zoneTestQueries) GetZone(ctx context.Context, zoneId int64) (database.Zone, error) {
	if zoneId != 3456 {
		return database.Zone{}, sql.ErrNoRows
	}

	return database.Zone{
		ID:      3456,
		Name:    "example.com",
		Serial:  2025062801,
		Admin:   "admin.example.com",
		Refresh: 10,
		Retry:   11,
		Expire:  12,
		Ttl:     13,
		Active:  true,
	}, nil
}

func TestAddZoneRoutes(t *testing.T) {
	r := chi.NewRouter()
	issuer, err := mjwt.NewIssuer("hello world", "1", jwt.SigningMethodRS256)
	if err != nil {
		t.Fatal(err)
	}
	AddZoneRoutes(r, &zoneTestQueries{}, issuer.KeyStore(), []string{"ns1.example.com", "ns2.example.com"})

	t.Run("/zones", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/zones", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/zones", nil)
		ps := auth.NewPermStorage()
		ps.Set("domain:owns=example.org")
		token, err := issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "[]\n", rec.Body.String())

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/zones", nil)
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "[{\"id\":3456,\"name\":\"example.com\",\"serial\":2025062801,\"admin\":\"admin.example.com\",\"refresh\":10,\"retry\":11,\"expire\":12,\"ttl\":13,\"active\":true,\"nameservers\":[\"ns1.example.com\",\"ns2.example.com\"]}]\n", rec.Body.String())
	})

	t.Run("/zones/{id}", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/zones/3456", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/zones/3456", nil)
		ps := auth.NewPermStorage()
		ps.Set("domain:owns=example.org")
		token, err := issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/zones/3456", nil)
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "{\"id\":3456,\"name\":\"example.com\",\"serial\":2025062801,\"admin\":\"admin.example.com\",\"refresh\":10,\"retry\":11,\"expire\":12,\"ttl\":13,\"active\":true,\"nameservers\":[\"ns1.example.com\",\"ns2.example.com\"]}\n", rec.Body.String())
	})
}
