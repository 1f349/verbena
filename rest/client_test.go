package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-cmp/cmp"
)

func TestNewClient(t *testing.T) {
	r := chi.NewRouter()
	issuer, err := mjwt.NewIssuer("Test", "1234", jwt.SigningMethodRS512)
	if err != nil {
		t.Fatal(err)
	}

	botTokenCall := 0
	r.Post("/refresh-bot-token", func(rw http.ResponseWriter, req *http.Request) {
		bearer, ok := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
		if !ok {
			t.Fatal("no bearer token found")
		}
		_, b, err := mjwt.ExtractClaims[auth.RefreshTokenClaims](issuer.KeyStore(), bearer)
		if err != nil {
			t.Fatal(err)
		}
		if b.Subject != "example.com" {
			t.Fatal("invalid bearer token subject")
		}
		botToken, err := auth.CreateAccessTokenWithDuration(issuer, time.Hour, "example.com", "5678", jwt.ClaimStrings{}, auth.NewPermStorage())
		if err != nil {
			t.Fatal(err)
		}
		botTokenCall++
		_ = json.NewEncoder(rw).Encode(struct {
			Token string `json:"token"`
		}{
			Token: botToken,
		})
	})
	r.Get("/zones", func(rw http.ResponseWriter, req *http.Request) {
		bearer, ok := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
		if !ok {
			t.Fatal("no bearer token found")
		}
		_, b, err := mjwt.ExtractClaims[auth.AccessTokenClaims](issuer.KeyStore(), bearer)
		if err != nil {
			t.Fatal(err)
		}
		if b.Subject != "example.com" {
			t.Fatal("invalid bearer token subject")
		}
		rw.Write([]byte(`[{"id":1,"name":"example.com","serial":2025010102,"admin":"hostmaster.example.com","refresh":3,"retry":4,"expire":5,"ttl":15,"active":true,"nameservers":["ns1.example.net","ns2.example.net","ns3.example.net"]}]`))
	})

	botToken, err := auth.CreateRefreshTokenWithDuration(issuer, time.Hour, "example.com", "3456", "3456", jwt.ClaimStrings{})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(r)
	client, err := NewClient(srv.URL, botToken)
	if err != nil {
		t.Fatal(err)
	}

	zones, err := client.GetZones()
	if err != nil {
		t.Fatal(err)
	}

	if len(zones) != 1 {
		t.Fatal("expected 1 zone")
	}

	if !cmp.Equal(zones[0], Zone{
		ID:          1,
		Name:        "example.com",
		Serial:      2025010102,
		Admin:       "hostmaster.example.com",
		Refresh:     3,
		Retry:       4,
		Expire:      5,
		Ttl:         15,
		Active:      true,
		Nameservers: []string{"ns1.example.net", "ns2.example.net", "ns3.example.net"},
	}) {
		t.Fatal("actual zone does not match expected zone")
	}
}
