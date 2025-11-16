package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/1f349/mjwt"
	"github.com/1f349/mjwt/auth"
	"github.com/1f349/verbena/conf"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/miekg/dns"
	"gopkg.in/yaml.v3"
)

var configPath = flag.String("conf", "", "Config file path")
var zone = flag.String("zone", "", "Zone to generate a token for")

func main() {
	flag.Parse()

	if *configPath == "" {
		logger.Logger.Fatal("Config flag is missing")
	}
	if *zone == "" {
		logger.Logger.Fatal("Zone flag is missing")
	}
	if _, isDomain := dns.IsDomainName(*zone); !isDomain {
		logger.Logger.Fatalf("Invalid zone %s", *zone)
		return
	}

	openConf, err := os.Open(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Logger.Fatal("Missing config file")
		} else {
			logger.Logger.Fatal("Open config file", "err", err)
		}
	}

	var config conf.Conf
	err = yaml.NewDecoder(openConf).Decode(&config)
	if err != nil {
		logger.Logger.Fatal("Parse config file", "err", err)
	}

	config.Cmd.LoadDefaults()

	wd := filepath.Dir(*configPath)

	keysPath := joinPath(wd, "keys")
	err = os.Mkdir(keysPath, 0700)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		logger.Logger.Fatal("Failed to create keys directory", "err", err)
	}

	// load the MJWT RSA public key from a pem encoded file
	apiKeystore, err := mjwt.NewKeyStoreFromPath(keysPath)
	if err != nil {
		logger.Logger.Fatal("Failed to load MJWT verifier public key from file", "path", filepath.Join(wd, "keys"), "err", err)
	}

	apiIssuer, err := mjwt.NewIssuerWithKeyStore("Verbena", config.TokenIssuer, jwt.SigningMethodRS512, apiKeystore)
	if err != nil {
		logger.Logger.Fatal("Failed to load MJWT issuer private key", "err", err)
	}

	db, err := database.InitDB(config.DB)
	if err != nil {
		logger.Logger.Fatal("Failed to open database", "err", err)
		return
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Minute)

	_, err = db.LookupZone(ctx, *zone)
	if err != nil {
		logger.Logger.Fatal("Failed to lookup zone", "err", err)
		return
	}

	tokenId := randomTokenId()
	if tokenId == 0 {
		logger.Logger.Fatal("Failed to generate random token id")
		return
	}

	tokenAti := strconv.FormatInt(tokenId, 16)

	botToken, err := auth.CreateRefreshTokenWithDuration(apiIssuer, 87600*time.Hour, *zone, tokenAti, tokenAti, jwt.ClaimStrings{})
	if err != nil {
		logger.Logger.Fatal("Failed to create bot token", "err", err)
		return
	}

	cancelCtx()

	fmt.Println("Bot Token: ", botToken)
}

func joinPath(base, option string) string {
	if filepath.IsAbs(option) {
		return filepath.Clean(option)
	}
	return filepath.Join(base, option)
}

func randomTokenId() int64 {
	const tokenReadLength = 8
	var b [tokenReadLength]byte
	least, err := io.ReadAtLeast(rand.Reader, b[:], tokenReadLength)
	if err != nil || least != tokenReadLength {
		return 0
	}
	return int64(binary.BigEndian.Uint64(b[:]))
}
