package routes

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

type zoneFileTestQueries struct{}

func (z *zoneFileTestQueries) GetZone(ctx context.Context, zoneId int64) (database.Zone, error) {
	if zoneId != 3456 {
		return database.Zone{}, sql.ErrNoRows
	}

	return database.Zone{
		ID:     3456,
		Name:   "example.com",
		Serial: 2025062801,
		Active: true,
	}, nil
}

func TestAddZoneFileRoutes(t *testing.T) {
	r := chi.NewRouter()
	issuer, err := mjwt.NewIssuer("hello world", "1", jwt.SigningMethodRS256)
	if err != nil {
		t.Fatal(err)
	}
	q := &zoneFileTestQueries{}
	AddZoneFileRoutes(r, q, issuer.KeyStore(), func(ctx context.Context, w io.Writer, zoneInfo database.Zone) error {
		_, err := fmt.Fprintln(w, "example zone file")
		return err
	})

	t.Run("GET /zones/3456/zone-file", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/zones/3456/zone-file", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("POST /zones/3456/zone-file", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/zones/3456/zone-file", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/zones/3456/zone-file", nil)
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
		req = httptest.NewRequest(http.MethodPost, "/zones/3456/zone-file", nil)
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "example zone file\n", rec.Body.String())
	})
}
