package routes

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/go-chi/chi/v5"
	"github.com/miekg/dns"
)

var nameserverResolver interface {
	LookupNS(context.Context, string) ([]*net.NS, error)
} = net.DefaultResolver

func AddNameserverRoutes(r chi.Router, keystore *mjwt.KeyStore) {
	r.Post("/zone-nameservers/{zone_name:[a-z0-9-.]+}", validateAuthToken(keystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
		zoneName := chi.URLParam(req, "zone_name")

		// Check if zone looks real
		_, validDomain := dns.IsDomainName(zoneName)
		if !validDomain {
			http.Error(rw, "Invalid zone name", http.StatusBadRequest)
			return
		}

		ns, err := nameserverResolver.LookupNS(req.Context(), zoneName)
		if err != nil {
			http.Error(rw, "Failed to lookup nameservers", http.StatusInternalServerError)
			return
		}

		out := make([]string, len(ns))
		for i := range ns {
			out[i] = ns[i].Host
		}
		_ = json.NewEncoder(rw).Encode(out)
	}))
}
