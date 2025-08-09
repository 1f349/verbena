package routes

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/gobuffalo/nulls"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

type recordTestQueries struct {
	records map[int64]database.Record
	nextId  atomic.Int64
}

func (r *recordTestQueries) GetZoneRecords(ctx context.Context, zoneId int64) ([]database.GetZoneRecordsRow, error) {
	if zoneId != 3456 {
		return nil, sql.ErrNoRows
	}

	rows := make([]database.GetZoneRecordsRow, len(r.records))
	for _, row := range r.records {
		rows = append(rows, database.GetZoneRecordsRow{
			Record: row,
			Name:   "example.com",
		})
	}

	return rows, nil
}

func (r *recordTestQueries) GetZoneRecord(ctx context.Context, row database.GetZoneRecordParams) (database.GetZoneRecordRow, error) {
	if row.ZoneID != 3456 {
		return database.GetZoneRecordRow{}, sql.ErrNoRows
	}

	record, ok := r.records[row.RecordID]
	if !ok {
		return database.GetZoneRecordRow{}, sql.ErrNoRows
	}

	return database.GetZoneRecordRow{
		Record: record,
		Name:   "example.com",
	}, nil
}

func (r *recordTestQueries) GetZone(ctx context.Context, zoneId int64) (database.Zone, error) {
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

func (r *recordTestQueries) InsertRecordFromApi(ctx context.Context, row database.InsertRecordFromApiParams) (int64, error) {
	if row.ZoneID != 3456 {
		return 0, sql.ErrNoRows
	}

	nextId := r.nextId.Add(1)
	r.records[nextId] = database.Record{
		ID:        nextId,
		Name:      row.Name,
		ZoneID:    row.ZoneID,
		Type:      row.Type,
		PreTtl:    row.PreTtl,
		PreValue:  row.PreValue,
		PreActive: row.PreActive,
		PreDelete: false,
	}
	return nextId, nil
}

func (r *recordTestQueries) UpdateRecordFromApi(ctx context.Context, row database.UpdateRecordFromApiParams) error {
	if row.ZoneID != 3456 {
		return sql.ErrNoRows
	}

	record, ok := r.records[row.ID]
	if !ok {
		return sql.ErrNoRows
	}

	record.PreTtl = row.PreTtl
	record.PreValue = row.PreValue
	record.PreActive = row.PreActive

	r.records[row.ID] = record

	return nil
}

func (r *recordTestQueries) DeleteRecordFromApi(ctx context.Context, row database.DeleteRecordFromApiParams) error {
	if row.ZoneID != 3456 {
		return sql.ErrNoRows
	}

	record, ok := r.records[row.RecordID]
	if !ok {
		return sql.ErrNoRows
	}

	record.PreDelete = true

	r.records[row.RecordID] = record

	return nil
}

func TestAddRecordRoutes(t *testing.T) {
	r := chi.NewRouter()
	issuer, err := mjwt.NewIssuer("hello world", "1", jwt.SigningMethodRS256)
	if err != nil {
		t.Fatal(err)
	}
	q := &recordTestQueries{
		records: make(map[int64]database.Record),
	}
	AddRecordRoutes(r, q, issuer.KeyStore())
	_, err = q.InsertRecordFromApi(t.Context(), database.InsertRecordFromApiParams{
		Name:      "",
		ZoneID:    3456,
		Type:      "AAAA",
		PreTtl:    nulls.Int32{},
		PreValue:  "2001:db8::5",
		PreActive: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("GET /zones/3456/records", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/zones/3456/records", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/zones/3456/records", nil)
		ps := auth.NewPermStorage()
		ps.Set("domain:owns=example.org")
		token, err := issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/zones/3456/records", nil)
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "[{\"id\":1,\"name\":\"\",\"zone_id\":3456,\"ttl\":null,\"type\":\"AAAA\",\"value\":{\"ip\":\"2001:db8::5\"},\"active\":true}]\n", rec.Body.String())
	})

	t.Run("GET /zones/3456/records/1", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/zones/3456/records/1", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/zones/3456/records/1", nil)
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
		req = httptest.NewRequest(http.MethodGet, "/zones/3456/records/1", nil)
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "{\"id\":1,\"name\":\"\",\"zone_id\":3456,\"ttl\":null,\"type\":\"AAAA\",\"value\":{\"ip\":\"2001:db8::5\"},\"active\":true}\n", rec.Body.String())
	})

	t.Run("POST /zones/3456/records", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/zones/3456/records", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/zones/3456/records", strings.NewReader(`{
  "name": "test",
	"ttl": null,
	"type": "AAAA",
	"value": {
		"ip": "2001:db8::6"
	},
	"active": true
}`))
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
		req = httptest.NewRequest(http.MethodPost, "/zones/3456/records", strings.NewReader(`{
  "name": "test",
	"ttl": null,
	"type": "AAAA",
	"value": {
		"ip": "2001:db8::6"
	},
	"active": true
}`))
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "{\"id\":2,\"name\":\"test\",\"zone_id\":3456,\"ttl\":null,\"type\":\"AAAA\",\"value\":{\"ip\":\"2001:db8::6\"},\"active\":true}\n", rec.Body.String())
	})

	t.Run("PUT /zones/3456/records/2", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/zones/3456/records/2", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPut, "/zones/3456/records/2", strings.NewReader(`{
	"ttl": null,
	"value": {
		"ip": "2001:db8::7"
	},
	"active": true
}`))
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
		req = httptest.NewRequest(http.MethodPut, "/zones/3456/records/2", strings.NewReader(`{
	"ttl": null,
	"value": {
		"ip": "2001:db8::7"
	},
	"active": true
}`))
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "{\"id\":2,\"name\":\"test\",\"zone_id\":3456,\"ttl\":null,\"type\":\"AAAA\",\"value\":{\"ip\":\"2001:db8::7\"},\"active\":true}\n", rec.Body.String())
	})

	t.Run("DELETE /zones/3456/records/2", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/zones/3456/records/2", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodDelete, "/zones/3456/records/2", nil)
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
		req = httptest.NewRequest(http.MethodDelete, "/zones/3456/records/2", nil)
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "", rec.Body.String())
	})
}
