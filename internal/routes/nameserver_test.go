package routes

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

type fakeNameserverResolver struct{}

func (f *fakeNameserverResolver) LookupNS(_ context.Context, zone string) ([]*net.NS, error) {
	if zone != "example.com" {
		return nil, fmt.Errorf("invalid zone %s", zone)
	}

	return []*net.NS{
		{Host: "b.iana-servers.net."},
		{Host: "a.iana-servers.net."},
	}, nil
}

func TestAddNameserverRoutes(t *testing.T) {
	r := chi.NewRouter()
	issuer, err := mjwt.NewIssuer("hello world", "1", jwt.SigningMethodRS256)
	if err != nil {
		t.Fatal(err)
	}
	AddNameserverRoutes(r, issuer.KeyStore())

	nameserverResolver = &fakeNameserverResolver{}

	t.Run("GET /zone-nameservers/example.com", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/zone-nameservers/example.com", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("POST /zone-nameservers/example.com", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/zone-nameservers/example.com", nil)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/zone-nameservers/example.com", nil)
		ps := auth.NewPermStorage()
		ps.Set("domain:owns=example.org")
		token, err := issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "[\"b.iana-servers.net.\",\"a.iana-servers.net.\"]\n", rec.Body.String())

		rec = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/zone-nameservers/example.com", nil)
		ps = auth.NewPermStorage()
		ps.Set("domain:owns=example.com")
		token, err = issuer.GenerateJwt("1234", "", jwt.ClaimStrings{}, time.Hour, auth.AccessTokenClaims{Perms: ps})
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "[\"b.iana-servers.net.\",\"a.iana-servers.net.\"]\n", rec.Body.String())
	})
}
