package routes

import (
	"context"
	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type authTestQueries struct {
}

func (a *authTestQueries) GetOwnerByUserIdAndZone(ctx context.Context, arg database.GetOwnerByUserIdAndZoneParams) (database.GetOwnerByUserIdAndZoneRow, error) {
	if arg.UserID != "1234" {
		panic("not allowed")
	}
	if arg.Name != "example.com" {
		panic("not allowed")
	}
	return database.GetOwnerByUserIdAndZoneRow{
		Owner: database.Owner{
			ID:     5678,
			ZoneID: 3456,
			UserID: "1234",
		},
		Zone: database.Zone{
			ID:      3456,
			Name:    "example.com",
			Serial:  2025010101,
			Admin:   "admin.example.com",
			Refresh: 300,
			Retry:   300,
			Expire:  300,
			Ttl:     300,
			Active:  true,
		},
	}, nil
}

func (a *authTestQueries) RegisterBotToken(ctx context.Context, arg database.RegisterBotTokenParams) (int64, error) {
	if arg.ZoneID != 3456 {
		panic("not allowed")
	}
	if arg.OwnerID != 5678 {
		panic("not allowed")
	}
	return 7890, nil
}

func (a *authTestQueries) BotTokenExists(ctx context.Context, id int64) (database.BotToken, error) {
	if id != 0x45 {
		panic("not allowed")
	}
	return database.BotToken{
		ID:      0x45,
		OwnerID: 5678,
		ZoneID:  3456,
	}, nil
}

func TestAddAuthRoutes(t *testing.T) {
	r := chi.NewRouter()
	q := &authTestQueries{}

	userIssuer, err := mjwt.NewIssuer("user issuer", "1", jwt.SigningMethodRS256)
	if err != nil {
		t.Fatal(err)
	}
	apiIssuer, err := mjwt.NewIssuer("api issuer", "2", jwt.SigningMethodRS256)
	if err != nil {
		t.Fatal(err)
	}

	AddAuthRoutes(r, q, userIssuer.KeyStore(), apiIssuer)

	t.Run("POST /bot-token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/bot-token", strings.NewReader(""))
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		ps := auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err := auth.CreateAccessToken(userIssuer, "1234", "aa", jwt.ClaimStrings{}, ps)
		if err != nil {
			t.Fatal(err)
		}
		authToken := "Bearer " + token

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/bot-token", strings.NewReader(""))
		req.Header.Set("Authorization", authToken)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/bot-token", strings.NewReader("{\"zone\":\"example.com..\"}"))
		req.Header.Set("Authorization", authToken)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/bot-token", strings.NewReader("{\"zone\":\"example.com\"}"))
		req.Header.Set("Authorization", authToken)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		if !strings.HasPrefix(s, "{\"token\":") {
			t.Fatal("invalid response body, should start with '{\"token\":'")
		}
		if !strings.HasSuffix(s, "\"}\n") {
			t.Fatal("invalid response body, should end with '\"}'")
		}
	})

	t.Run("POST /refresh-bot-token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refresh-bot-token", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		token, err := auth.CreateRefreshToken(apiIssuer, "example.com..", "aa", "23", jwt.ClaimStrings{})
		if err != nil {
			t.Fatal(err)
		}
		authToken := "Bearer " + token

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/refresh-bot-token", nil)
		req.Header.Set("Authorization", authToken)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		token, err = auth.CreateRefreshToken(apiIssuer, "example.com", "aa", "45", jwt.ClaimStrings{})
		if err != nil {
			t.Fatal(err)
		}
		authToken = "Bearer " + token

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/refresh-bot-token", nil)
		req.Header.Set("Authorization", authToken)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		s := rec.Body.String()
		if !strings.HasPrefix(s, "{\"token\":") {
			t.Fatal("invalid response body, should start with '{\"token\":'")
		}
		if !strings.HasSuffix(s, "\"}\n") {
			t.Fatal("invalid response body, should end with '\"}'")
		}
	})
}
