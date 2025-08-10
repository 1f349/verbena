package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/1f349/mjwt"
	"github.com/1f349/verbena/conf"
	"github.com/1f349/verbena/internal/builder"
	"github.com/1f349/verbena/internal/committer"
	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/internal/routes"
	"github.com/1f349/verbena/logger"
	"github.com/charmbracelet/log"
	"github.com/cloudflare/tableflip"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v4"
	"gopkg.in/yaml.v3"
)

var (
	configPath = flag.String("conf", "", "Config file path")
	debugLog   = flag.Bool("debug", false, "Enable debug logging")
	pidFile    = flag.String("pid-file", "", "Path to pid file")
)

func _() {
	// TODO: output zones files to update bind dns
	// sync between verbena services using grpc or similar
	// check status of other nodes and remove from dns when down
	// provide a maintenance mode to remove nodes early and retain uptime
}

func main() {
	flag.Parse()
	if *debugLog {
		logger.Logger.SetLevel(log.DebugLevel)
	}
	logger.Logger.Info("Starting...")

	upg, err := tableflip.New(tableflip.Options{
		PIDFile: *pidFile,
	})
	if err != nil {
		panic(err)
	}
	defer upg.Stop()

	if *configPath == "" {
		logger.Logger.Fatal("Config flag is missing")
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
		logger.Logger.Fatal("Invalid config file", "err", err)
	}

	wd := filepath.Dir(*configPath)

	zonesPath := filepath.Join(wd, config.ZonePath)
	err = os.Mkdir(zonesPath, 0700)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		logger.Logger.Fatal("Failed to create zone directory", "err", err)
	}

	// Do an upgrade on SIGHUP
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			err := upg.Upgrade()
			if err != nil {
				logger.Logger.Error("Failed upgrade", "err", err)
			}
		}
	}()

	keysPath := filepath.Join(wd, "keys")
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

	// Listen must be called before Ready
	lnApi, err := upg.Listen("tcp", config.Listen)
	if err != nil {
		logger.Logger.Fatal("Listen failed", "err", err)
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RealIP)
	if *debugLog {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Timeout(2 * time.Minute))
	r.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Add("Server", "Verbena")
			handler.ServeHTTP(rw, req)
		})
	})

	// Base endpoints
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Verbena API Endpoint", http.StatusOK)
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// TODO: add node health
		// TODO: maybe some cluster info too
	})

	// Add routes
	routes.AddZoneRoutes(r, db, apiKeystore, config.Nameservers)
	routes.AddRecordRoutes(r, db, apiKeystore)
	routes.AddAuthRoutes(r, db, apiKeystore, apiIssuer)

	zoneBuilder, err := builder.New(db, time.Duration(config.GeneratorTick), zonesPath, config.Nameservers)
	if err != nil {
		logger.Logger.Fatal("Failed to initialise zone builder", "err", err)
	}
	zoneBuilder.Start()

	commit := committer.New(db, time.Duration(config.CommitterTick), config.Primary, zoneBuilder)
	commit.Start()

	serverApi := &http.Server{
		Handler:           r,
		ReadTimeout:       1 * time.Minute,
		ReadHeaderTimeout: 1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
		IdleTimeout:       1 * time.Minute,
		MaxHeaderBytes:    4 * humanize.MiByte,
	}
	logger.Logger.Info("API server listening on", "addr", config.Listen)
	go func() {
		err := serverApi.Serve(lnApi)
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("API Serve failed", "err", err)
		}
	}()

	logger.Logger.Info("Ready")
	if err := upg.Ready(); err != nil {
		panic(err)
	}
	<-upg.Exit()

	time.AfterFunc(30*time.Second, func() {
		logger.Logger.Warn("Graceful shutdown timed out")
		os.Exit(1)
	})

	serverApi.Shutdown(context.Background())
}
