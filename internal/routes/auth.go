package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/logger"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/miekg/dns"
)

type authQueries interface {
	GetOwnerByUserIdAndZone(ctx context.Context, arg database.GetOwnerByUserIdAndZoneParams) (database.GetOwnerByUserIdAndZoneRow, error)
	RegisterBotToken(ctx context.Context, arg database.RegisterBotTokenParams) (int64, error)
	BotTokenExists(ctx context.Context, id int64) (database.BotToken, error)
}

func AddAuthRoutes(r *chi.Mux, db authQueries, userKeystore *mjwt.KeyStore, apiIssuer *mjwt.Issuer) {
	r.Post("/bot-token", validateAuthToken[auth.AccessTokenClaims](userKeystore, func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.AccessTokenClaims]) {
		var createBody struct {
			Zone string `json:"zone"`
		}

		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&createBody)
		if err != nil {
			http.Error(rw, "Failed to decode body", http.StatusBadRequest)
			return
		}

		_, isDomain := dns.IsDomainName(createBody.Zone)
		if !isDomain {
			http.Error(rw, "Invalid zone", http.StatusBadRequest)
			return
		}

		ownerRow, err := db.GetOwnerByUserIdAndZone(req.Context(), database.GetOwnerByUserIdAndZoneParams{
			UserID: b.Subject,
			Name:   createBody.Zone,
		})
		if err != nil {
			logger.Logger.Debug("Failed to get owner by user id and zone", "err", err)
			http.Error(rw, "Database error", http.StatusInternalServerError)
			return
		}

		tokenId, err := db.RegisterBotToken(req.Context(), database.RegisterBotTokenParams{
			OwnerID: ownerRow.Owner.ID,
			ZoneID:  ownerRow.Owner.ZoneID,
		})
		if err != nil {
			logger.Logger.Debug("Failed to register bot token", "err", err)
			http.Error(rw, "Database error", http.StatusInternalServerError)
			return
		}

		tokenAti := strconv.FormatInt(tokenId, 16)

		botToken, err := auth.CreateRefreshTokenWithDuration(apiIssuer, 87600*time.Hour, createBody.Zone, tokenAti, tokenAti, jwt.ClaimStrings{})
		if err != nil {
			logger.Logger.Debug("Failed to create refresh token", "err", err)
			http.Error(rw, "Failed to create refresh token", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(rw).Encode(struct {
			Token string `json:"token"`
		}{
			Token: botToken,
		})
	}))

	r.Post("/refresh-bot-token", validateAuthToken[auth.RefreshTokenClaims](apiIssuer.KeyStore(), func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[auth.RefreshTokenClaims]) {
		zone := b.Subject
		if _, isDomain := dns.IsDomainName(zone); !isDomain {
			http.Error(rw, "Invalid token", http.StatusBadRequest)
			return
		}

		accessTokenId, err := strconv.ParseInt(b.Claims.AccessTokenId, 16, 64)
		if err != nil {
			http.Error(rw, "Invalid token", http.StatusBadRequest)
			return
		}

		_, err = db.BotTokenExists(req.Context(), accessTokenId)
		switch {
		case err == nil:
			break
		case errors.Is(err, sql.ErrNoRows):
			http.Error(rw, "Invalid token", http.StatusUnauthorized)
			return
		case err != nil:
			logger.Logger.Debug("Failed to refresh bot token", "err", err)
			http.Error(rw, "Database error", http.StatusInternalServerError)
			return
		}

		ps := auth.NewPermStorage()
		ps.Set("domain:owns=" + zone)
		sessionToken, err := auth.CreateAccessToken(apiIssuer, "domain:owns="+zone, "", jwt.ClaimStrings{
			"verbena-bot-token",
		}, ps)
		if err != nil {
			http.Error(rw, "Failed to create token", http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(rw).Encode(struct {
			Token string `json:"token"`
		}{
			Token: sessionToken,
		})
	}))
}

type authHandler[T mjwt.Claims] func(rw http.ResponseWriter, req *http.Request, b mjwt.BaseTypeClaims[T])

func validateAuthToken[T mjwt.Claims](keystore *mjwt.KeyStore, next authHandler[T]) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rawAuth := req.Header.Get("Authorization")
		token, found := strings.CutPrefix(rawAuth, "Bearer ")
		if !found {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		_, b, err := mjwt.ExtractClaims[T](keystore, token)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next(rw, req, b)
	}
}
