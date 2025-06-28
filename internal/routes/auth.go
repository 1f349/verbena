package routes

import (
	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"net/http"
	"strings"
)

type authHandler func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims])

func validateAuthToken(keystore *mjwt.KeyStore, next authHandler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rawAuth := req.Header.Get("Authorization")
		token, found := strings.CutPrefix(rawAuth, "Bearer ")
		if !found {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		_, b, err := mjwt.ExtractClaims[auth.AccessTokenClaims](keystore, token)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next(rw, req, b)
	}
}
